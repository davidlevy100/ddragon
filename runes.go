package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

// getRuneNames reads from ddragon's 'Runeions.json' file
// the Rune names are the keys within the json file
func getRuneNames(client *http.Client, url, patch string) ([]string, error) {
	var results []string

	// the keys change in the json file
	// so not using marshalling or structs
	var names map[string]interface{}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &names)
	if err != nil {
		return nil, err
	}

	// this is used to grab the Rune keys
	data := names["data"].(map[string]interface{})

	for key := range data {
		results = append(results, key)
	}

	return results, nil

}

// getRuneImageFiles compiles a slice of imageFile structs
// from the names, paths, and the current patch
func getRuneImageInfo(names []string, patch string, paths map[string]string) ([]imageFile, error) {

	if len(names) == 0 {
		return nil, errors.New("no Runeion names found to process")
	}

	imageFiles := []imageFile{}

	for _, n := range names {

		// because the FiddleSticks icon is improperly named in ddragon
		if n == "Fiddlesticks" {
			n = "FiddleSticks"
		}

		urls := map[string]string{
			"splash":   fmt.Sprintf(splashURL, n),
			"centered": fmt.Sprintf(centeredURL, n),
			"icon":     fmt.Sprintf(iconURL, patch, n),
			"portrait": fmt.Sprintf(portraitURL, n),
		}

		for imageType, u := range urls {
			filePath := fmt.Sprintf("%s/%s.jpg", paths[imageType], n)

			i := imageFile{
				name: n,
				url:  u,
				path: filePath,
			}

			imageFiles = append(imageFiles, i)
		}
	}

	// because the FiddleSticks icon is improperly named in ddragon
	f := imageFile{
		name: "FiddleSticks",
		url:  fmt.Sprintf(iconURL, patch, "Fiddlesticks"),
		path: fmt.Sprintf("%s/%s.jpg", paths["icon"], "FiddleSticks"),
	}

	imageFiles = append(imageFiles, f)

	return imageFiles, nil

}

// createRuneFolders makes the folders in the current directory
// and returns the filepaths for later processing
func createRuneFolders(patch string) (map[string]string, error) {

	paths := map[string]string{
		"splash":   fmt.Sprintf(splashPath, patch),
		"centered": fmt.Sprintf(centeredPath, patch),
		"icon":     fmt.Sprintf(iconPath, patch),
		"portrait": fmt.Sprintf(portraitPath, patch),
	}

	newPath := fmt.Sprintf(assetPath, patch)

	err := os.MkdirAll(newPath, os.ModePerm)
	if err != nil {
		return nil, err
	}

	for _, thisPath := range paths {
		err := os.MkdirAll(thisPath, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}

	return paths, nil

}

func getRuneImageFiles(ddragonClient *http.Client, latestPatch string) ([]imageFile, error) {

	// get names of all Runes in latest patch
	Runes := fmt.Sprintf(runesURL, latestPatch)
	names, err := getRuneNames(ddragonClient, Runes, latestPatch)
	if err != nil {
		return nil, err
	}

	paths, err := createRuneFolders(latestPatch)
	if err != nil {
		return nil, err
	}

	imageFiles, err := getRuneImageInfo(names, latestPatch, paths)
	if err != nil {
		return nil, err
	} else {
		return imageFiles, nil
	}

}