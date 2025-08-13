// levels.go
//
// Package logging - Log Levels
//
// Defines human-readable log level strings and their corresponding
// internal numeric severities used by the logging subsystem.
//
// Author: Mark Oxley
// Company: DaggerTech
// Created: 2025
//
// Copyright (c) 2025 DaggerTech. All rights reserved.
// Package logging declares log level identifiers and their severities used by
// the logging subsystem.
package chronos

// Level names used throughout the logger and configuration.
const (
	INFO  = "INFO"
	DEBUG = "DEBUG"
	WARN  = "WARN"
	ERROR = "ERROR"
	FATAL = "FATAL"
)

// logLevels maps level names to their internal severity for filtering.
var logLevels map[string]int = map[string]int{
	INFO:  1,
	DEBUG: 2,
	WARN:  3,
	ERROR: 4,
	FATAL: 5,
}
