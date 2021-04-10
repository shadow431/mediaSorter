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

//TODO
/*
  Create a File struct?
    fileName
    sourceDir
	destDir
	make
	model
	serial
	dateTime

	generate struck for files in setDestPath.
	rename to setDestPath to procImage?
	Use struck all the way through then I'll have all the fields I want tell the end

*/

func main() {
	sourceDir := flag.String("sourceDir", "/tmp", "where to read the images from")
	destDir := flag.String("destDir", "/tmp", "where to read the images to")
	var files []string
	flag.Parse()
	*sourceDir = strings.TrimSuffix(*sourceDir, "/")
	*destDir = strings.TrimSuffix(*destDir, "/")

	e := filepath.Walk(*sourceDir, procDir(&files))
	if e != nil {
		log.Fatal(e)
	}

	et, err := exiftool.NewExiftool()
	/*
		et, err := exiftool.NewExiftool(func(s *exiftool.Exiftool) error {
			s.extraInitArgs = append(s.extraInitArgs, "-ee")
			return nil
		})
	*/
	if err != nil {
		fmt.Printf("Error initializing %v\n", err)
	}

	fileCount := len(files)
	fmt.Printf("%v files Total\n", fileCount)

	for _, file := range files {

		metaData := getExifInfo(et, file)
		imgPath := setDestPath(metaData)
		fileSplit := strings.Split(file, "/")
		filename := fileSplit[len(fileSplit)-1]

		destPath := *destDir + imgPath
		fullDstPath := destPath + filename

		if file != fullDstPath {
			fmt.Printf("%v ==> %v\n", file, fullDstPath)
			mvErr := mvMedia(file, destPath, filename)
			if mvErr != nil {
				fmt.Println(mvErr)
				os.Exit(1)
			}
		}
	}

	defer et.Close()

}

//TODO: Make this works
/*
func setExifArgs() exiftool.Exiftool {
	return func(e *exiftool.Exiftool) {
		e.extraInitArgs := []string{"-ee"}
	}
}*/

//TODO:  what if those fields dont exist?
func setDestPath(metaData map[string]interface{}) string {
	var newPath string
	strDateTime := fmt.Sprintf("%v", metaData["DateTimeOriginal"])
	arrDate := strings.Fields(strDateTime)
	newPath = fmt.Sprintf("/%v/", metaData["Model"])
	newPath += strings.ReplaceAll(arrDate[0], ":", "/")
	newPath += "/"

	return newPath
}

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

func mvMedia(source string, dest string, filename string) error {
	var dirMode os.FileMode
	destSplit := strings.Split(dest, "/")
	fullPath := dest + filename
	var di fs.FileInfo
	var err error

	dirMode, err = getParentMode(destSplit)
	if err != nil {
		return err
	}
	di, err = makeDir(dest, dirMode)
	if err != nil {
		return err
	}
	switch mode := di.Mode(); {
	case mode.IsDir():
		err := mvFile(source, fullPath)
		if err != nil {
			return err
		}
	case mode.IsRegular():
		fmt.Printf("%v is a file.... WTF!!!!", dest)
		return err
	}

	return nil
}

//WIP
func mvFile(source string, fullPath string) error {
	var serialAdded bool

	moved := false
	for !moved {
		if _, err := os.Stat(fullPath); errors.Is(err, fs.ErrNotExist) {
			os.Rename(source, fullPath)
			moved = true
		} else {
			srcHash := getHash(source)
			dstHash := getHash(fullPath)
			if bytes.Equal(srcHash, dstHash) && !serialAdded {
				fmt.Printf("Both files are the same")
				break
			}
			if serialAdded {
				fmt.Printf("FUCK!!! the files match even after adding the serial")
			}
			//Get Cam serial Number, add to start of file name
			//set Serial as added

		}

	}
	//TODO:
	/*
		         - var serialAdded boolean
			     - while not created:
				 -   if !exists:
				 -     os.Rename()
				 -   else:
				 -     getHashes:
				 -	    if Hash == Hash && !serialAdded:
				 - 	      dont care they are the same
				 -  	  break
						if serialAdded:
						  FUCK
					  getCamSerial:
					    add serial To file Name
						serialAdded = true
	*/

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

func makeDir(dest string, mode fs.FileMode) (fs.FileInfo, error) {
	var di fs.FileInfo
	var err error
	for di == nil {
		di, err = os.Stat(dest)
		if err != nil {
			err := os.MkdirAll(dest, mode)
			if err != nil {
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
