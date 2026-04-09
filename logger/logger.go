package logger

import (
	"github.com/sirupsen/logrus"
	"os"
)

var Log = logrus.New()

func Init() {
	Log.SetOutput(os.Stdout)
	Log.SetFormatter(&logrus.JSONFormatter{})
	Log.SetLevel(logrus.InfoLevel)
}
