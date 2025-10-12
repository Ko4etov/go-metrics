package pollcounter

import (
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

type PollCounter struct {
	filePath string
	mu *sync.Mutex
}

func New(filePath string) *PollCounter {
	if filePath == "" {
		filePath = "internal/agent/service/poll_counter/pollCounter.txt"
	}

	dir := filepath.Dir(filePath)
	os.MkdirAll(dir, 0755)

	return &PollCounter{
		filePath: filePath,
		mu: &sync.Mutex{},
	}
}

func (p *PollCounter) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	os.WriteFile(p.filePath, []byte("0"), 0666)
}

func (p *PollCounter) Get() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	data, err := os.ReadFile(p.filePath)
	if os.IsNotExist(err) {
		err := os.WriteFile(p.filePath, []byte("0"), 0666)
		if err != nil {
			panic(err)
		}
		return 0
	} else if err != nil {
		return 0
	}

	// Безопасное преобразование
	count, err := strconv.Atoi(string(data))
	if err != nil {
		os.WriteFile(p.filePath, []byte("0"), 0666)
		return 0
	}

	return count
}

func (p *PollCounter) Increment() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	data, err := os.ReadFile(p.filePath)
	current := 0
	if err == nil {
		current, _ = strconv.Atoi(string(data))
	}

	current++
	os.WriteFile(p.filePath, []byte(strconv.Itoa(current)), 0666)
	return current
}