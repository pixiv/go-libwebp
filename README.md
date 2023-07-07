go-libwebp
==========

[![ci](https://github.com/pixiv/go-libwebp/actions/workflows/ci.yml/badge.svg)](https://github.com/pixiv/go-libwebp/actions/workflows/ci.yml)
[![GoDoc](https://godoc.org/github.com/pixiv/go-libwebp/webp?status.svg)](https://godoc.org/github.com/pixiv/go-libwebp/webp)

A implementation of Go binding for [libwebp](https://developers.google.com/speed/webp/docs/api).

## Dependencies

- libwebp 0.4, 0.5

## Usage

The [examples](./examples) directory contains example codes and images.

### Decoding WebP into image.RGBA

```
package main

import (
	"github.com/pixiv/go-libwebp/test/util"
	"github.com/pixiv/go-libwebp/webp"
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

	"github.com/pixiv/go-libwebp/test/util"
	"github.com/pixiv/go-libwebp/webp"
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

## TODO

- Incremental decoding API
- Container API (Animation)

## License

This library is released under The BSD 2-Clause License.
See [LICENSE](./LICENSE).
