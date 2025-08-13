// main.go
//
// Package logging provides a simple asynchronous file and console logger for
// Nexus. It writes daily log files and colorizes console output. The logger
// is initialized from configuration and exposes convenience helpers for each
// log level and formatted variants.
//
// Author: Mark Oxley
// Company: DaggerTech
// Created: 2025
//
// Copyright (c) 2025 DaggerTech. All rights reserved.
// Package logging implements Nexus's lightweight logging subsystem with
// convenience level APIs and async file writing.
package chronos

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"
)

// Log represents a single log entry with timestamp, level, and message.
type Log struct {
	TimeStamp time.Time
	Level     string
	Message   string
}

// Logging is the logger instance handling level filtering and async writes.
// Use Init to configure the global logger used by package-level helpers.
type Logging struct {
	config   *Config
	path     string
	logChan  chan Log
	logLevel int
}

var logger *Logging
var mu sync.Mutex

// newLogging creates a new logger writing daily files to the given path and
// filtering below the provided log level.
func newLogging(cfg *Config, logLevel int) *Logging {
	os.MkdirAll(cfg.Location, 0755)
	l := &Logging{
		config:   cfg,
		path:     cfg.Location,
		logChan:  make(chan Log, 10000),
		logLevel: logLevel,
	}
	return l
}

// Init initializes the package-level logger from configuration and starts the
// background writer goroutine.
func Init(cfg *Config) error {
	if cfg == nil {
		cfg = &Config{}
	}
	if cfg.AppName == "" {
		return errors.New("AppName is required")
	}
	if cfg.Location == "" {
		if runtime.GOOS == "windows" {
			cfg.Location = fmt.Sprintf("C:\\ProgramData\\%s\\logs", cfg.AppName)
		} else {
			cfg.Location = fmt.Sprintf("/var/log/%s", cfg.AppName)
		}
	}
	if cfg.FilePeriod == "" {
		cfg.FilePeriod = LogPeriodHour
	}

	if cfg.Level == "" {
		cfg.Level = INFO
	}
	logLevel, ok := logLevels[cfg.Level]
	if !ok {
		return fmt.Errorf("invalid log level: %s", cfg.Level)
	}
	logger = newLogging(cfg, logLevel)
	os.Mkdir(cfg.Location, 0755)
	go logger.start()

	// Optionally install automatic graceful shutdown on common termination signals.
	if cfg.AutoStop {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sigc
			Stop()
		}()
	}
	return nil
}

// filename derives the log filename for the provided timestamp according to
// the configured rotation period (`Config.FilePeriod`).
//
// Formats by period:
// - LogPeriodHour  => nexus_YYYY-MM-DDTHH.log
// - LogPeriodDay   => nexus_YYYY-MM-DD.log
// - LogPeriodWeek  => nexus_YYYY-WW.log (ISO week number)
// - LogPeriodMonth => nexus_YYYY-MM.log
// - LogPeriodYear  => nexus_YYYY.log
//
// If an unknown period is configured, a daily filename is used as a fallback.
func (l *Logging) filename(t time.Time) string {
	datePart := ""
	switch l.config.FilePeriod {
	case LogPeriodHour:
		datePart = t.Format("2006-01-02T15")
	case LogPeriodDay:
		datePart = t.Format("2006-01-02")
	case LogPeriodWeek:
		y, w := t.ISOWeek()
		datePart = fmt.Sprintf("%04d-%02d", y, w)
	case LogPeriodMonth:
		datePart = t.Format("2006-01")
	case LogPeriodYear:
		datePart = t.Format("2006")
	default:
		return fmt.Sprintf("nexus_%s.log", t.Format("2006-01-02"))
	}
	return fmt.Sprintf("nexus_%s.log", datePart)
}

// start runs the background writer loop. It listens on l.logChan and appends
// formatted log lines to the appropriate file (as determined by filename()).
//
// Notes:
// - Files are opened in append mode and created if they don't exist.
// - I/O errors are written to stderr and the loop continues.
// - The loop terminates when the channel is closed by Stop().
func (l *Logging) start() {
	for log := range l.logChan {
		filename := l.filename(log.TimeStamp)
		fullpath := filepath.Join(l.path, filename)

		// Open the file in append mode, or create it if it doesn't exist.
		file, err := os.OpenFile(fullpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			// If the log file can't be opened, print an error to stderr and continue.
			fmt.Fprintf(os.Stderr, "ERROR: could not open log file %s: %v\n", fullpath, err)
			continue
		}

		output := fmt.Sprintf("%s\t%s\t%s\n", log.TimeStamp.Format("15:04:05"), log.Level, log.Message)

		// Write the log message to the file.
		if _, err := file.WriteString(output); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not write to log file %s: %v\n", fullpath, err)
		}

		// Close the file handle.
		file.Close()
	}
}

// addLog applies level filtering, writes to console with color, and enqueues
// the entry for async file persistence.
func (l *Logging) addLog(log Log) {
	if logger == nil {
		return
	}
	if logLevels[log.Level] < l.logLevel {
		return
	}

	// Color codes for terminal output
	const (
		colorRed    = "\033[31m"
		colorGreen  = "\033[32m"
		colorYellow = "\033[33m"
		colorBlue   = "\033[34m"
		colorPurple = "\033[35m"
		colorReset  = "\033[0m"
	)

	var color string
	switch log.Level {
	case "FATAL":
		color = colorPurple
	case "ERROR":
		color = colorRed
	case "WARN":
		color = colorYellow
	case "INFO":
		color = colorGreen
	case "DEBUG":
		color = colorBlue
	default:
		color = colorReset
	}

	fmt.Printf("%s%s\t%s\t%s%s\n", color, log.TimeStamp.Format("15:04:05"), log.Level, log.Message, colorReset)
	l.logChan <- log
}

// Stop gracefully shuts down the logger and releases the package-level logger.
func Stop() {
	mu.Lock()
	defer mu.Unlock()
	if logger == nil {
		return
	}
	close(logger.logChan)
	logger = nil
}

// Error logs a message at ERROR level.
func Error(msg string) {
	log := Log{
		TimeStamp: time.Now(),
		Level:     "ERROR",
		Message:   msg,
	}
	logger.addLog(log)
}

// Info logs a message at INFO level.
func Info(msg string) {
	log := Log{
		TimeStamp: time.Now(),
		Level:     "INFO",
		Message:   msg,
	}
	logger.addLog(log)
}

// Debug logs a message at DEBUG level.
func Debug(msg string) {
	log := Log{
		TimeStamp: time.Now(),
		Level:     "DEBUG",
		Message:   msg,
	}
	logger.addLog(log)
}

// Warn logs a message at WARN level.
func Warn(msg string) {
	log := Log{
		TimeStamp: time.Now(),
		Level:     "WARN",
		Message:   msg,
	}
	logger.addLog(log)
}

// Fatal logs a message at FATAL level.
func Fatal(msg string) {
	log := Log{
		TimeStamp: time.Now(),
		Level:     "FATAL",
		Message:   msg,
	}
	logger.addLog(log)
}

// Errorf logs a formatted message at ERROR level.
func Errorf(format string, args ...interface{}) {
	Error(fmt.Sprintf(format, args...))
}

// Infof logs a formatted message at INFO level.
func Infof(format string, args ...interface{}) {
	Info(fmt.Sprintf(format, args...))
}

// Debugf logs a formatted message at DEBUG level.
func Debugf(format string, args ...interface{}) {
	Debug(fmt.Sprintf(format, args...))
}

// Warnf logs a formatted message at WARN level.
func Warnf(format string, args ...interface{}) {
	Warn(fmt.Sprintf(format, args...))
}

// Fatalf logs a formatted message at FATAL level.
func Fatalf(format string, args ...interface{}) {
	Fatal(fmt.Sprintf(format, args...))
}
