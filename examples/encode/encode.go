package main

import (
	"bufio"
	"image"

	"github.com/harukasan/go-libwebp/examples/util"
	"github.com/harukasan/go-libwebp/webp"
)

func main() {
	img := util.ReadPNG("cosmos.png")

	// Create file and buffered writer
	io := util.CreateFile("encoded_cosmos.webp")
	w := bufio.NewWriter(io)
	defer func() {
		w.Flush()
		io.Close()
	}()

	config := webp.Config{
		Preset:  webp.PresetDefault,
		Quality: 90,
		Method:  6,
	}

	// Encode into WebP
	if err := webp.EncodeRGBA(w, img.(*image.RGBA), config); err != nil {
		panic(err)
	}
}
