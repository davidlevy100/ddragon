package main

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"
	"time"
)

func TestGetPatch(t *testing.T) {

	testClient := &http.Client{
		Timeout: time.Second * 10,
	}

	r, _ := regexp.Compile(`^\d+\.\d+\.\d$`)

	patchTests := []struct {
		name, url string
		result    bool
	}{
		{"ddragon", "http://ddragon.leagueoflegends.com/api/versions.json", true},
	}

	for _, thisTest := range patchTests {
		result, err := getPatch(testClient, thisTest.url)
		if err != nil {
			if !r.MatchString(result) {
				t.Errorf("recieved incompatible result from %s: %s", thisTest.name, result)
			} else {
				t.Error(err)
			}
		}
	}
}

func TestGetRuneData(t *testing.T) {

	testClient := &http.Client{
		Timeout: time.Second * 10,
	}

	patch, err := getPatch(testClient, versionURL)
	if err != nil {
		t.Error(err)
	}

	url := fmt.Sprintf(runesUrl, patch)

	data, err := getRuneData(testClient, url)
	if err != nil {
		t.Error(err)
	}

	t.Log(data)

}
