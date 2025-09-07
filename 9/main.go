package main

import (
	"fmt"
	"log"

	"github.com/ds124wfegd/WB_L2/9/unpackingString"
)

func main() {
	// Примеры использования
	testCases := []string{
		"a4bc2d5e",
		"abcd",
		"45",
		"",
		"qwe\\4\\5",
		"qwe\\45",
	}

	for _, test := range testCases {
		result, err := unpackingString.UnpackString(test)
		if err != nil {
			log.Printf("Ошибка для '%s': %v", test, err)
		} else {
			fmt.Printf("Входная строка: '%s'\n Выходная строка: '%s'\n", test, result)
		}
	}
}
