package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/disintegration/imaging"
	"image"
	"image/color"
	"image/jpeg"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

func main() {
	wg := &sync.WaitGroup{}
	file, err := os.Open("Aadhar_12.jpg")
	content, _ := ioutil.ReadAll(bufio.NewReader(file))
	srcImage, err := jpeg.Decode(bytes.NewReader(content))
	if err != nil {
		fmt.Println(err)
	}

	for i := 0.0; i <= 30; i += 5 {
		wg.Add(1)
		go saveRotatedImage(wg, srcImage, i)
	}

	wg.Wait()
}

func saveRotatedImage(wg *sync.WaitGroup, srcImage image.Image, angle float64) {
	defer wg.Done()
	newImage := imaging.Rotate(srcImage, angle, color.White)
	newImageFile, _ := os.Create(fmt.Sprintf("aadhar_tilt_%v.jpg", angle))
	defer newImageFile.Close()
	if err := jpeg.Encode(newImageFile, newImage, &jpeg.Options{100}); err != nil {
		log.Println(err)
	}
}
