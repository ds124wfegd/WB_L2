package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

func main() {

	fmt.Println("Введите слова в строке через пробел:")
	r := bufio.NewReader(os.Stdin)
	str, _ := r.ReadString('\n')
	str = strings.TrimSpace(str)

	arrWords := strings.Fields(str)
	res := anagrams(arrWords)

	fmt.Println("Анаграммы:")
	for key, values := range res {
		fmt.Printf("%s: %v\n", key, values)
	}
}

func anagrams(arrWords []string) map[string][]string {
	res := make(map[string][]string)
	anagramMap := make(map[string][]string)

	// группируем слова по сортированному ключу
	for _, word := range arrWords { // сложность O(n)
		sorted := sorting(word)                               // сложность O(m log m)
		anagramMap[sorted] = append(anagramMap[sorted], word) // сложность O(1)
	}

	// исключаем слова, у которых нет анаграм
	for _, group := range anagramMap { // сложность O(n)
		if len(group) > 1 {
			// используем первое слово как ключ
			res[group[0]] = group
		}
	}

	return res
}

func sorting(word string) string {
	letters := strings.Split(strings.ToLower(word), "") // разделение слова на буквы и преобразование их в нижний регистр, создание слайса из этих букв, сложность O(m)
	sort.Strings(letters)                               // сортировка слайса букв,  сложность O(m log m)
	return strings.Join(letters, "")                    // соединение букв в строку, сложность O(m)
}
