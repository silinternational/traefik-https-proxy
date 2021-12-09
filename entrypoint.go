package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// Replacement represents a key to find and value to replace it with
type Replacement struct {
	Key   string
	Value string
}

// EnvVar represents expected environment variables, whether they are required, and a description for error reporting
type EnvVar struct {
	Name     string
	Required bool
	Desc     string
	Default  string
}

func main() {

	var configFile string
	flag.StringVar(&configFile, "c", "/etc/traefik/traefik.toml", "Traefik config file to use, default: /etc/traefik/traefik.toml")
	flag.Parse()

	if _, err := os.Stat(configFile); err != nil {
		fmt.Println("Config file", configFile, "not found")
		os.Exit(1)
	}

	if len(os.Args) <= 1 {
		fmt.Println("You must provide a command to run after entrypoint process completes. You probably want: /traefik")
	}

	replacements, err := BuildReplacementsFromEnv()
	handleError(err)

	configToml, err := ReadTraefikToml(configFile)
	handleError(err)

	configToml = UpdateConfigContent(configToml, replacements)

	err = WriteTraefikToml(configFile, configToml)
	handleError(err)

	runCmd()
	os.Exit(0)
}

// Run CMD specified in Dockerfile or runtime and send output to stdout
func runCmd() {
	executable := os.Args[1]
	args := os.Args[2:]
	cmd := exec.Command(executable, args...)
	cmdStdout, err := cmd.StdoutPipe()
	handleError(err)

	scanner := bufio.NewScanner(cmdStdout)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()

	err = cmd.Start()
	handleError(err)

	err = cmd.Wait()
	handleError(err)
}

func handleError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// ReadTraefikToml reads the Traefik config file from filesystem and returns as byte array
func ReadTraefikToml(filename string) ([]byte, error) {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return []byte{}, fmt.Errorf("Unable to read config file at %s", filename)
	}

	return file, nil
}

// WriteTraefikToml writes updated Traefix config to filesystem
func WriteTraefikToml(filename string, contents []byte) error {
	err := ioutil.WriteFile(filename, contents, 0644)
	if err != nil {
		return err
	}

	return nil
}

// UpdateConfigContent replaces placeholders with values from environment variables
func UpdateConfigContent(config []byte, replacements []Replacement) []byte {
	for _, rep := range replacements {
		regex := regexp.MustCompile(rep.Key)
		config = regex.ReplaceAll(config, []byte(rep.Value))
	}

	return config
}

// BuildReplacementsFromEnv Build []Replacement from env vars
func BuildReplacementsFromEnv() ([]Replacement, error) {

	var configReplacements []Replacement

	envVars := GetEnvVarModels()
	for _, envvar := range envVars {
		value := os.Getenv(envvar.Name)
		if value == "" && envvar.Required {
			return configReplacements, fmt.Errorf("Missing required env var: %s. Description: %s", envvar.Name, envvar.Desc)
		}

		if value != "" {
			if envvar.Name == "LETS_ENCRYPT_CA" {
				if value == "staging" {
					value = "https://acme-staging.api.letsencrypt.org/directory"
				} else if value == "production" {
					value = "https://acme-v01.api.letsencrypt.org/directory"
				}
			} else if envvar.Name == "SANS" {
				sans := strings.Split(value, ",")
				value = ""
				for _, san := range sans {
					value += "\"" + san + "\", "
				}
				value = strings.TrimRight(value, ", ")
			}
			configReplacements = append(configReplacements, Replacement{
				Key:   envvar.Name,
				Value: value,
			})
		} else if value == "" && !envvar.Required {
			value = envvar.Default
		}
	}

	return configReplacements, nil
}

// GetEnvVarModels returns an array of EnvVar objects
func GetEnvVarModels() []EnvVar {
	envVars := []EnvVar{
		{
			Name:     "LETS_ENCRYPT_EMAIL",
			Required: true,
			Desc:     "An email address is required for LETS_ENCRYPT_EMAIL",
			Default:  "",
		},
		{
			Name:     "LETS_ENCRYPT_CA",
			Required: true,
			Desc:     "Which CA to use, either staging or production. Default: staging",
			Default:  "staging",
		},
		{
			Name:     "TLD",
			Required: true,
			Desc:     "TLD is required for use as main domain on certificate, ex: domain.com",
			Default:  "",
		},
		{
			Name:     "SANS",
			Required: true,
			Desc:     "SANS is required as comma separated list of FQDNs to list on SAN certificate, ex: app.domain.com,other.domain.com",
			Default:  "",
		},
		{
			Name:     "DNS_PROVIDER",
			Required: false,
			Desc:     "Which supported DNS provider to use with Lets Encrypt for validation. You must also set env vars for any other values the DNS provider needs",
			Default:  "cloudflare",
		},
		{
			Name:     "BACKEND1_URL",
			Required: true,
			Desc:     "Url to first backend, ex: http://app:80",
			Default:  "",
		},
		{
			Name:     "FRONTEND1_DOMAIN",
			Required: true,
			Desc:     "Domain for first frontend, ex: app.domain.com",
			Default:  "",
		},
		{
			Name:     "BACKEND2_URL",
			Required: false,
			Desc:     "Url to second backend, ex: http://other:80",
			Default:  "",
		},
		{
			Name:     "FRONTEND2_DOMAIN",
			Required: false,
			Desc:     "Domain for second frontend, ex: otherapp.domain.com",
			Default:  "",
		},
		{
			Name:     "BACKEND3_URL",
			Required: false,
			Desc:     "Url to third backend, ex: http://third:80",
			Default:  "",
		},
		{
			Name:     "FRONTEND3_DOMAIN",
			Required: false,
			Desc:     "Domain for third frontend, ex: thirdapp.domain.com",
			Default:  "",
		},
	}

	return envVars
}
