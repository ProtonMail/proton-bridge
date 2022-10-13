package restarter

import (
	"os"
	"strconv"
	"strings"

	"github.com/bradenaw/juniper/xslices"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/execabs"
)

const BridgeCrashCount = "BRIDGE_CRASH_COUNT"

type Restarter struct {
	restart bool
	crash   bool

	exe   string
	flags []string
}

func New(exe string) *Restarter {
	return &Restarter{exe: exe}
}

func (restarter *Restarter) Set(restart, crash bool) {
	restarter.restart = restart
	restarter.crash = crash
}

func (restarter *Restarter) Override(exe string) {
	restarter.exe = exe
}

func (restarter *Restarter) AddFlags(flags ...string) {
	restarter.flags = append(restarter.flags, flags...)
}

func (restarter *Restarter) Restart() {
	if !restarter.restart {
		return
	}

	if restarter.exe == "" {
		return
	}

	env := getEnvMap()

	if restarter.crash {
		env[BridgeCrashCount] = increment(env[BridgeCrashCount])
	} else {
		delete(env, BridgeCrashCount)
	}

	cmd := execabs.Command(restarter.exe, xslices.Join(os.Args[1:], restarter.flags)...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = getEnvList(env)

	if err := run(cmd); err != nil {
		logrus.WithError(err).Error("Failed to restart")
	}
}

func getEnvMap() map[string]string {
	env := make(map[string]string)

	for _, entry := range os.Environ() {
		if split := strings.SplitN(entry, "=", 2); len(split) == 2 {
			env[split[0]] = split[1]
		}
	}

	return env
}

func getEnvList(envMap map[string]string) []string {
	env := make([]string, 0, len(envMap))

	for key, value := range envMap {
		env = append(env, key+"="+value)
	}

	return env
}

func increment(value string) string {
	var valueInt int

	if parsed, err := strconv.Atoi(value); err == nil {
		valueInt = parsed
	}

	return strconv.Itoa(valueInt + 1)
}
