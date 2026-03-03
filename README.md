# Travel Weather Prediction

**Travel Weather Prediction** generates visual summaries to help interpret seasonal weather trends for upcoming travel periods.  
It processes and overlays monthly forecast maps provided by the U.S. National Oceanic and Atmospheric Administration (NOAA), Climate Prediction Center (CPC), to create consolidated visual insights.

> ⚠️ These maps are not precise weather forecasts.  
> They represent broader climate trends and probabilities across larger regions.

NOAA CPC model NCEP Climate Forecast System Version 2 (CFSv2) was developed at the Environmental Modeling Center at NCEP. It is a fully coupled model representing the interaction between the Earth's atmosphere, oceans, land and seaice. See https://cfs.ncep.noaa.gov/ 

The CFSv2 model provides expected variations in ground temperature and precipitation, as a difference to a standard mean as collected from the years 1990 - 2020. 
See https://journals.ametsoc.org/view/journals/clim/27/6/jcli-d-12-00823.1.xml 

Currently, map coverage is limited to Europe. Forecast data is typically available up to six months in advance from the current date.

---

## Purpose

NOAA provides multiple computational forecast outputs. While scientifically robust, reviewing them individually can make it difficult to quickly grasp overall climate tendencies.

This tool consolidates those results into a single, unified image to:

- Improve readability  
- Highlight dominant trends  
- Provide a quick seasonal overview for travel planning  

---

## Data Handling

- Source images downloaded from NOAA are stored in the user's home directory under `.noaa`. This allows for reuse of downloads even when the command is run from different locations.
- Downloaded images are reused in subsequent runs.
- Obsolete source images are automatically cleaned up when the program runs.

---

## Rendering Modes

The program supports different methods for consolidating NOAA forecast results.  
Each mode defines how dominant colors are determined and blended.

### 1. `all` (default)

This is the default mode. It runs all available rendering methods (`white`, `smooth`, and `confidence`) in sequence and generates a comprehensive interactive page.

---

### 2. `white`

The most dominant color at each pixel across all sample images is selected.

The opacity of that color is blended toward white according to its frequency.

**Example:**  
If a pixel is red in 75% of the images:
- Red is selected as the dominant color.
- It is rendered with 75% opacity blended toward white.

This mode visually expresses both dominance and strength of agreement.

---

### 2. `confidence`

Similar to `white`, but with a stricter dominance threshold and reduced opacity.

- A color must exceed a 50% baseline to be considered dominant.
- Opacity is normalized relative to that 50% threshold.
- The resulting image appears softer and emphasizes stronger consensus areas.

**Example:**  
If red appears in 75% of the images:
- Red is selected as dominant.
- It is rendered with 50% opacity (normalized against the 50% confidence baseline).

This mode emphasizes statistical confidence over visual intensity.

---

### 3. `smooth`

Instead of blending the dominant color toward white, this mode blends it toward the second most dominant color.

This produces:
- More saturated results  
- Smoother transitions  
- Stronger visual contrast  

It is typically more visually striking than the other modes.

---

## Interactive Forecast Page

The program automatically generates an `index.html` file in the output directory. This page provides a user-friendly interface to view the consolidated forecast maps:

- **Month Selector:** Buttons at the top allow you to switch between different forecast months (e.g., "Mar 2026", "Apr 2026").
- **Grid Layout:** For each month, it displays a comparison of different forecast variables (e.g., Temperature, Precipitation) across all available rendering modes.


---

## Installation & Usage

### Requirements

- Go must be installed on your system.

### Run directly without cloning

```bash
go run github.com/EverydayRoadster/TravelWeatherPrediction@latest
```

---

## Command Line Arguments (Detailed)

The following command line flags are available:

### `-renderMode`

Defines how forecast images are consolidated.

**Default:** `all`  
**Allowed values:** `all`, `white`, `smooth`, `confidence`

If an unsupported value is provided, the program prints:

```text
render mode <value> not supported.
```

and exits without processing.

#### Available Modes

| Mode | Description |
|------|------------|
| `all` | **(Default)** Runs all available modes (`white`, `smooth`, `confidence`) in sequence. |
| `white` | Selects the dominant color per pixel and blends it toward white according to frequency. |
| `smooth` | Blends the dominant color toward the second most dominant color, producing more saturated results. |
| `confidence` | Similar to `white`, but requires stronger dominance and applies reduced opacity to emphasize consensus areas. |

Example:

```bash
go run github.com/EverydayRoadster/TravelWeatherPrediction@latest -renderMode smooth
```

---

### `-input`

Specifies a custom directory for stoing PNG images downloaded from NOAA.

**Default:** `~/.noaa` (user's home directory)

Behavior:

- If the default home directory or a custom path is used, NOAA images may download on demand, automatically.

Example:

```bash
go run github.com/EverydayRoadster/TravelWeatherPrediction@latest -input ./forecast-data
```

---

### `-output`

Specifies the directory where generated PNG result images and the interactive index page are written.

**Default:** `.` (current directory)

The program:
- Appends the selected `renderMode` to the output path for PNG images.
- Generates an **interactive `index.html`** in the root of the output directory.

The `index.html` provides a consolidated view of all generated forecasts, allowing you to switch between months and view different rendering variables side-by-side.

Example:

```bash
go run github.com/EverydayRoadster/TravelWeatherPrediction@latest -output ./results
```

This will create:
- PNG images inside `./results/` (organized by month and variable).
- An interactive overview at `./results/index.html`.

---

## Example: Full Command

```bash
go run github.com/EverydayRoadster/TravelWeatherPrediction@latest \
    -output ./results \
    -renderMode all
```

---