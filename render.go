package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"text/template"
	"time"
)

//go:embed template.html
var templateFS embed.FS

type MonthData struct {
	Month        string
	DisplayMonth string
	Files        map[string]map[string]string // variable -> mode -> filename
}

func GenerateStaticForecastPage(outputDir string) error {
	files, err := os.ReadDir(outputDir)
	if err != nil {
		return err
	}

	re := regexp.MustCompile(`^(\d{6})_([^_]+)_([^_]+)_([^_]+)\.png$`)
	months := make(map[string]*MonthData)

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		// filter for file names matching calculated weather predictions pattern
		matches := re.FindStringSubmatch(f.Name())
		if matches == nil {
			continue
		}

		// read variables out from files again,
		// will allow to incorporate historically computed files as well,
		// in addition to currently computed ones
		month := matches[1]
		variable := matches[3]
		mode := matches[4]

		// Map short variable names to verbal names for display
		displayVariable := variable
		switch variable {
		case "Prec":
			displayVariable = "precipitation"
		case "T2m":
			displayVariable = "air temperature"
		}

		// build up logical structure for files
		if _, exists := months[month]; !exists {
			displayMonth := month
			if t, err := time.Parse("200601", month); err == nil {
				displayMonth = t.Format("Jan 2006")
			}
			months[month] = &MonthData{
				Month:        month,
				DisplayMonth: displayMonth,
				Files:        make(map[string]map[string]string),
			}
		}

		if _, exists := months[month].Files[displayVariable]; !exists {
			months[month].Files[displayVariable] = make(map[string]string)
		}

		months[month].Files[displayVariable][mode] = f.Name()
	}

	if len(months) == 0 {
		return fmt.Errorf("no forecast PNG files found")
	}

	var monthList []string
	for m := range months {
		monthList = append(monthList, m)
	}
	sort.Strings(monthList)

	var data []*MonthData
	for _, m := range monthList {
		data = append(data, months[m])
	}

	tmpl, err := template.ParseFS(templateFS, "template.html")
	if err != nil {
		return err
	}

	outFile := filepath.Join(outputDir, "index.html")
	file, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, data)
}
