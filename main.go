package main

import (
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"net"
	"os"
	"time"

	"github.com/disintegration/gift"
	"github.com/kbinani/screenshot"
	"github.com/pwaller/go-hexcolor"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/inconsolata"
	"golang.org/x/image/math/fixed"
)

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func main() {
	n := screenshot.NumActiveDisplays()

	for i := 0; i < n; i++ {
		bounds := screenshot.GetDisplayBounds(i)

		img, err := screenshot.CaptureRect(bounds)
		if err != nil {
			panic(err)
		}

		rawTime := time.Now()

		myIP := GetOutboundIP()

		curTime := rawTime.Unix()
		timeText := rawTime.Local()

		watermarkText := fmt.Sprintf("%s - %s", timeText, myIP.To4().String())
		rawFilepath := fmt.Sprintf("%d_%d_%dx%d.png", curTime, i, bounds.Dx(), bounds.Dy())
		format := "png"
		watermark := createWatermark(watermarkText, 2.0, parseColor("#FF0000FF"))

		sourceBounds := img.Bounds()
		watermarkBounds := watermark.Bounds()
		markedImage := image.NewRGBA(sourceBounds)
		draw.Draw(markedImage, sourceBounds, img, image.ZP, draw.Src)

		var offset image.Point
		for offset.X = watermarkBounds.Max.X / -2; offset.X < sourceBounds.Max.X; offset.X += watermarkBounds.Max.X {
			for offset.Y = watermarkBounds.Max.Y / -2; offset.Y < sourceBounds.Max.Y; offset.Y += watermarkBounds.Max.Y {
				draw.Draw(markedImage, watermarkBounds.Add(offset), watermark, image.ZP, draw.Over)
			}
		}

		file, _ := os.Create(rawFilepath)
		defer file.Close()

		switch format {
		case "png":
			err = png.Encode(file, markedImage)
		case "gif":
			err = gif.Encode(file, markedImage, &gif.Options{NumColors: 265})
		case "jpeg":
			err = jpeg.Encode(file, markedImage, &jpeg.Options{Quality: jpeg.DefaultQuality})
		default:
			log.Fatalf("unknown format %s", format)
		}
		if err != nil {
			log.Fatalf("unable to encode image: %s", err)
		}

		fmt.Printf("[*] Screenshot Written: %s\n", rawFilepath)
	}
}

func parseColor(str string) color.Color {
	r, g, b, a := hexcolor.HexToRGBA(hexcolor.Hex(str))
	return color.RGBA{
		A: a,
		R: r,
		G: g,
		B: b,
	}
}

func createWatermark(text string, scale float64, textColor color.Color) image.Image {
	var padding float64 = 2
	w := 8 * (float64(len(text)) + (padding * 2))
	h := 16 * padding
	img := image.NewRGBA(image.Rect(0, 0, int(w), int(h)))
	point := fixed.Point26_6{fixed.Int26_6(64 * padding), fixed.Int26_6(h * 64)}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(textColor),
		Face: inconsolata.Regular8x16,
		Dot:  point,
	}
	d.DrawString(text)

	bounds := img.Bounds()
	scaled := image.NewRGBA(image.Rect(0, 0, int(float64(bounds.Max.X)*scale), int(float64(bounds.Max.Y)*scale)))
	draw.BiLinear.Scale(scaled, scaled.Bounds(), img, bounds, draw.Src, nil)

	g := gift.New(
		gift.Rotate(45, color.Transparent, gift.CubicInterpolation),
	)
	rot := image.NewNRGBA(g.Bounds(scaled.Bounds()))
	g.Draw(rot, scaled)
	return rot
}
