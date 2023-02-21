package main

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("usage", os.Args[0], "cut_gzip_filename")
	}
	for i, filename := range os.Args {
		if i == 0 {
			continue
		}
		if err := dump(filename); err != nil {
			log.Fatal("unable to dump file", filename, err)
		}
	}
}

func dump(fileorpath string) error {
	file, err := os.Open(fileorpath)
	if err != nil {
		return fmt.Errorf("unable to open %s", err)
	}
	// This returns an *os.FileInfo type
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("unable to get stat %s", err)
	}
	// IsDir is short for fileInfo.Mode().IsDir()
	isDir := fileInfo.IsDir()
	file.Close()
	if isDir {
		// get all file names and sort from timesampt
		filenames := make([]string, 0)
		elsaped := time.Now()
		if err = filepath.Walk(fileorpath, func(subpath string, info fs.FileInfo, err error) error {
			if err != nil {
				log.Println("forgetting", subpath, err)
				return nil
			}
			if info.IsDir() {
				return nil
			}
			// keep only single full bow stats (trick)
			if !strings.HasPrefix(filepath.Base(subpath), "1") {
				return nil
			}
			filenames = append(filenames, subpath)
			if time.Since(elsaped).Seconds() > 10 {
				log.Println("found", len(filenames), "files")
				elsaped = time.Now()
			}
			return nil
		}); err != nil {
			return fmt.Errorf("error walking the path %s", err)
		}
		// sort it now
		sort.Slice(filenames, func(i, j int) bool {
			return filepath.Base(filenames[i]) < filepath.Base(filenames[j])
		})
		log.Println("Found", len(filenames), "files to unzip")
		for i, filename := range filenames {
			if err = dumpfile(filename); err != nil {
				return fmt.Errorf("dump %s %s", filename, err)
			}
			if time.Since(elsaped).Seconds() > 10 {
				log.Println("dumped", i, "files")
				elsaped = time.Now()
			}
		}
	} else {
		// file is not a directory
		return dumpfile(fileorpath)
	}
	return nil
}

func dumpfile(filename string) error {
	// Open and write the gzip file.
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("could not open file %s", err)
	}
	defer file.Close()

	// Create new reader to decompress gzip.
	reader, err := NewReader(file)
	if err != nil {
		return fmt.Errorf("could not create gzip reader %s", err)
	}

	// create buffer
	// large enough for one BoxState
	b := make([]byte, 65536*2)
	for {
		// read content to buffer
		readTotal, err := reader.Read(b)
		// print content from buffer
		if err == nil || err == io.EOF {
			fmt.Print(string(b[:readTotal]))
		}
		if err != nil {
			if err != io.EOF {
				log.Println(filename, err)
			}
			break
		}
	}
	return nil
}
