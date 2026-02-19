# Travel Weather Prediction

Travel Weather Prediction generates images for a better understanding to weather trends of an upcoming travel season. For this, it overlays monthly forecast images provided by NOAA. 

The maps right now are limited to the European region. Predictions are available for up half a year in advance of the current date. 

Purpose of this program is, to provide a consolidated single view across the many computational results provided by NOAA, for a better overview and understanding of expected climate. The maps should not be understood as a precise weather report, rather an indication to upcoming climate conditions for broader regions.

Travel Weather Prediction offers three methods, on how the results of NOAA are consolidated, known as "renderMode":

- white : a color indicates to most dominent across all sample images, respectively to occuranc is blended towards white. E.g. a dot being red in 75% of the images is selected as dominant color, but given with 75% opacity towards white. 

- confidence : the same as white, but with the blending performed to a 50% baseline. Meaning: there has to be more than 50% of a color to be dominant at a given dot, and the opacity will be more faint. E.g. a dot being red in 75% of the images is selected as dominant color, but given with 50% opacity towards white (which is normalized to a 50% level).

- smooth : the same as white, but the dominant color is blended towards the 2nd most dominant color. This usually generates more saturated images to look at.

## Usage

Travel Weather Prediction requires go installed on a computer.

go run https://github.com/EverydayRoadster/TravelWeatherPrediction

If no arguments are specified, the program will download images from NOAA into folder ".noaa" and store computed images to current directory.

Program arguments available:

- input $inputdirectory : input directory from where to read images from. For any other directories than ".noaa" (default), no image download will occure. Images will be processed for any subdirectory where there is no more subdirectory (leaf directory only).
- output $outputdirectory : Where the result image files stored to. Defaults to .
- renderMode [white|confidence|smooth]: render method (see above for details).
