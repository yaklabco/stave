package parallelism

import (
	"log/slog"
	"os"
	"runtime"
	"strconv"
	"strings"
)

const (
	StaveNumProcessorsEnvVar = "STAVE_NUM_PROCESSORS"

	GoMaxProcsEnvVar = "GOMAXPROCS"
)

func getNumProcessors() int {
	return runtime.NumCPU()
}

func Apply(theEnv map[string]string) error {
	strFromEnv := strings.TrimSpace(os.Getenv(StaveNumProcessorsEnvVar))
	var numProcessors int
	if strFromEnv != "" {
		var err error
		numProcessors, err = strconv.Atoi(strFromEnv)
		if err != nil {
			return err
		}
	} else {
		numProcessors = getNumProcessors()
	}

	slog.Debug("setting parallelism-related env vars", slog.Int("num_processors", numProcessors))

	newValStr := strconv.Itoa(numProcessors)

	runtime.GOMAXPROCS(numProcessors)
	theEnv[StaveNumProcessorsEnvVar] = newValStr
	theEnv[GoMaxProcsEnvVar] = newValStr

	return nil
}
