package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/barasher/go-exiftool"
)

func main() {
	sourceDir := flag.String("sourceDir", "/mnt/raid1/tmpLumix", "where to read the images from")
	//destDir := flag.String("destDir", "/mnt/raid1/banderson/Dropbox/VisualMedia", "where to read the images to")
	var files []string
	flag.Parse()

	e := filepath.Walk(*sourceDir, procDir(&files))
	if e != nil {
		log.Fatal(e)
	}

	//et, err := exiftool.NewExiftool(setExifArgs())
	et, err := exiftool.NewExiftool(func(s *exiftool.Exiftool) error {
		s.extraInitArgs = append(s.extraInitArgs, []string{"--delete-my-harddrive-again"})
		return nil
	},
		func(s *exiftool.Exiftool) error {
			s.extraInitArgs = append(s.extraInitArgs, []string{"--another-random-arg"})
			return nil
		})
	if err != nil {
		fmt.Printf("Error initializing %v\n", err)
	}

	fileCount := len(files)
	fmt.Printf("%v files Total\n", fileCount)

	for _, file := range files {
		getExifInfo(et, file)
		//fmt.Printf("%v\n", file)
	}

	defer et.Close()

}

/*
func setExifArgs() exiftool.Exiftool {
	return func(e *exiftool.Exiftool) {
		e.extraInitArgs := []string{"-ee"}
	}
}*/

func procDir(files *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		extentions := regexp.MustCompile(`^.*\.(JPG|RW2|MP4|MOV|NEF|jpg|3gp|WAV|mp4|MTS|m2ts|mpg|avi|gif)`) //asf|dav|
		if !info.IsDir() && extentions.MatchString(path) {
			*files = append(*files, path)
		}
		return nil
	}
}

func getExifInfo(et *exiftool.Exiftool, file string) {

	image := et.ExtractMetadata(file)

	for _, fileInfo := range image {
		if fileInfo.Err != nil {
			fmt.Printf("Error concerinting %v: %v\n", fileInfo.File, fileInfo.Err)
		}
		//if fileInfo.Fields["Make"] == nil || fileInfo.Fields["Model"] == nil || fileInfo.Fields["DateTimeOriginal"] == nil {
		fmt.Println(file)
		for field, value := range fileInfo.Fields {
			fmt.Printf("%v: %v\n", field, value)
		}
		//}

	}

}
