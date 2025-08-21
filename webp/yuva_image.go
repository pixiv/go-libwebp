package webp

import "image"

// YUVAImage represents a image of YUV colors with alpha channel image.
//
// YUVAImage contains decoded YCbCr image data with alpha channel,
// but it is not compatible with image.YCbCr. Because, the RGB-YCbCr conversion
// that used in WebP is following to ITU-R BT.601 standard.
// In contrast, the conversion of Image.YCbCr (and color.YCbCrModel) is following
// to the JPEG standard (JFIF). If you need the image as image.YCBCr, you will
// first convert from WebP to RGB image, then convert from RGB image to JPEG's
// YCbCr image.
//
// See: http://en.wikipedia.org/wiki/YCbCr
type YUVAImage struct {
	Y, Cb, Cr, A []uint8
	YStride      int
	CStride      int
	AStride      int
	ColorSpace   ColorSpace
	Rect         image.Rectangle
}

// NewYUVAImage creates and allocates image buffer.
func NewYUVAImage(r image.Rectangle, c ColorSpace) (image *YUVAImage) {
	yw, yh := r.Dx(), r.Dy()
	cw, ch := ((r.Max.X+1)/2 - r.Min.X/2), ((r.Max.Y+1)/2 - r.Min.Y/2)

	switch c {
	case YUV420:
		b := make([]byte, yw*yh+2*cw*ch)
		image = &YUVAImage{
			Y:          b[:yw*yh],
			Cb:         b[yw*yh+0*cw*ch : yw*yh+1*cw*ch],
			Cr:         b[yw*yh+1*cw*ch : yw*yh+2*cw*ch],
			A:          nil,
			YStride:    yw,
			CStride:    cw,
			AStride:    0,
			ColorSpace: c,
			Rect:       r,
		}

	case YUV420A:
		b := make([]byte, 2*yw*yh+2*cw*ch)
		image = &YUVAImage{
			Y:          b[:yw*yh],
			Cb:         b[yw*yh+0*cw*ch : yw*yh+1*cw*ch],
			Cr:         b[yw*yh+1*cw*ch : yw*yh+2*cw*ch],
			A:          b[yw*yh+2*cw*ch:],
			YStride:    yw,
			CStride:    cw,
			AStride:    yw,
			ColorSpace: c,
			Rect:       r,
		}
	}

	return
}
