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

// var ensemble = []string{"E1", "E2", "E3"}
var ensemble = []string{"1", "2", "3"}

// downloadImages is a stub that should download a predefined set of images
// and return the path to the directory where they are stored.
func getImages(inputDir string) (string, error) {
	now := time.Now().UTC()
	generationMonth := now.Format("200601")
	//	generationDayHour := now.Format("0215")
	generationDayHour := now.Format("02")

	for folderName, varCode := range variables {
		for lead := 1; lead <= 6; lead++ {

			forecastMonth := now.AddDate(0, lead, 0).Format("200601")
			for _, run := range ensemble {

				url := buildCurrentURL(varCode, run, lead)
				savePath := filepath.Join(
					inputDir,
					folderName,
					forecastMonth,
					fmt.Sprintf("%s%s_%s.png", generationMonth, generationDayHour, run),
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
			// download earlier predictions with relevant forecasts only
			if generationMonth <= forecastMonth {
				for _, run := range ensemble {
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
	}

	cleanupOldForecasts(inputDir, now)

	return inputDir, nil
}

func buildCurrentURL(variable, run string, lead int) string {
	return fmt.Sprintf(
		"%simagesInd%s/%sMonInd%d.gif",
		baseURL,
		run,
		variable,
		lead,
	)
}

func buildHistoryURL(variable, run string, lead int, historyMonth string) string {
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

func cleanupOldForecasts(inputDir string, now time.Time) {
	currentMonth := now.Format("200601")

	for folderName := range variables {
		varDir := filepath.Join(inputDir, folderName)
		entries, err := os.ReadDir(varDir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			forecastMonth := entry.Name()
			// If forecast month is before current month → delete
			if forecastMonth < currentMonth {
				fullPath := filepath.Join(varDir, forecastMonth)
				fmt.Println("Deleting old forecast folder:", fullPath)
				err := os.RemoveAll(fullPath)
				if err != nil {
					fmt.Println("Error deleting:", err)
				}
			}
			// If forecast month is current month → delete daily results
			/*			if forecastMonth == currentMonth {
						prevMonth := now.AddDate(0, -1, 0).Format("200601")
						matches, _ := filepath.Glob(filepath.Join(varDir, forecastMonth, prevMonth+"??_?.png"))
						for _, matching := range matches {
							fmt.Println("Deleting old daily forecast file:", matching)
							if err != os.Remove(matching) {
								fmt.Println("Error deleting:", err)
							}
						}
					} */
		}
	}
}
