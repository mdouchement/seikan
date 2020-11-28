package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type (
	// A Connection handles connection details.
	Connection struct {
		Address string `yaml:"address"`
		Secret  string `yaml:"secret"`
		Public  string `yaml:"public"`
	}

	// A Log handles logging details.
	Log struct {
		Level          string `yaml:"level"`
		ForceColor     bool   `yaml:"force_color"`
		ForceFormating bool   `yaml:"force_formating"`
	}

	// A Outbound handles tunneling details.
	Outbound struct {
		Identifier  string `yaml:"identifier"`
		Source      string `yaml:"source"`
		Destination string `yaml:"destination"`
	}
)

// A Server holds server's configuration fields.
type Server struct {
	Connection `yaml:",inline"`
	Clients    map[string]string `yaml:"clients"`
	Log        Log               `yaml:"log"`
	AllowList  []string          `yaml:"allow_list"`
	Outbounds  []Outbound        `yaml:"outbounds"`
}

// A Client holds client's configuration fields.
type Client struct {
	Identifier string     `yaml:"identifier"`
	Server     Connection `yaml:"server"`
	Secret     string     `yaml:"secret"`
	Public     string     `yaml:"public"`
	Log        Log        `yaml:"log"`
	Inbound    bool       `yaml:"inbound"`
	AllowList  []string   `yaml:"allow_list"`
	Outbounds  []Outbound `yaml:"outbounds"`
}

// Load loads a configuration file.
func Load(filename string, cfg interface{}) error {
	payload, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(payload, cfg)
}
