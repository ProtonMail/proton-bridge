package logs

import "github.com/sirupsen/logrus"

// GetLogEntry returns logrus.Entry with PID and `packageName`.
func GetLogEntry(packageName string) *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"pkg": packageName,
	})
}
