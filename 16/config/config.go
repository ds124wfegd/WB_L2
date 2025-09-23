package config

type Config struct {
	URL      string
	Output   string
	MaxDepth int
	Workers  int
}

func NewConfig(url, output string, depth, workers int) *Config {
	return &Config{
		URL:      url,
		Output:   output,
		MaxDepth: depth,
		Workers:  workers,
	}
}
