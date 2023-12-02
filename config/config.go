package config

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DbPath        string `yaml:"db_path"`
	LogPath       string `yaml:"log_path"`
	LogLevel      string `yaml:"log_level"`
	SessionLenSec int64  `yaml:"session_len"`
	Port          string `yaml:"port"`
}

func ParseConfig(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return Config{}, err
	}
	var conf Config
	err = yaml.Unmarshal(data, &conf)
	if err != nil {
		return Config{}, err
	}
	return conf, nil
}
