package config

import (
	"fmt"
	"os"
	"strconv"
)

// getVar will try to get an environment variable by name and return a helpful
// error message in case it is missing.
func getVar(name string) (string, error) {
	value, exists := os.LookupEnv(name)
	if !exists {
		return "", fmt.Errorf(`could not find environment variable %s`, name)
	}

	return value, nil
}

// getVarWithDefault will get an environment variable by name or return a
// default value in case it is missing.
func getVarWithDefault(name string, def string) string {
	value, err := getVar(name)
	if err != nil {
		return def
	}

	return value
}

// getIntVar will get an environment variable by name and convert it to an int.
func getIntVar(name string) (int, error) {
	value, err := getVar(name)
	if err != nil {
		return 0, err
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf(`environment variable %s is not an integer`, name)
	}

	return intValue, nil
}

// getIntVarWithDefault will get an environment variable as an int or return a
// default value if it is missing.
func getIntVarWithDefault(name string, def int) (int, error) {
	value, err := getVar(name)
	if err != nil {
		return def, nil
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf(`environment variable %s is not an integer`, name)
	}

	return intValue, nil
}
