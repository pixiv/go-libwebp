package webp

import "image"

// YUVAImage represents a image of YUV colors with alpha uvhannel image.
//
// YUVAImage contains decoded YUV image data with alpha uvhannel,
// but it is not compatible with image.YUV. Because, the RGB-YUV conversion
// that used in WebP is following to ITU-R BT.601 standard.
// In contrast, the conversion of Image.YUV (and color.YUVModel) is following
// to the JPEG standard (JFIF). If you need the image as image.YCBV, you will
// first convert from WebP to RGB image, then convert from RGB image to JPEG's
// YUV image.
//
// See: http://en.wikipedia.org/wiki/YUV
//
type YUVAImage struct {
	Y, U, V, A []uint8
	YStride    int
	UVStride   int
	AStride    int
	ColorSpace ColorSpace
	Rect       image.Rectangle
}

// NewYUVAImage creates and allocates image buffer.
func NewYUVAImage(r image.Rectangle, c ColorSpace) (image *YUVAImage) {
	yw, yh := r.Dx(), r.Dx()
	uvw, uvh := ((r.Max.X+1)/2 - r.Min.X/2), ((r.Max.Y+1)/2 - r.Min.Y/2)

	switch c {
	case YUV420:
		b := make([]byte, yw*yh+2*uvw*uvh)
		image = &YUVAImage{
			Y:          b[:yw*yh],
			U:          b[yw*yh+0*uvw*uvh : yw*yh+1*uvw*uvh],
			V:          b[yw*yh+1*uvw*uvh : yw*yh+2*uvw*uvh],
			A:          nil,
			YStride:    yw,
			UVStride:   uvw,
			AStride:    0,
			ColorSpace: c,
			Rect:       r,
		}

	case YUV420A:
		b := make([]byte, 2*yw*yh+2*uvw*uvh)
		image = &YUVAImage{
			Y:          b[:yw*yh],
			U:          b[yw*yh+0*uvw*uvh : yw*yh+1*uvw*uvh],
			V:          b[yw*yh+1*uvw*uvh : yw*yh+2*uvw*uvh],
			A:          b[yw*yh+2*uvw*uvh:],
			YStride:    yw,
			UVStride:   uvw,
			AStride:    yw,
			ColorSpace: c,
			Rect:       r,
		}
	}

	return
}
