package main

type customError struct {
	msg string
}

func (e *customError) Error() string {
	return e.msg
}

func test() *customError {
	// ... do something
	return nil
}

func main() {
	var err error // error это интерфейс, 	//Динамический тип = error (не nil!)
	//Динамическое значение = nil
	err = test()
	if err != nil { // // и сравнивая интерфейс с nil мы получим false
		println("error")
		return
	}
	println("ok")
}

/*
вывод программы: error
*/
