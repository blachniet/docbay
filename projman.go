package main

import (
	"archive/zip"
	log "github.com/Sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"path"
)

type ProjectManager struct {
	ProjectDir string
	TempDir    string
}

func (p ProjectManager) GetProjects() ([]string, error) {
	entries, err := ioutil.ReadDir(p.ProjectDir)
	if err != nil {
		return nil, err
	}
	result := []string{}
	for _, e := range entries {
		if e.IsDir() {
			result = append(result, e.Name())
		}
	}
	return result, nil
}

func (p ProjectManager) GetVersions(project string) ([]string, error) {
	entries, err := ioutil.ReadDir(p.GetProjectDir(project))
	if err != nil {
		return nil, err
	}
	result := []string{}
	for _, e := range entries {
		if e.IsDir() {
			result = append(result, e.Name())
		}
	}
	return result, nil
}

func (p ProjectManager) GetProjectVersionMap() (map[string][]string, error) {
	result := map[string][]string{}
	projects, err := p.GetProjects()
	if err != nil {
		return nil, err
	}
	for _, proj := range projects {
		versions, err := p.GetVersions(proj)
		if err != nil {
			return nil, err
		}
		projVers := []string{}
		for _, ver := range versions {
			projVers = append(projVers, ver)
		}
		result[proj] = projVers
	}

	return result, nil
}

func (p ProjectManager) AddProject(project, version string, input io.Reader) error {
	// Copy the input to a temporary location
	tmpFile, err := ioutil.TempFile(p.TempDir, "upfile_")
	if err != nil {
		log.WithField("err", err).Error("failed to create temp file")
		return err
	}
	_, err = io.Copy(tmpFile, input)
	if err != nil {
		log.WithField("err", err).Error("failed to copy to temp file")
		return err
	}
	tmpFile.Close()
	defer func() {
		err := os.Remove(tmpFile.Name())
		if err != nil {
			log.WithField("err", err).Error("failed to delete temp file")
		}
	}()

	// Remove the old version dir
	dstDir := p.GetVersionDir(project, version)
	err = os.RemoveAll(dstDir)
	if err != nil {
		log.WithField("err", err).Error("failed to remove old version dir")
		return err
	}

	// Unzip to the version dir
	zr, err := zip.OpenReader(tmpFile.Name())
	if err != nil {
		log.WithField("err", err).Error("failed to open zip reader")
		return err
	}
	defer zr.Close()

	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			err = os.MkdirAll(path.Join(dstDir, f.Name), 0755)
			if err != nil {
				return err
			}
		} else {
			dstPath := path.Join(dstDir, f.Name)
			err = os.MkdirAll(path.Dir(dstPath), 0755)
			if err != nil {
				return err
			}
			dstFile, err := os.Create(dstPath)
			if err != nil {
				return err
			}
			rc, err := f.Open()
			if err != nil {
				return err
			}
			_, err = io.Copy(dstFile, rc)
			if err != nil {
				return err
			}
			rc.Close()
			dstFile.Close()
		}
	}
	return nil
}

func (p ProjectManager) GetProjectDir(project string) string {
	return path.Join(p.ProjectDir, project)
}

func (p ProjectManager) GetVersionDir(project, version string) string {
	return path.Join(p.ProjectDir, project, version)
}
