go-libwebp
==========

[![Build Status](https://travis-ci.org/harukasan/go-libwebp.svg)](https://travis-ci.org/harukasan/go-libwebp)

A implementation of Go binding for [libwebp](https://developers.google.com/speed/webp/docs/api) written with cgo.

## Usage

The [examples](./examples) directory contains example codes and images.

### Decoding WebP into image.RGBA

```
package main

import (
	"github.com/harukasan/go-libwebp/test/util"
	"github.com/harukasan/go-libwebp/webp"
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

	"github.com/harukasan/go-libwebp/test/util"
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
	}

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

Copyright (c) 2014 MICHII Shunsuke. All rights reserved.

This library is released under The BSD 2-Clause License.
See [LICENSE](./LICENSE).
