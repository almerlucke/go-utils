package files

import (
	"fmt"
	"os"
	"strings"
)

// ReadDotEnvFile reads a .env file and returns a map with VAR=VAL pairs
// - no trimming of whitespace, VAR and VAL are read as is separated by a =
// - empty lines and lines that start with a # are skipped
// - vars can be added to the environment by setting addToEnv to true
func ReadDotEnvFile(filePath string, addToEnv bool) (map[string]string, error) {
	lines, err := ScanFile(filePath)
	if err != nil {
		return nil, err
	}

	m := map[string]string{}

	for line := range lines {
		if line.Error != nil {
			return nil, line.Error
		}

		if line.Line == "" || strings.HasPrefix(line.Line, "#") {
			continue
		}

		components := strings.SplitN(line.Line, "=", 2)
		if len(components) != 2 {
			return nil, fmt.Errorf("error on line %d: expected a var and value", line.Count)
		}

		m[components[0]] = components[1]
	}

	if addToEnv {
		for k, v := range m {
			err = os.Setenv(k, v)
			if err != nil {
				return nil, err
			}
		}
	}

	return m, nil
}
