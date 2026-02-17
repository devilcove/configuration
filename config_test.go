package configuration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Kairum-Labs/should"
)

// test struct.
type testConfig struct {
	Name  string `yaml:"name"`
	Count int    `yaml:"count"`
}

type otherConfig struct {
	Other string `yaml:"other"`
}

var (
	dir      string
	progName = "configurationTesting"
	config   *testConfig
)

func setup(t *testing.T) {
	t.Helper()
	dir = t.TempDir()
	err := os.MkdirAll(filepath.Join(dir, progName), 0o0750)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_CONFIG_HOME", dir)
	os.Args = []string{progName}
	config = &testConfig{Name: "test", Count: 42}
}

func reset(t *testing.T) {
	t.Helper()
	for k := range cached {
		delete(cached, k)
	}
}

func TestRead(t *testing.T) {
	setup(t)

	t.Run("fileMissing", func(t *testing.T) {
		config := &testConfig{}
		var pathError *os.PathError
		should.BeErrorAs(t, Get(config), &pathError)
	})

	t.Run("EnvVars", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "")
		t.Setenv("HOME", "")
		config := &testConfig{}
		err := Get(config)
		should.BeEqual(t, err.Error(), "configuration dir neither $XDG_CONFIG_HOME nor $HOME are defined")
	})

	t.Run("fromFile", func(t *testing.T) {
		var conf1, conf2 testConfig
		should.NotBeError(t, Save(config))
		should.NotBeError(t, Save(config, "testing"))
		// from file
		reset(t)
		should.NotBeError(t, Get(&conf1))
		should.BeEqual(t, conf1, *config)
		// from cache
		should.NotBeError(t, Get(&conf2, "testing"))
		should.BeEqual(t, conf1, conf2)
	})

	t.Run("invalidYaml", func(t *testing.T) {
		reset(t) // force read from file.
		should.NotBeError(t, os.WriteFile(filepath.Join(dir, progName, "config"), []byte("not: [valid"), permission))
		config := testConfig{}
		should.BeErrorIs(t, Get(&config), ErrYAMLMarshal)
	})

	t.Run("filepermission", func(t *testing.T) {
		should.NotBeError(t, Save(config))
		should.NotBeError(t, os.Chmod(filepath.Join(dir, progName), 0o0400))
		result := testConfig{}
		reset(t) // force to read from file.
		var pathError *os.PathError
		should.BeErrorAs(t, Get(&result), &pathError)
		should.NotBeError(t, os.Chmod(filepath.Join(dir, progName), 0o0755))
	})
}

func TestSave(t *testing.T) {
	setup(t)
	t.Run("EnvVars", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "")
		t.Setenv("HOME", "")
		err := Save(config)
		should.BeEqual(t, err.Error(), "configuration dir neither $XDG_CONFIG_HOME nor $HOME are defined")
	})

	t.Run("invalidYaml", func(t *testing.T) {
		type bad struct {
			A int
			B map[string]int `yaml:",inline"`
		}
		value := bad{A: 1, B: map[string]int{"a": 2}}
		should.BeErrorIs(t, Save(&value), ErrYAMLMarshal)
		var pathError *os.PathError
		should.BeErrorAs(t, Get(&value), &pathError)
	})

	t.Run("success", func(t *testing.T) {
		var config1, config2 testConfig
		junk := otherConfig{}
		should.NotBeError(t, Save(config))
		bytes, err := os.ReadFile(filepath.Join(os.Getenv("XDG_CONFIG_HOME"), progName, "config"))
		should.NotBeError(t, err)
		should.ContainSubstring(t, string(bytes), "name: test")
		should.ContainSubstring(t, string(bytes), "count: 42")
		// get val from file
		should.NotBeError(t, Get(&config1))
		should.BeEqual(t, config, &config1)
		// get val from cache
		should.NotBeError(t, Get(&config2))
		should.BeEqual(t, config1, config2)
		// invalid type
		should.BeErrorIs(t, Get(&junk), ErrInterfaceConversion)
	})

	t.Run("permission", func(t *testing.T) {
		should.NotBeError(t, os.Chmod(filepath.Join(dir, progName), 0o0400))
		var pathError *os.PathError
		should.BeErrorAs(t, Save(config), &pathError)
		should.NotBeError(t, os.Chmod(filepath.Join(dir, progName), 0o0755))
	})
}
