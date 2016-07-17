package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/janekolszak/idp/core"
	"github.com/janekolszak/idp/providers/cookie"
	"github.com/julienschmidt/httprouter"
)

type HandlerConfig struct {
	IDP            *core.IDP
	Provider       core.Provider // interface, not pointer
	CookieProvider *cookie.CookieAuth
	ConsentForm    string
	StaticFiles    string
}

type IdpHandler struct {
	HandlerConfig

	consentTemplate *template.Template
	router          *httprouter.Router
}

func CreateHandler(config HandlerConfig) (*IdpHandler, error) {
	h := IdpHandler{HandlerConfig: config}

	var err error
	h.consentTemplate, err = template.New("tmpl").Parse(h.ConsentForm)
	if err != nil {
		return nil, err
	}

	return &h, nil
}

func (h *IdpHandler) Attach(router *httprouter.Router) {
	router.GET("/", h.HandleChallenge())
	router.POST("/", h.HandleChallenge())
	router.GET("/consent", h.HandleConsentGET())
	router.POST("/consent", h.HandleConsentPOST())
	if h.StaticFiles != "" {
		router.ServeFiles("/static/*filepath", http.Dir(h.StaticFiles))
	}
}

func (h *IdpHandler) HandleChallenge() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		fmt.Println("-> HandleChallenge")
		defer fmt.Println("<- HandleChallenge")

		fmt.Println(sessions.GetRegistry(r))

		selector, user, err := h.CookieProvider.Check(r)
		if err == nil {
			fmt.Println("Authenticated with Cookie")
			err = h.CookieProvider.UpdateCookie(w, r, selector, user)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		} else {
			// Can't authenticate with "Remember Me" cookie,
			// so try with another provider:
			user, err = h.Provider.Check(r)
			if err != nil {
				// Authentication failed, or any other error
				fmt.Println(err.Error())

				// for "form" provider GET, this just displays the form
				h.Provider.WriteError(w, r, err)
				return
			}
			fmt.Println("Authenticated with Form Auth")

			// Save the RememberMe cookie
			err = h.CookieProvider.SetCookie(w, r, user)
			if err != nil {
				fmt.Println(err.Error())
			}
		}

		challenge, err := h.IDP.NewChallenge(r, user)
		if err != nil {
			fmt.Println(err.Error())
			h.Provider.WriteError(w, r, err)
			return
		}

		fmt.Println("Challenge: ", challenge)

		err = challenge.Save(w, r)
		if err != nil {
			fmt.Println(err.Error())
			h.Provider.WriteError(w, r, err)
			return
		}

		fmt.Println("Registry: ", sessions.GetRegistry(r))

		http.Redirect(w, r, "/consent", http.StatusFound)
	}
}

func (h *IdpHandler) HandleConsentGET() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		fmt.Println("-> HandleConsentGET")
		defer fmt.Println("<- HandleConsentGET")

		fmt.Println("Registry: ", sessions.GetRegistry(r))

		challenge, err := h.IDP.GetChallenge(r)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		h.consentTemplate.Execute(w, challenge)
	}
}

func (h *IdpHandler) HandleConsentPOST() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

		fmt.Println("-> HandleConsentPOST")
		defer fmt.Println("<- HandleConsentPOST")

		challenge, err := h.IDP.GetChallenge(r)
		if err != nil {
			fmt.Println(err.Error())
			h.Provider.WriteError(w, r, err)
			return
		}

		answer := r.FormValue("answer")
		fmt.Println("Answer: ", answer)
		if answer != "y" {
			err = challenge.RefuseAccess(w, r)
			if err != nil {
				fmt.Println(err.Error())
				h.Provider.WriteError(w, r, err)
			}
			return
		}

		err = challenge.GrantAccessToAll(w, r)
		if err != nil {
			// Server error
			fmt.Println(err.Error())
			h.Provider.WriteError(w, r, err)
			return
		}
	}
}
