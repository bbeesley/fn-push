package zip

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/bmatcuk/doublestar/v4"
)

func getFsys(path string) fs.FS {
	var fsys fs.FS
	f, getWdErr := os.Getwd()
	if getWdErr != nil {
		log.Fatal(getWdErr)
	}
	if path == "." {
		path = f
	} else {
		path = filepath.Join(f, path)
	}
	fsys = os.DirFS(path)
	return fsys
}

func sliceIndex(limit int, predicate func(i int) bool) int {
	for i := 0; i < limit; i++ {
		if predicate(i) {
			return i
		}
	}
	return -1
}

func remove(s []string, i int) []string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func addSymlinkToZip(zipWriter *zip.Writer, linkPath string, targetPath string) error {
	symlinkContent := targetPath
	symlinkFile := &zip.FileHeader{
		Name:     linkPath,
		Method:   zip.Store,
		Modified: time.Now(),
	}
	symlinkFile.SetMode(0777 | os.ModeSymlink)
	writer, err := zipWriter.CreateHeader(symlinkFile)
	if err != nil {
		return err
	}
	_, err = writer.Write([]byte(symlinkContent))
	if err != nil {
		return err
	}

	return nil
}

// BuildFileList uses a base path along with arrays on include and exclude globs
// to build a list of files which must be added to the archive
func BuildFileList(path string, include []string, exclude []string) []string {
	var matches []string
	var results []string
	fsys := getFsys(path)
	for i := 0; i < len(include); i++ {
		fileSet, error := doublestar.Glob(fsys, include[i], doublestar.WithFilesOnly())
		if error != nil {
			fmt.Printf("Failed to get files for glob: %v\n", include[i])
		}
		matches = append(matches, fileSet...)
	}
	results = append(results, matches...)
	for i := 0; i < len(exclude); i++ {
		for j := range matches {
			exclude, matchError := doublestar.Match(exclude[i], matches[j])
			if matchError != nil {
				fmt.Printf("Error while checking file agains excludes: %v\n", matchError)
			}
			if exclude {
				index := sliceIndex(len(results), func(ix int) bool { return results[ix] == matches[j] })
				if index != -1 {
					fmt.Printf("Removing: %v\n", matches[j])
					results = remove(results, index)
				}
			}
		}
	}
	return results
}

func addFilesToZip(path string, files []string, rootDir string, symlinkNodeModules bool) *bytes.Buffer {
	fsys := getFsys(path)
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	if symlinkNodeModules {
		err := addSymlinkToZip(w, "node_modules", "/opt/nodejs/node_modules")
		if err != nil {
			log.Fatal("Failed to create symlink in zip archive", err)
		}
	}
	for _, file := range files {
		if rootDir != "" {
			file = filepath.Join(rootDir, file)
		}
		f, createWriterError := w.Create(file)
		if createWriterError != nil {
			log.Fatal(createWriterError)
		}
		source, readError := fsys.Open(file)
		if readError != nil {
			log.Fatal(readError)
		}
		_, copyError := io.Copy(f, source)
		if copyError != nil {
			log.Fatal(createWriterError)
		}
	}

	err := w.Close()
	if err != nil {
		log.Fatal(err)
	}
	return buf
}

// Create takes a base path, include and exclude arrays of glob patterns, a rootDir which defines a base path within
// the zip archive, and a boolean to indicate whether it should create a symlink from the lambada layer path to the
// function's node_modules path. It uses these arguments to create a list of files to be added to the archive,
// creates the archive, and returns it as a buffer.
func Create(path string, include []string, exclude []string, rootDir string, symlinkNodeModules bool) *bytes.Buffer {
	fileList := BuildFileList(path, include, exclude)
	zip := addFilesToZip(path, fileList, rootDir, symlinkNodeModules)
	return zip
}
