package sortUtilitie

// Flags представляет флаги командной строки для сортировки
type Flags struct {
	Key          int    // -k: колонка для сортировки
	Numeric      bool   // -n: числовая сортировка
	Reverse      bool   // -r: обратный порядок
	Unique       bool   // -u: только уникальные строки
	MonthSort    bool   // -M: сортировка по месяцам
	IgnoreBlanks bool   // -b: игнорировать хвостовые пробелы
	CheckSorted  bool   // -c: проверить отсортированность
	HumanNumeric bool   // -h: человекочитаемые числа
	ColumnSep    string // разделитель колонок
}
