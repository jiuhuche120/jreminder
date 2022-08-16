package logger

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/jiuhuche120/jreminder/pkg/config"
	"github.com/sirupsen/logrus"
)

func NewLogger(cfg *config.Config) *logrus.Logger {
	logger := logrus.New()
	level := parseLevel(cfg.Log.Level)
	logger.SetLevel(level)
	formatter := getTextFormatter()
	logger.SetFormatter(formatter)
	logger.SetReportCaller(cfg.Log.ReportCaller)
	return logger
}

func getTextFormatter() logrus.Formatter {
	return &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02T15:04:05.000",
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			_, filename := filepath.Split(f.File)
			return "", fmt.Sprintf("%12s:%-4d", filename, f.Line)
		},
	}
}

func parseLevel(level string) logrus.Level {
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		lvl = logrus.ErrorLevel
	}
	return lvl
}
