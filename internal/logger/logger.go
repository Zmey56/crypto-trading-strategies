package logger

import (
	"fmt"
	"log"
	"os"
)

// Level представляет уровень логирования
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// Logger представляет логгер
type Logger struct {
	level  Level
	logger *log.Logger
}

// New создает новый логгер
func New(level Level) *Logger {
	return &Logger{
		level:  level,
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

// NewWithFile создает логгер с записью в файл
func NewWithFile(level Level, filename string) (*Logger, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &Logger{
		level:  level,
		logger: log.New(file, "", log.LstdFlags),
	}, nil
}

// Debug логирует отладочное сообщение
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level <= LevelDebug {
		l.logger.Printf("[DEBUG] "+format, args...)
	}
}

// Info логирует информационное сообщение
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level <= LevelInfo {
		l.logger.Printf("[INFO] "+format, args...)
	}
}

// Warn логирует предупреждение
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.level <= LevelWarn {
		l.logger.Printf("[WARN] "+format, args...)
	}
}

// Error логирует ошибку
func (l *Logger) Error(format string, args ...interface{}) {
	if l.level <= LevelError {
		l.logger.Printf("[ERROR] "+format, args...)
	}
}

// Fatal логирует критическую ошибку и завершает программу
func (l *Logger) Fatal(format string, args ...interface{}) {
	if l.level <= LevelFatal {
		l.logger.Printf("[FATAL] "+format, args...)
		os.Exit(1)
	}
}

// WithFields создает логгер с дополнительными полями
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	// Простая реализация - в реальном проекте можно использовать структурированное логирование
	return l
}

// SetLevel устанавливает уровень логирования
func (l *Logger) SetLevel(level Level) {
	l.level = level
} 