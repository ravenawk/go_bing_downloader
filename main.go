package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type BingResponse struct {
	Images []Image
}
type Image struct {
	Startdate string `json:"startdate"`
	URL       string `json:"url"`
	Title     string `json:"title"`
}

var imageTitlere = regexp.MustCompile(`[^a-zA-Z0-9_-]`)

const bingURL string = "https://www.bing.com"

func main() {
	httpClient := &http.Client{Timeout: 10 * time.Second}
	imageMetadataPath := "/HPImageArchive.aspx?format=js&idx=0&n=1&mkt=en-US"
	imageMetadataURL := bingURL + imageMetadataPath
	imageJSON, err := fetchURLBody(httpClient, imageMetadataURL)
	if err != nil {
		log.Fatal(err)
	}

	imageInfo, err := transformJSON(imageJSON)
	if err != nil {
		log.Fatal(err)
	}
	imageBytes, err := fetchURLBody(httpClient, bingURL+imageInfo.URL)
	if err != nil {
		log.Fatal(err)
	}

	cleanTitle := sanitizeTitle(imageInfo.Title)
	err = saveImage(imageBytes, cleanTitle, imageInfo.Startdate)
	if err != nil {
		log.Fatal(err)
	}

}

func fetchURLBody(client *http.Client, imageURL string) ([]byte, error) {
	response, err := client.Get(imageURL)

	// log.Fatal shouldn't be used in a non-main function it should return the error to main
	if err != nil {
		return nil, err
	}

	// Defer in a function needs to return any data open files before the return statement which causes the defer to fire and close the data steam (whatever that might be)
	defer response.Body.Close()

	// For things like status codes things like fmt.Errorf (to include vars - %d for example) or errors.New(to include just text) should be used
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code - %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return body, nil

}

func transformJSON(imageInfo []byte) (Image, error) {
	var imageOTD BingResponse
	err := json.Unmarshal(imageInfo, &imageOTD)
	if err != nil {
		return Image{}, err
	}
	if len(imageOTD.Images) == 0 {
		return Image{}, errors.New("something went wrong, no data pulled")
	}

	return imageOTD.Images[0], nil

}

func sanitizeTitle(title string) string {
	cleanTitle := strings.ReplaceAll(title, " ", "_")
	return imageTitlere.ReplaceAllString(cleanTitle, "")
}

func saveImage(image []byte, title string, startdate string) error {

	imagePath := filepath.Join("Pictures", "BingWallpapers")
	fileName := title + "_" + startdate + ".jpg"
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	savePath := filepath.Join(homeDir, imagePath)
	filePath := filepath.Join(savePath, fileName)
	err = os.MkdirAll(savePath, 0700)
	if err != nil {
		return err
	}
	err = os.WriteFile(filePath, image, 0600)
	if err != nil {
		return err
	}
	fmt.Printf("Image %s has been downloaded and saved!\n", fileName)

	return nil

}
