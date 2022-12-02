package isglb

import "github.com/sirupsen/logrus"

var log = logrus.StandardLogger()

func SetLogger(logger *logrus.Logger) {
	log = logger
}

func GetLogger() *logrus.Logger {
	return log
}
