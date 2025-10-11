package poll_metrics_counter

import (
	"os"
	"strconv"
)

type PollMetricsCounter struct {
	filePath string
}

func New() *PollMetricsCounter {
	return &PollMetricsCounter{
		filePath: "internal/agent/service/poll_metrics_counter/pollCounter.txt",
	}
}

func (p *PollMetricsCounter) Reset() {
	os.WriteFile(p.filePath, []byte("0"), 0666)
}

func (p *PollMetricsCounter) Get() int {
	data, err := os.ReadFile(p.filePath)
	if err != nil {
		p.Reset()
		return 0
	}

	// Безопасное преобразование
	count, err := strconv.Atoi(string(data))
	if err != nil {
		p.Reset()
		return 0
	}

	return count
}

func (p *PollMetricsCounter) Increment() int {
	current := p.Get()
	current++
	os.WriteFile(p.filePath, []byte(strconv.Itoa(current)), 0666)
	return current
}