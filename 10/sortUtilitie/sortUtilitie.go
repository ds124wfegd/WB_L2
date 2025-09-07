// Упрощённый аналог UNIX-утилиты sort
package sortUtilitie

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	// ErrNotSorted возващается когда данные не отсортированы
	ErrNotSorted = errors.New("data is not sorted")
	// ErrInvalidColumn возвращается при неверном номере колонки
	ErrInvalidColumn = errors.New("invalid column number")
)

// месяцы для сортировки
var months = map[string]time.Month{
	"jan": time.January, "feb": time.February, "mar": time.March,
	"apr": time.April, "may": time.May, "jun": time.June,
	"jul": time.July, "aug": time.August, "sep": time.September,
	"oct": time.October, "nov": time.November, "dec": time.December,
}

// человекочитаемые суффиксы
var humanSuffixes = map[string]float64{
	"K": 1e3, "M": 1e6, "G": 1e9, "T": 1e12,
	"KB": 1e3, "MB": 1e6, "GB": 1e9, "TB": 1e12,
	"KiB": 1024, "MiB": 1024 * 1024, "GiB": 1024 * 1024 * 1024,
}

// Sort выполняет сортировку строк согласно флагам
func Sort(input io.Reader, output io.Writer, flags Flags) error {
	lines, err := readLines(input, flags)
	if err != nil {
		return err
	}

	if flags.CheckSorted {
		return checkSorted(lines, flags)
	}

	sortLines(lines, flags)

	return writeLines(output, lines, flags)
}

// readLines читает стрки из входного потока
func readLines(input io.Reader, flags Flags) ([]string, error) {
	scanner := bufio.NewScanner(input)
	var lines []string
	seen := make(map[string]bool)

	for scanner.Scan() {
		line := scanner.Text()
		if flags.IgnoreBlanks {
			line = strings.TrimSpace(line)
		}

		if flags.Unique {
			if seen[line] {
				continue
			}
			seen[line] = true
		}

		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading input: %w", err)
	}

	return lines, nil
}

// writeLines записывает строки в выходной поток
func writeLines(output io.Writer, lines []string, flags Flags) error {
	writer := bufio.NewWriter(output)
	defer writer.Flush()

	for _, line := range lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("error writing output: %w", err)
		}
	}
	return nil
}

// compare сравнивает две строки согласно флагам
func compare(a, b string, flags Flags) int {
	if flags.Key > 0 {
		aCol := getColumn(a, flags.Key, flags.ColumnSep)
		bCol := getColumn(b, flags.Key, flags.ColumnSep)
		a, b = aCol, bCol
	}

	switch {
	case flags.Numeric:
		return compareNumeric(a, b)
	case flags.MonthSort:
		return compareMonths(a, b)
	case flags.HumanNumeric:
		return compareHumanNumeric(a, b)
	default:
		return strings.Compare(a, b)
	}
}

// sortLines сортирует строки согласно флагам
func sortLines(lines []string, flags Flags) {
	sort.SliceStable(lines, func(i, j int) bool {
		return compare(lines[i], lines[j], flags) < 0
	})

	if flags.Reverse {
		reverse(lines)
	}
}

// getColumn извлекает колонку из строки
func getColumn(line string, colNum int, sep string) string {
	columns := strings.Split(line, sep)
	if colNum <= 0 || colNum > len(columns) {
		return ""
	}
	return columns[colNum-1]
}

// compareNumeric сравнивает строки как числа
func compareNumeric(a, b string) int {
	numA, errA := strconv.ParseFloat(a, 64)
	numB, errB := strconv.ParseFloat(b, 64)

	// Если обе строки - числа, сравниваем как числа
	if errA == nil && errB == nil {
		if numA < numB {
			return -1
		}
		if numA > numB {
			return 1
		}
		return 0
	}

	// Если одна из строк не число, сравниваем как строки
	return strings.Compare(a, b)
}

// compareHumanNumeric сравнивает человекочитаемые числа
func compareHumanNumeric(a, b string) int {
	numA, errA := parseHumanNumber(a)
	numB, errB := parseHumanNumber(b)

	if errA == nil && errB == nil {
		if numA < numB {
			return -1
		}
		if numA > numB {
			return 1
		}
		return 0
	}

	return strings.Compare(a, b)
}

// compareMonths сравнивает строки как названия месяцев
func compareMonths(a, b string) int {
	monthA, okA := parseMonth(a)
	monthB, okB := parseMonth(b)

	if okA && okB {
		if monthA < monthB {
			return -1
		}
		if monthA > monthB {
			return 1
		}
		return 0
	}

	return strings.Compare(a, b)
}

// reverse переворачивает порядок элементов
func reverse(lines []string) {
	for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
		lines[i], lines[j] = lines[j], lines[i]
	}
}

// parseMonth парсит название месяца
func parseMonth(s string) (time.Month, bool) {
	s = strings.ToLower(s[:3])
	month, ok := months[s]
	return month, ok
}

// checkSorted проверяет, отсортированы ли данные
func checkSorted(lines []string, flags Flags) error {
	for i := 1; i < len(lines); i++ {
		if compare(lines[i-1], lines[i], flags) > 0 {
			return ErrNotSorted
		}
	}
	return nil
}

// parseHumanNumber парсит человекочитаемое число
func parseHumanNumber(s string) (float64, error) {
	s = strings.TrimSpace(s)
	s = strings.ToUpper(s)

	re := regexp.MustCompile(`^([0-9.]+)\s*([A-Za-z]+)?$`)
	matches := re.FindStringSubmatch(s)

	if len(matches) == 0 {
		return strconv.ParseFloat(s, 64)
	}

	number, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, err
	}

	if len(matches) > 2 && matches[2] != "" {
		suffix := matches[2]
		if multiplier, ok := humanSuffixes[suffix]; ok {
			number *= multiplier
		}
	}

	return number, nil
}
