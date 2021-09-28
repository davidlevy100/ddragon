package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	versionURL  = "https://ddragon.leagueoflegends.com/api/versions.json"
	champURL    = "https://ddragon.leagueoflegends.com/cdn/%s/data/en_US/champion.json"
	centeredURL = "https://cdn.communitydragon.org/%s/champion/%s/splash-art/centered"
	splashURL   = "https://cdn.communitydragon.org/%s/champion/%s/splash-art"
	iconURL     = "https://cdn.communitydragon.org/%s/champion/%s/square"
	portraitURL = "https://cdn.communitydragon.org/%s/champion/%s/portrait"
	waitTime    = 75
)

func main() {

	//create global client for speed
	cdragonClient := &http.Client{
		Timeout: time.Second * 10,
	}

	//
	patch, err := getPatch(cdragonClient)
	if err != nil {
		fmt.Println("Could not find patch info")
		return
	}

	names, err := getNames(cdragonClient, patch)
	if err != nil {
		fmt.Println("Could not get champ names")
		return
	}

	paths := map[string]string{
		"splash":   fmt.Sprintf("%s/Splash", patch),
		"centered": fmt.Sprintf("%s/SplashCentered", patch),
		"icon":     fmt.Sprintf("%s/Icon", patch),
		"portrait": fmt.Sprintf("%s/Portrait", patch),
	}

	err = os.MkdirAll(patch, os.ModePerm)
	if err != nil {
		fmt.Println("Could not create required folder:", patch)
		return
	}

	for _, thisPath := range paths {
		err := os.MkdirAll(thisPath, os.ModePerm)
		if err != nil {
			fmt.Println("Could not create required folder:", thisPath)
			return
		}

	}

	logname := fmt.Sprintf("%s/logs.txt", patch)

	logfile, err := os.OpenFile(logname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println("Could not create log file")
		return
	}
	log.SetOutput(logfile)

	var wg sync.WaitGroup

	for _, thisName := range names {

		urls := map[string]string{
			"splash":   fmt.Sprintf(splashURL, patch, thisName),
			"centered": fmt.Sprintf(centeredURL, patch, thisName),
			"icon":     fmt.Sprintf(iconURL, patch, thisName),
			"portrait": fmt.Sprintf(portraitURL, patch, thisName),
		}

		for imageType, url := range urls {
			filePath := fmt.Sprintf("%s/%s.jpg", paths[imageType], thisName)
			wg.Add(1)
			go getImage(cdragonClient, url, filePath, &wg)
		}

		time.Sleep(waitTime * time.Millisecond)

	}

	wg.Wait()

	fmt.Println("Done")
}

func getPatch(client *http.Client) (string, error) {

	resp, err := client.Get(versionURL)
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

func getNames(client *http.Client, patch string) ([]string, error) {
	var results []string
	var names map[string]interface{}

	url := fmt.Sprintf(champURL, patch)

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

	data := names["data"].(map[string]interface{})

	for key := range data {
		results = append(results, key)
	}

	return results, nil

}

func getImage(client *http.Client, url, path string, wg *sync.WaitGroup) {

	resp, err := client.Get(url)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Println(resp.StatusCode, url)
		return
	}

	var data bytes.Buffer

	_, err = io.Copy(&data, resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	file, err := os.Create(path)
	if err != nil {
		log.Printf("could not save %s, %v\n", path, err)
		return
	}

	defer wg.Done()
	defer file.Close()

	file.Write(data.Bytes())
	fmt.Println("writing", path)

}
