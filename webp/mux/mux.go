// Package mux provides Go bindings for libwebpmux and libwebpdemux.
package mux

/*
#cgo LDFLAGS: -lwebpmux -lwebpdemux -lwebp -lsharpyuv -lm

#include <stdlib.h>
#include <string.h>
#include <webp/demux.h>
#include <webp/mux.h>

static const char kICCP[] = "ICCP";

static WebPMuxError webpGetICCProfile(const uint8_t* data, size_t size, uint8_t** out, size_t* out_size) {
	uint8_t* copy;
	WebPData bitstream;
	WebPDemuxer* demux;
	WebPChunkIterator iter;
	int found;

	*out = NULL;
	*out_size = 0;

	copy = (uint8_t*)WebPMalloc(size);
	if (copy == NULL) {
		return WEBP_MUX_MEMORY_ERROR;
	}
	memcpy(copy, data, size);

	bitstream.bytes = copy;
	bitstream.size = size;
	demux = WebPDemux(&bitstream);
	if (demux == NULL) {
		WebPFree(copy);
		return WEBP_MUX_BAD_DATA;
	}

	found = WebPDemuxGetChunk(demux, kICCP, 1, &iter);
	if (!found || iter.chunk.bytes == NULL || iter.chunk.size == 0) {
		WebPDemuxDelete(demux);
		WebPFree(copy);
		return WEBP_MUX_OK;
	}

	*out = (uint8_t*)WebPMalloc(iter.chunk.size);
	if (*out == NULL) {
		WebPDemuxReleaseChunkIterator(&iter);
		WebPDemuxDelete(demux);
		WebPFree(copy);
		return WEBP_MUX_MEMORY_ERROR;
	}
	memcpy(*out, iter.chunk.bytes, iter.chunk.size);
	*out_size = iter.chunk.size;
	WebPDemuxReleaseChunkIterator(&iter);
	WebPDemuxDelete(demux);
	WebPFree(copy);
	return WEBP_MUX_OK;
}

static WebPMuxError webpSetICCProfile(const uint8_t* data, size_t size, const uint8_t* icc, size_t icc_size, uint8_t** out, size_t* out_size) {
	WebPData bitstream;
	WebPData icc_data;
	WebPData assembled;
	WebPMux* mux;
	WebPMuxError err;

	*out = NULL;
	*out_size = 0;

	bitstream.bytes = data;
	bitstream.size = size;
	mux = WebPMuxCreate(&bitstream, 1);
	if (mux == NULL) {
		return WEBP_MUX_BAD_DATA;
	}

	icc_data.bytes = icc;
	icc_data.size = icc_size;
	err = WebPMuxSetChunk(mux, kICCP, &icc_data, 1);
	if (err != WEBP_MUX_OK) {
		WebPMuxDelete(mux);
		return err;
	}

	WebPDataInit(&assembled);
	err = WebPMuxAssemble(mux, &assembled);
	WebPMuxDelete(mux);
	if (err != WEBP_MUX_OK) {
		WebPDataClear(&assembled);
		return err;
	}

	*out = (uint8_t*)assembled.bytes;
	*out_size = assembled.size;
	return WEBP_MUX_OK;
}
*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

var errEmptyWebPBitstream = errors.New("empty webp bitstream")
var errWebPMuxAssemble = errors.New("Could not assemble webp bitstream")

// GetICCProfile extracts the ICCP chunk from a WebP bitstream.
// It returns nil, nil when the bitstream has no ICC profile.
// Reading uses libwebpdemux so files with inconsistent VP8X flags
// (for example ICCP present but ALPHA_FLAG missing) still yield the profile.
func GetICCProfile(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errEmptyWebPBitstream
	}

	var out *C.uint8_t
	var outSize C.size_t
	status := C.webpGetICCProfile((*C.uint8_t)(unsafe.Pointer(&data[0])), C.size_t(len(data)), &out, &outSize)
	if status != C.WEBP_MUX_OK {
		return nil, fmt.Errorf("WebPDemuxGetChunk returns unexpected status: %s", statusString(status))
	}
	if out == nil || outSize == 0 {
		return nil, nil
	}
	defer C.WebPFree(unsafe.Pointer(out))
	return C.GoBytes(unsafe.Pointer(out), C.int(outSize)), nil
}

// SetICCProfile returns a WebP bitstream with the given ICC profile.
// An empty profile leaves the input unchanged.
func SetICCProfile(data, icc []byte) ([]byte, error) {
	if len(icc) == 0 {
		return data, nil
	}
	if len(data) == 0 {
		return nil, errEmptyWebPBitstream
	}

	var out *C.uint8_t
	var outSize C.size_t
	status := C.webpSetICCProfile(
		(*C.uint8_t)(unsafe.Pointer(&data[0])), C.size_t(len(data)),
		(*C.uint8_t)(unsafe.Pointer(&icc[0])), C.size_t(len(icc)),
		&out, &outSize,
	)
	if status != C.WEBP_MUX_OK {
		return nil, fmt.Errorf("WebPMuxSetChunk returns unexpected status: %s", statusString(status))
	}
	if out == nil {
		return nil, errWebPMuxAssemble
	}
	defer C.WebPFree(unsafe.Pointer(out))
	return C.GoBytes(unsafe.Pointer(out), C.int(outSize)), nil
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
