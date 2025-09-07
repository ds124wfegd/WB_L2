// Package ntp предоставляет возможностьь для работы с NTP серверами.
package ntpTime

import (
	"fmt"
	"os"
	"time"

	"github.com/beevik/ntp"
)

// Client представляет NTP клиент.
type Client struct {
	server string
}

// New создает новый экземпляр NTP клиента.
func New(server string) *Client {
	return &Client{server: server}
}

// DefaultClient создает клиент с сервером по умолчанию.
func DefaultClient() *Client {
	return New("pool.ntp.org")
}

// GetTime возвращает текущее время с NTP сервера.
func (c *Client) GetTime() (time.Time, error) {
	return ntp.Time(c.server)
}

// PrintTime получает и выводит текущее время.
// Возвращает код выхода: 0 при успехе, 1 при ошибке.
func (c *Client) PrintTime() int {
	currentTime, err := c.GetTime()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка получения времени: %v\n", err)
		return 1
	}

	fmt.Printf("Точное время: %s\n", currentTime.Format(time.RFC3339))
	return 0
}
