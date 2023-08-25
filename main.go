package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	versionURL = "http://ddragon.leagueoflegends.com/api/versions.json"
	assetPath  = "assets/%s"
	waitTime   = 10
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

	newPath := fmt.Sprintf(assetPath, latestPatch)

	err = os.MkdirAll(newPath, os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}

	logname := fmt.Sprintf("%s/logs.txt", newPath)

	logfile, err := os.OpenFile(logname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	log.SetOutput(logfile)

	var imageFiles []imageFile

	champFiles, err := getChampImageFiles(ddragonClient, latestPatch)
	if err != nil {
		fmt.Println(err)
	} else {
		imageFiles = append(imageFiles, champFiles...)
	}

	runeFiles, err := getRuneImageFiles(ddragonClient, latestPatch)
	if err != nil {
		fmt.Println(err)
	} else {
		imageFiles = append(imageFiles, runeFiles...)
	}

	itemFiles, err := getItemImageFiles(ddragonClient, latestPatch)
	if err != nil {
		fmt.Println(err)
	} else {
		imageFiles = append(imageFiles, itemFiles...)
	}

	if len(imageFiles) == 0 {
		fmt.Println("could not download any files")
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
