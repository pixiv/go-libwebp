// Package mux provides Go bindings for libwebpmux and libwebpdemux.
package mux

/*
#cgo LDFLAGS: -lwebpmux -lwebpdemux -lwebp -lsharpyuv -lm

#include <webp/demux.h>
#include <webp/mux.h>
*/
import "C"

import (
	"bytes"
	"errors"
	"fmt"
	"unsafe"
)

// FourCC is a RIFF chunk identifier. libwebp currently supports "ICCP", "EXIF",
// and "XMP " (trailing space) for metadata via WebPMuxSetChunk / WebPDemuxGetChunk.
type FourCC string

const (
	ICCP FourCC = "ICCP"
	EXIF FourCC = "EXIF"
	XMP  FourCC = "XMP "
)

var errEmptyWebPBitstream = errors.New("empty webp bitstream")
var errNilChunk = errors.New("nil chunk")
var errFourCCLengthMustBe4 = errors.New("fourcc must be 4 bytes")

// GetChunk extracts the first chunk with the given fourcc from a WebP bitstream.
// It returns nil, nil when the bitstream has no such chunk.
// Reading uses libwebpdemux so files with inconsistent VP8X flags
// (for example ICCP present but ALPHA_FLAG missing) still yield the chunk.
func GetChunk(data []byte, fourcc FourCC) ([]byte, error) {
	if len(data) == 0 {
		return nil, errEmptyWebPBitstream
	}
	cFourcc, err := cFourCC(fourcc)
	if err != nil {
		return nil, err
	}

	bitstream, status := cWebPData(data)
	if status != C.WEBP_MUX_OK {
		return nil, muxErrorf("WebPDemuxGetChunk", status)
	}
	defer C.WebPFree(unsafe.Pointer(bitstream.bytes))

	demux := C.WebPDemux(&bitstream)
	if demux == nil {
		return nil, muxErrorf("WebPDemuxGetChunk", C.WEBP_MUX_BAD_DATA)
	}
	defer C.WebPDemuxDelete(demux)

	var iter C.WebPChunkIterator
	found := C.WebPDemuxGetChunk(demux, (*C.char)(unsafe.Pointer(&cFourcc[0])), 1, &iter)
	if found == 0 {
		return nil, nil
	}
	defer C.WebPDemuxReleaseChunkIterator(&iter)
	if iter.chunk.size == 0 {
		return []byte{}, nil
	}
	return bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(iter.chunk.bytes)), iter.chunk.size)), nil
}

// SetChunk returns a WebP bitstream with the given chunk set for fourcc.
// An existing chunk of the same fourcc is replaced. A nil chunk is rejected
// because WebPMuxSetChunk rejects NULL pointers; an empty chunk is allowed.
func SetChunk(data []byte, fourcc FourCC, chunk []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errEmptyWebPBitstream
	}
	cFourcc, err := cFourCC(fourcc)
	if err != nil {
		return nil, err
	}
	if chunk == nil {
		return nil, errNilChunk
	}

	bitstream, status := cWebPData(data)
	if status != C.WEBP_MUX_OK {
		return nil, muxErrorf("WebPMuxSetChunk", status)
	}
	defer C.WebPFree(unsafe.Pointer(bitstream.bytes))

	chunkData, status := cWebPData(chunk)
	if status != C.WEBP_MUX_OK {
		return nil, muxErrorf("WebPMuxSetChunk", status)
	}
	defer C.WebPFree(unsafe.Pointer(chunkData.bytes))

	// Use copy_data=0 to avoid copying bitstream and chunkData a second time.
	// The mux borrows both buffers. Freeing either one before WebPMuxDelete would
	// cause a use-after-free. The defer order is safe: WebPMuxDelete runs before
	// both WebPFree calls.
	webpMux := C.WebPMuxCreate(&bitstream, 0)
	if webpMux == nil {
		return nil, muxErrorf("WebPMuxSetChunk", C.WEBP_MUX_BAD_DATA)
	}
	defer C.WebPMuxDelete(webpMux)

	status = C.WebPMuxSetChunk(webpMux, (*C.char)(unsafe.Pointer(&cFourcc[0])), &chunkData, 0)
	if status != C.WEBP_MUX_OK {
		return nil, muxErrorf("WebPMuxSetChunk", status)
	}

	var assembled C.WebPData
	status = C.WebPMuxAssemble(webpMux, &assembled)
	defer C.WebPFree(unsafe.Pointer(assembled.bytes))
	if status != C.WEBP_MUX_OK {
		return nil, muxErrorf("WebPMuxSetChunk", status)
	}
	if assembled.bytes == nil {
		return nil, muxErrorf("WebPMuxSetChunk", C.WEBP_MUX_MEMORY_ERROR)
	}
	return bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(assembled.bytes)), assembled.size)), nil
}

func cFourCC(fourcc FourCC) ([4]byte, error) {
	if len(fourcc) != 4 {
		return [4]byte{}, errFourCCLengthMustBe4
	}
	var out [4]byte
	copy(out[:], fourcc)
	return out, nil
}

// cWebPData copies b into C memory. cgo forbids passing a WebPData whose bytes
// pointer refers to Go memory (a Go pointer inside a Go-allocated struct).
// An empty slice still gets a non-NULL pointer: WebPMuxSetChunk rejects NULL
// even when size is 0, and WebPMalloc(0) may return NULL.
func cWebPData(b []byte) (C.WebPData, C.WebPMuxError) {
	n := len(b)
	alloc := n
	if alloc == 0 {
		alloc = 1
	}
	p := C.WebPMalloc(C.size_t(alloc))
	if p == nil {
		return C.WebPData{}, C.WEBP_MUX_MEMORY_ERROR
	}
	copy(unsafe.Slice((*byte)(p), n), b)
	return C.WebPData{
		bytes: (*C.uint8_t)(p),
		size:  C.size_t(n),
	}, C.WEBP_MUX_OK
}

func muxErrorf(op string, status C.WebPMuxError) error {
	return fmt.Errorf("%s caught unexpected status: %s", op, statusString(status))
}

func statusString(status C.WebPMuxError) string {
	switch status {
	case C.WEBP_MUX_OK:
		return "WEBP_MUX_OK"
	case C.WEBP_MUX_NOT_FOUND:
		return "WEBP_MUX_NOT_FOUND"
	case C.WEBP_MUX_INVALID_ARGUMENT:
		return "WEBP_MUX_INVALID_ARGUMENT"
	case C.WEBP_MUX_BAD_DATA:
		return "WEBP_MUX_BAD_DATA"
	case C.WEBP_MUX_MEMORY_ERROR:
		return "WEBP_MUX_MEMORY_ERROR"
	case C.WEBP_MUX_NOT_ENOUGH_DATA:
		return "WEBP_MUX_NOT_ENOUGH_DATA"
	}
	return fmt.Sprintf("unexpected mux status %d", int(status))
}
