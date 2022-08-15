package logger

import (
	"github.com/jiuhuche120/jreminder/pkg/config"
	"github.com/sirupsen/logrus"
)

func NewLogger(cfg *config.Config) *logrus.Logger {
	return logrus.New()
}
