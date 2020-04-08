// Copyright (c) 2020 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/pprof"
	"sort"
	"strconv"
	"time"

	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/sirupsen/logrus"
)

type logConfiger interface {
	GetLogDir() string
	GetLogPrefix() string
}

const (
	// Zendesk now has a file size limit of 20MB. When the last N log files
	// are zipped, it should fit under 20MB. Value in MB (average file has
	// few hundreds kB).
	maxLogFileSize = 10 * 1024 * 1024 //nolint[gochecknoglobals]
	// Including the current logfile.
	maxNumberLogFiles = 3 //nolint[gochecknoglobals]
)

// logFile is pointer to currently open file used by logrus.
var logFile *os.File //nolint[gochecknoglobals]

var logFileRgx = regexp.MustCompile("^v.*\\.log$")           //nolint[gochecknoglobals]
var logCrashRgx = regexp.MustCompile("^v.*_crash_.*\\.log$") //nolint[gochecknoglobals]

// GetLogEntry returns logrus.Entry with PID and `packageName`.
func GetLogEntry(packageName string) *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"pkg": packageName,
	})
}

// HandlePanic reports the crash to sentry or local file when sentry fails.
func HandlePanic(cfg *Config, output string) {
	if !cfg.IsDevMode() {
		// TODO: Is it okay to just create a throwaway client like this?
		c := pmapi.NewClientManager(cfg.GetAPIConfig()).GetAnonymousClient()
		defer c.Logout()

		if err := c.ReportSentryCrash(fmt.Errorf(output)); err != nil {
			log.Error("Sentry crash report failed: ", err)
		}
	}

	filename := getLogFilename(cfg.GetLogPrefix() + "_crash_")
	filepath := filepath.Join(cfg.GetLogDir(), filename)
	f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		log.Error("Cannot open file to write crash report: ", err)
		return
	}

	_, _ = f.WriteString(output)
	_ = pprof.Lookup("goroutine").WriteTo(f, 2)

	log.Warn("Crash report saved to ", filepath)
}

// GetGID returns goroutine number which can be used to distiguish logs from
// the concurent processes. Keep in mind that it returns the number of routine
// which executes the function.
func GetGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

// SetupLog set up log level, formatter and output (file or stdout).
// Returns whether should be used debug for IMAP and SMTP servers.
func SetupLog(cfg logConfiger, levelFlag string) (debugClient, debugServer bool) {
	level, useFile := getLogLevelAndFile(levelFlag)

	logrus.SetLevel(level)

	if useFile {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		setLogFile(cfg.GetLogDir(), cfg.GetLogPrefix())
		watchLogFileSize(cfg.GetLogDir(), cfg.GetLogPrefix())
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			ForceColors:     true,
			FullTimestamp:   true,
			TimestampFormat: time.StampMilli,
		})
		logrus.SetOutput(os.Stdout)
	}

	switch levelFlag {
	case "debug-client", "debug-client-json":
		debugClient = true
	case "debug-server", "debug-server-json", "trace":
		fmt.Println("THE LOG WILL CONTAIN **DECRYPTED** MESSAGE DATA")
		log.Warning("================================================")
		log.Warning("THIS LOG WILL CONTAIN **DECRYPTED** MESSAGE DATA")
		log.Warning("================================================")
		debugClient = true
		debugServer = true
	}

	return debugClient, debugServer
}

func setLogFile(logDir, logPrefix string) {
	if logFile != nil {
		return
	}

	filename := getLogFilename(logPrefix)
	var err error
	logFile, err = os.OpenFile(filepath.Join(logDir, filename), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	logrus.SetOutput(logFile)

	// Users sometimes change the name of the log file. We want to always log
	// information about bridge version (included in log prefix) and OS.
	log.Warn("Bridge version: ", logPrefix, " ", runtime.GOOS)
}

func getLogFilename(logPrefix string) string {
	currentTime := strconv.Itoa(int(time.Now().Unix()))
	return logPrefix + "_" + currentTime + ".log"
}

func watchLogFileSize(logDir, logPrefix string) {
	go func() {
		for {
			time.Sleep(60 * time.Second)
			checkLogFileSize(logDir, logPrefix)
		}
	}()
}

func checkLogFileSize(logDir, logPrefix string) {
	if logFile == nil {
		return
	}

	stat, err := logFile.Stat()
	if err != nil {
		log.Error("Log file size check failed: ", err)
		return
	}

	if stat.Size() >= maxLogFileSize {
		log.Warn("Current log file ", logFile.Name(), " is too big, opening new file")
		closeLogFile()
		setLogFile(logDir, logPrefix)
	}

	if err := clearLogs(logDir); err != nil {
		log.Error("Cannot clear logs ", err)
	}
}

func closeLogFile() {
	if logFile != nil {
		_ = logFile.Close()
		logFile = nil
	}
}

func clearLogs(logDir string) error {
	files, err := ioutil.ReadDir(logDir)
	if err != nil {
		return err
	}

	var logsWithPrefix []string
	var crashesWithPrefix []string

	for _, file := range files {
		if logFileRgx.MatchString(file.Name()) {
			if logCrashRgx.MatchString(file.Name()) {
				crashesWithPrefix = append(crashesWithPrefix, file.Name())
			} else {
				logsWithPrefix = append(logsWithPrefix, file.Name())
			}
		} else {
			// Older versions of Bridge stored logs in subfolders for each version.
			// That also has to be cleared and the functionality can be removed after some time.
			if file.IsDir() {
				if err := clearLogs(filepath.Join(logDir, file.Name())); err != nil {
					return err
				}
			} else {
				removeLog(logDir, file.Name())
			}
		}
	}

	removeOldLogs(logDir, logsWithPrefix)
	removeOldLogs(logDir, crashesWithPrefix)
	return nil
}

func removeOldLogs(logDir string, filenames []string) {
	count := len(filenames)
	if count <= maxNumberLogFiles {
		return
	}

	sort.Strings(filenames) // Sorted by timestamp: oldest first.
	for _, filename := range filenames[:count-maxNumberLogFiles] {
		removeLog(logDir, filename)
	}
}

func removeLog(logDir, filename string) {
	// We need to be sure to delete only log files.
	// Directory with logs can also contain other files.
	if !logFileRgx.MatchString(filename) {
		return
	}
	if err := os.RemoveAll(filepath.Join(logDir, filename)); err != nil {
		log.Error("Cannot remove old logs ", err)
	}
}
