package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

func main() {
	var (
		day time.Time
		err error

		ptrPath = flag.String("path", "", "file or path to Magiline {epoch}.json.gz files")
		ptrDay  = flag.String(
			"day", "", "filter to keep only files {epoch}.json[.gz] where d(utc) <= epoch < d+1(utc) ")
	)
	flag.Parse()
	if *ptrPath == "" {
		fmt.Println("Please provide a path where {epoch}.json.gz can be found")
		return
	}
	if *ptrDay != "" {
		if day, err = time.ParseInLocation("20060102", *ptrDay, time.UTC); err != nil {
			log.Fatal("unabme to parse day", *ptrDay, err)
		}
	}

	if err := Dump(*ptrPath, day); err != nil {
		log.Fatal("unable to dump file", *ptrPath, err)
	}
}

const Suffix = ".json.gz"

func Dump(fileorpath string, day time.Time) error {
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
	var epochMin, epochMax int
	if !day.IsZero() {
		epochMin = int(day.Unix())
		epochMax = int(day.Add(24 * time.Hour).Unix())
	}
	if isDir {
		// get all file names and sort from timestamp
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
			base := filepath.Base(subpath)
			if !strings.HasSuffix(base, Suffix) {
				return nil
			}
			epoch, eerr := strconv.Atoi(base[0 : len(base)-len(Suffix)])
			if eerr != nil {
				// not an {epoch}.§json.gzip file
				return nil
			}
			if epochMin != 0 && (epoch < epochMin || epoch >= epochMax) {
				// not the right day
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
		// sort it now using epoch
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
