package main

import (
	"errors"
	"fmt"
	"image"
	"image/png"
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
	fmt.Println("Downloading: ", url)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == 404 {
			return fmt.Errorf("not (yet available): %s", resp.Request.URL.Path)
		}
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

	image, err := cropImage(resp.Body)
	if err != nil {
		return err
	}

	err = png.Encode(out, image)
	if err != nil {
		return err
	}
	return nil
}

// cropImage crops a PNG image by removing 50px from left/right and 30px from top.
func cropImage(reader io.Reader) (image.Image, error) {

	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	if w <= 100 || h <= 30 {
		return nil, fmt.Errorf("image too small to crop: %dx%d", w, h)
	}

	cropRect := image.Rect(40, 30, w-50, h-50)
	cropped := img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(cropRect)

	return cropped, nil
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
		}
	}
}
