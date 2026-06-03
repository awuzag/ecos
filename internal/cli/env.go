package cli

import (
	"bufio"
	"os"
	"strings"

	"github.com/samber/oops"
)

const ecosAPIKeyEnv = "ECOS_API_KEY"

func resolveAPIKey(flagValue string, envFile string, envFileExplicit bool) (string, error) {
	if value := strings.TrimSpace(flagValue); value != "" {
		return value, nil
	}
	if value := strings.TrimSpace(os.Getenv(ecosAPIKeyEnv)); value != "" {
		return value, nil
	}
	if strings.TrimSpace(envFile) == "" {
		return "", nil
	}
	return readAPIKeyFromEnvFile(envFile, envFileExplicit)
}

func readAPIKeyFromEnvFile(path string, explicit bool) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) && !explicit {
			return "", nil
		}
		return "", oops.In("cli").
			With("env_file", path).
			Wrap(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, found := strings.Cut(line, "=")
		if !found || strings.TrimSpace(key) != ecosAPIKeyEnv {
			continue
		}
		return strings.Trim(strings.TrimSpace(value), `"'`), nil
	}
	if err := scanner.Err(); err != nil {
		return "", oops.In("cli").
			With("env_file", path).
			Wrap(err)
	}
	return "", nil
}
