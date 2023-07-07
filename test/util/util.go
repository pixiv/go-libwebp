// Package util contains utility code for demosntration of go-libwebp.
package util

import (
	"bufio"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
)

func examplesDir() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic(errors.New("could not retrieve the directory"))
	}
	result, err := filepath.Abs(filepath.Join(filepath.Dir(file), "..", "..", "examples"))
	if err != nil {
		panic(err)
	}
	return result
}

// GetExFilePath returns the path of specified example file.
func GetExFilePath(name string) string {
	path := filepath.Join(examplesDir(), "images", name)
	if _, err := os.Stat(path); err == nil {
		return path
	}
	panic(fmt.Errorf("%v does not exist in any directory which contains in $GOPATH", name))
}

// GetOutFilePath returns the path of specified out file.
func GetOutFilePath(name string) string {
	path := filepath.Join(examplesDir(), "out")
	if _, err := os.Stat(path); err == nil {
		return filepath.Join(path, name)
	}
	panic(fmt.Errorf("out directory does not exist in any directory which contains in $GOPATH"))
}

// OpenFile opens specified example file
func OpenFile(name string) (io io.Reader) {
	io, err := os.Open(GetExFilePath(name))
	if err != nil {
		panic(err)
	}
	return
}

// ReadFile reads and returns data bytes of specified example file.
func ReadFile(name string) (data []byte) {
	data, err := ioutil.ReadFile(GetExFilePath(name))
	if err != nil {
		panic(err)
	}
	return
}

// CreateFile opens specified example file
func CreateFile(name string) (f *os.File) {
	f, err := os.Create(GetOutFilePath(name))
	if err != nil {
		panic(err)
	}
	return
}

// WritePNG encodes and writes image into PNG file.
func WritePNG(img image.Image, name string) {
	f, err := os.Create(GetOutFilePath(name))
	if err != nil {
		panic(err)
	}
	b := bufio.NewWriter(f)
	defer func() {
		b.Flush()
		f.Close()
	}()

	if err := png.Encode(b, img); err != nil {
		panic(err)
	}
	return
}

// ReadPNG reads and decodes png data into image.Image
func ReadPNG(name string) (img image.Image) {
	io, err := os.Open(GetExFilePath(name))
	if err != nil {
		panic(err)
	}
	img, err = png.Decode(io)
	if err != nil {
		panic(err)
	}
	return
}
