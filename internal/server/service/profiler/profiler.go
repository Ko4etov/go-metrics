// Package profiler предоставляет инструменты для профилирования приложения.
//
// Пакет включает:
// - Запуск HTTP-сервера для pprof
// - Сохранение профилей памяти на диск
// - Автоматическое создание дампов памяти
package profiler

import (
	"fmt"
	"net/http"
	_ "net/http/pprof" // автоматически добавляет handlers для pprof
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/Ko4etov/go-metrics/internal/server/service/logger"
)

// StartProfiling запускает HTTP-сервер для профилирования pprof.
//
// Параметр addr определяет адрес, на котором будет запущен сервер.
func StartProfiling(addr string) {
	go func() {
		logger.Logger.Infof("Starting pprof server on %s", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			logger.Logger.Errorf("pprof server failed: %v", err)
		}
	}()
}

// SaveProfiling сохраняет профиль памяти после указанной длительности.
//
// Параметры:
//   - dir: директория для сохранения профиля
//   - duration: время через которое будет сделан дамп памяти
func SaveProfiling(dir string, duration time.Duration) {
	go func() {
		time.Sleep(duration)

		baseProfile := filepath.Join(dir, "result.pprof")
		if err := SaveHeapProfile(baseProfile); err != nil {
			logger.Logger.Warnf("could not save base profile: %v", err)
		}
	}()
}

// SaveHeapProfile сохраняет профиль памяти в файл.
//
// Параметр filename определяет путь к файлу для сохранения профиля.
// Возвращает ошибку, если не удалось создать файл или записать профиль.
func SaveHeapProfile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("could not create memory profile: %w", err)
	}
	defer f.Close()

	runtime.GC()

	if err := pprof.WriteHeapProfile(f); err != nil {
		return fmt.Errorf("could not write memory profile: %w", err)
	}

	return nil
}