// main.go
// Entry point for the Weather Forecast Pattern Analyzer.
// The program reads PNG images from a directory, calculates the most dominant
// color per pixel across all images, and writes a composite image.
//
// Usage: go run . -input path/to/dir -output path/to/output.png

package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
)

type Constant int

const (
	RENDER_WHITE Constant = iota
	RENDER_SMOOTH
	RENDER_CONFIDENCE
	DEFAULT_NOAA_DIR
)

var CNST = []string{"white", "smooth", "confidence", ".noaa"}

func main() {
	var inputDir, outputDir string
	var renderMode string

	flag.StringVar(&renderMode, "renderMode", CNST[RENDER_WHITE], "enable render mode")
	flag.StringVar(&inputDir, "input", CNST[DEFAULT_NOAA_DIR], "directory containing PNG images or directories of PNG images")
	flag.StringVar(&outputDir, "output", ".", "directory for result PNG images")
	flag.Parse()

	if slices.Contains([]string{CNST[DEFAULT_NOAA_DIR], ""}, inputDir) {
		// No argument – trigger download
		var err error
		inputDir, err = getImages(inputDir)
		if err != nil {
			log.Fatalf("failed to download images: %v", err)
		}
	}

	err := filepath.WalkDir(inputDir, func(path string, dir fs.DirEntry, err error) error {
		if err != nil {
			return err // permission errors etc.
		}
		if !dir.IsDir() {
			return nil
		}

		// Check whether this directory has subdirectories
		hasSubdirs := false
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}

		for _, e := range entries {
			if e.IsDir() {
				hasSubdirs = true
				break
			}
		}
		// Leaf directory → do the work
		if !hasSubdirs {
			doRender(path, renderMode, outputDir)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("can't walk: %v", err)
	}
}

func doRender(inputDir, renderMode, outputDir string) {
	files, err := filepath.Glob(filepath.Join(inputDir, "*.png"))
	if err != nil || len(files) == 0 {
		return
	}

	var baseW, baseH int
	var imgList []image.Image
	for idx, f := range files {
		fHandle, err := os.Open(f)
		if err != nil {
			log.Fatalf("failed to open %s: %v", f, err)
		}
		img, err := png.Decode(fHandle)
		fHandle.Close()
		if err != nil {
			log.Fatalf("failed to decode %s: %v", f, err)
		}
		if idx == 0 {
			baseW, baseH = img.Bounds().Dx(), img.Bounds().Dy()
		} else {
			if img.Bounds().Dx() != baseW || img.Bounds().Dy() != baseH {
				log.Fatalf("image %s dimensions (%dx%d) differ from base (%dx%d)", f, img.Bounds().Dx(), img.Bounds().Dy(), baseW, baseH)
			}
		}
		imgList = append(imgList, img)
	}

	// Prepare result image with RGBA to allow alpha blending.
	result := image.NewRGBA(image.Rect(0, 0, baseW, baseH))

	// For each pixel position compute dominant color.

	for y := 0; y < baseH; y++ {
		for x := 0; x < baseW; x++ {
			freq := make(map[color.RGBA]int)
			for _, img := range imgList {
				c := img.At(x, y)
				r, g, b, a := c.RGBA()
				key := color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
				freq[key]++
			}
			// Find color with max count.
			var maxC color.RGBA = color.RGBA{0xff, 0xff, 0xff, 0xff}
			var max2C color.RGBA = color.RGBA{0xff, 0xff, 0xff, 0xff}
			maxCnt := -1
			max2Cnt := -1
			for col, cnt := range freq {
				if cnt > maxCnt {
					max2Cnt = maxCnt
					max2C = maxC
					maxCnt = cnt
					maxC = col
				}
				if (cnt > max2Cnt) && (maxC != col) {
					max2Cnt = cnt
					max2C = col
				}
			}
			if max2Cnt == -1 {
				max2C = maxC
			}
			if renderMode == CNST[RENDER_SMOOTH] {
				w1 := float64(maxCnt) / float64(maxCnt+max2Cnt)
				w2 := float64(max2Cnt) / float64(maxCnt+max2Cnt)

				r := uint8(float64(maxC.R)*w1 + float64(max2C.R)*w2)
				g := uint8(float64(maxC.G)*w1 + float64(max2C.G)*w2)
				b := uint8(float64(maxC.B)*w1 + float64(max2C.B)*w2)

				result.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
			} else {
				confidence := float64(maxCnt) / float64(len(imgList)) // towards white
				if renderMode == CNST[RENDER_CONFIDENCE] {            // towards 50%
					confidence = 2*confidence - 1
					if confidence < 0 {
						confidence = 0
					}
				}
				scaledR := uint8(float64(maxC.R)*confidence + 255*(1-confidence))
				scaledG := uint8(float64(maxC.G)*confidence + 255*(1-confidence))
				scaledB := uint8(float64(maxC.B)*confidence + 255*(1-confidence))

				result.SetRGBA(x, y, color.RGBA{R: scaledR, G: scaledG, B: scaledB, A: 255})
			}
		}
	}
	outputPath := filepath.Join(filepath.Dir(outputDir), filepath.Base(filepath.Dir(inputDir)))
	err = os.MkdirAll(outputPath, 0755)
	if err != nil {
		log.Fatalf("failed to create output directory: %v", err)
	}

	outFile, err := os.Create(filepath.Join(outputPath, filepath.Base(inputDir)+".png"))
	if err != nil {
		log.Fatalf("failed to create output file: %v", err)
	}
	defer outFile.Close()

	if err := png.Encode(outFile, result); err != nil {
		log.Fatalf("failed to encode PNG: %v", err)
	}

	fmt.Printf("Result image written to %s\n", outputPath)
}
