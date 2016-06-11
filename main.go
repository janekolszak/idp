package main

import (
	"github.com/janekolszak/idp/core"
	"github.com/janekolszak/idp/providers"

	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
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

	// Command line options
	// clientID     = flag.String("id", "dupa", "OAuth2 client ID of the IdP")
	// clientSecret = flag.String("secret", "asdf", "OAuth2 client secret")
	hydraURL   = flag.String("hydra", "https://hydra:4444", "Hydra's URL")
	configPath = flag.String("conf", ".hydra.yml", "Path to Hydra's configuration")
)

// IdP has its credentials preconfigured by Hydra.
// This function parses the yaml file with that information
func readConfig(path string) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		panic(err)
	}
}

func main() {
	fmt.Println("Identity Provider started")

	flag.Parse()
	// fmt.Println("Client ID      ", *clientID)
	// fmt.Println("Client Secret: ", *clientSecret)
	// fmt.Println("Hydra:         ", *hydraURL)

	// Read the configuration file
	readConfig(*configPath)

	idp := core.IdP{
		HydraAddress: *hydraURL,
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Port:         3000,
		Provider:     providers.Const{Answer: true},
	}

	err := idp.Connect()
	if err != nil {
		panic(err)
	}

	idp.Run()
}
