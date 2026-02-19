package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const baseURL = "https://www.cpc.ncep.noaa.gov/products/CFSv2/"
const historyURL = "https://www.cpc.ncep.noaa.gov/products/CFSv2/cfsv2_fcst_history/"

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
	generationMonth := now.Format("200601")
	generationDay := now.Format("02")

	for folderName, varCode := range variables {
		for lead := 1; lead <= 6; lead++ {

			forecastMonth := now.AddDate(0, lead, 0).Format("200601")
			for _, run := range runs {

				url := buildCurrentURL(varCode, run, lead)
				savePath := filepath.Join(
					inputDir,
					folderName,
					forecastMonth,
					fmt.Sprintf("%s%s_%s.png", generationMonth, generationDay, run),
				)
				_, err := os.Stat(savePath)
				if errors.Is(err, os.ErrNotExist) {
					err := download(url, savePath)
					if err != nil {
						fmt.Println("Error:", err)
					}
				}

			}
		}
	}
	for history := 0; history < 6; history++ {
		historyDate := now.AddDate(0, -history, 0)
		historyMonth := historyDate.Format("200601")
		for lead := 1; lead <= 6; lead++ {
			forecastMonth := historyDate.AddDate(0, lead-1, 0).Format("200601")
			for _, run := range runs {
				for folderName, varCode := range variables {
					url := buildHistoryURL(varCode, run, lead, historyMonth)
					savePath := filepath.Join(
						inputDir,
						folderName,
						forecastMonth,
						fmt.Sprintf("%s_%s.png", historyMonth, run),
					)
					_, err := os.Stat(savePath)
					if errors.Is(err, os.ErrNotExist) {
						err := download(url, savePath)
						if err != nil {
							fmt.Println("Error:", err)
						}
					}
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

func buildHistoryURL(variable, run string, lead int, historyMonth string) string {
	// https://www.cpc.ncep.noaa.gov/products/CFSv2/cfsv2_fcst_history/202602/imagesInd1/euT2mMonInd1.gif
	//	<baseURL><historyMonth>/imagesInd<run>/<variable>MonInd<lead>.gif

	return fmt.Sprintf(
		"%s%s/imagesInd%s/%sMonInd%d.gif",
		historyURL,
		historyMonth,
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
