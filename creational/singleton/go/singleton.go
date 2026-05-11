package main

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type Logger interface {
	Debug(string)
	Info(string)
}

type logger struct {
	mu sync.Mutex
}

func (l *logger) Debug(msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintf(os.Stdout, "[DEBUG] %s %s\n", time.Now().Format(time.DateTime), msg)
}

func (l *logger) Info(msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintf(os.Stdout, "[INFO]  %s %s\n", time.Now().Format(time.DateTime), msg)
}

var (
	once      sync.Once
	singleton *logger
)

func GetLogger() Logger {
	once.Do(func() {
		singleton = &logger{}
	})
	return singleton
}

func main() {
	var wg sync.WaitGroup

	for i := range 5 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			l := GetLogger()
			fmt.Printf("goroutine %d → addr: %p\n", n, l)
			l.Info(fmt.Sprintf("message from goroutine %d", n))
		}(i)
	}

	wg.Wait()
}
