package main

import (
	"bytes"
	"encoding/json"
	"errors"
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
	versionURL   = "http://ddragon.leagueoflegends.com/api/versions.json"
	champURL     = "http://ddragon.leagueoflegends.com/cdn/%s/data/en_US/champion.json"
	centeredURL  = "http://ddragon.leagueoflegends.com/cdn/img/champion/centered/%s_0.jpg"
	splashURL    = "http://ddragon.leagueoflegends.com/cdn/img/champion/splash/%s_0.jpg"
	iconURL      = "http://ddragon.leagueoflegends.com/cdn/%s/img/champion/%s.png"
	portraitURL  = "http://ddragon.leagueoflegends.com/cdn/img/champion/loading/%s_0.jpg"
	splashPath   = "assets/%s/Splash"
	centeredPath = "assets/%s/SplashCentered"
	iconPath     = "assets/%s/Icon"
	portraitPath = "assets/%s/Portrait"
	assetPath    = "assets/%s"
	waitTime     = 0
)

type imageFile struct {
	name string
	url  string
	data bytes.Buffer
	path string
}

func main() {

	// create global client for speed
	ddragonClient := &http.Client{
		Timeout: time.Second * 10,
	}

	// get latest patch version number
	latestPatch, err := getPatch(ddragonClient, versionURL)
	if err != nil {
		fmt.Println(err)
		return
	}

	// get names of all champs in latest patch
	champs := fmt.Sprintf(champURL, latestPatch)
	names, err := getNames(ddragonClient, champs, latestPatch)
	if err != nil {
		fmt.Println(err)
		return
	}

	paths, err := createFolders(latestPatch)
	if err != nil {
		fmt.Println(err)
		return
	}

	imageFiles, err := getImageFiles(names, latestPatch, paths)
	if err != nil {
		fmt.Println(err)
		return
	}

	// collector that creates, owns sends to and closes the channel
	ch := collector(ddragonClient, imageFiles)

	// channel reciever
	sink(ch)

}

// collector creates and owns the imageFile channel
// and is responsible for closing it
// collector greates a goroutine for each image to download
// and puts the result on the imageFile channel
func collector(client *http.Client, imageFiles []imageFile) <-chan imageFile {

	var wg sync.WaitGroup

	out := make(chan imageFile)

	for _, i := range imageFiles {
		wg.Add(1)
		go func(i imageFile) {
			resp, err := client.Get(i.url)
			defer wg.Done()
			defer resp.Body.Close()
			if err != nil {
				log.Println(err)
				return
			}
			if resp.StatusCode != http.StatusOK {
				log.Println(resp.StatusCode, i.url)
				return
			}

			_, err = io.Copy(&i.data, resp.Body)
			if err != nil {
				log.Println(err)
				return
			}

			out <- i

		}(i)
		time.Sleep(waitTime)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// sink recieves image files on the read-only imageFile channel
// the range loop will automaticaly stop when the channel's owner
// closes the channel
func sink(ch <-chan imageFile) {

	for f := range ch {
		go func(f imageFile) {

			file, err := os.Create(f.path)
			if err != nil {
				log.Printf("could not save %s, %v\n", f.path, err)
				return
			} else {
				file.Write(f.data.Bytes())
			}

			defer file.Close()

		}(f)
	}

}

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

// getNames reads from ddragon's 'champions.json' file
// the champ names are the keys within the json file
func getNames(client *http.Client, url, patch string) ([]string, error) {
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

	// this is used to grab the champ keys
	data := names["data"].(map[string]interface{})

	for key := range data {
		results = append(results, key)
	}

	return results, nil

}

// getImageFiles compiles a slice of imageFile structs
// from the names, paths, and the current patch
func getImageFiles(names []string, patch string, paths map[string]string) ([]imageFile, error) {

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

// createFolders makes the folders in the current directory
// and returns the filepaths for later processing
func createFolders(patch string) (map[string]string, error) {

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

	logname := fmt.Sprintf("assets/%s/logs.txt", patch)

	logfile, err := os.OpenFile(logname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}
	log.SetOutput(logfile)

	return paths, nil

}
