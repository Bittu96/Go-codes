package main

import (
	"fmt"
	"github.com/gen2brain/go-fitz"
	"image/png"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func panicIfError(err error) {
	if err != nil {
		panic(err)
	}
}

func myConverter(inputRootDir, outputRootDir string) {
	var files []string
	err := filepath.Walk(inputRootDir, func(path string, info os.FileInfo, err error) error {
		fmt.Println(path, info, err)
		if filepath.Ext(path) == ".pdf" {
			files = append(files, path)
		}
		return nil
	})
	panicIfError(err)

	for _, file := range files {
		doc, err := fitz.New(file)
		panicIfError(err)

		folderName := strings.TrimSuffix(path.Base(file), filepath.Ext(path.Base(file)))
		for n := 0; n < doc.NumPage(); n++ {
			img, err := doc.Image(n)
			panicIfError(err)

			err = os.MkdirAll(outputRootDir+folderName, 0755)
			panicIfError(err)

			f, err := os.Create(filepath.Join(outputRootDir+folderName+"/", fmt.Sprintf("%s_PAGE_%0d.jpg", folderName, n+1)))
			panicIfError(err)

			//err = jpeg.Encode(f, img, &jpeg.Options{Quality: jpeg.DefaultQuality})
			err = png.Encode(f, img)

			panicIfError(err)

			err = f.Close()
			panicIfError(err)
		}
	}
}

func main() {
	inputRootDir := "pdfToImageTrials/pdf/"
	outputRootDir := "pdfToImageTrials/img/"
	myConverter(inputRootDir, outputRootDir)
}
