package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const baseURL = "https://www.cpc.ncep.noaa.gov/products/CFSv2/"

var variables = map[string]string{
	"Europe_T2m":  "euT2m",
	"Europe_Prec": "euPrec",
}

// var runs = []string{"E1", "E2", "E3"}
var runs = []string{"1", "2", "3"}

// downloadImages is a stub that should download a predefined set of images
// and return the path to the directory where they are stored.
// The actual implementation will be provided later.
func getImages(inputDir string) (string, error) {
	now := time.Now().UTC()
	generationMonth := now.Format("2006-01")

	for folderName, varCode := range variables {

		for lead := 1; lead <= 6; lead++ {

			forecastMonth := now.AddDate(0, lead, 0).Format("2006-01")

			for _, run := range runs {

				url := buildCurrentURL(varCode, run, lead)

				savePath := filepath.Join(
					inputDir,
					folderName,
					forecastMonth,
					fmt.Sprintf("%s_%s.png", generationMonth, run),
				)

				err := download(url, savePath)
				if err != nil {
					fmt.Println("Error:", err)
				}
			}
		}
	}
	return inputDir, nil
}

func buildCurrentURL(variable, run string, lead int) string {
	// https://www.cpc.ncep.noaa.gov/products/CFSv2/imagesInd1/euT2mMonInd1.gif
	//	<baseURL>imagesInd<run>/<variable>MonInd<lead>.gif

	return fmt.Sprintf(
		"%simagesInd%s/%sMonInd%d.gif",
		baseURL,
		run,
		variable,
		lead,
	)
}

func download(url, path string) error {

	fmt.Println("Downloading:", url)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
