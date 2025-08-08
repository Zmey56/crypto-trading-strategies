package logger

import (
	"fmt"
	"log"
	"os"
)

// Level is a logging verbosity level
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// Logger is a minimal logger wrapper
type Logger struct {
	level  Level
	logger *log.Logger
}

// New creates a new logger writing to stdout
func New(level Level) *Logger {
	return &Logger{
		level:  level,
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

// NewWithFile creates a logger writing to a file
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

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level <= LevelDebug {
		l.logger.Printf("[DEBUG] "+format, args...)
	}
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level <= LevelInfo {
		l.logger.Printf("[INFO] "+format, args...)
	}
}

// Warn logs a warning
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.level <= LevelWarn {
		l.logger.Printf("[WARN] "+format, args...)
	}
}

// Error logs an error
func (l *Logger) Error(format string, args ...interface{}) {
	if l.level <= LevelError {
		l.logger.Printf("[ERROR] "+format, args...)
	}
}

// Fatal logs a fatal error and exits
func (l *Logger) Fatal(format string, args ...interface{}) {
	if l.level <= LevelFatal {
		l.logger.Printf("[FATAL] "+format, args...)
		os.Exit(1)
	}
}

// WithFields returns the same logger; placeholder for structured logging
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	// Простая реализация - в реальном проекте можно использовать структурированное логирование
	return l
}

// SetLevel adjusts logging verbosity level
func (l *Logger) SetLevel(level Level) {
	l.level = level
}
