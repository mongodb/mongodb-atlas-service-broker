package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

// Poll will run f every 10 seconds until it returns true or the timout is reached.
func Poll(timeoutMinutes int, f func() (bool, error)) error {
	pollInterval := 10

	for i := 0; i < timeoutMinutes*60; i++ {
		res, err := f()
		if err != nil {
			return err
		}

		if res {
			return nil
		}

		i += pollInterval
		time.Sleep(time.Duration(pollInterval) * time.Second)
	}

	return fmt.Errorf("timeout while polling (waited %d minutes)", timeoutMinutes)
}

// GetEnvOrPanic will get an environment variable, panicking if it does not exist.
func GetEnvOrPanic(name string) string {
	value, exists := os.LookupEnv(name)
	if !exists {
		panic(fmt.Sprintf(`Could not find environment variable "%s"`, name))
	}

	return value
}

// ReadInYAMLFileAndConvert reads in the yaml file given by the path given
func ReadInYAMLFileAndConvert(pathToYamlFile string, crd interface{}) interface{} {
	// Read in the yaml file at the path given
	yamlFile, err := ioutil.ReadFile(pathToYamlFile)
	if err != nil {
		log.Printf("Error while parsing YAML file %v, error: %s", pathToYamlFile, err)
	}

	// Map yamlFile to interface
	var body interface{}
	if err := yaml.Unmarshal([]byte(yamlFile), &body); err != nil {
		panic(err)
	}

	// Convert yaml to its json counterpart
	body = ConvertYAMLtoJSONHelper(body)

	// Generate json string from data structure
	jsonFormat, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(jsonFormat, &crd); err != nil {
		panic(err)
	}

	return crd
}

// ConvertYAMLtoJSONHelper converts the yaml to json recursively
func ConvertYAMLtoJSONHelper(i interface{}) interface{} {
	switch item := i.(type) {
	case map[interface{}]interface{}:
		document := map[string]interface{}{}
		for k, v := range item {
			document[k.(string)] = ConvertYAMLtoJSONHelper(v)
		}
		return document
	case []interface{}:
		for i, arr := range item {
			item[i] = ConvertYAMLtoJSONHelper(arr)
		}
	}

	return i
}
