/*
	imgObj struct methods
*/

package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"

	"golang.org/x/image/draw"
	"golang.org/x/image/tiff"
	"golang.org/x/image/webp"
)

type dims struct {
	x int
	y int
}

// primary image struct
type ImgObj struct {
	src_path  string
	src_dims  dims
	src_img   image.Image
	red_dims  dims
	red_img   *image.RGBA
	to_copy   bool
	dest_path string
}

// image constructor
func newImgObj(src_path string, dest_path string) *ImgObj {
	img := new(ImgObj)
	img.src_path = src_path
	img.dest_path = dest_path
	return img
}

func (im *ImgObj) print(str io.Writer) {

	out_str := fmt.Sprintf("src_path: %s\ndest_path: %s\n", im.src_path, im.dest_path)
	out_str += fmt.Sprintf("src_dims: [%d, %d]\nred_dims: [%d, %d]\n", im.src_dims.x, im.src_dims.y, im.red_dims.x, im.red_dims.y)
	out_str += fmt.Sprintf("to_copy: %t\n", im.to_copy)
	out_str += fmt.Sprintf("src_img: %v\n", im.src_img)
	out_str += fmt.Sprintf("red_img: %v", im.red_img)

	if _, e := fmt.Fprintf(str, "%s\n", out_str); e != nil {
		log.Fatal(e)
	}

}

// decode image based on extension, sets src_img
func (im *ImgObj) decode_img() error {
	if im.src_path == "" {
		return fmt.Errorf("no filepath in ImgObj")
	}

	src_path := im.src_path
	img, _ := os.Open(src_path)
	defer img.Close()

	ext := filepath.Ext(src_path)[1:]
	if ext == "" {
		return fmt.Errorf("no extension for file: %s", filepath.Base(src_path))
	}

	// 	switch on extension
	switch ext {
	case "jpeg":
		dec, e := jpeg.Decode(img)
		if e != nil {
			return fmt.Errorf("error decoding file: %s\n%s", src_path, e)
		}
		im.src_img = dec
	case "jpg":
		dec, e := jpeg.Decode(img)
		if e != nil {
			return fmt.Errorf("error decoding file: %s\n%s", src_path, e)
		}
		im.src_img = dec
	case "png":
		dec, e := png.Decode(img)
		if e != nil {
			return fmt.Errorf("error decoding file: %s\n%s", src_path, e)
		}
		im.src_img = dec
	case "tiff":
		dec, e := tiff.Decode(img)
		if e != nil {
			return fmt.Errorf("error decoding file: %s\n%s", src_path, e)
		}
		im.src_img = dec
	case "bmp":
		dec, e := tiff.Decode(img)
		if e != nil {
			return fmt.Errorf("error decoding file: %s\n%s", src_path, e)
		}
		im.src_img = dec
	case "webp":
		dec, e := webp.Decode(img)
		if e != nil {
			return fmt.Errorf("error decoding file: %s\n%s", src_path, e)
		}
		im.src_img = dec
	default:
		return fmt.Errorf("did not recognize format: %s", ext)
	}

	return nil

}

// determines and sets dims for scaled image; or sets to_copy flag to true if in range
func (im *ImgObj) calc_dims(target int) error {
	if im.src_img == nil || im.src_img.Bounds() == image.Rect(0, 0, 0, 0) {
		return fmt.Errorf("no image loaded")
	}

	bnds := im.src_img.Bounds()
	w := bnds.Dx()
	h := bnds.Dy()
	im.src_dims = dims{x: w, y: h}
	tot := w * h

	calc_ratio := math.Sqrt(float64(target) / float64(tot))
	ratio := min(calc_ratio, 1.5)

	if ratio > 0.99 && ratio < 1.025 {
		im.to_copy = true
		return nil
	}
	im.to_copy = false

	im.red_dims = dims{
		x: int(math.Round(float64(w) * ratio)),
		y: int(math.Round(float64(h) * ratio)),
	}

	return nil
}

// resize image using CatmullRom to red_dims
func (im *ImgObj) resize_img() {
	if im.red_dims.x == 0 || im.red_dims.y == 0 {
		panic("yikes, called resize_img with red_dim x or y == 0!")
	}

	// set the expected size
	im.red_img = image.NewRGBA(image.Rect(0, 0, im.red_dims.x, im.red_dims.y))

	// resize:
	draw.CatmullRom.Scale(im.red_img, im.red_img.Rect, im.src_img, im.src_img.Bounds(), draw.Over, nil)

}

// encode resized image to jpeg using passed quality and write to dest_path
func (im *ImgObj) encode_and_write(qual int) error {
	out, err := os.Create(im.dest_path)
	if err != nil {
		return err
	}
	defer out.Close()

	// always make jpegs
	jpeg.Encode(out, im.red_img, &jpeg.Options{Quality: qual})
	return nil

}

// copy original file to dest_path
func (im *ImgObj) copy_file() error {
	srcFile, err := os.Open(im.src_path)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(im.dest_path)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}

	return nil
}

// process image object writing original or reduced image as jpeg to dest_path
func (img *ImgObj) proc_img(a Args, trk Tracking) error {

	if err := img.decode_img(); err != nil {
		return err
	}

	if err := img.calc_dims(a.target); err != nil {
		// sets to_copy property
		return err
	}

	if !img.to_copy { // reduce and write
		img.resize_img()
		if err := img.encode_and_write(a.qual); err != nil {
			return err
		}
		LOCK.Lock()
		trk["scaled_cnt"]++
		LOCK.Unlock()

	} else { // copy as-is
		if err := img.copy_file(); err != nil {
			return err
		}
		LOCK.Lock()
		trk["copied_cnt"]++
		LOCK.Unlock()
	}

	return nil

}
