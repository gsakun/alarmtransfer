package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

func LoadFile(filename string) (*Config, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	// If the entire config body is empty the UnmarshalYAML method is
	// never called. We thus have to set the DefaultConfig at the entry
	// point as well.
	err = yaml.UnmarshalStrict(content, &cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

type Config struct {
	DbConfig *DbConfig `yaml:"dbconfig"`
}

type DbConfig struct {
	Database string `yaml:"database"`
	Maxconn  int    `yaml:"maxconn"`
	Maxidle  int    `yaml:"maxidle"`
}
