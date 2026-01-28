// levels.go
//
// # Chronos Logging - Log Levels
//
// Defines human-readable log level strings and their corresponding
// internal numeric severities used throughout the logging subsystem.
//
// Author: Mark Oxley
// Company: DaggerTech
// Created: 2025
//
// Copyright (c) 2025 DaggerTech. All rights reserved.
//
// This file declares the set of supported log levels and their relative
// severities. The numeric severities are used for filtering in
// `Logging.addLog()` and when initializing the logger in `Init()`.
//
// Severity ordering (low -> high):
//
//	INFO(1) < DEBUG(2) < WARN(3) < ERROR(4) < FATAL(5)
//
// Note: Higher numbers are treated as more severe. Messages are emitted when
// their severity is greater than or equal to the configured threshold.
package chronos

// Level names used throughout the logger and configuration.
//
// These constants are the canonical string representations written to the
// console and log files, and accepted via `Config.Level`.
const (
	INFO  = "INFO"
	DEBUG = "DEBUG"
	WARN  = "WARN"
	ERROR = "ERROR"
	FATAL = "FATAL"
)

// logLevels maps level names to their internal severity for filtering.
//
// The values are intentionally monotonic increasing to reflect severity.
// They are used in comparisons like:
//   if logLevels[log.Level] < l.logLevel { return }
// so any message with a severity lower than the configured threshold is
// dropped before printing or enqueuing for file persistence.
var logLevels map[string]int = map[string]int{
	DEBUG: 1,
	INFO:  2,
	WARN:  3,
	ERROR: 4,
	FATAL: 5,
}
