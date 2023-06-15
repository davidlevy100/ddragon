package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// getPatch reads from ddragons `versions.json` file.
// getPatch assumes the file is in sorted order
// and the zeroth element is the latest patch number
func getPatch(client *http.Client, url string) (string, error) {

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	var versions []string

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(body, &versions)
	if err != nil {
		return "", err
	}

	return versions[0], nil

}
