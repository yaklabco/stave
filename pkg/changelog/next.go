package changelog

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/caarlos0/svu/v3/pkg/svu"
	"github.com/spf13/viper"
)

// ErrEmptyVersion is returned when tools returns an empty version.
var ErrEmptyVersion = errors.New("tool returned empty version")

// NextVersion returns the next semantic version.
// It strips off the leading 'v', if present, to match CHANGELOG heading format.
func NextVersion() (string, error) {
	version, err := NextTag()
	if err != nil {
		return "", err
	}

	// Strip leading 'v' to match CHANGELOG format
	version = strings.TrimPrefix(version, "v")

	return version, nil
}

// NextTag returns the next tag.
// It does not strip off the leading 'v' if present.
func NextTag() (string, error) {
	viperInstance, err := loadSVUConfig()
	if err != nil {
		return "", fmt.Errorf("loading svu config: %w", err)
	}

	// All of the following reflection shenanigans are necessary because the `svu.option` type is... unexported.
	fn := reflect.ValueOf(svu.Next)
	fnType := fn.Type()
	optionType := fnType.In(fnType.NumIn() - 1).Elem()

	svuOpts := make([]any, 0, 5) //nolint:mnd // This is the maximum number of options we expect to append (see below).
	svuOpts = append(
		svuOpts,
		svu.WithPrefix(viperInstance.GetString("prefix")),
		svu.WithPattern(viperInstance.GetString("pattern")),
	)
	if len(viperInstance.GetStringSlice("log.directory")) > 0 {
		svuOpts = append(svuOpts, svu.WithDirectories(viperInstance.GetStringSlice("log.directory")...))
	}
	if viperInstance.GetBool("always") {
		svuOpts = append(svuOpts, svu.Always())
	}
	if viperInstance.GetBool("v0") {
		svuOpts = append(svuOpts, svu.KeepV0())
	}

	reflectedSlice := reflect.MakeSlice(reflect.SliceOf(optionType), len(svuOpts), len(svuOpts))
	for i, opt := range svuOpts {
		reflectedSlice.Index(i).Set(reflect.ValueOf(opt))
	}

	results := fn.CallSlice([]reflect.Value{reflectedSlice})
	if len(results) != 2 {
		return "", fmt.Errorf("svu.Next call returned more than two results: %v", results)
	}

	out, ok := results[0].Interface().(string)
	if !ok {
		return "", fmt.Errorf("svu.Next call returned non-string result: %v", results)
	}

	err = nil
	if !results[1].IsNil() {
		err, ok = results[1].Interface().(error)
		if !ok {
			return "", fmt.Errorf("svu.Next call returned non-error result: %v", results)
		}
	}

	if err != nil {
		return "", fmt.Errorf("svu.Next: %w", err)
	}

	version := strings.TrimSpace(out)
	if version == "" {
		return "", fmt.Errorf("svu.Next: %w", ErrEmptyVersion)
	}

	return version, nil
}

func loadSVUConfig() (*viper.Viper, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("getting home directory: %w", err)
	}

	config, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("getting config directory: %w", err)
	}

	viperInstance := viper.New()
	viperInstance.AutomaticEnv()
	viperInstance.SetEnvPrefix("svu")
	viperInstance.AddConfigPath(".")
	viperInstance.AddConfigPath(config)
	viperInstance.AddConfigPath(home)
	viperInstance.SetConfigName(".svu")
	viperInstance.SetConfigType("yaml")
	err = viperInstance.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("reading svu config: %w", err)
	}

	return viperInstance, nil
}
