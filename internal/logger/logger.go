package logger

import (
	"context"
	"fmt"
	"log"
)

type DefaultLogger struct{}

func NewDefaultLogger() DefaultLogger {
	return DefaultLogger{}
}

func (DefaultLogger) Debugf(ctx context.Context, format string, v ...any) {
	log.Printf("[DEBUG] %v\n", fmt.Sprintf(format, v...))
}

func (DefaultLogger) Infof(ctx context.Context, format string, v ...any) {
	log.Printf("[INFO] %v\n", fmt.Sprintf(format, v...))
}

func (DefaultLogger) Warnf(ctx context.Context, format string, v ...any) {
	log.Printf("[WARN] %v\n", fmt.Sprintf(format, v...))
}

func (DefaultLogger) Errorf(ctx context.Context, format string, v ...any) {
	log.Printf("[ERROR] %v\n", fmt.Sprintf(format, v...))
}
