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
	"log"
	"os"
	"path/filepath"
)

// pixelColor holds a color and its occurrence count.
type pixelColor struct {
	c   color.Color
	cnt int
}

func main() {
	var inputDir string

	flag.StringVar(&inputDir, "input", "examples/26-04", "directory containing PNG images")
	flag.Parse()

	// Allow positional argument to override input directory
	if len(flag.Args()) > 0 {
		inputDir = flag.Args()[0]
	}

	if inputDir == "" {
		log.Fatalf("input directory required")
	}

	// Derive output path from input directory name
	outputPath := filepath.Join(filepath.Dir(inputDir), filepath.Base(inputDir)+".png")

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		log.Fatalf("failed to create output directory: %v", err)
	}

	files, err := filepath.Glob(filepath.Join(inputDir, "*.png"))
	if err != nil || len(files) == 0 {
		log.Fatalf("no PNG files found in %s", inputDir)
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
	totalImages := len(imgList)
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
			var maxC color.RGBA
			maxCnt := -1
			for col, cnt := range freq {
				if cnt > maxCnt {
					maxCnt = cnt
					maxC = col
				}
			}
			// Compute dominance as percentage of dominance.
			dominance := float64(maxCnt) / float64(totalImages)
			scaledR := uint8(float64(maxC.R)*dominance + 255*(1-dominance))
			scaledG := uint8(float64(maxC.G)*dominance + 255*(1-dominance))
			scaledB := uint8(float64(maxC.B)*dominance + 255*(1-dominance))

			result.SetRGBA(x, y, color.RGBA{R: scaledR, G: scaledG, B: scaledB, A: 255})
		}
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("failed to create output file: %v", err)
	}
	defer outFile.Close()

	if err := png.Encode(outFile, result); err != nil {
		log.Fatalf("failed to encode PNG: %v", err)
	}

	fmt.Printf("Result image written to %s\n", outputPath)
}
