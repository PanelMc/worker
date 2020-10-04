package infra

import (
	"github.com/sirupsen/logrus"

	nested "github.com/antonfisher/nested-logrus-formatter"
)

// InitializeLogger initializes the global logger with the default configuration
func InitializeLogger() {
	logrus.SetFormatter(&logrus.TextFormatter{})

	logrus.SetFormatter(&nested.Formatter{
		HideKeys: true,
		NoColors: true,
	})
	logrus.SetLevel(logrus.TraceLevel)
}
