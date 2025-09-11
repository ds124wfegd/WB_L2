package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// FieldSet хранит номера полей, которые нужно вывести
type FieldSet map[int]bool

// parseFields парсит строку с указанием полей
func parseFields(fieldsSpec string) (FieldSet, error) {
	fields := make(FieldSet)
	if fieldsSpec == "" {
		return fields, nil
	}

	parts := strings.Split(fieldsSpec, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "-") {
			// Обработка диапазона (например, "3-5")
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("неверный формат диапазона: %s", part)
			}

			start, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			if err != nil {
				return nil, fmt.Errorf("неверное начало диапазона: %s", rangeParts[0])
			}

			end, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if err != nil {
				return nil, fmt.Errorf("неверный конец диапазона: %s", rangeParts[1])
			}

			if start > end {
				return nil, fmt.Errorf("начало диапазона больше конца: %d-%d", start, end)
			}

			for i := start; i <= end; i++ {
				if i > 0 { // Номера полей должны быть положительными
					fields[i] = true
				}
			}
		} else {
			// Обработка отдельного номера поля
			fieldNum, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("неверный номер поля: %s", part)
			}
			if fieldNum > 0 {
				fields[fieldNum] = true
			}
		}
	}

	return fields, nil
}

// processLine обрабатывает одну строку
func processLine(line string, delimiter string, fields FieldSet, separatedOnly bool) string {
	// Если установлен флаг -s и строка не содержит разделитель, возвращаем пустую строку
	if separatedOnly && !strings.Contains(line, delimiter) {
		return ""
	}

	// Разбиваем строку по разделителю
	parts := strings.Split(line, delimiter)

	// Если не указаны поля для вывода, возвращаем всю строку
	if len(fields) == 0 {
		return line
	}

	// Собираем только указанные поля
	var resultParts []string
	for i, part := range parts {
		fieldNum := i + 1 // Нумерация полей начинается с 1
		if fields[fieldNum] {
			resultParts = append(resultParts, part)
		}
	}

	return strings.Join(resultParts, delimiter)
}

func main() {
	// Парсинг аргументов командной строки
	fieldsFlag := flag.String("f", "", "Номера полей для вывода (например: 1,3-5)")
	delimiterFlag := flag.String("d", "\t", "Разделитель полей")
	separatedFlag := flag.Bool("s", false, "Выводить только строки с разделителем")
	flag.Parse()

	// Парсинг полей для вывода
	fields, err := parseFields(*fieldsFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка парсинга полей: %v\n", err)
		os.Exit(1)
	}

	// Чтение из STDIN и обработка строк
	scanner := bufio.NewScanner(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

	for scanner.Scan() {
		line := scanner.Text()
		processed := processLine(line, *delimiterFlag, fields, *separatedFlag)
		if processed != "" {
			fmt.Fprintln(writer, processed)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка чтения ввода: %v\n", err)
		os.Exit(1)
	}
}
