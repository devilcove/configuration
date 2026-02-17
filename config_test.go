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
	cached = nil
}

func TestRead(t *testing.T) {
	setup(t)

	t.Run("fileMissing", func(t *testing.T) {
		config := &testConfig{}
		err := Get(config)
		var pathError *os.PathError
		should.BeErrorAs(t, err, &pathError)
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
		err := Save(config)
		should.BeNil(t, err)
		// from file
		reset(t)
		err = Get(&conf1)
		should.BeNil(t, err)
		should.BeEqual(t, conf1, *config)
		// from cache
		err = Get(&conf2)
		should.BeNil(t, err)
		should.BeEqual(t, conf1, conf2)
	})

	t.Run("invalidYaml", func(t *testing.T) {
		reset(t) // force read from file.
		err := os.WriteFile(filepath.Join(dir, progName, "config"), []byte("not: [valid"), permission)
		should.BeNil(t, err)
		config := testConfig{}
		should.BeErrorIs(t, Get(&config), ErrYAMLMarshal)
	})

	t.Run("filepermission", func(t *testing.T) {
		err := Save(config)
		should.BeNil(t, err, nil)
		err = os.Chmod(filepath.Join(dir, progName), 0o0400)
		should.BeNil(t, err)
		result := testConfig{}
		reset(t) // force to read from file.
		err = Get(&result)
		var pathError *os.PathError
		should.BeErrorAs(t, err, &pathError)
		err = os.Chmod(filepath.Join(dir, progName), 0o0755)
		should.BeNil(t, err)
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
		err := Save(&value)
		should.BeErrorIs(t, err, ErrYAMLMarshal)
		err = Get(&value)
		var pathError *os.PathError
		should.BeErrorAs(t, err, &pathError)
	})

	t.Run("success", func(t *testing.T) {
		var config1, config2 testConfig
		junk := otherConfig{}
		err := Save(config)
		should.BeNil(t, err)
		bytes, err := os.ReadFile(filepath.Join(os.Getenv("XDG_CONFIG_HOME"), progName, "config"))
		should.BeNil(t, err)
		should.ContainSubstring(t, string(bytes), "name: test")
		should.ContainSubstring(t, string(bytes), "count: 42")
		// get val from file
		err = Get(&config1)
		should.BeNil(t, err)
		should.BeEqual(t, config, &config1)
		// get val from cache
		err = Get(&config2)
		should.BeNil(t, err)
		should.BeEqual(t, config1, config2)
		// invalid type
		err = Get(&junk)
		should.BeErrorIs(t, err, ErrInterfaceConversion)
	})

	t.Run("permission", func(t *testing.T) {
		err := os.Chmod(filepath.Join(dir, progName), 0o0400)
		should.BeNil(t, err)
		err = Save(config)
		var pathError *os.PathError
		should.BeErrorAs(t, err, &pathError)
		err = os.Chmod(filepath.Join(dir, progName), 0o0755)
		should.BeNil(t, err)
	})
}
