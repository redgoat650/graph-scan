package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

func readImage(filePath string) (image.Image, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	image, _, err := image.Decode(f)
	return image, err
}

func writeImage(img image.Image, filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}

	defer f.Close()

	return png.Encode(f, img)
}

func isCloseTo(v uint32, ref uint32, ff uint32) bool {
	return uint32(math.Abs(float64(v)-float64(ref))) < ff
}

// func checkProximity(img image.Image, x, y, rb, gb, bb, ff int) bool {
// 	deltas := []int{-2, -1, 1, 2}

// 	for _, dX := range deltas {
// 		px := img.At(x+dX, y)
// 		r, g, b, _ := px.RGBA()
// 		if filterColor(int(r), int(g), int(b), rb, gb, bb, ff) {
// 			// Found a similar pixel in the x direction
// 			return true
// 		}
// 	}
// 	for _, dY := range deltas {
// 		py := img.At(x, y+dY)
// 		r, g, b, _ := py.RGBA()
// 		if filterColor(int(r), int(g), int(b), rb, gb, bb, ff) {
// 			// Found a similar pixel in the y direction
// 			return true
// 		}
// 	}

// 	return false
// }

func filterColor(c, baseline color.Color, ff uint32) bool {
	r, g, b, _ := c.RGBA()
	rb, gb, bb, _ := baseline.RGBA()

	return isCloseTo(r, rb, ff) &&
		isCloseTo(g, gb, ff) &&
		isCloseTo(b, bb, ff)
}

func countPxByColor(img image.Image, c color.Color) (cnt int) {
	rb, gb, bb, _ := c.RGBA()
	bnds := img.Bounds()
	for x := bnds.Min.X; x < bnds.Max.X-1; x++ {
		for y := bnds.Min.Y; y < bnds.Max.Y-1; y++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r == rb && g == gb && b == bb {
				cnt++
			}
		}
	}

	return cnt
}

const (
	imageBottomOffsetPx = 1
)

func main() {
	img, err := readImage("./test2.png")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	bnds := img.Bounds()

	redFilter := image.NewRGBA(bnds)
	notRedFilter := image.NewRGBA(bnds)
	redBlueFilter := image.NewRGBA(bnds)
	notRedBlueFilter := image.NewRGBA(bnds)

	// var blueColor, redColor *color.Color

	// Filter RED
	redColor := color.RGBA{0xf1, 0x75, 0x72, 0xff}
	redFudgeFactor := uint32(0x3c00)

	// Filter BLUE
	blueColor := color.RGBA{0x00, 0xC0, 0xC5, 0xff}
	blueFudgeFactor := uint32(0x4000)

	for x := bnds.Min.X; x < bnds.Max.X; x++ {
		hitBlue := false
		hitRed := false

		for y := bnds.Min.Y; y < bnds.Max.Y-imageBottomOffsetPx; y++ {
			c := img.At(x, y)

			if !hitBlue {
				// Did we hit blue?
				hitBlue = filterColor(c, blueColor, blueFudgeFactor)
			}

			if !hitRed {
				// Did we hit red?
				hitRed = filterColor(c, redColor, redFudgeFactor)
			}

			if hitRed {
				redFilter.Set(x, y, redColor)
				redBlueFilter.Set(x, y, redColor)
			} else {
				notRedFilter.Set(x, y, c)
			}

			if hitBlue {
				if !hitRed {
					redBlueFilter.Set(x, y, blueColor)
				}
			} else if !hitRed {
				notRedBlueFilter.Set(x, y, c)
			}
		}
	}

	err = writeImage(redFilter, "./red.png")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = writeImage(notRedFilter, "./notRed.png")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = writeImage(redBlueFilter, "./redBlue.png")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = writeImage(notRedBlueFilter, "./notRedBlue.png")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	rPx := countPxByColor(redFilter, redColor)
	bPx := countPxByColor(redBlueFilter, blueColor)
	rbPx := rPx + bPx
	totalPx := bnds.Dx() * (bnds.Dy() - imageBottomOffsetPx)

	fmt.Println(bnds.Dx(), bnds.Dy())
	fmt.Println(rPx, "/", totalPx)

	hrPerTradingDay := 6.5
	halfHoursPerTradingDay := hrPerTradingDay * 2
	tradingDays := 14.0

	volumeAtFullHeightPerHalfHour := 10000000.0
	volumeForFullImage := volumeAtFullHeightPerHalfHour * halfHoursPerTradingDay * tradingDays

	volumePerPixel := volumeForFullImage / float64(bnds.Dx()*bnds.Dy())

	totalVolumeRed := rPx * int(volumePerPixel)

	totalVolumeBlue := rbPx * int(volumePerPixel)

	fmt.Println("Total volume red", totalVolumeRed/1000000, "M")
	fmt.Println("Total volume blue", totalVolumeBlue/1000000, "M")
}
