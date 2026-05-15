package services

import (
	"github.com/alpyxn/aeterna/backend/internal/config/common"
)

type LoggingModule struct{}

func (LoggingModule) Name() string { return "LoggingModule" }
func (LoggingModule) Section() string {
	return "logging"
}

func init() {
	common.Register(LoggingModule{})
}

type LoggingSection struct {
	Level      string
	Format     string
	File       string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
}

func (LoggingModule) LoadAndValidate() (LoggingSection, error) {
	return LoggingSection{
		Level:      common.GetenvTrim("LOG_LEVEL"),
		Format:     common.GetenvTrim("LOG_FORMAT"),
		File:       common.GetenvTrim("LOG_FILE"),
		MaxSize:    common.GetInt("LOG_MAX_SIZE", common.DefaultLogMaxSize),
		MaxBackups: common.GetInt("LOG_MAX_BACKUPS", common.DefaultLogMaxBackups),
		MaxAge:     common.GetInt("LOG_MAX_AGE", common.DefaultLogMaxAge),
		Compress:   common.GetBool("LOG_COMPRESS", common.DefaultLogCompress),
	}, nil
}
