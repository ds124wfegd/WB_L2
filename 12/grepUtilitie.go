package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// config хранит все флаги для работы grep
type config struct {
	after      int    // количество строк после совпадения (флаг -A)
	before     int    // количество строк до совпадения (флаг -B)
	context    int    // количество строк контекста (флаг -C)
	count      bool   // только подсчет совпадений (флаг -c)
	ignoreCase bool   // игнорировать регистр (флаг -i)
	invert     bool   // инвертировать поиск (флаг -v)
	fixed      bool   // фиксированная строка вместо regex (флаг -F)
	lineNum    bool   // выводить номера строк (флаг -n)
	pattern    string // шаблон для поиска
	filename   string // имя файла (если не указан, читаем из STDIN)
}

func main() {
	cfg := parseFlags()
	if err := runGrep(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
		os.Exit(1)
	}
}

// parseFlags обрабатывает аргументы командной строки и возвращает конфигурацию (config)
func parseFlags() config {
	var cfg config

	// регистрируем флаги
	flag.IntVar(&cfg.after, "A", 0, "вывести N строк после совпадения")
	flag.IntVar(&cfg.before, "B", 0, "вывести N строк до совпадения")
	flag.IntVar(&cfg.context, "C", 0, "вывести N строк контекста (вокруг совпадения)")
	flag.BoolVar(&cfg.count, "c", false, "только подсчет совпадающих строк")
	flag.BoolVar(&cfg.ignoreCase, "i", false, "игнорировать регистр")
	flag.BoolVar(&cfg.invert, "v", false, "инвертировать поиск (выводить несовпадающие строки)")
	flag.BoolVar(&cfg.fixed, "F", false, "фиксированная строка (не регулярное выражение)")
	flag.BoolVar(&cfg.lineNum, "n", false, "выводить номера строк")

	flag.Parse()

	// обработка флага контекста
	if cfg.context > 0 {
		cfg.after = cfg.context
		cfg.before = cfg.context
	}

	// получение аргументов PATTERN и [FILE]
	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Использование: grep [ОПЦИИ] ШАБЛОН [ФАЙЛ]")
		os.Exit(1)
	}

	cfg.pattern = args[0]
	if len(args) > 1 {
		cfg.filename = args[1]
	}

	return cfg
}

// runGrep основная функция, выполняющая поиск
func runGrep(cfg config) error {
	var scanner *bufio.Scanner

	// открытие файла/использование STDIN
	if cfg.filename == "" {
		scanner = bufio.NewScanner(os.Stdin)
	} else {
		file, err := os.Open(cfg.filename)
		if err != nil {
			return err
		}
		defer file.Close()
		scanner = bufio.NewScanner(file)
	}

	// подготовка шаблона с учетом регистра
	pattern := cfg.pattern
	if cfg.ignoreCase && !cfg.fixed {
		// добавление флага игноирования регистра для regex
		pattern = "(?i)" + pattern
	}

	var re *regexp.Regexp
	// использование регулярного выражение, если не используется фиксированная строка
	if !cfg.fixed {
		var err error
		re, err = regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("неверное регулярное выражение: %v", err)
		}
	}

	// чтение из входного потока
	lines := readLines(scanner)
	// индексы совпадающих строк
	matches := findMatches(lines, re, cfg)

	// флаг -c, просто выводим количество совпадений
	if cfg.count {
		fmt.Println(len(matches))
		return nil
	}

	// вывод результатов
	printMatches(lines, matches, cfg)
	return nil
}

// readLines читает все строи из сканера и возвращает их как слайс
func readLines(scanner *bufio.Scanner) []string {
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

// findMatches ищет строки, соответствующие шаблону, и возвращает их индексы
func findMatches(lines []string, re *regexp.Regexp, cfg config) []int {
	var matches []int

	for i, line := range lines {
		matched := false

		if cfg.fixed {
			// поиск фиксированной строки
			if cfg.ignoreCase {
				matched = strings.Contains(strings.ToLower(line), strings.ToLower(cfg.pattern))
			} else {
				matched = strings.Contains(line, cfg.pattern)
			}
		} else {
			// поиск по регулярному выражению
			matched = re.MatchString(line)
		}

		// инвертирование результата,при флаге -v
		if cfg.invert {
			matched = !matched
		}

		if matched {
			matches = append(matches, i)
		}
	}

	return matches
}

// printMatches выводит найденные строки с учетом контекста и других флагов
func printMatches(lines []string, matches []int, cfg config) {
	printed := make(map[int]bool) // для отслеживания уже выведенных строк
	lastPrinted := -1             // индекс последней выведенной строки

	for _, matchIdx := range matches {
		// если нужно выводить контекст (до/после)
		if cfg.before > 0 || cfg.after > 0 {
			printWithContext(lines, matchIdx, cfg, printed, &lastPrinted)
		} else {
			// вывод одной строки
			printLine(lines, matchIdx, cfg, printed, &lastPrinted)
		}
	}
}

// printWithContext выводит строку с контекстом (строки до и после)
func printWithContext(lines []string, matchIdx int, cfg config, printed map[int]bool, lastPrinted *int) {
	// оределение диапазона строк для вывода
	start := max(0, matchIdx-cfg.before)
	end := min(len(lines)-1, matchIdx+cfg.after)

	// вывод строки в диапазоне
	for i := start; i <= end; i++ {
		// добавление разделителя, если есть разрыв между группами контекста
		if i > *lastPrinted+1 && *lastPrinted != -1 && i != start {
			fmt.Println("--")
		}
		printLine(lines, i, cfg, printed, lastPrinted)
	}
}

// printLine выводит одну строку с учетом флагов
func printLine(lines []string, idx int, cfg config, printed map[int]bool, lastPrinted *int) {
	// пропуск уже выведенных строк (чтобы избежать дублирования)
	if printed[idx] {
		return
	}

	// вывод номер строки, если установлен флаг -n
	if cfg.lineNum {
		fmt.Printf("%d:", idx+1)
	}

	// вывод самой строку
	fmt.Println(lines[idx])

	// отметка строки, как выведенной и обновление последнего индекса
	printed[idx] = true
	*lastPrinted = idx
}

// max возвращает большее из двух чисел
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min возвращает меньшее из двух чисел
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
