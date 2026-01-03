// Package configuration provides functions to save and read configuration structures.
//
// Configuration files are read from/save to XDG_CONFIG_HOME/<executable name>/config.
package configuration

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v4"
)

const permission = 0o0600

var (
	cached any
	// ErrInterfaceConversion indicates that supplied T is different from cached type.
	ErrInterfaceConversion = errors.New("interface conversion")
	// ErrYAMLMarshal indicates error marshalling supplied data to YAML.
	ErrYAMLMarshal = errors.New("unable to marshal data to yaml")
)

// Get returns the configuration data for the supplied configuration struct type T, caching it after first retrieval.
// Error will be returned if:
//
// - both XDG_CONFIG_HOME and HOME env vars not set.
//
// - user lacks permission to read from XDG_CONFIG_HOME/<executable name>/config.
//
// - on subsequent calls, supplied T must be same as original T.
//
// - config file cannot be converted to supplied T.
func Get[T any](config *T) error {
	if cached == nil {
		d, err := fromFile[T]()
		if err != nil {
			return err
		}
		cached = &d
	}
	data, ok := cached.(*T)
	if !ok {
		return fmt.Errorf(
			"%w: wanted %T but cached type is %T",
			ErrInterfaceConversion,
			config,
			cached,
		)
	}
	*config = *data
	return nil
}

// Save saves the provided struct as a yaml config file in $XDG_CONFIG_HOME/executable name/config
// and updates the cache.
// Error will be returned if:
//
// - both XDG_CONFIG_HOME and HOME env vars not set.
//
// - user lacks permission to write to XDG_CONFIG_HOME/<executable name>/config.
//
// - supplied strut T cannot be marshalled to yaml.
func Save[T any](config *T) (err error) {
	// yaml.Marshal will panic with invalid data
	defer func() {
		if v := recover(); v != nil {
			err = fmt.Errorf("%w: %v", ErrYAMLMarshal, v)
		}
	}()
	progName := filepath.Base(os.Args[0])
	xdg, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("configuration dir %w", err)
	}
	cfgfile := filepath.Join(xdg, progName, "/config")
	bytes, err := yaml.Marshal(config)
	// this err check is unnecessary, yaml.Marshal will panic with invalid data
	if err != nil {
		return fmt.Errorf("%w: %w", ErrYAMLMarshal, err)
	}
	if err := os.WriteFile(cfgfile, bytes, permission); err != nil {
		return fmt.Errorf("unable to write file %s: %w", cfgfile, err)
	}
	cached = config
	return nil
}

// func fromFile reads the yaml configuration file and unmarshals it into a struct of type T
// config file location is $XDG_CONFIG_HOME/executable name/config.
func fromFile[T any]() (T, error) {
	var data T
	progName := filepath.Base(os.Args[0])
	xdg, err := os.UserConfigDir()
	if err != nil {
		return data, fmt.Errorf("configuration dir %w", err)
	}
	bytes, err := os.ReadFile(filepath.Join(xdg, progName, "config"))
	if err != nil {
		return data, fmt.Errorf("read config file %w", err)
	}
	if err := yaml.Unmarshal(bytes, &data); err != nil {
		return data, fmt.Errorf("unmarshal %w", err)
	}
	return data, nil
}
