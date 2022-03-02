go-libwebp
==========

[![GoDoc](https://godoc.org/github.com/tidbyt/go-libwebp/webp?status.svg)](https://godoc.org/github.com/tidbyt/go-libwebp/webp)

A implementation of Go binding for [libwebp](https://developers.google.com/speed/webp/docs/api).

## Dependencies

- libwebp 0.5+, compiled with `--enable-libwebpmux`

## Usage

The [examples](./examples) directory contains example codes and images.

### Decoding WebP into image.RGBA

```
package main

import (
	"github.com/tidbyt/go-libwebp/test/util"
	"github.com/tidbyt/go-libwebp/webp"
)

func main() {
	var err error

	// Read binary data
	data := util.ReadFile("cosmos.webp")

	// Decode
	options := &webp.DecoderOptions{}
	img, err := webp.DecodeRGBA(data, options)
	if err != nil {
		panic(err)
	}

	util.WritePNG(img, "encoded_cosmos.png")
}
```

You can set more decoding options such as cropping, flipping and scaling.

### Encoding WebP from image.RGBA

```
package main

import (
	"bufio"
	"image"

	"github.com/tidbyt/go-libwebp/test/util"
	"github.com/tidbyt/go-libwebp/webp"
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

	config := webp.ConfigPreset(webp.PresetDefault, 90)

	// Encode into WebP
	if err := webp.EncodeRGBA(w, img.(*image.RGBA), config); err != nil {
		panic(err)
	}
}
```


### Encoding animations from a series of frames

```
package main

import (
	"image"
	"time"

	"github.com/tidbyt/go-libwebp/test/util"
	"github.com/tidbyt/go-libwebp/webp"
)

func main() {
	// Get some frames
	img := []image.Image{
		util.ReadPNG("butterfly.png"),
		util.ReadPNG("checkerboard.png"),
		util.ReadPNG("yellow-rose-3.png"),
	}

	// Initialize the animation encoder
	width, height := 24, 24
	anim, err := webp.NewAnimationEncoder(width, height, 0, 0)
	if err != nil {
		panic(err)
	}
	defer anim.Close()

	// Add each frame to the animation
	for i, im := range img {
		// all frames of an animation must have the same dimensions
		cropped := im.(interface {
			SubImage(r image.Rectangle) image.Image
		}).SubImage(image.Rect(0, 0, width, height))

		if err := anim.AddFrame(cropped, 100*time.Millisecond); err != nil {
			panic(err)
		}
	}

	// Assemble the final animation
	buf, err := anim.Assemble()
	if err != nil {
		panic(err)
	}

	// Write to disk
	f := util.CreateFile("animation.webp")
	defer f.Close()
	f.Write(buf)
}

```

## TODO

- Incremental decoding API

## License

Copyright (c) 2016 MICHII Shunsuke. All rights reserved.

This library is released under The BSD 2-Clause License.
See [LICENSE](./LICENSE).
