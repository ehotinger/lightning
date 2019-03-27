package config

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

// Config stores configuration details.
type Config struct {
	AzureAccountName string `yaml:"accountName"`
	AzureAccountKey  string `yaml:"accountKey"`
	CachePath        string `yaml:"cachePath"`
	ContainerName    string `yaml:"containerName"`
}

// NewConfig creates a new Config object.
func NewConfig(
	accountName string,
	accountKey string,
	containerName string,
	cachePath string) *Config {
	return &Config{
		AzureAccountName: accountName,
		AzureAccountKey:  accountKey,
		ContainerName:    containerName,
		CachePath:        cachePath,
	}
}

// NewConfigFromBytes unmarshals a Config from the specified bytes.
func NewConfigFromBytes(data []byte) (*Config, error) {
	c := &Config{}
	err := yaml.Unmarshal(data, c)
	return c, err
}

// NewConfigFromFile creates a new Config from a file.
func NewConfigFromFile(file string) (*Config, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return NewConfigFromBytes(data)
}
