package webp

/*
#cgo LDFLAGS: -lwebpdemux

#include <stdlib.h>
#include <webp/demux.h>
*/
import "C"

import (
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	"unsafe"
)

// AnimationDecoder decodes an animated WebP.
type AnimationDecoder struct {
	opts  C.WebPAnimDecoderOptions
	c     *C.WebPAnimDecoder
	cData *C.WebPData
}

// AnimationInfo represents properties of an animation.
type AnimationInfo struct {
	CanvasWidth     int
	CanvasHeight    int
	LoopCount       int
	FrameCount      int
	BackgroundColor color.RGBA
}

// Animation represents a decoded WebP animation.
type Animation struct {
	AnimationInfo

	// Image is the list of decoded frames.
	Image []*image.RGBA

	// Timestamp of each frame in milliseconds.
	Timestamp []int
}

// NewAnimationDecoder initializes a new decoder.
func NewAnimationDecoder(data []byte) (*AnimationDecoder, error) {
	ad := &AnimationDecoder{}

	if C.WebPAnimDecoderOptionsInit(&ad.opts) == 0 {
		return nil, errors.New("failed to initialize animation decoder config")
	}
	ad.opts.color_mode = C.MODE_RGBA

	ad.cData = &C.WebPData{}
	C.WebPDataInit(ad.cData)

	ad.cData.bytes = (*C.uint8_t)(C.CBytes(data))
	ad.cData.size = (C.size_t)(len(data))

	ad.c = C.WebPAnimDecoderNew(ad.cData, &ad.opts)
	if ad.c == nil {
		C.free(unsafe.Pointer(ad.cData.bytes))
		return nil, errors.New("failed to initialize animation decoder")
	}

	return ad, nil
}

// GetInfo retrieves properties of the animation.
func (ad *AnimationDecoder) GetInfo() (*AnimationInfo, error) {
	info := &C.WebPAnimInfo{}

	if C.WebPAnimDecoderGetInfo(ad.c, info) == 0 {
		return nil, errors.New("error in WebPAnimDecoderGetInfo")
	}

	b := make([]uint8, 4)
	binary.BigEndian.PutUint32(b, uint32(info.bgcolor))

	return &AnimationInfo{
		CanvasWidth:     int(info.canvas_width),
		CanvasHeight:    int(info.canvas_height),
		LoopCount:       int(info.loop_count),
		FrameCount:      int(info.frame_count),
		BackgroundColor: color.RGBA{b[0], b[1], b[2], b[3]},
	}, nil
}

// Decode decodes a WebP animation.
func (ad *AnimationDecoder) Decode() (*Animation, error) {
	info, err := ad.GetInfo()
	if err != nil {
		return nil, err
	}

	anim := &Animation{
		AnimationInfo: *info,
	}

	for C.WebPAnimDecoderHasMoreFrames(ad.c) != 0 {
		var ts C.int
		var pix *C.uint8_t

		if C.WebPAnimDecoderGetNext(ad.c, &pix, &ts) == 0 {
			return nil, errors.New("error in WebPAnimDecoderGetNext")
		}

		img := image.NewRGBA(image.Rect(0, 0, info.CanvasWidth, info.CanvasHeight))
		C.memcpy(unsafe.Pointer(&img.Pix[0]), unsafe.Pointer(pix), C.size_t(len(img.Pix)))

		anim.Image = append(anim.Image, img)
		anim.Timestamp = append(anim.Timestamp, int(ts))
	}

	return anim, nil
}

// Close deletes the decoder and frees resources.
func (ad *AnimationDecoder) Close() {
	C.free(unsafe.Pointer(ad.cData.bytes))
	C.WebPAnimDecoderDelete(ad.c)
}
