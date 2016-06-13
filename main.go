package main

import (
	"github.com/janekolszak/idp/core"
	"github.com/janekolszak/idp/helpers"
	"github.com/janekolszak/idp/providers"

	"flag"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

var (
	// Configuration file
	config *helpers.HydraConfig

	// Command line options
	// clientID     = flag.String("id", "dupa", "OAuth2 client ID of the IdP")
	// clientSecret = flag.String("secret", "asdf", "OAuth2 client secret")
	hydraURL     = flag.String("hydra", "https://hydra:4444", "Hydra's URL")
	configPath   = flag.String("conf", ".hydra.yml", "Path to Hydra's configuration")
	htpasswdPath = flag.String("htpasswd", "/etc/idp/htpasswd", "Path to credentials in htpasswd format")
)

func Handler(h http.HandlerFunc) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		h.ServeHTTP(w, r)
	}
}

func main() {
	fmt.Println("Identity Provider started")

	flag.Parse()
	// Read the configuration file
	config = helpers.NewHydraConfig(*configPath)

	// Setup the provider
	provider, err := providers.NewBasicAuth(*htpasswdPath, "localhost")
	if err != nil {
		panic(err)
	}

	idp := core.IdP{
		HydraAddress: *hydraURL,
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Port:         3000,
		Provider:     provider,
	}

	err = idp.Connect()
	if err != nil {
		panic(err)
	}

	router := httprouter.New()
	router.GET("/", Handler(idp.GetConsentGET()))
	router.POST("/", Handler(idp.GetConsentPOST()))
	http.ListenAndServe(":3000", router)

}
