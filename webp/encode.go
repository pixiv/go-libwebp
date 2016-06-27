package webp

/*
#include <stdlib.h>
#include <webp/encode.h>

int writeWebP(uint8_t*, size_t, struct WebPPicture*);

static WebPPicture *malloc_WebPPicture(void) {
	return malloc(sizeof(WebPPicture));
}

static void free_WebPPicture(WebPPicture* webpPicture) {
	free(webpPicture);
}

static int getNearLossless(WebPConfig* webpConfig, int* value) {
#if WEBP_ENCODER_ABI_VERSION < 0x206
	return 0;
#else
	*value = webpConfig->near_lossless;
	return 1;
#endif
}

static int setNearLossless(WebPConfig* webpConfig, int value) {
#if WEBP_ENCODER_ABI_VERSION < 0x206
	return 0;
#else
	webpConfig->near_lossless = value;
	return 1;
#endif
}

static int getExact(WebPConfig* webpConfig, int* value) {
#if WEBP_ENCODER_ABI_VERSION < 0x209
	return 0;
#else
	*value = webpConfig->exact;
	return 1;
#endif
}

static int setExact(WebPConfig* webpConfig, int value) {
#if WEBP_ENCODER_ABI_VERSION < 0x209
	return 0;
#else
	webpConfig->exact = value;
	return 1;
#endif
}

*/
import "C"

import (
	"errors"
	"fmt"
	"image"
	"io"
	"sync"
	"unsafe"
)

// Config specifies WebP encoding configuration.
type Config struct {
	c C.WebPConfig
}

// ConfigPreset returns initialized configuration with given preset and quality
// factor.
func ConfigPreset(preset Preset, quality float32) (*Config, error) {
	c := &Config{}
	if C.WebPConfigPreset(&c.c, C.WebPPreset(preset), C.float(quality)) == 0 {
		return nil, errors.New("failed to initialize webp config")
	}
	return c, nil
}

// ConfigLosslessPreset returns initialized configuration for lossless encoding.
// Given level specifies desired efficiency level between 0 (fastest, lowest
// compression) and 9 (slower, best compression).
func ConfigLosslessPreset(level int) (*Config, error) {
	c := &Config{}
	if C.WebPConfigLosslessPreset(&c.c, C.int(level)) == 0 {
		return nil, errors.New("failed to initialize webp config")
	}
	return c, nil
}

// SetLossless sets lossless parameter that specifies whether to enable lossless
// encoding.
func (c *Config) SetLossless(v bool) {
	c.c.autofilter = boolToValue(v)
}

// Lossless returns lossless parameter flag whether to enable lossless encoding.
func (c *Config) Lossless() bool {
	return valueToBool(c.c.lossless)
}

// SetQuality sets encoding quality factor between 0 (smallest file) and 100
// (biggest).
func (c *Config) SetQuality(v float32) {
	c.c.quality = C.float(v)
}

// Quality returns encoding quality factor.
func (c *Config) Quality() float32 {
	return float32(c.c.quality)
}

// SetMethod sets method parameter that specifies quality/speed trade-off
// (0=fast, 6=slower-better).
func (c *Config) SetMethod(v float32) {
	c.c.method = C.int(v)
}

// Method returns method parameter.
func (c *Config) Method() int {
	return int(c.c.method)
}

// SetImageHint sets hint for image type. It is used to only lossless encoding
// for now.
func (c *Config) SetImageHint(v ImageHint) {
	c.c.image_hint = C.WebPImageHint(v)
}

// ImageHint returns hint parameter for image type.
func (c *Config) ImageHint() ImageHint {
	return ImageHint(c.c.image_hint)
}

// SetTargetPSNR sets target PSNR value that specifies the minimal distortion to
// try to achieve. If it sets 0, disable target PSNR.
func (c *Config) SetTargetPSNR(v float32) {
	c.c.target_PSNR = C.float(v)
}

// TargetPSNR returns target PSNR value.
func (c *Config) TargetPSNR() float32 {
	return float32(c.c.target_PSNR)
}

// SetSegments sets segments parameter that specifies the maximum number of
// segments to use, in [1..4].
func (c *Config) SetSegments(v int) {
	c.c.segments = C.int(v)
}

// Segments returns segments parameter.
func (c *Config) Segments() int {
	return int(c.c.segments)
}

// SetSNSStrength sets SNS strength parameter between 0 (off) and 100 (maximum).
func (c *Config) SetSNSStrength(v int) {
	c.c.sns_strength = C.int(v)
}

// SNSStrength returns SNS strength parameter.
func (c *Config) SNSStrength() int {
	return int(c.c.sns_strength)
}

// SetFilterStrength sets filter strength parameter between 0 (off) and 100
// (strongest).
func (c *Config) SetFilterStrength(v int) {
	c.c.filter_strength = C.int(v)
}

// FilterStrength returns filter strength parameter.
func (c *Config) FilterStrength() int {
	return int(c.c.filter_strength)
}

// SetFilterSharpness sets filter sharpness parameter between 0 (off) and 7
// (least sharp).
func (c *Config) SetFilterSharpness(v int) {
	c.c.filter_sharpness = C.int(v)
}

// FilterSharpness returns filter sharpness parameter.
func (c *Config) FilterSharpness() int {
	return int(c.c.filter_sharpness)
}

// SetFilterType sets filter type parameter.
func (c *Config) SetFilterType(v FilterType) {
	c.c.filter_type = C.int(v)
}

// FilterType returns filter type parameter.
func (c *Config) FilterType() FilterType {
	return FilterType(c.c.filter_type)
}

// SetAutoFilter sets auto filter flag that specifies whether to auto adjust
// filter strength.
func (c *Config) SetAutoFilter(v bool) {
	c.c.autofilter = boolToValue(v)
}

// AutoFilter returns auto filter flag.
func (c *Config) AutoFilter() bool {
	return valueToBool(c.c.autofilter)
}

// SetAlphaCompression sets alpha compression parameter.
func (c *Config) SetAlphaCompression(v int) {
	c.c.alpha_compression = C.int(v)
}

// AlphaCompression returns alpha compression parameter.
func (c *Config) AlphaCompression() int {
	return int(c.c.alpha_compression)
}

// SetAlphaFiltering sets alpha filtering parameter.
func (c *Config) SetAlphaFiltering(v int) {
	c.c.alpha_filtering = C.int(v)
}

// AlphaFiltering returns alpha filtering parameter.
func (c *Config) AlphaFiltering() int {
	return int(c.c.alpha_filtering)
}

// SetPass sets pass parameter that specifies number of entropy-analysis passes
// between 1 and 10.
func (c *Config) SetPass(v int) {
	c.c.pass = C.int(v)
}

// Pass returns pass parameter.
func (c *Config) Pass() int {
	return int(c.c.pass)
}

// SetPreprocessing sets preprocessing filter.
func (c *Config) SetPreprocessing(v Preprocessing) {
	c.c.preprocessing = C.int(v)
}

// Preprocessing returns preprocessing filter.
func (c *Config) Preprocessing() Preprocessing {
	return Preprocessing(c.c.preprocessing)
}

// SetPartitions sets partitions parameter.
func (c *Config) SetPartitions(v int) {
	c.c.partitions = C.int(v)
}

// Partitions returns partitions parameter.
func (c *Config) Partitions() int {
	return int(c.c.partitions)
}

// SetPartitionLimit returns partition limit parameter.
func (c *Config) SetPartitionLimit(v int) {
	c.c.partition_limit = C.int(v)
}

// PartitionLimit returns partition limit parameter.
func (c *Config) PartitionLimit() int {
	return int(c.c.partition_limit)
}

// SetEmulateJPEGSize sets flag whether the compression parameters remaps to
// match the expected output size from JPEG compression.
func (c *Config) SetEmulateJPEGSize(v bool) {
	c.c.emulate_jpeg_size = boolToValue(v)
}

// EmulateJPEGSize returns the flag whether to enable emulating JPEG size.
func (c *Config) EmulateJPEGSize() bool {
	return valueToBool(c.c.emulate_jpeg_size)
}

// SetThreadLevel sets thread level parameter. If non-zero value is specified,
// try and use multi-threaded encoding.
func (c *Config) SetThreadLevel(v int) {
	c.c.thread_level = C.int(v)
}

// ThreadLevel returns thread level parameter.
func (c *Config) ThreadLevel() int {
	return int(c.c.thread_level)
}

// SetLowMemory sets flag whether to reduce memory usage.
func (c *Config) SetLowMemory(v bool) {
	c.c.low_memory = boolToValue(v)
}

// LowMemory returns low memory flag.
func (c *Config) LowMemory() bool {
	return valueToBool(c.c.low_memory)
}

// SetNearLossless sets near lossless encoding factor between 0 (max loss) and
// 100 (disable near lossless encoding, default).
func (c *Config) SetNearLossless(v int) {
	if C.setNearLossless(&c.c, C.int(v)) == 0 {
		panic("near_lossless parameter is not supported")
	}
}

// NearLossless returns near lossless encoding factor.
func (c *Config) NearLossless() int {
	var v C.int
	if C.getNearLossless(&c.c, &v) == 0 {
		panic("near_lossless parameter is not supported")
	}
	return int(v)
}

// SetExact sets the flag whether to preserve the exact RGB values under
// transparent area.
func (c *Config) SetExact(v bool) {
	if C.setExact(&c.c, boolToValue(v)) == 0 {
		panic("exact parameter is not supported")
	}
}

// Exact returns exact flag.
func (c *Config) Exact() bool {
	var v C.int
	if C.getExact(&c.c, &v) == 0 {
		panic("exact parameter is not supported")
	}
	return valueToBool(v)
}

func boolToValue(v bool) C.int {
	if v {
		return 1
	}
	return 0
}

func valueToBool(v C.int) bool {
	if v > 0 || v < 0 {
		return true
	}
	return false
}

type destinationManager struct {
	writer io.Writer
}

var destinationManagerMapMutex sync.RWMutex
var destinationManagerMap = make(map[uintptr]*destinationManager)

// GetDestinationManagerMapLen returns the number of globally working sourceManagers for debug
func GetDestinationManagerMapLen() int {
	destinationManagerMapMutex.RLock()
	defer destinationManagerMapMutex.RUnlock()
	return len(destinationManagerMap)
}

func makeDestinationManager(w io.Writer, pic *C.WebPPicture) (mgr *destinationManager) {
	mgr = &destinationManager{writer: w}
	destinationManagerMapMutex.Lock()
	defer destinationManagerMapMutex.Unlock()
	destinationManagerMap[uintptr(unsafe.Pointer(pic))] = mgr
	return
}

func releaseDestinationManager(pic *C.WebPPicture) {
	destinationManagerMapMutex.Lock()
	defer destinationManagerMapMutex.Unlock()
	delete(destinationManagerMap, uintptr(unsafe.Pointer(pic)))
}

func getDestinationManager(pic *C.WebPPicture) *destinationManager {
	destinationManagerMapMutex.RLock()
	defer destinationManagerMapMutex.RUnlock()
	return destinationManagerMap[uintptr(unsafe.Pointer(pic))]
}

//export writeWebP
func writeWebP(data *C.uint8_t, size C.size_t, pic *C.WebPPicture) C.int {
	mgr := getDestinationManager(pic)
	bytes := C.GoBytes(unsafe.Pointer(data), C.int(size))
	_, err := mgr.writer.Write(bytes)
	if err != nil {
		return 0 // TODO: can't pass error message
	}
	return 1
}

// EncodeRGBA encodes and writes image.Image into the writer as WebP.
// Now supports image.RGBA or image.NRGBA.
func EncodeRGBA(w io.Writer, img image.Image, c *Config) (err error) {
	if err = validateConfig(c); err != nil {
		return
	}

	pic := C.malloc_WebPPicture()
	if pic == nil {
		return errors.New("Could not allocate webp picture")
	}
	defer C.free_WebPPicture(pic)

	makeDestinationManager(w, pic)
	defer releaseDestinationManager(pic)

	if C.WebPPictureInit(pic) == 0 {
		return errors.New("Could not initialize webp picture")
	}
	defer C.WebPPictureFree(pic)

	pic.use_argb = 1

	pic.width = C.int(img.Bounds().Dx())
	pic.height = C.int(img.Bounds().Dy())

	pic.writer = C.WebPWriterFunction(C.writeWebP)

	switch p := img.(type) {
	case *RGBImage:
		C.WebPPictureImportRGB(pic, (*C.uint8_t)(&p.Pix[0]), C.int(p.Stride))
	case *image.RGBA:
		C.WebPPictureImportRGBA(pic, (*C.uint8_t)(&p.Pix[0]), C.int(p.Stride))
	case *image.NRGBA:
		C.WebPPictureImportRGBA(pic, (*C.uint8_t)(&p.Pix[0]), C.int(p.Stride))
	default:
		return errors.New("unsupported image type")
	}

	if C.WebPEncode(&c.c, pic) == 0 {
		return fmt.Errorf("Encoding error: %d", pic.error_code)
	}

	return
}

// EncodeYUVA encodes and writes YUVA Image data into the writer as WebP.
func EncodeYUVA(w io.Writer, img *YUVAImage, c *Config) (err error) {
	if err = validateConfig(c); err != nil {
		return
	}

	pic := C.malloc_WebPPicture()
	if pic == nil {
		return errors.New("Could not allocate webp picture")
	}
	defer C.free_WebPPicture(pic)

	makeDestinationManager(w, pic)
	defer releaseDestinationManager(pic)

	if C.WebPPictureInit(pic) == 0 {
		return errors.New("Could not initialize webp picture")
	}
	defer C.WebPPictureFree(pic)

	pic.use_argb = 0
	pic.colorspace = C.WebPEncCSP(img.ColorSpace)
	pic.width = C.int(img.Rect.Dx())
	pic.height = C.int(img.Rect.Dy())
	pic.y = (*C.uint8_t)(&img.Y[0])
	pic.u = (*C.uint8_t)(&img.Cb[0])
	pic.v = (*C.uint8_t)(&img.Cr[0])
	pic.y_stride = C.int(img.YStride)
	pic.uv_stride = C.int(img.CStride)

	if img.ColorSpace == YUV420A {
		pic.a = (*C.uint8_t)(&img.A[0])
		pic.a_stride = C.int(img.AStride)
	}

	pic.writer = C.WebPWriterFunction(C.writeWebP)
	pic.custom_ptr = unsafe.Pointer(&destinationManager{writer: w})

	if C.WebPEncode(&c.c, pic) == 0 {
		return fmt.Errorf("Encoding error: %d", pic.error_code)
	}

	return
}

func validateConfig(c *Config) error {
	if C.WebPValidateConfig(&c.c) == 0 {
		return errors.New("invalid configuration")
	}
	return nil
}
