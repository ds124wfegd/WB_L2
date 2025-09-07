// Программа для получения точного времени через NTP сервер
// NTP (Network Time Protocol) - сетевой протокол
// для синхронизации внутренних часов компьютеров
// с использованием сетей с переменной латентностью.
package main

import (
	"os"

	"github.com/ds124wfegd/WB_L2/8/ntpTime"
)

func main() {

	client := ntpTime.DefaultClient()
	exitCode := client.PrintTime()
	os.Exit(exitCode)
}
