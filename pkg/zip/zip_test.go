package zip

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"testing"
)

func TestIncludesStarStar(t *testing.T) {
	files := BuildFileList(".", []string{"**"}, []string{})
	if len(files) != 2 {
		t.Fatal("length", len(files))
	}
}

func TestIncludesStarDotGo(t *testing.T) {
	files := BuildFileList(".", []string{"*.go"}, []string{})
	if len(files) != 2 {
		t.Fatal("length", len(files))
	}
}

func TestIncludesTestDotGo(t *testing.T) {
	files := BuildFileList(".", []string{"*test.go"}, []string{})
	if len(files) != 1 {
		t.Fatal("length", len(files))
	}
}
func TestMultipleIncludes(t *testing.T) {
	files := BuildFileList(".", []string{"*test.go", "zip.go"}, []string{})
	if len(files) != 2 {
		t.Fatal("length", len(files))
	}
}
func TestIncludeDepth(t *testing.T) {
	files := BuildFileList("../", []string{"**/zip*"}, []string{})
	if len(files) != 2 {
		t.Fatal("length", len(files))
	}
}
func TestExcludes(t *testing.T) {
	files := BuildFileList("../", []string{"**/zip*"}, []string{"**/*test*"})
	if len(files) != 1 {
		t.Fatal("length", len(files))
	}
}
func TestExcludesStarStar(t *testing.T) {
	files := BuildFileList("../", []string{"**/zip*"}, []string{"**"})
	if len(files) != 0 {
		t.Fatal("length", len(files))
	}
}

func TestMultipleExcludes(t *testing.T) {
	files := BuildFileList("../", []string{"**/zip*"}, []string{"**/*test*", "**/zip.go"})
	if len(files) != 0 {
		t.Fatal("length", len(files))
	}
}

func TestCreateZip(t *testing.T) {
	// Open the file for reading
	thisFile, err := os.Open("zip_test.go")
	if err != nil {
		t.Fatal("Error opening file", err)
	}
	defer thisFile.Close()

	// Read the file contents
	expected, err := io.ReadAll(thisFile)
	if err != nil {
		t.Fatal("Error reading file", err)
	}
	zipData := Create("../", []string{"**/zip*"}, []string{"**/zip.go"}, "", false)
	r, err := zip.NewReader(bytes.NewReader(zipData.Bytes()), int64(zipData.Len()))
	if err != nil {
		t.Fatal("Error opening zip archive", err)
	}
	var fileNames []string

	// Read the contents of each file in the zip archive
	for _, f := range r.File {
		fileNames = append(fileNames, f.Name)
		rc, err := f.Open()
		if err != nil {
			t.Fatal("Error opening file", f.Name, err)
		}
		defer rc.Close()

		actual, err := io.ReadAll(rc)
		if err != nil {
			t.Fatal("Error reading file", f.Name, err)
		}

		if !bytes.Equal(expected, actual) {
			t.Fatal("expected", expected, "actual", actual)
		}
	}
	if len(fileNames) != 1 {
		t.Fatal("length", len(fileNames))
	}
}
