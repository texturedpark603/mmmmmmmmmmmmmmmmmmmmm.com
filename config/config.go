package config

import (
	"bufio"
	"os"
	"strings"
)

type Config struct {
	LastFMKey  string
	LastFMUser string
	Port       string
}

func Load(path string) *Config {
	readEnvFile(path)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	user := os.Getenv("LASTFM_USER")
	if user == "" {
		user = "qm"
	}

	return &Config{
		LastFMKey:  os.Getenv("LASTFM_API_KEY"),
		LastFMUser: user,
		Port:       port,
	}
}

func (c *Config) Valid() bool {
	return c.LastFMKey != ""
}

func readEnvFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}
