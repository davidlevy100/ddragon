package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"unicode"
)

const (
	itemsUrl     = "http://ddragon.leagueoflegends.com/cdn/%s/data/en_US/item.json"
	itemImageURL = "http://ddragon.leagueoflegends.com/cdn/%s/img/item/%s"
	itemsPath    = "assets/%s/items"
)

type Item struct {
	Name  string `json:"name"`
	Image struct {
		Full string `json:"full"`
	} `json:"image"`
}

// getItemData reads from ddragon's 'item.json' file
// the item names are the keys within the json file
func getItemData(client *http.Client, url string, patch string) (map[string]string, error) {

	// the keys change in the json file
	// so not using marshalling or structs
	var data map[string]interface{}

	body, err := getJson(client, url)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	items := data["data"].(map[string]interface{})

	var results = make(map[string]string)

	for _, thisItem := range items {
		jsonData, err := json.Marshal(thisItem)
		if err != nil {
			continue
		}

		var i Item

		err = json.Unmarshal(jsonData, &i)
		if err != nil {
			continue
		}

		name := parseName(i.Name)

		results[name] = fmt.Sprintf(itemImageURL, patch, i.Image.Full)

	}

	return results, nil

}

// getItemImageFiles compiles a slice of imageFile structs
// from the names, paths, and the current patch
func getItemImageInfo(itemImageUrls map[string]string, path string) ([]imageFile, error) {

	if len(itemImageUrls) == 0 {
		return nil, errors.New("no item names found to process")
	}

	imageFiles := []imageFile{}

	for key, url := range itemImageUrls {

		i := imageFile{
			name: key,
			url:  url,
			path: fmt.Sprintf("%s/%s.png", path, key),
		}

		imageFiles = append(imageFiles, i)
	}

	return imageFiles, nil

}

// createItemFolders makes the folders in the current directory
// and returns the filepaths for later processing
func createItemFolder(patch string) (string, error) {

	path := fmt.Sprintf(itemsPath, patch)

	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return "", err
	}

	return path, nil

}

func getItemImageFiles(ddragonClient *http.Client, latestPatch string) ([]imageFile, error) {

	url := fmt.Sprintf(itemsUrl, latestPatch)

	itemImageUrls, err := getItemData(ddragonClient, url, latestPatch)
	if err != nil {
		return nil, err
	}

	path, err := createItemFolder(latestPatch)
	if err != nil {
		return nil, err
	}

	imageFiles, err := getItemImageInfo(itemImageUrls, path)
	if err != nil {
		return nil, err
	}
	return imageFiles, nil

}

func parseName(name string) string {

	var builder strings.Builder

	// Iterate over each character in the input string
	for _, char := range name {
		if unicode.IsLetter(char) {
			builder.WriteRune(char)
		}
	}

	output := builder.String()

	return output

}
