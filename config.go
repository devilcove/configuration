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
	cached = map[string]any{}
	// ErrInterfaceConversion indicates that supplied T is different from cached type.
	ErrInterfaceConversion = errors.New("interface conversion")
	// ErrYAMLMarshal indicates error marshalling supplied data to YAML.
	ErrYAMLMarshal = errors.New("unable to (un)marshal data to/from yaml")
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
func Get[T any](config *T, names ...string) error {
	name := "config"
	if len(names) > 0 {
		name = names[0]
	}

	if cached[name] == nil {
		d, err := fromFile[T](name)
		if err != nil {
			return err
		}
		cached[name] = &d
	}

	data, ok := cached[name].(*T)
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

// Save saves the provided struct as a yaml config file in $XDG_CONFIG_HOME/executable name/
// and updates the cache. Config dir and file willl be created if it dosen't exist
// Error will be returned if:
//
// - both XDG_CONFIG_HOME and HOME env vars not set.
//
// - user lacks permission to write to XDG_CONFIG_HOME/<executable name>/.
//
// - supplied strut T cannot be marshalled to yaml.
func Save[T any](config *T, names ...string) (err error) {
	// yaml.Marshal will panic with invalid data
	defer func() {
		if v := recover(); v != nil {
			err = fmt.Errorf("%w: %v", ErrYAMLMarshal, v)
		}
	}()

	progName := filepath.Base(os.Args[0])
	fileName := "config"
	if len(names) > 0 {
		fileName = names[0]
	}

	xdg, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("configuration dir %w", err)
	}

	cfgfile := filepath.Join(xdg, progName, fileName)
	// check if file exists and create if necessary
	if _, err := os.Stat(cfgfile); err != nil {
		if err := createFile(xdg, progName, fileName); err != nil {
			return fmt.Errorf("unable to create file %s/%s/%s %w", xdg, progName, fileName, err)
		}
	}

	bytes, err := yaml.Marshal(config)
	// this err check is unnecessary, yaml.Marshal will panic with invalid data
	if err != nil {
		return fmt.Errorf("%w: %w", ErrYAMLMarshal, err)
	}
	if err := os.WriteFile(cfgfile, bytes, permission); err != nil {
		return fmt.Errorf("unable to write file %s: %w", cfgfile, err)
	}
	cached[fileName] = config
	return nil
}

// func fromFile reads the yaml configuration file and unmarshals it into a struct of type T
// config file location is $XDG_CONFIG_HOME/executable name/config.
func fromFile[T any](fileName string) (T, error) {
	var data T
	progName := filepath.Base(os.Args[0])
	xdg, err := os.UserConfigDir()
	if err != nil {
		return data, fmt.Errorf("configuration dir %w", err)
	}

	bytes, err := os.ReadFile(filepath.Join(xdg, progName, fileName))
	if err != nil {
		return data, fmt.Errorf("read config file %w", err)
	}

	if err := yaml.Unmarshal(bytes, &data); err != nil {
		return data, fmt.Errorf("%w %w", ErrYAMLMarshal, err)
	}
	return data, nil
}

// createFile creates any empty configfile. only call if config dir/file does not exist.
func createFile(xdg, prog, file string) error {
	if err := os.MkdirAll(filepath.Join(xdg, prog), 0o700); err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(xdg, prog, file))
	if err != nil {
		return err
	}

	f.Close()
	return nil
}
