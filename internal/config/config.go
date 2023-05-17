package config

import (
	"io/ioutil"
	"regexp"

	"github.com/pkg/errors"
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

	// An Allow is a list of allowed endpoints wit options.
	Allow struct {
		Type               string           `yaml:"type"`
		Endpoint           string           `yaml:"endpoint"`
		IgnoreErrors       []string         `yaml:"ignore_errors"`
		IgnoreErrorsRegexp []*regexp.Regexp `yaml:"-"`
	}

	AllowWrapper struct {
		Allow
	}
)

// A Server holds server's configuration fields.
type Server struct {
	Connection `yaml:",inline"`
	Clients    map[string]string `yaml:"clients"`
	Log        Log               `yaml:"log"`
	AllowList  []AllowWrapper    `yaml:"allow_list"`
	Outbounds  []Outbound        `yaml:"outbounds"`
}

// A Client holds client's configuration fields.
type Client struct {
	Identifier string         `yaml:"identifier"`
	Server     Connection     `yaml:"server"`
	Secret     string         `yaml:"secret"`
	Public     string         `yaml:"public"`
	Log        Log            `yaml:"log"`
	Inbound    bool           `yaml:"inbound"`
	AllowList  []AllowWrapper `yaml:"allow_list"`
	Outbounds  []Outbound     `yaml:"outbounds"`
}

// Load loads a configuration file.
func Load(filename string, cfg interface{}) error {
	payload, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(payload, cfg)
}

func (a *AllowWrapper) UnmarshalYAML(value *yaml.Node) error {
	if value.Tag == "!!str" {
		return value.Decode(&a.Endpoint)
	}

	if err := value.Decode(&a.Allow); err != nil {
		return err
	}

	if a.Type != "" && a.Type != "cidr" {
		return errors.New("type must be empty or a cidr")
	}

	if a.Type == "cidr" && len(a.IgnoreErrors) > 0 {
		return errors.New("ignore_errors cannot be be used for a cidr allow_list")
	}

	for _, expr := range a.IgnoreErrors {
		re, err := regexp.Compile(expr)
		if err != nil {
			return errors.Wrapf(err, "`%s`", expr)
		}

		a.IgnoreErrorsRegexp = append(a.IgnoreErrorsRegexp, re)
	}

	return nil
}
