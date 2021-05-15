package main

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/barasher/go-exiftool"
)

type MediaFile struct {
	fileName string      //DSC005.JPG
	source   string      // /tmp/DSC005.JPG
	destPath string      // /final/dest/
	make     interface{} // 'Lumix'
	model    interface{} // 'GH5S'
	serial   interface{} // 'CAMSERIALHERE'
	dateTime interface{} // 'Date and Time of Photo'
}

func main() {
	sourceDir := flag.String("sourceDir", "/tmp", "where to read the images from")
	destDir := flag.String("destDir", "/tmp", "where to read the images to")
	info := flag.Bool("info", false, "only print info about the files")
	metadata := flag.Bool("metadata", false, "only print metadata about the files")
	dry := flag.Bool("dry-run", false, "Will go through the motions but not create the directories or move the files")
	var files []string
	flag.Parse()
	*sourceDir = strings.TrimSuffix(*sourceDir, "/")
	*destDir = strings.TrimSuffix(*destDir, "/")

	e := filepath.Walk(*sourceDir, procDir(&files))
	if e != nil {
		log.Fatal(e)
	}

	et, err := exiftool.NewExiftool(exiftool.ExtractEmbedded())

	/*
		et, err := exiftool.NewExiftool(func(s *exiftool.Exiftool) error {
			s.extraInitArgs = append(s.extraInitArgs, "-ee")
			return nil
		})
	*/
	if err != nil {
		fmt.Printf("Error initializing %v\n", err)
	}
	defer et.Close()

	fileCount := len(files)
	fmt.Printf("%v files Total\n", fileCount)

	for _, file := range files {
		if !*metadata {
			fileInfo := setupFileInfo(et, file, *destDir)

			fullDstPath := fileInfo.destPath + fileInfo.fileName

			if !*info {
				if file != fullDstPath {
					fmt.Printf("%v ==> %v\n", file, fullDstPath)
					mvErr := mvMedia(fileInfo, *dry)
					if mvErr != nil {
						fmt.Println(mvErr)
						os.Exit(1)
					}
				}
			} else {
				fmt.Println(fileInfo)
			}
		} else {
			for k, v := range getExifInfo(et, file) {
				fmt.Printf("%v => %v\n", k, v)
			}
		}
	}
}

//TODO: can this be more clever?
func setupFileInfo(et *exiftool.Exiftool, file string, destDir string) MediaFile {

	metaData := getExifInfo(et, file)
	fileSplit := strings.Split(file, "/")
	filename := fileSplit[len(fileSplit)-1]
	var serialNumber interface{}
	var model interface{}
	var taken interface{}

	if metaData["Model"] != nil && metaData["Model"] != "" { //GH5(s), Nikon D5300
		model = metaData["Model"]
	} else if metaData["Originator"] != nil { //ZoomH6
		model = metaData["Originator"]
	} else if metaData["OtherSerialNumber"] != nil { //GoPro Hero4Silver
		model = metaData["OtherSerialNumber"]
	}
	if metaData["DateTimeOriginal"] != nil { //GH5(s), Nikon D5300, ZoomH6
		taken = metaData["DateTimeOriginal"]
	} else if metaData["CreateDate"] != nil { //GoPro Hero4Silver
		taken = metaData["CreateDate"]
	}

	if metaData["SerialNumber"] != nil { //GH5(s), Nikon D5300, GoPro Hero4Silver
		serialNumber = metaData["SerialNumber"] //ToDo: Probably needs some work due to nikon model comming back as scientific notation
	} else if metaData["InternalSerialNumber"] != nil { //ZoomH6
		serialNumber = metaData["InternalSerialNumber"]
	}
	imgPath := setDestPath(model, taken)
	destPath := destDir + imgPath

	fileInfo := MediaFile{fileName: filename, source: file, destPath: destPath, make: metaData["Make"], model: model, serial: serialNumber, dateTime: taken}
	return fileInfo
}

//TODO:  what if those fields dont exist?
func setDestPath(model interface{}, taken interface{}) string {
	var newPath string
	strDateTime := fmt.Sprintf("%v", taken)
	arrDate := strings.Fields(strDateTime)
	newPath = fmt.Sprintf("/%v/", model)
	newPath += strings.ReplaceAll(arrDate[0], ":", "/")
	newPath += "/"

	return newPath
}

func procDir(files *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		extentions := regexp.MustCompile(`^.*\.(JPG|RW2|MP4|MOV|NEF|jpg|3gp|WAV|mp4|MTS|m2ts|mpg|avi|gif|LRV)`) //asf|dav|
		if !info.IsDir() && extentions.MatchString(path) {
			*files = append(*files, path)
		}
		return nil
	}
}

func getExifInfo(et *exiftool.Exiftool, file string) map[string]interface{} {
	var fields map[string]interface{}
	//fmt.Printf("%v", file)

	image := et.ExtractMetadata(file)

	for _, fileInfo := range image {
		if fileInfo.Err != nil {
			fmt.Printf("Error concerinting %v: %v\n", fileInfo.File, fileInfo.Err)
		}
		fields = fileInfo.Fields

	}
	return fields
}

func mvMedia(fileInfo MediaFile, dry bool) error {
	var dirMode os.FileMode
	dest := fileInfo.destPath
	destSplit := strings.Split(dest, "/")
	var di fs.FileInfo
	var err error

	dirMode, err = getParentMode(destSplit)
	if err != nil {
		return err
	}

	di, err = makeDir(dest, dirMode, dry)
	if err != nil && !dry {
		return err
	}

	switch mode := di.Mode(); {
	case mode.IsDir():
		err := mvFile(fileInfo, dry)
		if err != nil {
			return err
		}
	case mode.IsRegular():
		fmt.Printf("%v is a file.... WTF!!!!\n", dest)
		return err
	default:
		fmt.Println("dir is none of these")
	}

	return nil
}

func mvFile(fileInfo MediaFile, dry bool) error {
	var serialAdded bool
	source := fileInfo.source

	moved := false
	for !moved {
		fullPath := fileInfo.destPath + fileInfo.fileName
		if _, err := os.Stat(fullPath); errors.Is(err, fs.ErrNotExist) {
			if !dry {
				os.Rename(source, fullPath)
			} else {
				fmt.Printf("DRY-RUN: Would have moved %v => %v\n", source, fullPath)
			}
			moved = true
		} else {
			srcHash := getHash(source)
			dstHash := getHash(fullPath)
			if bytes.Equal(srcHash, dstHash) {
				if serialAdded {
					fmt.Printf("Both files are the same w/ the serial Number\n")
				} else {
					fmt.Printf("Both files are the same w/o the serail Number\n")
				}
				break
			}
			if serialAdded {
				fmt.Printf("FUCK!!! the files Don't match even after adding the serial\n")
				break
			} else {
				fileInfo.fileName = fmt.Sprintf("%v-%v", fileInfo.serial, fileInfo.fileName)
				serialAdded = true
			}
		}
	}
	return nil
}

func getHash(file string) []byte {
	hasher := sha256.New()
	f, err := os.Open(file)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	defer f.Close()
	if _, err := io.Copy(hasher, f); err != nil {
		fmt.Printf("%v\n", err)
	}
	return hasher.Sum(nil)
}

func makeDir(dest string, mode fs.FileMode, dry bool) (fs.FileInfo, error) {
	var di fs.FileInfo
	var err error
	for di == nil {
		di, err = os.Stat(dest)
		if err != nil {
			if !dry {
				err := os.MkdirAll(dest, mode)
				if err != nil {
					return di, err
				}
			} else {
				fmt.Printf("Dry-Run: Would have created %v\n", dest)
				return di, err
			}
		}
	}
	return di, err
}

func getParentMode(destSlice []string) (fs.FileMode, error) {
	forPath := "/"
	var dirMode os.FileMode
	var err error
	for _, path := range destSlice {
		forPath += path + "/"
		fi, err := os.Stat(forPath)
		if err != nil {
			break
		}
		dirMode = fi.Mode()
	}
	return dirMode, err
}
