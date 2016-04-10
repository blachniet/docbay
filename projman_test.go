package main

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
)

func TestGetProjects(t *testing.T) {
	pm := ProjectManager{"_test_data/t0", "_test_data/tmp"}
	projs, err := pm.GetProjects()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	expected := map[int]string{
		0: "proj0",
		1: "proj1",
	}
	if len(projs) != len(expected) {
		t.Errorf("expected %v projects but there were %v", len(expected), len(projs))
	}
	for i, p := range projs {
		if p != expected[i] {
			t.Errorf("expected %v but was %v", expected[i], p)
		}
	}
}

func TestProjectDirDNE(t *testing.T) {
	pm := ProjectManager{"-test_data/dne", "_test_data/tmp"}
	_, err := pm.GetProjects()
	if err == nil {
		t.Errorf("expected err when asking for projects")
	}

	_, err = pm.GetVersions("proj0")
	if err == nil {
		t.Errorf("expected err when asking for versions")
	}

	_, err = pm.GetProjectVersionMap()
	if err == nil {
		t.Errorf("expected err when asking for map")
	}
}

func TestGetVersions(t *testing.T) {
	pm := ProjectManager{"_test_data/t0", "_test_data/tmp"}
	versions, err := pm.GetVersions("proj0")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	expected := map[int]string{
		0: "latest",
		1: "stable",
	}
	if len(versions) != len(expected) {
		t.Errorf("expected %v versions but there were %v", len(expected), len(versions))
	}
	for i, v := range versions {
		if v != expected[i] {
			t.Errorf("expected %v but was %v", expected[i], v)
		}
	}
}

func TestGetVersionsProjDNE(t *testing.T) {
	pm := ProjectManager{"_test_data/t0", "_test_data/tmp"}
	_, err := pm.GetVersions("projdne")
	if err == nil {
		t.Error("expected err")
	}
}

func TestGetProjectVersionMap(t *testing.T) {
	pm := ProjectManager{"_test_data/t0", "_test_data/tmp"}
	m, err := pm.GetProjectVersionMap()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	expected := map[string][]string{
		"proj0": []string{"latest", "stable"},
		"proj1": []string{"latest"},
	}
	if len(expected) != len(m) {
		t.Errorf("expected %v entries but there were %v", len(expected), len(m))
	}
	for k, v := range expected {
		for i, ver := range v {
			if ver != m[k][i] {
				t.Errorf("expected %v but was %v", ver, m[k][i])
			}
		}
	}
}

func TestGetProjectDir(t *testing.T) {
	pm := ProjectManager{"_test_data/t0", "_test_data/tmp"}
	dir := pm.GetProjectDir("foobar")
	expected := "_test_data/t0/foobar"
	if expected != dir {
		t.Errorf("expected %v but was %v", expected, dir)
	}
}

func TestGetVersionDir(t *testing.T) {
	pm := ProjectManager{"_test_data/t0", "_test_data/tmp"}
	dir := pm.GetVersionDir("foobar", "latest")
	expected := "_test_data/t0/foobar/latest"
	if expected != dir {
		t.Errorf("expected %v but was %v", expected, dir)
	}
}

func TestAddProject(t *testing.T) {
	pm := ProjectManager{"_test_data/t0", "_test_data/tmp"}

	// Ensure the version dir gets deleted
	defer os.RemoveAll(pm.GetVersionDir("proj0", "v1"))

	// Make sure the temp dir exists
	err := os.MkdirAll(pm.TempDir, 0755)
	if err != nil {
		t.Fatal("unexpected err creating temp dir:", err)
	}

	// Open the input zip
	input, err := os.Open("_test_data/v1.zip")
	if err != nil {
		t.Fatal("unexpected err opening file:", err)
	}
	defer input.Close()

	// Add it
	err = pm.AddProject("proj0", "v1", input)
	if err != nil {
		t.Fatal("unexpected error adding project: ", err)
	}

	// Ensure it's there
	found := false
	versions, err := pm.GetVersions("proj0")
	if err != nil {
		t.Fatal("unexpected err getting versions")
	}
	for _, ver := range versions {
		if ver == "v1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("did not find project version")
	}

	entries, err := ioutil.ReadDir(pm.GetVersionDir("proj0", "v1"))
	if err != nil {
		t.Fatal("unexpected error checking version entries:", err)
	}
	expected := []string{
		".DS_Store", "foobar.png", "index.html",
	}
	if len(entries) != len(expected) {
		t.Fatalf("expected %v entries but there were %v", len(expected), len(entries))
	}
	for i, e := range entries {
		if e.Name() != expected[i] {
			t.Errorf("expected %v but was %v", expected[i], e.Name())
		}
	}

	contents, err := ioutil.ReadFile(path.Join(pm.GetVersionDir("proj0", "v1"), "index.html"))
	if err != nil {
		t.Fatal("err reading file contents:", err)
	}
	expectedContents := "proj0-v1"
	if strings.TrimSpace(string(contents)) != expectedContents {
		t.Fatalf("expect '%v' in contents but was '%v'", expectedContents, strings.TrimSpace(string(contents)))
	}
}
