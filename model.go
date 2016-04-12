package main

import (
	log "github.com/Sirupsen/logrus"
	"mime/multipart"
	"net/http"
)

type IndexModel struct {
	ProjectVersions map[string][]string
	UploadErrors    map[string]string
	Project         string
	Version         string
	Content         multipart.File
	ContentHeader   *multipart.FileHeader
}

func NewIndexModel(app *App) (*IndexModel, error) {
	m := &IndexModel{
		ProjectVersions: map[string][]string{},
		UploadErrors:    map[string]string{},
	}
	v, err := app.Projects.GetProjectVersionMap()
	m.ProjectVersions = v
	return m, err
}

func (i *IndexModel) ParseForm(r *http.Request) bool {
	err := r.ParseMultipartForm(0)
	if err != nil {
		log.WithField("err", err).Error("error parsing multipart form")
		i.UploadErrors["General"] = "Failed to process your request. Try again or review the logs."
		return false
	}
	i.Project = r.Form.Get("project")
	if i.Project == "" {

		i.UploadErrors["Project"] = "You must provide project name."
	}
	i.Version = r.Form.Get("version")
	if i.Version == "" {
		i.UploadErrors["Version"] = "You must provide version."
	}
	file, fileHeader, err := r.FormFile("content")
	if err != nil {
		log.WithField("err", err).Error("upload file error")
		i.UploadErrors["Content"] = "Error uploading file. Try again or review the logs."
	} else {
		i.Content = file
		i.ContentHeader = fileHeader
	}
	return len(i.UploadErrors) == 0
}
