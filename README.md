# configuration
[![Go Reference](https://pkg.go.dev/badge/github.com/devilcove/configuration?status.svg)](https://pkg.go.dev/github.com/devilcove/configuration?tab=doc)  
A small Go package for loading and saving application configuration as YAML using the XDG Base Directory specification.

Configuration is stored per executable under:

$XDG_CONFIG_HOME/<executable-name>/config


The package provides transparent in-memory caching, and enforces strict type consistency across reads.

## Features

* YAML-based configuration files using [go.yaml.in/yaml/v4](https://pkg.go.dev/go.yaml.in/yaml/v4) lib
* XDG-compliant configuration directory resolution
* Generic API (Get[T], Save[T])
* Automatic in-memory caching
* Strict type safety across calls
* Secure file permissions (0600)

## Installation  
```
go get github.com/devilcove/configuration
```

## Usage  
Define a Configuration Struct  
```
type Config struct {
    Host string `yaml:"host"`
    Port int    `yaml:"port"`
}
```
## Load Configuration
```
var cfg Config
if err := configuration.Get(&cfg); err != nil {
    log.Fatal(err)
}
```

* On first call, the configuration is read from disk and cached.
* Subsequent calls return the cached value.
* The supplied type must match the type used on the first call.

## Save Configuration
```
cfg := Config{
    Host: "localhost",
    Port: 8080,
}

if err := configuration.Save(&cfg); err != nil {
    log.Fatal(err)
}
```

* The configuration is serialized to YAML.
* The file is written with permissions 0600.
* The saved configuration replaces the in-memory cache.

## Configuration File Location

The configuration file is always stored at:
```
$XDG_CONFIG_HOME/<executable-name>/config
```


Where:
* `XDG_CONFIG_HOME` is resolved using os.UserConfigDir()
* `<executable-name>` is derived from os.Args[0]

### Examples:

| OS	| Path Example|
| --    | ------------|
|Linux	|~/.config/myapp/config|
|macOS	|~/Library/Application Support/myapp/config|
|Windows	|%AppData%\myapp\config|

## Caching Behavior

* Configuration is loaded once per process.
* All subsequent Get calls return the cached value.
* Mixing different configuration types in the same process is not allowed.

Example of invalid usage:
```
var a ConfigA
var b ConfigB

configuration.Get(&a)
configuration.Get(&b) // returns ErrInterfaceConversion
```

## Error Handling
The package exposes the following sentinel errors:  
| Name | Error |
| ---- | ----- |
| ErrInterfaceConverion | the cached type does not match |
| ErrYAMLMarshal | unable to marshal data |

Errors are returned when:
* No configuration directory can be resolved
* The config file cannot be read or written
* The YAML cannot be unmarshaled into the target type

## Licence

MIT Licence. See [LICENCE](LICENCE) for details.