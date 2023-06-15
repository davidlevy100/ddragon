package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
)

// getChampNames reads from ddragon's 'champions.json' file
// the champ names are the keys within the json file
func getChampNames(client *http.Client, url string) ([]string, error) {
	var results []string

	// the keys change in the json file
	// so not using marshalling or structs
	var names map[string]interface{}

	body, err := getJson(client, url)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &names)
	if err != nil {
		return nil, err
	}

	// this is used to grab the champ keys
	data := names["data"].(map[string]interface{})

	for key := range data {
		results = append(results, key)
	}

	return results, nil

}

// getChampImageFiles compiles a slice of imageFile structs
// from the names, paths, and the current patch
func getChampImageInfo(names []string, patch string, paths map[string]string) ([]imageFile, error) {

	if len(names) == 0 {
		return nil, errors.New("no champion names found to process")
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

// createChampFolders makes the folders in the current directory
// and returns the filepaths for later processing
func createChampFolders(patch string) (map[string]string, error) {

	paths := map[string]string{
		"splash":   fmt.Sprintf(splashPath, patch),
		"centered": fmt.Sprintf(centeredPath, patch),
		"icon":     fmt.Sprintf(iconPath, patch),
		"portrait": fmt.Sprintf(portraitPath, patch),
	}

	for _, thisPath := range paths {
		err := os.MkdirAll(thisPath, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}

	return paths, nil

}

func getChampImageFiles(ddragonClient *http.Client, latestPatch string) ([]imageFile, error) {

	// get names of all champs in latest patch
	champPath := fmt.Sprintf(champURL, latestPatch)
	names, err := getChampNames(ddragonClient, champPath)
	if err != nil {
		return nil, err
	}

	paths, err := createChampFolders(latestPatch)
	if err != nil {
		return nil, err
	}

	imageFiles, err := getChampImageInfo(names, latestPatch, paths)
	if err != nil {
		return nil, err
	}
	return imageFiles, nil
}
