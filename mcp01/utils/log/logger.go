// add logger with zap
package log

import (
	"fmt"
	"io"
	"os"
	"time"

	"log/slog"
)

var Logger *slog.Logger

func InitLogger(logLevel slog.Level, writer io.Writer) {
	handler := NewTextHandler(writer, &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: logLevel <= slog.LevelDebug,
	})
	Logger = slog.New(handler)
	slog.SetDefault(Logger)
}

func LogAndExit(err error) {
	if err == nil {
		return
	}
	Logger.Warn(err.Error())
	os.Exit(1)
}

// log time
func LogExeTime(name string) func() {
	start := time.Now()
	return func() {
		if Logger == nil {
			return
		}
		Logger.Info(fmt.Sprintf("%s execution time: %v", name, time.Since(start)))
	}
}
