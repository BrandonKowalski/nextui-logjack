package internal

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Directory struct {
	Name string
	Path string
}

type Config struct {
	Directories []Directory
}

var configInstance *Config

func LoadConfig() (*Config, error) {
	configPath := getConfigPath()

	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			config := &Config{
				Directories: []Directory{
					{Name: "Logs", Path: "/tmp/logjack-files"},
				},
			}
			configInstance = config
			return config, nil
		}
		return nil, err
	}
	defer file.Close()

	var directories []Directory
	scanner := bufio.NewScanner(file)
	inDirectories := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section := strings.ToLower(strings.Trim(line, "[]"))
			inDirectories = section == "directories"
			continue
		}

		if inDirectories {
			if idx := strings.Index(line, "="); idx > 0 {
				name := strings.TrimSpace(line[:idx])
				path := expandEnvVars(strings.TrimSpace(line[idx+1:]))
				if name != "" && path != "" {
					directories = append(directories, Directory{Name: name, Path: path})
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(directories) == 0 {
		directories = []Directory{
			{Name: "Logs", Path: "/tmp/logjack-files"},
		}
	}

	config := &Config{Directories: directories}
	configInstance = config
	return config, nil
}

func GetConfig() *Config {
	if configInstance == nil {
		config, err := LoadConfig()
		if err != nil {
			return &Config{
				Directories: []Directory{
					{Name: "Logs", Path: "/tmp/logjack-files"},
				},
			}
		}
		return config
	}
	return configInstance
}

func getConfigPath() string {
	if os.Getenv("ENVIRONMENT") == "DEV" {
		return "logjack.ini"
	}
	return filepath.Join(filepath.Dir(os.Args[0]), "logjack.ini")
}

var envVarPattern = regexp.MustCompile(`\{([^}]+)\}`)

func expandEnvVars(s string) string {
	return envVarPattern.ReplaceAllStringFunc(s, func(match string) string {
		varName := match[1 : len(match)-1]
		if val := os.Getenv(varName); val != "" {
			return val
		}
		return match
	})
}
