package trunks

import (
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// ParseConf read the yaml file and populate the Config instancce
func ParseConf(file string) (*TrunksConfig, error) {
	var trunksConfig *TrunksConfig
	path, err := filepath.Abs(file)
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(yamlFile, &trunksConfig); err != nil {
		return nil, err
	}
	return trunksConfig, nil
}

// Interfaces struct
type NIC struct {
	ST string `yaml:"st"`
	GW string `yaml:"gw"`
}

// Bandwidth struct
type Bandwidth struct {
	Forward float64 `yaml:"forward"`
	Return  float64 `yaml:"return"`
}

// Delay struct
type Delay struct {
	Value  float64 `yaml:"value"`
	Offset float64 `yaml:"offset"`
}

// ACM struct
type ACM struct {
	Weight   float64 `yaml:"weight"`
	Duration int     `yaml:"duration"`
}

// TrunksConfig struct
type TrunksConfig struct {
	NIC                NIC       `yaml:"nic"`
	Bandwidth          Bandwidth `yaml:"bandwidth"`
	Delay              Delay     `yaml:"delay"`
	ACMList            []*ACM    `yaml:"acm"`
	QoS                bool
	Logs               string
	KernelVersionCheck bool
	ACMEnabled         bool
	ACMCounter         int
	ACMIndex           int
	CurrentACM         *ACM
}
