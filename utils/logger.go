package utils

import (
	"path/filepath"
	"strings"

	"im/mlog"
	"im/model"
	"im/utils/fileutils"
)

const (
	LOG_ROTATE_SIZE = 10000
	LOG_FILENAME    = "log"
)

func MloggerConfigFromLoggerConfig(s *model.LogSettings) *mlog.LoggerConfiguration {
	return &mlog.LoggerConfiguration{
		EnableConsole: *s.EnableConsole,
		ConsoleJson:   *s.ConsoleJson,
		ConsoleLevel:  strings.ToLower(*s.ConsoleLevel),
		EnableFile:    *s.EnableFile,
		FileJson:      *s.FileJson,
		FileLevel:     strings.ToLower(*s.FileLevel),
		FileLocation:  GetLogFileLocation(*s.FileLocation),
	}
}

func GetLogFileLocation(fileLocation string) string {
	if fileLocation == "" {
		fileLocation, _ = fileutils.FindDir("logs")
	}

	return filepath.Join(fileLocation, LOG_FILENAME)
}

// DON'T USE THIS Modify the level on the app logger
func DisableDebugLogForTest() {
	mlog.GloballyDisableDebugLogForTest()
}

// DON'T USE THIS Modify the level on the app logger
func EnableDebugLogForTest() {
	mlog.GloballyEnableDebugLogForTest()
}
