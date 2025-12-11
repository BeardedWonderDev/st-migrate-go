package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
)

func main() {
	const size = 256
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	bg := color.RGBA{R: 11, G: 30, B: 45, A: 255}  // deep slate
	accent := color.RGBA{R: 0, G: 185, B: 156, A: 255} // teal accent
	light := color.RGBA{R: 233, G: 245, B: 255, A: 255}

	// background
	draw.Draw(img, img.Bounds(), &image.Uniform{bg}, image.Point{}, draw.Src)

	// circle accent
	drawCircle(img, size/2, size/2, 96, color.RGBA{R: 0, G: 95, B: 115, A: 255})
	drawCircle(img, size/2, size/2, 80, color.RGBA{R: 0, G: 140, B: 140, A: 255})

	// block letter S
	draw.Draw(img, image.Rect(70, 70, 170, 90), &image.Uniform{light}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(70, 120, 170, 140), &image.Uniform{light}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(70, 170, 170, 190), &image.Uniform{light}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(70, 90, 90, 120), &image.Uniform{light}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(150, 140, 170, 170), &image.Uniform{light}, image.Point{}, draw.Src)

	// block letter T
	draw.Draw(img, image.Rect(180, 70, 240, 90), &image.Uniform{accent}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(204, 90, 216, 190), &image.Uniform{accent}, image.Point{}, draw.Src)

	out, err := os.Create("images/logo.png")
	if err != nil {
		panic(err)
	}
	defer out.Close()
	if err := png.Encode(out, img); err != nil {
		panic(err)
	}
}

// drawCircle fills a circle with the given radius and color.
func drawCircle(img *image.RGBA, cx, cy, r int, c color.Color) {
	for x := -r; x <= r; x++ {
		for y := -r; y <= r; y++ {
			if x*x+y*y <= r*r {
				img.Set(cx+x, cy+y, c)
			}
		}
	}
}
