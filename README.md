# Identity Provider (IdP) for Hydra [![Build Status](https://travis-ci.org/janekolszak/idp.svg?branch=master)](https://travis-ci.org/janekolszak/idp) [![Code Climate](https://codeclimate.com/github/janekolszak/idp/badges/gpa.svg)](https://codeclimate.com/github/janekolszak/idp) [![Gitter](https://img.shields.io/gitter/room/nwjs/nw.js.svg?maxAge=2592000)](https://gitter.im/janekolszak/idp)

This is a helper library for handling *challenge* requests from [Hydra](https://github.com/ory-am/hydra).
IdP handles:
- Storing challenge in a short lived cookie
- Passing user's consent to Hydra
- Retriving keys from Hydra and using them for JWT verification

## About

Let's say we have an Identity Provider with:
- */login* endpoint that accepts Hydra's challenges
- */consent* endpoint that handles getting consent from the user

This is how challenge request is hadled with the IdP library:

![Sequence Diagram](https://raw.githubusercontent.com/janekolszak/idp/master/doc/sequenceDiagram.png)

## Initialization

```go
import (
	"github.com/janekolszak/idp"
	"github.com/boj/rethinkstore"
	"time"
)

func main() {
	challengeCookieStore, err = rethinkstore.NewRethinkStore(/* RethinkDB address */,
	                                                         /* Database name */,
	                                                         "challengeCookies", 5, 5, []byte("something-very-secret"))
	// Return on error

	// How long do Challenge cookies live?
	challengeCookieStore.MaxAge(60 * 5) // 5 min

	// Create the IDP
	IDP = idp.NewIDP(&idp.IDPConfig{
		ClusterURL:            /* Hydra's address */,
		ClientID:              /* IDP's client ID */,
		ClientSecret:          /* IDP's client secret */,
		KeyCacheExpiration:    time.Duration(/* Key expiration time */) * time.Second,
		ClientCacheExpiration: time.Duration(/* Client info expiration */) * time.Second,
		CacheCleanupInterval:  time.Duration(/* Cache cleanup interval. Eg. 30 */) * time.Second,
		ChallengeStore:        challengeCookieStore,
	})

	// Connects with Hydra and fills caches
	err = IDP.Connect()
	// Return on error

}

```

## Usage

```go

func HandleChallengeGET(w http.ResponseWriter, r *http.Request) {
	// 0. Render HTML page with a login form
}

func HandleChallengePOST(w http.ResponseWriter, r *http.Request) {
	// 0. Parse and validate login data (username:password, login cookie etc)
	//    Return on error

	// 1. Verify user's credentials (eg. check username:password).
	//    Return on error
	//    Obtain userid

	// 2. Create a Challenge
	challenge, err := IDP.NewChallenge(r, userid)
	//    Return on error

	// 3. Save the Challenge to a cookie with a small TTL
	err = challenge.Save(w, r)
	//    Return on error

	// 4. Redirect to the consent endpoint
}

// Displays Consent screen. Here user agrees for listed scopes
func HandleConsentGET(w http.ResponseWriter, r *http.Request) {

	// 0. Get the Challenge from the cookie
	challenge, err := IDP.GetChallenge(r)
	//    Return on error

	// 1. Display consent screen
	//    Use challenge.User to get user's ID
	//    Use challenge.Scopes to display requested scopes

	// 2. If any error occured delete the Challenge cookie (optional)
	if err != nil {
		err = challenge.Delete(c.Writer, c.Request)
	}

	// 3. Render the HTML consent page
}

func HandleConsentPOST(w http.ResponseWriter, r *http.Request) {
	// 0. Get the Challenge from the cookie
	challenge, err := model.IDP.GetChallenge(c.Request)
	//    Return on error

    // 1. Parse and validate consent data (eg. form answer=y or list of scopes)
	//    Return on error

	// 2. If user refused access
	err = challenge.RefuseAccess(w, r)
	//    Return

	// 3. If userf agreed to grant access
	err = challenge.GrantAccessToAll(w, r)
	//    Return
}

```
