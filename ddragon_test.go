package main

import (
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
