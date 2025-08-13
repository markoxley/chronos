// config.go
 //
 // Package configuration types for Chronos logging.
 //
 // Author: Mark Oxley
 // Company: DaggerTech
 // Created: 2025
 //
 // This file defines the configuration contract used to initialize and
 // control the Chronos logger. The configuration is typically provided by
 // your application and read during startup to configure log destinations,
 // rotation cadence, and default verbosity.
 package chronos

 // Config describes how the Chronos logger should operate.
 //
 // Typical usage:
 //  cfg := &Config{
 //      AppName:    "nexus",
 //      Location:   "/var/log/nexus", // or C:\\ProgramData\\Nexus\\logs on Windows
 //      FilePeriod: LogPeriodDay,       // Hour, Day, Week, Month, or Year
 //      Level:      INFO,               // DEBUG < INFO < WARN < ERROR < FATAL
 //  }
 //  if err := Init(cfg); err != nil { /* handle error */ }
 //
 // Notes:
 // - If Location is empty, a sensible OS-specific default is chosen based on
 //   AppName (see Init in main.go).
 // - If FilePeriod is empty, it defaults to Hourly rotation.
 // - If Level is empty, it defaults to INFO.
 type Config struct {
     // AppName is the logical name of your application. It is used to derive
     // OS-appropriate default log locations when Location is not explicitly set
     // (e.g., /var/log/<AppName> on Unix-like systems or
     // C:\\ProgramData\\<AppName>\\logs on Windows).
     AppName string `json:"app_name"`

     // Location is the absolute directory path where log files are written.
     // If left empty, Chronos chooses a platform-specific default derived from
     // AppName. The directory will be created with 0755 permissions if it does
     // not exist.
     Location string `json:"location"`

     // FilePeriod controls the log file rotation cadence by determining the
     // timestamp granularity embedded in the filename. Supported values are
     // LogPeriodHour, LogPeriodDay, LogPeriodWeek, LogPeriodMonth, and
     // LogPeriodYear.
     FilePeriod LogPeriod `json:"file_period"`

     // Level is the minimum log severity that will be emitted. Messages below
     // this level are filtered before being printed or enqueued for file
     // persistence. Valid values are DEBUG, INFO, WARN, ERROR, and FATAL.
     Level string `json:"level"`

     // AutoStop, when true, installs an OS signal handler (e.g., SIGINT/Ctrl-C
     // and SIGTERM) to automatically invoke Stop() so the logger flushes and
     // closes gracefully during application shutdown. Default is false to avoid
     // interfering with host application's own signal handling. Enable this if
     // you do not already manage Stop() explicitly.
     AutoStop bool `json:"auto_stop"`
 }
