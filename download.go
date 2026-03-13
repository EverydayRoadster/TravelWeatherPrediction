package main

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const baseURL = "https://www.cpc.ncep.noaa.gov/products/CFSv2/"
const historyURL = "https://www.cpc.ncep.noaa.gov/products/CFSv2/cfsv2_fcst_history/"

var parameters = map[string]string{
	"Europe_T2m":  "euT2m",
	"Europe_Prec": "euPrec",
}

var ensemble = []string{"1", "2", "3"}

func getImages(inputDir string, cleanupDays int) error {
	now := time.Now().UTC()

	currentMonth := now.Format("200601")
	currentDay := now.Format("02")

	cleanupOldForecastImages(inputDir, currentMonth, now, cleanupDays)

	// current calculation run (daily) images
	for paramName, paramCode := range parameters {
		for lead := 0; lead < 6; lead++ {
			monthAdjustment := 0
			if now.Day() > 11 { // this logic may need to become more sophisticated
				monthAdjustment = 1
			}
			forecastMonth := now.AddDate(0, lead+monthAdjustment, 0).Format("200601")
			for _, run := range ensemble {

				url := buildCurrentURL(paramCode, run, lead+1)
				savePath := filepath.Join(
					inputDir,
					paramName,
					forecastMonth,
					fmt.Sprintf("%s%s_%s.png", currentMonth, currentDay, run),
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

	// history calculated images
	skipHistory := -1
	for history := 1; history <= 6; history++ {
		historyDate := now.AddDate(0, -history, 0)
		historyMonth := historyDate.Format("200601")
		for lead := 1; lead <= 6; lead++ {
			if skipHistory == history {
				break
			}
			forecastMonth := historyDate.AddDate(0, lead-1, 0).Format("200601")
			// download earlier predictions with relevant forecasts only
			if currentMonth <= forecastMonth {
				for _, run := range ensemble {
					if skipHistory == history {
						break
					}
					for paramName, varCode := range parameters {
						url := buildHistoryURL(varCode, run, lead, historyMonth)
						savePath := filepath.Join(
							inputDir,
							paramName,
							forecastMonth,
							fmt.Sprintf("%s_%s.png", historyMonth, run),
						)
						_, err := os.Stat(savePath)
						if errors.Is(err, os.ErrNotExist) {
							if !(skipHistory == history) {
								err := download(url, savePath)
								if err != nil {
									fmt.Println("Error:", err)
									skipHistory = history
									break
								}
							}
						}
					}
				}
			}
		}
	}
	return nil
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
			return fmt.Errorf("This month data not yet available, skipping this month: %s", resp.Request.URL.Path)
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
		return fmt.Errorf("unable to crop: %s", out.Name())
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

func cleanupOldForecastImages(inputDir, currentMonth string, now time.Time, cleanupDays int) {
	latestCleanupDay := now.AddDate(0, 0, -cleanupDays).Format("20060102")

	for paramName := range parameters {
		paramDir := filepath.Join(inputDir, paramName)
		forecastMonths, err := listFiles(paramDir, "??????")
		if err != nil {
			continue
		}
		for _, forecastMonth := range forecastMonths {
			fullMonthPath := filepath.Join(paramDir, forecastMonth)

			// If forecast month is before current month → delete entire folder
			if forecastMonth < currentMonth {
				fmt.Println("Deleting old forecast folder:", fullMonthPath)
				err := os.RemoveAll(fullMonthPath)
				if err != nil {
					fmt.Println("Error deleting folder:", err)
				}
				continue
			}

			// Clean up daily downloads (yyyyMMdd_run.png) older than latestCleanupDay
			dailyImageFiles, err := listFiles(fullMonthPath, "????????_?.png")
			if err != nil {
				continue
			}
			for _, fileName := range dailyImageFiles {
				if fileName[:8] > latestCleanupDay {
					break
				}
				filePath := filepath.Join(fullMonthPath, fileName)
				fmt.Println("Deleting old daily forecast file:", filePath)
				err := os.Remove(filePath)
				if err != nil {
					fmt.Println("Error deleting file:", err)
				}
			}
		}
	}
}

func listFiles(dir, pattern string) ([]string, error) {
	return fs.Glob(os.DirFS(dir), pattern)
}
