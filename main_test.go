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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNewLogging(t *testing.T) {
	path := "/tmp/test.log"
	logLevel := 2

	l := NewLogging(path, logLevel)

	if l.path != path {
		t.Errorf("Expected path to be %s, but got %s", path, l.path)
	}

	if l.level != logLevel {
		t.Errorf("Expected level to be %d, but got %d", logLevel, l.level)
	}

	if l.logChan == nil {
		t.Error("Expected logChan to be initialized, but it was nil")
	}
}

func TestAddLog(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "logtest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	logger = NewLogging(tempDir, logLevels[INFO])
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.start()
	}()

	Info("test message")

	time.Sleep(100 * time.Millisecond)

	Stop()
	wg.Wait()

	filename := fmt.Sprintf("nexus_%s.log", time.Now().Format("2006-01-02"))
	fullpath := filepath.Join(tempDir, filename)

	content, err := ioutil.ReadFile(fullpath)
	if err != nil {
		t.Fatalf("could not read log file: %v", err)
	}

	if !strings.Contains(string(content), "test message") {
		t.Errorf("log file does not contain expected message")
	}
}

func TestLoggingLevels(t *testing.T) {
	logger = NewLogging("/tmp", logLevels[WARN])

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

func TestStop(t *testing.T) {
	logger = NewLogging("/tmp", logLevels[INFO])
	go logger.start()

	Stop()

	if logger != nil {
		t.Error("Expected logger to be nil after Stop(), but it was not")
	}
}

func setupBenchmark(b *testing.B) func() {
	tempDir, err := ioutil.TempDir("", "benchlog")
	if err != nil {
		b.Fatal(err)
	}
	logger = NewLogging(tempDir, logLevels[DEBUG])
	go logger.start()

	return func() {
		Stop()
		os.RemoveAll(tempDir)
	}
}

func BenchmarkInfo(b *testing.B) {
	teardown := setupBenchmark(b)
	defer teardown()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Info("info message")
	}
}

func BenchmarkInfof(b *testing.B) {
	teardown := setupBenchmark(b)
	defer teardown()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Infof("info message %d", i)
	}
}

func BenchmarkDebug(b *testing.B) {
	teardown := setupBenchmark(b)
	defer teardown()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Debug("debug message")
	}
}

func BenchmarkDebugf(b *testing.B) {
	teardown := setupBenchmark(b)
	defer teardown()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Debugf("debug message %d", i)
	}
}

func BenchmarkWarn(b *testing.B) {
	teardown := setupBenchmark(b)
	defer teardown()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Warn("warn message")
	}
}

func BenchmarkWarnf(b *testing.B) {
	teardown := setupBenchmark(b)
	defer teardown()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Warnf("warn message %d", i)
	}
}
