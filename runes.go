package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
)

const (
	runesUrl     = "http://ddragon.leagueoflegends.com/cdn/%s/data/en_US/runesReforged.json"
	runeImageURL = "http://ddragon.leagueoflegends.com/cdn/img/%s"
	runesPath    = "assets/%s/runes"
)

type RuneData []struct {
	ID    int    `json:"id"`
	Key   string `json:"key"`
	Icon  string `json:"icon"`
	Name  string `json:"name"`
	Slots []struct {
		Runes []struct {
			ID   int    `json:"id"`
			Key  string `json:"key"`
			Icon string `json:"icon"`
			Name string `json:"name"`
		} `json:"runes"`
	} `json:"slots"`
}

// getRuneNames reads from ddragon's 'RunesReforged.json' file
// the Rune names are the keys within the json file
func getRuneData(client *http.Client, url string) (map[string]string, error) {

	var data RuneData

	body, err := getJson(client, url)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	var results = make(map[string]string)

	for _, thisRuneFamily := range data {
		results[thisRuneFamily.Key] = fmt.Sprintf(runeImageURL, thisRuneFamily.Icon)

		for _, thisSlot := range thisRuneFamily.Slots {
			for _, thisRune := range thisSlot.Runes {
				results[thisRune.Key] = fmt.Sprintf(runeImageURL, thisRune.Icon)
			}
		}
	}

	return results, nil

}

// getRuneImageFiles compiles a slice of imageFile structs
// from the names, paths, and the current patch
func getRuneImageInfo(runeImageUrls map[string]string, path string) ([]imageFile, error) {

	if len(runeImageUrls) == 0 {
		return nil, errors.New("no rune names found to process")
	}

	imageFiles := []imageFile{}

	for key, url := range runeImageUrls {

		i := imageFile{
			name: key,
			url:  url,
			path: fmt.Sprintf("%s/%s.png", path, key),
		}

		imageFiles = append(imageFiles, i)
	}

	return imageFiles, nil

}

// createRuneFolders makes the folders in the current directory
// and returns the filepaths for later processing
func createRuneFolder(patch string) (string, error) {

	path := fmt.Sprintf(runesPath, patch)

	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return "", err
	}

	return path, nil

}

func getRuneImageFiles(ddragonClient *http.Client, latestPatch string) ([]imageFile, error) {

	url := fmt.Sprintf(runesUrl, latestPatch)

	runeImageUrls, err := getRuneData(ddragonClient, url)
	if err != nil {
		return nil, err
	}

	path, err := createRuneFolder(latestPatch)
	if err != nil {
		return nil, err
	}

	imageFiles, err := getRuneImageInfo(runeImageUrls, path)
	if err != nil {
		return nil, err
	}
	return imageFiles, nil

}
