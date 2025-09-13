package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

// Глобальные переменные для управления запущенными процессами
var (
	processMutex     sync.Mutex  // Мьютекс для защиты доступа к runningProcesses
	runningProcesses []*exec.Cmd // Список запущенных команд
)

func main() {
	flag.Parse()

	// Настройка обработки сигналов прерывания
	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, os.Interrupt)
	go func() {
		for range interruptChannel {
			interruptRunningProcesses()
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	// Увеличиваем буфер для чтения длинных строк
	buffer := make([]byte, 0, 64*1024)
	scanner.Buffer(buffer, 1024*1024)

	// Основной цикл чтения и выполнения команд
	for {
		if !scanner.Scan() {
			if scanner.Err() != nil {
				fmt.Fprintln(os.Stderr, scanner.Err())
			}
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		_ = processCommandLine(line)
	}
}

// Устанавливает список запущенных процессов
func setRunningProcesses(commands []*exec.Cmd) {
	processMutex.Lock()
	runningProcesses = commands
	processMutex.Unlock()
}

// Очищает список запущенных процессов
func clearRunningProcesses() {
	setRunningProcesses(nil)
}

// Посылает сигнал прерывания всем запущенным процессам
func interruptRunningProcesses() {
	processMutex.Lock()
	commands := append([]*exec.Cmd(nil), runningProcesses...)
	processMutex.Unlock()

	for _, cmd := range commands {
		if cmd == nil || cmd.Process == nil {
			continue
		}
		_ = cmd.Process.Signal(os.Interrupt)
	}
}

// Заменяет переменные окружения в строке на их значения
func expandEnvironmentVariables(input string) string {
	envVarRegex := regexp.MustCompile(`\$[A-Za-z_][A-Za-z0-9_]*`)
	return envVarRegex.ReplaceAllStringFunc(input, func(match string) string {
		variableName := match[1:] // Убираем символ '$'
		return os.Getenv(variableName)
	})
}

// Разбивает строку на токены, обрабатывая операторы и переменные окружения
func tokenizeCommandLine(line string) []string {
	// Добавляем пробелы вокруг операторов для корректного разбиения
	line = strings.ReplaceAll(line, "&&", " && ")
	line = strings.ReplaceAll(line, "||", " || ")
	line = strings.ReplaceAll(line, "|", " | ")
	line = strings.ReplaceAll(line, ">", " > ")
	line = strings.ReplaceAll(line, "<", " < ")

	tokens := strings.Fields(line)

	// Заменяем переменные окружения на их значения
	for i := range tokens {
		tokens[i] = expandEnvironmentVariables(tokens[i])
	}

	return tokens
}

// Проверяет, является ли команда встроенной
func isBuiltinCommand(commandName string) bool {
	switch commandName {
	case "cd", "pwd", "echo", "kill", "ps":
		return true
	}
	return false
}

// Выполняет встроенную команду оболочки
func executeBuiltinCommand(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) (int, error) {
	if len(args) == 0 {
		return 0, nil
	}

	switch args[0] {
	case "cd":
		if len(args) < 2 {
			return 1, errors.New("cd: missing path")
		}
		path := args[1]

		// Обработка домашней директории
		if path == "~" {
			if homeDir, ok := os.LookupEnv("HOME"); ok {
				path = homeDir
			}
		}

		// Преобразование относительного пути в абсолютный
		if !filepath.IsAbs(path) {
			currentDir, _ := os.Getwd()
			path = filepath.Join(currentDir, path)
		}

		if err := os.Chdir(path); err != nil {
			return 1, err
		}
		return 0, nil

	case "pwd":
		currentDir, err := os.Getwd()
		if err != nil {
			return 1, err
		}
		fmt.Fprintln(stdout, currentDir)
		return 0, nil

	case "echo":
		fmt.Fprintln(stdout, strings.Join(args[1:], " "))
		return 0, nil

	case "kill":
		if len(args) < 2 {
			return 1, errors.New("kill: missing process ID")
		}
		processID, err := strconv.Atoi(args[1])
		if err != nil {
			return 1, err
		}
		process, err := os.FindProcess(processID)
		if err != nil {
			return 1, err
		}
		if err := process.Kill(); err != nil {
			return 1, err
		}
		return 0, nil

	case "ps":
		var command *exec.Cmd
		if runtime.GOOS == "windows" {
			command = exec.Command("tasklist")
		} else {
			command = exec.Command("ps", "-e", "-o", "pid,comm")
		}
		command.Stdin = stdin
		command.Stdout = stdout
		command.Stderr = stderr
		if err := command.Run(); err != nil {
			return 1, err
		}
		return 0, nil
	}

	return 127, fmt.Errorf("unknown builtin command: %s", args[0])
}

// Разбивает токены на этапы пайплайна по оператору '|'
func parsePipelineStages(tokens []string) ([][]string, error) {
	var stages [][]string
	var currentStage []string

	for _, token := range tokens {
		if token == "|" {
			if len(currentStage) == 0 {
				return nil, errors.New("empty pipeline stage")
			}
			stages = append(stages, currentStage)
			currentStage = nil
			continue
		}
		currentStage = append(currentStage, token)
	}

	if len(currentStage) > 0 {
		stages = append(stages, currentStage)
	}

	return stages, nil
}

// Обрабатывает перенаправления ввода/вывода в этапе пайплайна
func processRedirections(stageTokens []string) (commandArgs []string, input io.Reader, output io.Writer, err error) {
	commandArgs = []string{}
	input = nil
	output = nil

	for i := 0; i < len(stageTokens); i++ {
		token := stageTokens[i]

		// Обработка перенаправления вывода
		if token == ">" && i+1 < len(stageTokens) {
			filename := stageTokens[i+1]
			file, fileErr := os.Create(filename)
			if fileErr != nil {
				return nil, nil, nil, fileErr
			}
			output = file
			i++ // Пропускаем следующий токен (имя файла)
			continue
		}

		// Обработка перенаправления ввода
		if token == "<" && i+1 < len(stageTokens) {
			filename := stageTokens[i+1]
			file, fileErr := os.Open(filename)
			if fileErr != nil {
				return nil, nil, nil, fileErr
			}
			input = file
			i++ // Пропускаем следующий токен (имя файла)
			continue
		}

		commandArgs = append(commandArgs, token)
	}

	return commandArgs, input, output, nil
}

// Выполняет пайплайн команд
func executePipeline(tokens []string) (int, error) {
	stages, err := parsePipelineStages(tokens)
	if err != nil {
		return 1, err
	}
	if len(stages) == 0 {
		return 0, nil
	}

	// Обработка одиночной команды (без пайплайна) с возможными перенаправлениями
	if len(stages) == 1 {
		args, inputReader, outputWriter, err := processRedirections(stages[0])
		if err != nil {
			return 1, err
		}
		if len(args) == 0 {
			return 0, nil
		}

		// Выполнение встроенной команды
		if isBuiltinCommand(args[0]) {
			if inputReader == nil {
				inputReader = os.Stdin
			}
			if outputWriter == nil {
				outputWriter = os.Stdout
			}
			exitCode, err := executeBuiltinCommand(args, inputReader, outputWriter, os.Stderr)
			// Закрываем файл вывода, если он был открыт
			if closer, ok := outputWriter.(io.Closer); ok && closer != os.Stdout {
				_ = closer.Close()
			}
			return exitCode, err
		}
	}

	var commands []*exec.Cmd
	var previousOutput io.Reader = os.Stdin
	var finalOutput io.Writer = os.Stdout
	var filesToClose []io.Closer

	// Создаем команды для каждого этапа пайплайна
	for stageIndex, stageTokens := range stages {
		args, inputRedirect, outputRedirect, err := processRedirections(stageTokens)
		if err != nil {
			return 1, err
		}
		if len(args) == 0 {
			return 1, errors.New("empty command")
		}

		commandName := args[0]
		commandArgs := args[1:]
		command := exec.Command(commandName, commandArgs...)

		// Настройка ввода
		if inputRedirect != nil {
			command.Stdin = inputRedirect
			if closer, ok := inputRedirect.(io.Closer); ok {
				filesToClose = append(filesToClose, closer)
			}
		} else {
			command.Stdin = previousOutput
		}

		// Настройка вывода
		if stageIndex < len(stages)-1 {
			// Промежуточный этап - перенаправляем вывод на следующий этап
			outputPipe, err := command.StdoutPipe()
			if err != nil {
				return 1, err
			}
			previousOutput = outputPipe
		} else {
			// Последний этап - используем перенаправление или stdout
			if outputRedirect != nil {
				command.Stdout = outputRedirect
				if closer, ok := outputRedirect.(io.Closer); ok {
					filesToClose = append(filesToClose, closer)
				}
			} else {
				command.Stdout = finalOutput
			}
		}

		command.Stderr = os.Stderr
		commands = append(commands, command)
	}

	// Сохраняем список запущенных процессов для возможного прерывания
	setRunningProcesses(commands)
	defer clearRunningProcesses()

	// Запускаем все команды
	for _, cmd := range commands {
		if err := cmd.Start(); err != nil {
			return 1, err
		}
	}

	// Ожидаем завершения всех команд
	var waitError error
	for i := len(commands) - 1; i >= 0; i-- {
		if err := commands[i].Wait(); err != nil && waitError == nil {
			waitError = err
		}
	}

	// Закрываем все открытые файлы
	for _, closer := range filesToClose {
		_ = closer.Close()
	}

	if waitError != nil {
		return 1, waitError
	}
	return 0, nil
}

// Обрабатывает строку команд с операторами && и ||
func processCommandLine(line string) int {
	tokens := tokenizeCommandLine(line)
	if len(tokens) == 0 {
		return 0
	}

	var commandSegments [][]string
	var logicalOperators []string
	currentSegment := []string{}

	// Разбиваем на сегменты по логическим операторам
	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		if token == "&&" || token == "||" {
			if len(currentSegment) == 0 {
				return 1
			}
			commandSegments = append(commandSegments, currentSegment)
			logicalOperators = append(logicalOperators, token)
			currentSegment = nil
			continue
		}
		currentSegment = append(currentSegment, token)
	}

	if len(currentSegment) > 0 {
		commandSegments = append(commandSegments, currentSegment)
	}
	if len(commandSegments) == 0 {
		return 0
	}

	previousExitCode := 0
	// Выполняем сегменты с учетом логических операторов
	for segmentIndex, segment := range commandSegments {
		if segmentIndex > 0 {
			operator := logicalOperators[segmentIndex-1]
			// Пропускаем выполнение если условия не выполняются
			if operator == "&&" && previousExitCode != 0 {
				continue
			}
			if operator == "||" && previousExitCode == 0 {
				continue
			}
		}
		exitCode, _ := executePipeline(segment)
		previousExitCode = exitCode
	}
	return previousExitCode
}
