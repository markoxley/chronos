// main_test.go
//
// # Package logging - Tests and Benchmarks
//
// Covers initialization, log writing to daily files, level filtering,
// graceful shutdown, and performance of common logging functions.
//
// Author: Mark Oxley
// Company: DaggerTech
// Created: 2025
//
// Copyright (c) 2025 DaggerTech. All rights reserved.
package chronos

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// getConfig returns a baseline configuration used across tests and benchmarks.
// Individual tests may override fields (e.g., Location) to run in temp dirs.
func getConfig() *Config {
	return &Config{
		AppName:    "test",
		Location:   "/tmp",
		FilePeriod: LogPeriodHour,
		Level:      INFO,
	}
}

// TestNewLogging verifies that a new logger initializes the path, level, and
// channel correctly using the provided configuration and log level.
func TestNewLogging(t *testing.T) {
	path := "/tmp"
	logLevel := 2

	l := newLogging(getConfig(), logLevel)

	if l.path != path {
		t.Errorf("Expected path to be %s, but got %s", path, l.path)
	}

	if l.logLevel != logLevel {
		t.Errorf("Expected level to be %d, but got %d", logLevel, l.logLevel)
	}

	if l.logChan == nil {
		t.Error("Expected logChan to be initialized, but it was nil")
	}
}

// TestAddLog validates that an INFO message is printed and persisted to the
// expected file, using the configured rotation period and temp directory.
func TestAddLog(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "logtest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	cfg := getConfig()
	cfg.Location = tempDir
	logger = newLogging(cfg, logLevels[INFO])
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.start()
	}()

	Info("test message")

	time.Sleep(100 * time.Millisecond)

	// capture expected filename before stopping logger to avoid nil deref
	filename := logger.filename(time.Now())

	Stop()
	wg.Wait()

	fullpath := filepath.Join(tempDir, filename)

	content, err := os.ReadFile(fullpath)
	if err != nil {
		t.Fatalf("could not read log file: %v", err)
	}

	if !strings.Contains(string(content), "test message") {
		t.Errorf("log file does not contain expected message")
	}
}

// TestLoggingLevels ensures level-based filtering works: DEBUG/INFO are
// filtered when threshold is WARN, while WARN/ERROR are accepted.
func TestLoggingLevels(t *testing.T) {
	logger = newLogging(getConfig(), logLevels[WARN])

	Debug("debug message")
	Info("info message")
	if len(logger.logChan) != 0 {
		t.Errorf("Expected 0 messages in log channel, but got %d", len(logger.logChan))
	}

	Warn("warn message")
	if len(logger.logChan) != 1 {
		t.Errorf("Expected 1 message in log channel, but got %d", len(logger.logChan))
	}

	Error("error message")
	if len(logger.logChan) != 2 {
		t.Errorf("Expected 2 messages in log channel, but got %d", len(logger.logChan))
	}
	Stop()
}

// TestStop confirms Stop() closes the channel and clears the global logger.
func TestStop(t *testing.T) {
	cfg := getConfig()
	cfg.Location = "/tmp"
	logger = newLogging(cfg, logLevels[INFO])
	go logger.start()

	Stop()

	if logger != nil {
		t.Error("Expected logger to be nil after Stop(), but it was not")
	}
}

// TestAutoStopOnSIGTERM verifies that enabling AutoStop causes the logger to
// gracefully stop when the process receives SIGTERM.
func TestAutoStopOnSIGTERM(t *testing.T) {
	// Ensure a clean slate
	Stop()

	tempDir, err := os.MkdirTemp("", "autostop-term")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	cfg := getConfig()
	cfg.Location = tempDir
	cfg.AutoStop = true
	if err := Init(cfg); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Send SIGTERM to ourselves
	if runtime.GOOS == "windows" {
		t.Skip("SIGTERM not supported on Windows; skipping test")
	}
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("failed to find process: %v", err)
	}
	if err := p.Signal(syscall.SIGTERM); err != nil {
		t.Fatalf("failed to send SIGTERM: %v", err)
	}

	// Allow handler to run
	time.Sleep(100 * time.Millisecond)

	if logger != nil {
		t.Error("expected logger to be nil after SIGTERM AutoStop, but it was not")
	}
}

// TestAutoStopThenManualStop ensures calling Stop() after AutoStop has already
// executed is safe and remains idempotent.
func TestAutoStopThenManualStop(t *testing.T) {
	// Ensure a clean slate
	Stop()

	tempDir, err := os.MkdirTemp("", "autostop-int")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	cfg := getConfig()
	cfg.Location = tempDir
	cfg.AutoStop = true
	if err := Init(cfg); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Trigger AutoStop via SIGINT (Ctrl-C)
	if runtime.GOOS == "windows" {
		t.Skip("SIGINT delivery not supported on Windows; skipping test")
	}
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("failed to find process: %v", err)
	}
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatalf("failed to send SIGINT: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	// Now call Stop() manually — should be a no-op and not panic.
	Stop()

	if logger != nil {
		t.Error("expected logger to be nil after AutoStop + Stop")
	}
}

// TestExternalHandlerInvoked ensures the external handler receives the
// correct log data when set and a log entry is emitted.
func TestExternalHandlerInvoked(t *testing.T) {
	// Ensure a clean slate
	Stop()
	SetHandler(nil)

	logger = newLogging(getConfig(), logLevels[INFO])
	defer func() {
		Stop()
		SetHandler(nil)
	}()

	called := make(chan Log, 1)
	SetHandler(func(ts time.Time, level, message string) {
		called <- Log{TimeStamp: ts, Level: level, Message: message}
	})

	start := time.Now()
	Info("external handler message")
	end := time.Now()

	select {
	case got := <-called:
		if got.Level != "INFO" {
			t.Fatalf("expected level INFO, got %s", got.Level)
		}
		if got.Message != "external handler message" {
			t.Fatalf("expected message to match, got %s", got.Message)
		}
		if got.TimeStamp.Before(start) || got.TimeStamp.After(end) {
			t.Fatalf("timestamp %v not within expected range", got.TimeStamp)
		}
	default:
		t.Fatal("expected external handler to be called")
	}
}

// TestExternalHandlerNotInvokedBelowLevel ensures the external handler is not
// called when the log entry is filtered out by the current log level.
func TestExternalHandlerNotInvokedBelowLevel(t *testing.T) {
	// Ensure a clean slate
	Stop()
	SetHandler(nil)

	logger = newLogging(getConfig(), logLevels[WARN])
	defer func() {
		Stop()
		SetHandler(nil)
	}()

	called := make(chan struct{}, 1)
	SetHandler(func(ts time.Time, level, message string) {
		called <- struct{}{}
	})

	Info("should be filtered")

	select {
	case <-called:
		t.Fatal("external handler should not be called for filtered log")
	default:
		// expected
	}
}

// setupBenchmark creates a temporary logger at DEBUG level and returns a
// teardown function that stops the logger and cleans up the temp directory.
func setupBenchmark(b *testing.B) func() {
	tempDir, err := os.MkdirTemp("", "benchlog")
	if err != nil {
		b.Fatal(err)
	}
	logger = newLogging(getConfig(), logLevels[DEBUG])
	go logger.start()

	return func() {
		Stop()
		os.RemoveAll(tempDir)
	}
}

// BenchmarkInfo measures the throughput of logging INFO messages.
func BenchmarkInfo(b *testing.B) {
	teardown := setupBenchmark(b)
	defer teardown()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Info("info message")
	}
}

// BenchmarkInfof measures the throughput of formatted INFO messages.
func BenchmarkInfof(b *testing.B) {
	teardown := setupBenchmark(b)
	defer teardown()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Infof("info message %d", i)
	}
}

// BenchmarkDebug measures the throughput of logging DEBUG messages.
func BenchmarkDebug(b *testing.B) {
	teardown := setupBenchmark(b)
	defer teardown()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Debug("debug message")
	}
}

// BenchmarkDebugf measures the throughput of formatted DEBUG messages.
func BenchmarkDebugf(b *testing.B) {
	teardown := setupBenchmark(b)
	defer teardown()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Debugf("debug message %d", i)
	}
}

// BenchmarkWarn measures the throughput of logging WARN messages.
func BenchmarkWarn(b *testing.B) {
	teardown := setupBenchmark(b)
	defer teardown()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Warn("warn message")
	}
}

// BenchmarkWarnf measures the throughput of formatted WARN messages.
func BenchmarkWarnf(b *testing.B) {
	teardown := setupBenchmark(b)
	defer teardown()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Warnf("warn message %d", i)
	}
}
