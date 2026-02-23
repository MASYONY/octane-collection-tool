package config

import (
	"os"
	"path/filepath"
	"strings"
)

func ParseProperties(content string) map[string]string {
	result := map[string]string{}
	for _, rawLine := range strings.Split(content, "\n") {
		line := strings.TrimSpace(strings.TrimSuffix(rawLine, "\r"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		sep := strings.Index(line, "=")
		if sep < 0 {
			continue
		}
		key := strings.TrimSpace(line[:sep])
		value := strings.TrimSpace(line[sep+1:])
		result[key] = value
	}
	return result
}

func LoadConfig(configPath string) (map[string]string, error) {
	resolvedPath := configPath
	if resolvedPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		resolvedPath = filepath.Join(cwd, "config.properties")
	}
	resolvedPath, err := filepath.Abs(resolvedPath)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(resolvedPath); err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, err
	}
	content, err := os.ReadFile(resolvedPath)
	if err != nil {
		return nil, err
	}
	return ParseProperties(string(content)), nil
}
