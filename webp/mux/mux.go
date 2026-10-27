package mux

/*
#cgo LDFLAGS: -lwebpmux -lwebpdemux -lwebp -lsharpyuv -lm

#include <webp/demux.h>
#include <webp/mux.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// FourCC is a four-byte RIFF chunk identifier.
type FourCC string

const (
	ICCP FourCC = "ICCP"
	EXIF FourCC = "EXIF"
	XMP  FourCC = "XMP "
)

// MuxError corresponds to WebPMuxError, including the OK success status.
type MuxError int

const (
	MuxOK              MuxError = C.WEBP_MUX_OK
	MuxNotFound        MuxError = C.WEBP_MUX_NOT_FOUND
	MuxInvalidArgument MuxError = C.WEBP_MUX_INVALID_ARGUMENT
	MuxBadData         MuxError = C.WEBP_MUX_BAD_DATA
	MuxMemoryError     MuxError = C.WEBP_MUX_MEMORY_ERROR
	MuxNotEnoughData   MuxError = C.WEBP_MUX_NOT_ENOUGH_DATA
)

func (s MuxError) String() string {
	switch s {
	case MuxOK:
		return "WEBP_MUX_OK"
	case MuxNotFound:
		return "WEBP_MUX_NOT_FOUND"
	case MuxInvalidArgument:
		return "WEBP_MUX_INVALID_ARGUMENT"
	case MuxBadData:
		return "WEBP_MUX_BAD_DATA"
	case MuxMemoryError:
		return "WEBP_MUX_MEMORY_ERROR"
	case MuxNotEnoughData:
		return "WEBP_MUX_NOT_ENOUGH_DATA"
	default:
		return fmt.Sprintf("WebPMuxError(%d)", int(s))
	}
}

// Data corresponds to WebPData. Its zero value is initialized and empty.
// Owned data (from NewData or MuxAssemble) must be released with DataClear.
// Borrowed data (from MuxGetChunk or ChunkIterator.Chunk) must not be cleared.
// Copying a Data value does not copy its underlying storage.
type Data C.WebPData

// NewData copies b into libwebp-allocated memory. It returns nil on allocation
// failure. Even an empty input receives a non-NULL buffer, for MuxSetChunk.
func NewData(b []byte) *Data {
	n := len(b)
	if n == 0 {
		n = 1
	}
	p := C.WebPMalloc(C.size_t(n))
	if p == nil {
		return nil
	}
	copy(unsafe.Slice((*byte)(p), len(b)), b)
	return &Data{bytes: (*C.uint8_t)(p), size: C.size_t(len(b))}
}

// Bytes returns a view of the data, without copying. The view is valid only
// while the underlying storage is alive and unchanged. Treat borrowed views
// as read-only; clone them if they need to outlive their owner.
func (d *Data) Bytes() []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(d.bytes)), d.size)
}

// DataClear frees owned storage and resets data. Do not use on borrowed data.
func DataClear(data *Data) { C.WebPDataClear((*C.WebPData)(data)) }

// DataCopy copies src into dst, which must not already own storage.
// The caller owns the resulting data. It returns false on failure.
func DataCopy(src, dst *Data) bool {
	return C.WebPDataCopy((*C.WebPData)(src), (*C.WebPData)(dst)) != 0
}

// Mux corresponds to the opaque WebPMux object.
type Mux C.WebPMux

func GetMuxVersion() int { return int(C.WebPGetMuxVersion()) }

// MuxNew creates an empty mux, or returns nil on allocation failure.
func MuxNew() *Mux { return (*Mux)(C.WebPMuxNew()) }

// MuxCreate parses a WebP container, returning nil on invalid data or allocation
// failure. With copyData=false, bitstream's storage must live until MuxDelete.
func MuxCreate(bitstream *Data, copyData bool) *Mux {
	return (*Mux)(C.WebPMuxCreate((*C.WebPData)(bitstream), cBool(copyData)))
}

// MuxDelete releases mux. All borrowed chunk data becomes invalid.
func MuxDelete(mux *Mux) { C.WebPMuxDelete((*C.WebPMux)(mux)) }

// MuxSetChunk sets or replaces a non-image chunk. With copyData=false, chunk's
// storage must remain alive and unchanged until assembly and all subsequent
// uses of the mux have finished (or the chunk is replaced/deleted).
func MuxSetChunk(mux *Mux, fourcc FourCC, chunk *Data, copyData bool) MuxError {
	id, ok := cFourCC(fourcc)
	if !ok {
		return MuxInvalidArgument
	}
	return MuxError(C.WebPMuxSetChunk((*C.WebPMux)(mux), &id[0], (*C.WebPData)(chunk), cBool(copyData)))
}

// MuxGetChunk writes a borrowed reference into chunk, which must not own
// storage. The reference is valid until mutation or deletion of the mux.
// A missing chunk returns MuxNotFound.
func MuxGetChunk(mux *Mux, fourcc FourCC, chunk *Data) MuxError {
	id, ok := cFourCC(fourcc)
	if !ok {
		return MuxInvalidArgument
	}
	return MuxError(C.WebPMuxGetChunk((*C.WebPMux)(mux), &id[0], (*C.WebPData)(chunk)))
}

func MuxDeleteChunk(mux *Mux, fourcc FourCC) MuxError {
	id, ok := cFourCC(fourcc)
	if !ok {
		return MuxInvalidArgument
	}
	return MuxError(C.WebPMuxDeleteChunk((*C.WebPMux)(mux), &id[0]))
}

// MuxSetImage sets a non-animated image, replacing existing images/frames.
// With copyData=false, bitstream has the same lifetime requirements as chunk
// data passed to MuxSetChunk.
func MuxSetImage(mux *Mux, bitstream *Data, copyData bool) MuxError {
	return MuxError(C.WebPMuxSetImage((*C.WebPMux)(mux), (*C.WebPData)(bitstream), cBool(copyData)))
}

// MuxAssemble writes a WebP container into output, which must not already own
// storage. On success the caller owns output and must call DataClear on it.
func MuxAssemble(mux *Mux, output *Data) MuxError {
	return MuxError(C.WebPMuxAssemble((*C.WebPMux)(mux), (*C.WebPData)(output)))
}

func cFourCC(fourcc FourCC) ([4]C.char, bool) {
	var id [4]C.char
	if len(fourcc) != len(id) {
		return id, false
	}
	for i := range id {
		id[i] = C.char(fourcc[i])
	}
	return id, true
}

func cBool(b bool) C.int {
	if b {
		return 1
	}
	return 0
}
