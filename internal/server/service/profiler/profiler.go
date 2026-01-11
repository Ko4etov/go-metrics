package profiler

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" // автоматически добавляет handlers для pprof
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/Ko4etov/go-metrics/internal/server/service/logger"
)

func StartProfiling(addr string) {
	go func() {
		log.Printf("Starting pprof server on %s", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Printf("pprof server failed: %v", err)
		}
	}()
}

func SaveProfiling(dir string, duration time.Duration) {
	go func() {
		time.Sleep(duration)

		baseProfile := filepath.Join(dir, "result.pprof")
		if err := SaveHeapProfile(baseProfile); err != nil {
			logger.Logger.Warnf("could not save base profile: %v", err)
		}
	}()
}

func SaveHeapProfile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("could not create memory profile: %w", err)
	}
	defer f.Close()
	
	// Собираем данные о памяти
	runtime.GC()
	
	if err := pprof.WriteHeapProfile(f); err != nil {
		return fmt.Errorf("could not write memory profile: %w", err)
	}
	
	log.Printf("Heap profile saved to %s", filename)
	return nil
}