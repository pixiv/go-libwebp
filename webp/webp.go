package webp

/*
#cgo LDFLAGS: -lwebp

#include <stdlib.h>
#include <webp/encode.h>

*/
import "C"

// ColorSpace represents encoding color space in WebP
type ColorSpace int

const (
	// YUV420 specifies YUV4:2:0
	YUV420 ColorSpace = C.WEBP_YUV420
	// YUV420A specifies YUV4:2:0 with alpha channel
	YUV420A ColorSpace = C.WEBP_YUV420A
)

// Preset corresponds to C.WebPPreset
type Preset int

const (
	// PresetDefault corresponds to WEBP_PRESET_DEFAULT, for default preset.
	PresetDefault Preset = C.WEBP_PRESET_DEFAULT
	// PresetPicture corresponds to WEBP_PRESET_PICTURE, for digital picture, like portrait, inner shot
	PresetPicture Preset = C.WEBP_PRESET_PICTURE
	// PresetPhoto corresponds to WEBP_PRESET_PHOTO, for outdoor photograph, with natural lighting
	PresetPhoto Preset = C.WEBP_PRESET_PHOTO
	// PresetDrawing corresponds to WEBP_PRESET_DRAWING, for hand or line drawing, with high-contrast details
	PresetDrawing Preset = C.WEBP_PRESET_DRAWING
	// PresetIcon corresponds to WEBP_PRESET_ICON, for small-sized colorful images
	PresetIcon Preset = C.WEBP_PRESET_ICON
	// PresetText corresponds to WEBP_PRESET_TEXT, for text-like
	PresetText Preset = C.WEBP_PRESET_TEXT
)

// FilterType corresponds to filter types in compression parameters.
type FilterType int

const (
	// SimpleFilter (=0, default)
	SimpleFilter FilterType = iota
	// StrongFilter (=1)
	StrongFilter
)
