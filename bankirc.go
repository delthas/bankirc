package bankirc

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

type Config struct {
	ClientID     string    `yaml:"client_id"`
	ClientSecret string    `yaml:"client_secret"`
	Accounts     []Account `yaml:"accounts"`
	IRCServer    string    `yaml:"irc_server"`
	Channel      string    `yaml:"irc_channel"`
	Nick         string    `yaml:"irc_nick"`
}

type Account struct {
	Bank string `yaml:"bank"`
	Name string `yaml:"name"`
	ID   string `yaml:"id"`
}

func ReadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("opening config: %v", err)
	}
	defer f.Close()
	yd := yaml.NewDecoder(f)
	var cfg Config
	if err := yd.Decode(&cfg); err != nil {
		if err == io.EOF {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("decoding config: %v", err)
	}
	return &cfg, nil
}

func WriteConfig(path string, cfg *Config) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("opening config: %v", err)
	}
	defer f.Close()
	ye := yaml.NewEncoder(f)
	if err := ye.Encode(cfg); err != nil {
		return fmt.Errorf("encoding config: %v", err)
	}
	return nil
}
