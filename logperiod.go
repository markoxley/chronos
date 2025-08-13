// logperiod.go
//
// Chronos Logging - Log File Rotation Periods
//
// Defines the `LogPeriod` type and supported rotation cadences used to derive
// log filenames in `(*Logging).filename()` and control how frequently new log
// files are created.
//
// Author: Mark Oxley
// Company: DaggerTech
// Created: 2025
//
// Copyright (c) 2025 DaggerTech. All rights reserved.
package chronos

// LogPeriod represents the cadence at which log files rotate and the
// timestamp granularity embedded into the log filename.
//
// It is configured via `Config.FilePeriod`. If not set, it defaults to
// `LogPeriodHour` in `Init()`.
type LogPeriod string

// Supported rotation cadences. These impact the filename format used by
// `(*Logging).filename(t time.Time)`:
//
// - LogPeriodHour  => nexus_YYYY-MM-DDTHH.log
// - LogPeriodDay   => nexus_YYYY-MM-DD.log
// - LogPeriodWeek  => nexus_YYYY-WW.log (ISO Week: YYYY is ISO year; WW is week)
// - LogPeriodMonth => nexus_YYYY-MM.log
// - LogPeriodYear  => nexus_YYYY.log
//
// Notes about weeks:
// - Week uses ISO 8601 week-numbering (`time.Time.ISOWeek()`), which may differ
//   from the calendar year around year boundaries. This behavior is by design
//   to group logs according to ISO weeks.
const (
    LogPeriodHour  LogPeriod = "Hour"
    LogPeriodDay   LogPeriod = "Day"
    LogPeriodWeek  LogPeriod = "Week"
    LogPeriodMonth LogPeriod = "Month"
    LogPeriodYear  LogPeriod = "Year"
)
