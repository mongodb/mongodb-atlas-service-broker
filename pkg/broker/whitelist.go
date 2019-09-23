package broker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Whitelist map[string][]string

func ReadWhitelistFile(path string) (Whitelist, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	whitelist := Whitelist{}
	if err := json.Unmarshal([]byte(bytes), &whitelist); err != nil {
		return nil, err
	}

	for whitelistProviderName, _ := range whitelist {
		var isValid bool
		for _, providerName := range providerNames {
			if whitelistProviderName == providerName {
				isValid = true
			}
		}
		if !isValid {
			return nil, fmt.Errorf("invalid whitelist")
		}
	}

	return whitelist, nil
}
