package main

import (
	"fmt"
	"sync"
)

type Config struct {
	DSN      string
	LogLevel string
	Port     int
}

var (
	config     *Config
	configOnce sync.Once
)

func GetConfig() *Config {
	configOnce.Do(func() {
		config = &Config{
			DSN:      "postgres://localhost:5432/singleton",
			LogLevel: "info",
			Port:     8080,
		}
		fmt.Println("Config loaded")
	})
	return config
}

func main() {
	var wg sync.WaitGroup

	for range 5 {
		wg.Go(func() {
			cfg := GetConfig()
			fmt.Printf("addr: %p, port: %d\n", cfg, cfg.Port)
		})
	}

	wg.Wait()
}
