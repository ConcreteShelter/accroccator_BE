package config

import (
	"bufio"
	"os"
	"strings"
)

func SetEnvVariablesFromFile(filePath string) error {

	file, err := os.Open(filePath)

	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "#") {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}

		err = os.Setenv(key, value)
		if err != nil {
			return err
		}

	}

	return scanner.Err()

}
