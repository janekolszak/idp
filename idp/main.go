package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

// Structure of the .hydra.yml file shared with Hydra via a volume
type Config struct {
	Port         int    `yaml:"port"`
	Issuer       string `yaml:"issuer"`
	ConsentURL   string `yaml:"consent_url"`
	ClusterURL   string `yaml:"cluster_url"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
}

var (
	// Configuration file
	config Config
)

// IdP has its credentials preconfigured by Hydra.
// This function parses the yaml file with that information
func readConfig() {
	data, err := ioutil.ReadFile("/root/.hydra.yml")
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		panic(err)
	}
}

func main() {
	// Read the configuration file
	readConfig()

	idp := IdP{
		HydraAddress: os.Getenv("HYDRA_URL"),
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Port:         3000,
	}

	err := idp.Connect()
	if err != nil {
		panic(err)
	}

	idp.Run()
}
