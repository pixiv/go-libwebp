package util_test

import (
	"testing"

	"github.com/pixiv/go-libwebp/test/util"
)

var PNGFiles = []string{
	"butterfly.png",
	"cosmos.png",
	"fizyplankton.png",
	"kinkaku.png",
	"yellow-rose-3.png",
}

var WebPFiles = []string{
	"butterfly.webp",
	"cosmos.webp",
	"fizyplankton.webp",
	"kinkaku.webp",
	"yellow-rose-3.webp",
}

func TestOpenFile(t *testing.T) {
	for _, file := range PNGFiles {
		util.OpenFile(file)
	}
	for _, file := range WebPFiles {
		util.OpenFile(file)
	}
}

func TestReadFile(t *testing.T) {
	for _, file := range PNGFiles {
		util.ReadFile(file)
	}
	for _, file := range WebPFiles {
		util.ReadFile(file)
	}
}

func TestCreateFile(t *testing.T) {
	f := util.CreateFile("util_test")
	f.Write([]byte{'o', 'k'})
	f.Close()
}

func TestReadWritePNG(t *testing.T) {
	for _, file := range PNGFiles {
		png := util.ReadPNG(file)
		util.WritePNG(png, "util_test_"+file)
	}
}
