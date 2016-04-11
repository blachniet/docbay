package main

import (
	"flag"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
	"os"
	"path"
	"strconv"
)

func main() {
	var debug bool
	var docbayDir string
	var port int
	flag.BoolVar(&debug, "debug", false, "Enable debug logging")
	flag.StringVar(&docbayDir, "dir", "", "Directory used by docbay to store sites and settings")
	flag.IntVar(&port, "port", -1, "Port to server on (overrides env var PORT)")
	flag.Parse()

	if debug {
		log.SetLevel(log.DebugLevel)
	}

	if docbayDir == "" {
		docbayDir = os.Getenv("DOCBAY_DIR")
	}
	if docbayDir == "" {
		docbayDir = "_docbayDir"
	}
	projDir := path.Join(docbayDir, "proj")
	if err := os.MkdirAll(projDir, 0755); err != nil {
		log.WithField("err", err).Fatal("Error creating docbay project dir")
	}
	tempDir := path.Join(docbayDir, "tmp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		log.WithField("err", err).Fatal("Error creating docbay temp dir")
	}

	if port == -1 {
		portStr := os.Getenv("PORT")
		if portStr != "" {
			var err error
			if port, err = strconv.Atoi(portStr); err != nil {
				log.WithField("err", err).Fatal("Error parsing port from environment variable")
			}
		}
	}
	if port == -1 {
		port = 3000
	}

	// parse templates
	tmpl, err := template.ParseGlob("templates/*.tmpl")
	if err != nil {
		log.WithField("err", err).Fatal("Failed to parse templates")
	}

	app := &App{
		Template: tmpl,
		RootDir:  docbayDir,
		Projects: &ProjectManager{projDir, tempDir},
	}
	r := mux.NewRouter()
	r.Handle("/", AppHandler{app, GetIndex})
	r.Handle("/_/upload", AppHandler{app, PostProjectDocs})
	r.Handle("/_/delete", AppHandler{app, DeleteDocs})
	r.Handle("/{project}", AppHandler{app, GetDefaultVersion})
	r.Handle("/{project}/", AppHandler{app, GetDefaultVersion})
	r.PathPrefix("/{project}/{version}/").Handler(AppHandler{app, GetProjectDocs})

	log.WithField("port", port).Infoln("Serving on port", port)
	http.ListenAndServe(fmt.Sprintf(":%v", port), r)
}

func GetIndex(app *App, w http.ResponseWriter, r *http.Request) *AppError {
	data, err := app.Projects.GetProjectVersionMap()
	if err != nil {
		return &AppError{err, "failed to read project tree", http.StatusInternalServerError}
	}

	err = app.Template.ExecuteTemplate(w, "index.tmpl", map[string]interface{}{
		"ProjectVersions": data,
	})
	if err != nil {
		return &AppError{err, "render err", http.StatusInternalServerError}
	}
	return nil
}

func GetDefaultVersion(app *App, w http.ResponseWriter, r *http.Request) *AppError {
	vars := mux.Vars(r)
	projectID := vars["project"]
	http.Redirect(w, r, fmt.Sprintf("/%v/latest/", projectID), http.StatusFound)
	return nil
}

func GetProjectDocs(app *App, w http.ResponseWriter, r *http.Request) *AppError {
	vars := mux.Vars(r)
	project := vars["project"]
	version := vars["version"]
	vDir := app.Projects.GetVersionDir(project, version)

	// TODO: Err if not directory or not exists
	prefix := fmt.Sprintf("/%v/%v/", project, version)
	fs := http.FileServer(http.Dir(vDir))
	http.StripPrefix(prefix, fs).ServeHTTP(w, r)
	return nil
}

func PostProjectDocs(app *App, w http.ResponseWriter, r *http.Request) *AppError {
	err := r.ParseMultipartForm(0)
	if err != nil {
		return &AppError{err, "Could not understand request", http.StatusBadRequest}
	}

	project := r.Form.Get("project")
	version := r.Form.Get("version")
	if project == "" || version == "" {
		return &AppError{err, "Could not understand request", http.StatusBadRequest}
	}

	formFile, _, err := r.FormFile("content")
	if err != nil {
		return &AppError{err, "Could not open file", http.StatusInternalServerError}
	}
	defer formFile.Close()

	err = app.Projects.SetVersionDocs(project, version, formFile)
	if err != nil {
		return &AppError{err, "failed to add version", http.StatusInternalServerError}
	}

	http.Redirect(w, r, "/", http.StatusFound)
	return nil
}

func DeleteDocs(app *App, w http.ResponseWriter, r *http.Request) *AppError {
	project := r.URL.Query().Get("project")
	version := r.URL.Query().Get("version")
	if project == "" || version == "" {
		return &AppError{nil, "Could not understand request", http.StatusBadRequest}
	}

	err := app.Projects.DeleteVersionDocs(project, version)
	if err != nil {
		return &AppError{err, "Error deleting docs", http.StatusInternalServerError}
	}

	http.Redirect(w, r, "/", http.StatusFound)
	return nil
}
