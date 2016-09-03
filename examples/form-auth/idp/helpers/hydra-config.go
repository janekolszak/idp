package helpers

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

// Structure of the .hydra.yml file shared with Hydra via a volume
type HydraConfig struct {
	Port         int    `yaml:"port"`
	Issuer       string `yaml:"issuer"`
	ConsentURL   string `yaml:"consent_url"`
	ClusterURL   string `yaml:"cluster_url"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
}

// IdP has its credentials preconfigured by Hydra.
// This function parses the yaml file with that information
func NewHydraConfig(path string) *HydraConfig {
	config := new(HydraConfig)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		panic(err)
	}
	return config
}
