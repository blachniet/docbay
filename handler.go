package main

import (
	log "github.com/Sirupsen/logrus"
	"html/template"
	"net/http"
)

// AppError represents an error that occurs in a handler.
type AppError struct {
	Error   error
	Message string
	Code    int
}

// App contains objects and services commonly required by handlers.
type App struct {
	Template *template.Template
	RootDir  string
	Projects *ProjectManager
}

// AppHandler is a specialized HTTP handler that returns an error and accepts
// and app environment
type AppHandler struct {
	*App
	H func(app *App, w http.ResponseWriter, r *http.Request) *AppError
}

func (h AppHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.H(h.App, w, r)
	if err != nil {
		log.WithFields(log.Fields{
			"err":         err.Error,
			"message":     err.Message,
			"http.status": err.Code,
			"http.method": r.Method,
			"http.path":   r.URL.Path,
		}).Error("http.err")

		// TODO: Customize this
		http.Error(w, err.Message, err.Code)
	}
}
