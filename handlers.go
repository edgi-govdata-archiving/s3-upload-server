package main

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/julienschmidt/httprouter"
)

// templates is a collection of views for rendering with the renderTemplate function
// see homeHandler for an example
var templates = template.Must(template.ParseFiles("views/index.html", "views/expired.html", "views/accessDenied.html", "views/notFound.html"))

// CORSHandler is an empty 200 response for OPTIONS requests that responds with
// headers set in addCorsHeaders
func CORSHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	addCorsHeaders(w, r)
	w.WriteHeader(http.StatusOK)
}

// HomeHandler renders the home page
func HomeHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	renderTemplate(w, "index.html")
}

// renderTemplate renders a template with the values of cfg.TemplateData
func renderTemplate(w http.ResponseWriter, tmpl string) {
	err := templates.ExecuteTemplate(w, tmpl, cfg.TemplateData)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// middleware handles request logging, expiry & authentication if set
func middleware(handler httprouter.Handle) httprouter.Handle {
	// return auth middleware if configuration settings are present
	if cfg.HttpAuthUsername != "" && cfg.HttpAuthPassword != "" {
		return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
			// poor man's logging:
			fmt.Println(r.Method, r.URL.Path, time.Now())

			user, pass, ok := r.BasicAuth()
			fmt.Println(user, pass, ok)
			if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(cfg.HttpAuthUsername)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(cfg.HttpAuthPassword)) != 1 {
				w.Header().Set("WWW-Authenticate", `Basic realm="Please enter your username and password for this site"`)
				w.WriteHeader(http.StatusUnauthorized)
				renderTemplate(w, "accessDenied.html")
				return
			}

			if cfg.Deadline != nil {
				if time.Now().After(*cfg.Deadline) {
					w.WriteHeader(http.StatusForbidden)
					renderTemplate(w, "expired.html")
					return
				}
			}

			addCorsHeaders(w, r)
			handler(w, r, p)
		}
	}

	// no-auth middware func
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// poor man's logging:
		fmt.Println(r.Method, r.URL.Path, time.Now())

		if cfg.Deadline != nil {
			if time.Now().After(*cfg.Deadline) {
				w.WriteHeader(http.StatusForbidden)
				renderTemplate(w, "expired.html")
				return
			}
		}

		addCorsHeaders(w, r)

		handler(w, r, p)
	}
}

func addCorsHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	for _, o := range cfg.AllowedOrigins {
		if origin == o {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			return
		}
	}
}
