package main

import (
	"flag"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/gorilla/sessions"
	"github.com/janekolszak/idp/examples/basic-auth/idp/helpers"
	"github.com/janekolszak/idp/examples/basic-auth/idp/providers/basic"
	"github.com/janekolszak/idp/examples/basic-auth/idp/providers/cookie"
	"github.com/julienschmidt/httprouter"

	_ "github.com/mattn/go-sqlite3"
)

const (
	consent = `<html><head></head><body>
	<p>User:        {{.User}} </p>
	<p>Client Name: {{.Client.Name}} </p>
	<p>Scopes:      {{range .Scopes}} {{.}} {{end}} </p>
	<p>Do you agree to grant access to those scopes? </p>
	<p>
		<form method="post">
			<input type="submit" name="answer" value="y">
			<input type="submit" name="answer" value="n">
		</form>
	</p>
 	</body></html>
	`
)

var (
	idp            *idp.IDP
	provider       idp.Provider
	cookieProvider *cookie.CookieAuth

	// Command line options
	// clientID     = flag.String("id", "someid", "OAuth2 client ID of the IdP")
	// clientSecret = flag.String("secret", "somesecret", "OAuth2 client secret")
	hydraURL     = flag.String("hydra", "https://hydra:4444", "Hydra's URL")
	configPath   = flag.String("conf", ".hydra.yml", "Path to Hydra's configuration")
	htpasswdPath = flag.String("htpasswd", "/etc/idp/htpasswd", "Path to credentials in htpasswd format")
	cookieDBPath = flag.String("cookie-db", "/etc/idp/remember.db3", "Path to a database with remember me cookies")
)

func HandleChallengeGET() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		fmt.Println("Challenge!")

		selector, user, err := cookieProvider.Check(r)
		if err == nil {
			fmt.Println("Authenticated with Cookie")
			err = cookieProvider.UpdateCookie(w, r, selector, user)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		} else {
			// Can't authenticate with "Remember Me" cookie,
			// so try with another provider:
			user, err = provider.Check(r)
			if err != nil {
				// Authentication failed, or any other error
				fmt.Println(err.Error())
				provider.WriteError(w, r, err)
				return
			}
			fmt.Println("Authenticated with Basic Auth")

			// Save the RememberMe cookie
			err = cookieProvider.SetCookie(w, r, user)
			if err != nil {
				fmt.Println(err.Error())
			}
		}

		challenge, err := idp.NewChallenge(r, user)
		if err != nil {
			fmt.Println(err.Error())
			provider.WriteError(w, r, err)
			return
		}

		err = challenge.Save(w, r)
		if err != nil {
			fmt.Println(err.Error())
			provider.WriteError(w, r, err)
			return
		}

		http.Redirect(w, r, "/consent", http.StatusFound)
	}
}

func HandleConsentGET() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		challenge, err := idp.GetChallenge(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Println("Data ", challenge.User)

		t := template.Must(template.New("tmpl").Parse(consent))

		t.Execute(w, challenge)
	}
}

func HandleConsentPOST() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

		fmt.Println("Consent POST!")
		challenge, err := idp.GetChallenge(r)
		if err != nil {
			fmt.Println(err.Error())
			provider.WriteError(w, r, err)
			return
		}

		answer := r.FormValue("answer")
		fmt.Println("Answer: ", answer)

		if answer != "y" {
			// No challenge token
			// TODO: Handle negative answer
			challenge.RefuseAccess(w, r)
			return
		}

		err = challenge.GrantAccessToAll(w, r)
		if err != nil {
			// Server error
			fmt.Println(err.Error())
			provider.WriteError(w, r, err)
			return
		}
	}
}

func main() {
	fmt.Println("Identity Provider started!")

	flag.Parse()
	// Read the configuration file
	hydraConfig := helpers.NewHydraConfig(*configPath)

	// Setup the providers
	var err error
	provider, err = basic.NewBasicAuth(*htpasswdPath, "localhost")
	if err != nil {
		panic(err)
	}

	dbCookieStore, err := cookie.NewDBStore("sqlite3", *cookieDBPath)
	if err != nil {
		panic(err)
	}

	cookieProvider = &cookie.CookieAuth{
		Store:  dbCookieStore,
		MaxAge: time.Second * 30,
	}

	config := idp.IDPConfig{
		ClusterURL:            *hydraURL,
		ClientID:              hydraConfig.ClientID,
		ClientSecret:          hydraConfig.ClientSecret,
		KeyCacheExpiration:    10 * time.Minute,
		ClientCacheExpiration: 10 * time.Minute,
		CacheCleanupInterval:  30 * time.Second,

		// TODO: [IMPORTANT] Don't use CookieStore here
		ChallengeStore: sessions.NewCookieStore([]byte("something-very-secret")),
	}

	idp = idp.NewIDP(&config)

	// Connect with Hydra
	err = idp.Connect()
	if err != nil {
		panic(err)
	}

	router := httprouter.New()
	router.GET("/", HandleChallengeGET())
	router.POST("/", HandleChallengeGET())
	router.GET("/consent", HandleConsentGET())
	router.POST("/consent", HandleConsentPOST())
	http.ListenAndServe(":3000", router)

	idp.Close()
}
