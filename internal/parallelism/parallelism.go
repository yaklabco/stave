package parallelism

import (
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

func Apply(envMap map[string]string) error {
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

	newValStr := strconv.Itoa(numProcessors)

	runtime.GOMAXPROCS(numProcessors)
	envMap[StaveNumProcessorsEnvVar] = newValStr
	envMap[GoMaxProcsEnvVar] = newValStr

	return nil
}
