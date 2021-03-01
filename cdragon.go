// Package cdragon offers methods to download cdragon images
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func getPatch() (string, error) {

	var versions []string

	url := "https://ddragon.leagueoflegends.com/api/versions.json"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return "", err
	}

	res, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf("%s", err), err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Sprintf("%s", err), err
	}

	err = json.Unmarshal(body, &versions)
	if err != nil {
		return fmt.Sprintf("%s", err), err
	}

	return versions[0], nil

}

func getNames(patch string) ([]string, error) {

	var results []string

	var names map[string]interface{}

	url := fmt.Sprintf("http://ddragon.leagueoflegends.com/cdn/%s/data/en_US/champion.json", patch)
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return []string{}, err
	}

	res, err := client.Do(req)
	if err != nil {
		return []string{}, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return []string{}, err
	}

	err = json.Unmarshal(body, &names)
	if err != nil {
		return []string{}, err
	}

	d, _ := names["data"]

	for key, _ := range d.(map[string]interface{}) {
		results = append(results, key)
	}

	return results, nil

}

func getImages(patch string, names []string) {

	for _, name := range names {

		url := fmt.Sprintf("https://cdn.communitydragon.org/%s/champion/%s/splash-art/centered", patch, name)

		getImage(name, url)

	}

}

func getImage(name, url string) {

	response, e := http.Get(url)
	if e != nil {
		log.Fatal(e)
	}
	defer response.Body.Close()

	saveImage(name, response.Body)

}

func saveImage(name string, imageBytes io.ReadCloser) {

	os.Mkdir("SplashCentered", 0777)

	file, err := os.Create(fmt.Sprintf("SplashCentered/%s.jpg", name))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	_, err = io.Copy(file, imageBytes)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("saved %s.jpg\n", name)

}

func main() {

	patch, err := getPatch()
	if err != nil {
		fmt.Println("Could not find patch info")
		return
	}

	names, err := getNames(patch)
	if err != nil {
		fmt.Println("Could not get champ names")
		return
	}

	getImages(patch, names)

}
