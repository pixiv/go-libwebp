go-libwebp
==========

[![ci](https://github.com/pixiv/go-libwebp/actions/workflows/ci.yml/badge.svg)](https://github.com/pixiv/go-libwebp/actions/workflows/ci.yml)
[![GoDoc](https://godoc.org/github.com/pixiv/go-libwebp/webp?status.svg)](https://godoc.org/github.com/pixiv/go-libwebp/webp)

A implementation of Go binding for [libwebp](https://developers.google.com/speed/webp/docs/api).

## Dependencies

- libwebp 1.3.2 and above

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

### Container chunks (mux / demux)

Encode/decode stay in `github.com/pixiv/go-libwebp/webp` and link only `libwebp`.
Chunk access (`ICCP`, `EXIF`, `XMP `, …) is a separate package so callers that
do not need the container API do not link `libwebpmux` / `libwebpdemux`.

```
import "github.com/pixiv/go-libwebp/webp/mux"

// Copy Go bytes into owned C storage.
input := mux.NewData(data)
if input == nil {
    panic("allocation failed")
}
defer mux.DataClear(input)

m := mux.MuxCreate(input, false)
if m == nil {
    panic("invalid WebP or allocation failure")
}
defer mux.MuxDelete(m)

icc := mux.NewData(iccBytes)
if icc == nil {
    panic("allocation failed")
}
// copyData=true lets us release icc immediately after setting the chunk.
status := mux.MuxSetChunk(m, mux.ICCP, icc, true)
mux.DataClear(icc)
if status != mux.MuxOK {
    panic(status)
}

var output mux.Data
if status := mux.MuxAssemble(m, &output); status != mux.MuxOK {
    panic(status)
}
defer mux.DataClear(&output)
// Consume output.Bytes() before output is cleared.
```

The bindings expose individual C operations (without the `WebP` prefix),
including explicit creation, assembly and deletion. `MuxError` preserves C
status codes, including `MuxNotFound`. `NewData` copies Go bytes into C memory;
`Data.Bytes()` returns a view without copying.

For reading chunks, use `Demux(input)`, then
`DemuxGetChunk(d, mux.ICCP, 1, &iter)` with a `mux.ChunkIterator`.
It returns a boolean; `iter.Chunk()` gives borrowed `Data`.
Call `DemuxReleaseChunkIterator(&iter)` before `DemuxDelete(d)` and keep
`input` alive until the demuxer is deleted. Borrowed data must not be passed to
`DataClear`. With `copyData=false`, mux input buffers must also stay alive for
all operations that use them; see the function documentation for details.

## TODO

- Incremental decoding API
- Container API (Animation)

## License

This library is released under The BSD 2-Clause License.
See [LICENSE](./LICENSE).
