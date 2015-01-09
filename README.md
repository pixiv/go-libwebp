go-libwebp
==========

A implementation of Go binding for [libwebp](https://developers.google.com/speed/webp/docs/api) written with cgo.

## Usage

The [examples](./examples) directory contains example codes and images.

### Decoding WebP into image.RGBA

```
import (
	"io/ioutil"

	"github.com/harukasan/go-libwebp/examples/util"
	"github.com/harukasan/go-libwebp/webp"
)

func main() {
	var err error

	// Read binary data
	data, err := ioutil.ReadFile("examples/cosmos.webp")
	if err != nil {
		panic(err)
	}

	// Decode
	options := &webp.DecoderOptions{}
	img, err := webp.DecodeRGBA(data, options)
	if err != nil {
		panic(err)
	}

	err = util.WritePNG(img, "out/encoded_cosmos.png")
	if err != nil {
		panic(err)
	}
}
```

You can set more decoding options such as cropping, flipping and scaling.

### Encoding WebP from image.RGBA

```
package main

import (
	"bufio"
	"image"

	"github.com/harukasan/go-libwebp/examples/util"
	"github.com/harukasan/go-libwebp/webp"
)

func main() {
	err := util.ReadPNG("examples/cosmos.png")
	if err != nil {
		panic()
	}

	// Encode to WebP
	io := ("out.webp")
	w := bufio.NewWriter(f)
	defer func() {
		w.Flush()
		f.Close()
	}()

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
