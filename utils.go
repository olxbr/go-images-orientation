package orientation

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"strings"

	"github.com/disintegration/imaging" // go get -u github.com/disintegration/imaging
	"github.com/rwcarlsen/goexif/exif"  // go get -u github.com/rwcarlsen/goexif/exif
	"github.com/sirupsen/logrus"        // go get -u github.com/sirupsen/logrus

	"bytes"
)

// ReadImage makes a copy of image (jpg,png or gif) and applies
// all necessary operation to reverse its orientation to 1 or none
// The result is a image with corrected orientation and without
// exif data.

func ReadImage(imgBody []byte, logger *logrus.Entry, imageId string) (imagebody []byte) {
	imgBodyReader := bytes.NewReader(imgBody)

	// deal with exif
	img, imgExtension, err := image.Decode(imgBodyReader)
	if err != nil {
		logger.Errorf("error when decoding image, %s \n Image ID: %s", err.Error(), imageId)
	}
	if imgExtension != "png" && imgExtension != "jpg" && imgExtension != "jpeg" && imgExtension != "gif" {
		logger.Infof("image type %s has no exif to check for orientation \n Image ID: %s", imgExtension, imageId)
		return imgBody
	}
	x := GetExif(imgBody, logger, imageId)

	if x != nil {
		orient, _ := x.Get(exif.Orientation)
		if orient != nil {
			if orient.String() == "1" || orient.String() == "0" {
				logger.Infof("image already has correct orientation %s, no further exif manipulation is needed \n Image ID: %s", orient, imageId)
				return imgBody
			}
			logger.Infof("image had orientation %s \n Image ID: %s", orient.String(), imageId)

			img = reverseOrientation(img, orient.String(), logger, imageId)
			switch imgExtension {
			case "png":
				buffer := new(bytes.Buffer)
				err := png.Encode(buffer, img)
				if err != nil {
					logger.Errorf("error while encoding corrected image: %s \n Image ID: %s", err, imageId)
				}
				imgBody = buffer.Bytes()
				return imgBody
			case "gif":
				buffer := new(bytes.Buffer)
				err := gif.Encode(buffer, img, nil)
				if err != nil {
					logger.Errorf("error while encoding corrected image: %s \n Image ID: %s", err, imageId)
				}
				imgBody = buffer.Bytes()
				return imgBody
			case "jpeg", "jpg":
				buffer := new(bytes.Buffer)
				err := jpeg.Encode(buffer, img, nil)
				if err != nil {
					logger.Errorf("error while encoding corrected image: %s \n Image ID: %s", err, imageId)
				}
				imgBody = buffer.Bytes()
				return imgBody
			}
		} else {
			logger.Infof("image has no orientation data - implying 1, no further exif manipulation is needed \n Image ID: %s", imageId)
			return imgBody
		}
	}
	return imgBody
}

// reverseOrientation amply`s what ever operation is necessary to transform given orientation
// to the orientation 1
func reverseOrientation(img image.Image, o string, logger *logrus.Entry, imageId string) *image.NRGBA {
	switch o {
	case "2":
		return imaging.FlipV(img)
	case "3":
		return imaging.Rotate180(img)
	case "4":
		return imaging.Rotate180(imaging.FlipV(img))
	case "5":
		return imaging.Rotate270(imaging.FlipV(img))
	case "6":
		return imaging.Rotate270(img)
	case "7":
		return imaging.Rotate90(imaging.FlipV(img))
	case "8":
		return imaging.Rotate90(img)
	}
	logger.Errorf("unknown orientation: %s, when attempting to rotate, expected 2-8 \n Image ID: %s", o, imageId)
	return imaging.Clone(img)
}

func GetExif(imgBody []byte, logger *logrus.Entry, imageId string) *exif.Exif {
	//dont know why, but exif needs this "hack" to decode properly sometimes
	imgBodyStringReader := strings.NewReader(string(imgBody))
	x, err := exif.Decode(imgBodyStringReader)
	if fmt.Sprint(err) == "EOF" {
		logger.Infof("Image is clean of exif data")
		return nil
	}
	if err != nil {
		if x == nil {
			logger.Infof("Unable to read exif data, might imply that orientation is correct and no manipulation is needed, error found: %s \n Image ID: %s", err, imageId)
			return nil
		}
		logger.Errorf("failed reading exif data: %s \n Image ID: %s", err.Error(), imageId)
	}
	return x
}

func GetExifOrientation(exifData *exif.Exif) (string, error) {
	if exifData == nil {
		return "none", nil
	}
	o, err := exifData.Get(exif.Orientation)
	if o == nil {
		return "none", nil
	}
	return o.String(), err
}
