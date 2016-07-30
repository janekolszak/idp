package main

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/boj/rethinkstore"
	"github.com/janekolszak/idp/core"
	"github.com/janekolszak/idp/helpers"
	"github.com/janekolszak/idp/providers/cookie"
	"github.com/janekolszak/idp/providers/form"
	"github.com/janekolszak/idp/userdb"
	"github.com/janekolszak/idp/userdb/rethinkdb/store"
	"github.com/janekolszak/idp/userdb/rethinkdb/verifier"
	"github.com/julienschmidt/httprouter"
	r "gopkg.in/dancannon/gorethink.v2"
)

const (
	consent = `<html>
<head></head>
<body>
<p>User:        {{.User}} </p>
<p>Client Name: {{.Client.Name}} </p>
<p>Scopes:      {{range .Scopes}} {{.}} {{end}} </p>
<p>Do you agree to grant access to those scopes? </p>
<p><form method="post">
	<input type="submit" name="answer" value="y">
	<input type="submit" name="answer" value="n">
</form></p>
</body></html>
`

	loginform = `
<html>
<head></head>
<body>
<form method="post">
	<p>Example App</p>
	<p>username <input type="text" name="username"></p>
	<p>password <input type="password" name="password" autocomplete="off"></p>
	<input type="submit">
	<a href="{{.RegisterURI}}">Register</a>
</form>
<hr>
{{.Msg}}
<body>
</html>
`
)

var (
	hydraURL     = flag.String("hydra", "https://hydra:4444", "Hydra's URL")
	configPath   = flag.String("conf", ".hydra.yml", "Path to Hydra's configuration")
	htpasswdPath = flag.String("htpasswd", "/etc/idp/htpasswd", "Path to credentials in htpasswd format")
	cookieDBPath = flag.String("cookie-db", "/etc/idp/remember.db3", "Path to a database with remember me cookies")
	staticFiles  = flag.String("static", "", "directory to serve as /static (for CSS/JS/images etc)")
)

func main() {
	fmt.Println("Identity Provider started!")

	flag.Parse()
	// Read the configuration file
	hydraConfig := helpers.NewHydraConfig(*configPath)

	session, err := r.Connect(r.ConnectOpts{
		Address:  os.Getenv("DATABASE_URL"),
		Database: os.Getenv("DATABASE_NAME"),
	})
	if err != nil {
		panic(err)
	}

	// Setup the providers
	db, err := store.NewStore(session)
	if err != nil {
		panic(err)
	}

	testUser := &userdb.User{
		FirstName: "Joe",
		LastName:  "Doe",
		Username:  "u",
		Email:     "joe@example.com",
	}

	db.Insert(testUser, "p")

	ver, err := verifier.NewVerifier(session)
	if err != nil {
		panic(err)
	}

	provider, err := form.NewFormAuth(form.Config{
		LoginForm:          loginform,
		LoginUsernameField: "username",
		LoginPasswordField: "password",

		TemplateDir: os.Getenv("TEMPLATE_DIR"),

		// Store for
		UserStore:    db,
		UserVerifier: ver,

		// Validation options:
		Username: form.Complexity{
			MinLength: 1,
			MaxLength: 100,
			Patterns:  []string{".*"},
		},
		Password: form.Complexity{
			MinLength: 1,
			MaxLength: 100,
			Patterns:  []string{".*"},
		},
	})
	if err != nil {
		panic(err)
	}

	cookieStore, err := cookie.NewRethinkDBStore(os.Getenv("DATABASE_URL"), os.Getenv("DATABASE_NAME"))
	if err != nil {
		panic(err)
	}
	defer cookieStore.Close()

	cookieProvider := &cookie.CookieAuth{
		Store:  cookieStore,
		MaxAge: time.Minute * 1,
	}

	challengeCookieStore, err := rethinkstore.NewRethinkStore(os.Getenv("DATABASE_URL"), os.Getenv("DATABASE_NAME"), "challenges", 5, 5, []byte("something-very-secret"))
	if err != nil {
		panic(err)
	}
	defer challengeCookieStore.Close()
	challengeCookieStore.MaxAge(60 * 5) // 5 min

	idp := core.NewIDP(&core.IDPConfig{
		ClusterURL:            *hydraURL,
		ClientID:              hydraConfig.ClientID,
		ClientSecret:          hydraConfig.ClientSecret,
		KeyCacheExpiration:    10 * time.Minute,
		ClientCacheExpiration: 10 * time.Minute,
		CacheCleanupInterval:  30 * time.Second,

		// TODO: [IMPORTANT] Don't use CookieStore here
		ChallengeStore: challengeCookieStore,
	})

	// Connect with Hydra
	err = idp.Connect()
	if err != nil {
		panic(err)
	}

	handler, err := CreateHandler(HandlerConfig{
		IDP:            idp,
		Provider:       provider,
		CookieProvider: cookieProvider,
		ConsentForm:    consent,
		StaticFiles:    *staticFiles,
	})

	emailPort, err := strconv.Atoi(os.Getenv("EMAIL_PORT"))
	if err != nil {
		panic(err)
	}

	testWorkerOpts := verifier.WorkerOpts{
		Session:         session,
		EndpointAddress: os.Getenv("CONSENT_URL") + "/verify",
		RequestMaxAge:   time.Minute * 1,
		CleanupInterval: time.Minute * 60,

		EmailerOpts: helpers.EmailerOpts{
			Host:         os.Getenv("EMAIL_HOST"),
			Port:         emailPort,
			User:         os.Getenv("EMAIL_USER"),
			Password:     os.Getenv("EMAIL_PASS"),
			From:         os.Getenv("EMAIL_FROM"),
			TextTemplate: template.Must(template.New("tmpl").Parse("Hi! {{.Username}}, visit {{.URL}} to verify!")),
			HtmlTemplate: template.Must(template.New("tmpl").Parse("Hi! {{.Username}}, click <a href={{.URL}}> here </a> to verify!")),
			Domain:       "localhost:3000",
		},
	}
	verifierWorker, err := verifier.NewWorker(testWorkerOpts)
	if err != nil {
		panic(err)
	}
	err = verifierWorker.Start()
	if err != nil {
		panic(err)
	}

	router := httprouter.New()
	handler.Attach(router)
	http.ListenAndServe(":3000", router)

	verifierWorker.Stop()

	idp.Close()
}
