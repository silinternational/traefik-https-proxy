package main

import (
	"os"
	"regexp"
	"testing"
)

func TestUpdateConfigContent(t *testing.T) {
	original := `
example TEST
another FIELD	
`
	expected := `
example val
another green	
`
	replacements := []Replacement{
		Replacement{
			Key:   "TEST",
			Value: "val",
		},
		Replacement{
			Key:   "FIELD",
			Value: "green",
		},
	}
	results := UpdateConfigContent([]byte(original), replacements)

	if string(results) != expected {
		t.Error("Results to not match expected. Results:", results)
		t.FailNow()
	}
}

func TestBuildReplacementsFromEnv(t *testing.T) {
	// Test failure for required env var
	_, err := BuildReplacementsFromEnv()
	if err == nil {
		t.Error("BuildReplacementsFromEnv should have failed because no env vars have been set")
		t.FailNow()
	}

	setRequiredEnvVars()

	replacements, err := BuildReplacementsFromEnv()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	replacementsCount := len(replacements)
	if replacementsCount != 6 {
		t.Error("Replacements did not have enough entries, only found", replacementsCount, "but expected 6")
		t.FailNow()
	}
}

func TestReadUpdateWrite(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Error("Unable to get current working directory for TestReadUpdateWrite")
		t.FailNow()
	}
	readFile := dir + "/traefik.toml"
	writeFile := dir + "/traefik_test.toml"

	// If writeFile already exists, delete it
	if _, err := os.Stat(writeFile); err == nil {
		os.Remove(writeFile)
	}

	configToml, err := ReadTraefikToml(readFile)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// Make sure placeholders for env vars exist
	envVars := GetEnvVarModels()
	for _, envvar := range envVars {
		search := regexp.MustCompile(envvar.Name)
		found := search.Find(configToml)
		if found == nil {
			t.Error("Did not find key in configToml template for env var", envvar.Name)
			t.FailNow()
		}
	}

	// Update config with required env var values
	setRequiredEnvVars()
	replacements, err := BuildReplacementsFromEnv()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	configToml = UpdateConfigContent(configToml, replacements)

	// Make sure placeholders for required env vars no longer exist in configToml
	for _, envvar := range envVars {
		if envvar.Required {
			search := regexp.MustCompile(envvar.Name)
			found := search.Find(configToml)
			if found != nil {
				t.Error("Uh oh, placeholder for required env var still present after update for env var:", envvar.Name)
				t.FailNow()
			}
		}
	}

	// Write out test file for manual reivew
	err = WriteTraefikToml(writeFile, configToml)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

}

func setRequiredEnvVars() {
	os.Setenv("LETS_ENCRYPT_EMAIL", "test@testing.com")
	os.Setenv("LETS_ENCRYPT_CA", "staging")
	os.Setenv("TLD", "testing.com")
	os.Setenv("SANS", "test.testing.com,another.testing.com")
	os.Setenv("BACKEND1_URL", "http://app:80")
	os.Setenv("FRONTEND1_DOMAIN", "test.testing.com")
}
