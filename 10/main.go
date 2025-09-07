package main

import (
	"flag"
	"fmt"
	"os"



func main() {
	flags := parseFlags()

	if err := run(flags); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// parseFlags парсит флаги командной строки
func parseFlags() sortUtilitie.Flags {
	var flags sortUtilitie.Flags

	flag.IntVar(&flags.Key, "k", 0, "sort by column number")
	flag.BoolVar(&flags.Numeric, "n", false, "sort numerically")
	flag.BoolVar(&flags.Reverse, "r", false, "reverse sort order")
	flag.BoolVar(&flags.Unique, "u", false, "output only unique lines")
	flag.BoolVar(&flags.MonthSort, "M", false, "sort by month names")
	flag.BoolVar(&flags.IgnoreBlanks, "b", false, "ignore trailing blanks")
	flag.BoolVar(&flags.CheckSorted, "c", false, "check if data is sorted")
	flag.BoolVar(&flags.HumanNumeric, "h", false, "sort human-readable numbers")

	flag.StringVar(&flags.ColumnSep, "sep", "\t", "column separator")
	flag.StringVar(&flags.ColumnSep, "t", "\t", "column separator (short)")

	flag.Parse()

	return flags
}

// run выполняет основную логику программы
func run(flags sortUtilitie.Flags) error {
	var input *os.File
	var err error

	if flag.NArg() == 0 {
		input = os.Stdin
	} else {
		input, err = os.Open(flag.Arg(0))
		if err != nil {
			return fmt.Errorf("error opening file: %w", err)
		}
		defer input.Close()
	}

	return sortUtilitie.Sort(input, os.Stdout, flags)
}
