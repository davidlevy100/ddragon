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
	versionURL  = "http://ddragon.leagueoflegends.com/api/versions.json"
	champURL    = "http://ddragon.leagueoflegends.com/cdn/%s/data/en_US/champion.json"
	centeredURL = "http://ddragon.leagueoflegends.com/cdn/img/champion/centered/%s_0.jpg"
	splashURL   = "http://ddragon.leagueoflegends.com/cdn/img/champion/splash/%s_0.jpg"
	iconURL     = "http://ddragon.leagueoflegends.com/cdn/%s/img/champion/%s.png"
	portraitURL = "http://ddragon.leagueoflegends.com/cdn/img/champion/loading/%s_0.jpg"
	waitTime    = 0
)

type imageFile struct {
	name string
	url  string
	data bytes.Buffer
	path string
}

func main() {
	// create global client for speed
	cdragonClient := &http.Client{
		Timeout: time.Second * 10,
	}

	// get latest patch version number
	patch, err := getPatch(cdragonClient)
	if err != nil {
		fmt.Println("Could not find patch info")
		return
	}

	paths := map[string]string{
		"splash":   fmt.Sprintf("assets/%s/Splash", patch),
		"centered": fmt.Sprintf("assets/%s/SplashCentered", patch),
		"icon":     fmt.Sprintf("assets/%s/Icon", patch),
		"portrait": fmt.Sprintf("assets/%s/Portrait", patch),
	}

	newPath := fmt.Sprintf("assets/%s", patch)

	err = os.MkdirAll(newPath, os.ModePerm)
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

	logname := fmt.Sprintf("assets/%s/logs.txt", patch)

	logfile, err := os.OpenFile(logname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println("Could not create log file")
		return
	}
	log.SetOutput(logfile)

	names, err := getNames(cdragonClient, patch)
	if err != nil {
		fmt.Println("Could not get champ names")
		return
	}

	imageFiles := []imageFile{}

	for _, n := range names {

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

	ch := collector(cdragonClient, imageFiles)
	sink(ch)

}

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

func sink(ch <-chan imageFile) {

	for f := range ch {
		go func(f imageFile) {

			file, err := os.Create(f.path)
			if err != nil {
				log.Printf("could not save %s, %v\n", f.path, err)
				return
			} else {
				file.Write(f.data.Bytes())
				fmt.Println("writing", f.path)
			}

			defer file.Close()

		}(f)
	}

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
