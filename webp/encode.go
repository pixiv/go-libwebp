package webp

/*
#include <stdlib.h>
#include <string.h>
#include <webp/encode.h>

int golibwebpWriteWebP(uint8_t*, size_t, struct WebPPicture*);
int golibwebpProgressHook(int, struct WebPPicture*);

static WebPPicture *calloc_WebPPicture(void) {
	return calloc(sizeof(WebPPicture), 1);
}

static void free_WebPPicture(WebPPicture* webpPicture) {
	free(webpPicture);
}

static int webpEncodeYUVA(const WebPConfig *config, WebPPicture *picture, uint8_t *y, uint8_t *u, uint8_t *v, uint8_t *a) {
	picture->y = y;
	picture->u = u;
	picture->v = v;
	if (picture->colorspace == WEBP_YUV420A) {
		picture->a = a;
	}
	picture->writer = (WebPWriterFunction)golibwebpWriteWebP;
	picture->progress_hook = (WebPProgressHook)golibwebpProgressHook;

  return WebPEncode(config, picture);
}

static int webpEncodeGray(const WebPConfig *config, WebPPicture *picture, uint8_t *y) {
	int ok = 0;
	const int c_width = (picture->width + 1) >> 1;
	const int c_height = (picture->height + 1) >> 1;
	const int c_stride = c_width;
	const int c_size = c_stride * c_height;
	const int gray = 128;
	uint8_t* chroma;

	chroma = malloc(c_size);
	if (!chroma) {
		return 0;
	}
	memset(chroma, gray, c_size);

	picture->y = y;
	picture->u = chroma;
	picture->v = chroma;
	picture->uv_stride = c_stride;
	picture->writer = (WebPWriterFunction)golibwebpWriteWebP;
	picture->progress_hook = (WebPProgressHook)golibwebpProgressHook;

	ok = WebPEncode(config, picture);

	free(chroma);

	return ok;
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

type ProgressHook func(int) bool

type EncodeError struct {
	encodeErrorCode EncodeErrorCode
}

func (e *EncodeError) Error() string {
	return fmt.Sprintf("Encoding error: %d", e.encodeErrorCode)
}

func (e *EncodeError) EncodeErrorCode() EncodeErrorCode {
	return e.encodeErrorCode
}

var _ error = &EncodeError{}

type EncodeErrorCode int

const (
	EncodeErrorCodeVP8EncOK                        EncodeErrorCode = C.VP8_ENC_OK
	EncodeErrorCodeVP8EncErrorOutOfMemory          EncodeErrorCode = C.VP8_ENC_ERROR_OUT_OF_MEMORY
	EncodeErrorCodeVP8EncErrorBitstreamOutOfMemory EncodeErrorCode = C.VP8_ENC_ERROR_BITSTREAM_OUT_OF_MEMORY
	EncodeErrorCodeVP8EncErrorNullParameter        EncodeErrorCode = C.VP8_ENC_ERROR_NULL_PARAMETER
	EncodeErrorCodeVP8EncErrorInvalidConfiguration EncodeErrorCode = C.VP8_ENC_ERROR_INVALID_CONFIGURATION
	EncodeErrorCodeVP8EncErrorBadDimension         EncodeErrorCode = C.VP8_ENC_ERROR_BAD_DIMENSION
	EncodeErrorCodeVP8EncErrorPartition0Overflow   EncodeErrorCode = C.VP8_ENC_ERROR_PARTITION0_OVERFLOW
	EncodeErrorCodeVP8EncErrorPartitionOverflow    EncodeErrorCode = C.VP8_ENC_ERROR_PARTITION_OVERFLOW
	EncodeErrorCodeVP8EncErrorBadWrite             EncodeErrorCode = C.VP8_ENC_ERROR_BAD_WRITE
	EncodeErrorCodeVP8EncErrorFileTooBig           EncodeErrorCode = C.VP8_ENC_ERROR_FILE_TOO_BIG
	EncodeErrorCodeVP8EncErrorUserAbort            EncodeErrorCode = C.VP8_ENC_ERROR_USER_ABORT
	EncodeErrorCodeVP8ErrorLast                    EncodeErrorCode = C.VP8_ENC_ERROR_LAST
)

var errWebPPictureAllocate = errors.New("Could not allocate webp picture")
var errWebPPictureInitialize = errors.New("Could not initialize webp picture")
var errUnsupportedImageType = errors.New("unsupported image type")
var errInvalidConfiguration = errors.New("invalid configuration")
var errInitializeWebPConfig = errors.New("failed to initialize webp config")

// ConfigPreset returns initialized configuration with given preset and quality
// factor.
func ConfigPreset(preset Preset, quality float32) (*Config, error) {
	c := &Config{}
	if C.WebPConfigPreset(&c.c, C.WebPPreset(preset), C.float(quality)) == 0 {
		return nil, errInitializeWebPConfig
	}
	return c, nil
}

// ConfigLosslessPreset returns initialized configuration for lossless encoding.
// Given level specifies desired efficiency level between 0 (fastest, lowest
// compression) and 9 (slower, best compression).
func ConfigLosslessPreset(level int) (*Config, error) {
	c := &Config{}
	if C.WebPConfigPreset(&c.c, C.WebPPreset(PresetDefault), C.float(0)) == 0 {
		return nil, errInitializeWebPConfig
	}
	if C.WebPConfigLosslessPreset(&c.c, C.int(level)) == 0 {
		return nil, errInitializeWebPConfig
	}
	return c, nil
}

// SetLossless sets lossless parameter that specifies whether to enable lossless
// encoding.
func (c *Config) SetLossless(v bool) {
	c.c.lossless = boolToValue(v)
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
func (c *Config) SetMethod(v int) {
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

// SetAlphaQuality sets alpha quality parameter.
func (c *Config) SetAlphaQuality(v int) {
	c.c.alpha_quality = C.int(v)
}

// AlphaQuality returns alpha quality parameter.
func (c *Config) AlphaQuality() int {
	return int(c.c.alpha_quality)
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
	c.c.near_lossless = C.int(v)
}

// NearLossless returns near lossless encoding factor.
func (c *Config) NearLossless() int {
	return int(c.c.near_lossless)
}

// SetExact sets the flag whether to preserve the exact RGB values under
// transparent area.
func (c *Config) SetExact(v bool) {
	c.c.exact = boolToValue(v)
}

// Exact returns exact flag.
func (c *Config) Exact() bool {
	return valueToBool(c.c.exact)
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
	writer       io.Writer
	progressHook ProgressHook
}

var destinationManagerMapMutex sync.RWMutex
var destinationManagerMap = make(map[uintptr]*destinationManager)

// GetDestinationManagerMapLen returns the number of globally working sourceManagers for debug
func GetDestinationManagerMapLen() int {
	destinationManagerMapMutex.RLock()
	defer destinationManagerMapMutex.RUnlock()
	return len(destinationManagerMap)
}

func makeDestinationManager(w io.Writer, progressHook ProgressHook, pic *C.WebPPicture) (mgr *destinationManager) {
	mgr = &destinationManager{writer: w, progressHook: progressHook}
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

//export golibwebpWriteWebP
func golibwebpWriteWebP(data *C.uint8_t, size C.size_t, pic *C.WebPPicture) C.int {
	mgr := getDestinationManager(pic)
	bytes := C.GoBytes(unsafe.Pointer(data), C.int(size))
	_, err := mgr.writer.Write(bytes)
	if err != nil {
		return 0 // TODO: can't pass error message
	}
	return 1
}

//export golibwebpProgressHook
func golibwebpProgressHook(percent C.int, pic *C.WebPPicture) C.int {
	mgr := getDestinationManager(pic)
	shouldContinue := true
	if mgr.progressHook != nil {
		func() {
			defer func() {
				if r := recover(); r != nil {
					shouldContinue = false
				}
			}()
			shouldContinue = mgr.progressHook(int(percent))
		}()
	}

	return boolToValue(shouldContinue)
}

// EncodeRGBA encodes and writes image.Image into the writer as WebP.
// Now supports image.RGBA or image.NRGBA.
func EncodeRGBA(w io.Writer, img image.Image, c *Config) (err error) {
	return EncodeRGBAWithProgress(w, img, c, nil)
}

// EncodeRGBAWithProgress encodes and writes image.Image into the writer as WebP.
// Now supports image.RGBA or image.NRGBA.
// This function accepts progress hook function and supports cancellation.
func EncodeRGBAWithProgress(w io.Writer, img image.Image, c *Config, progressHook ProgressHook) (err error) {
	if err = ValidateConfig(c); err != nil {
		return
	}

	pic := C.calloc_WebPPicture()
	if pic == nil {
		return errWebPPictureAllocate
	}
	defer C.free_WebPPicture(pic)

	makeDestinationManager(w, progressHook, pic)
	defer releaseDestinationManager(pic)

	if C.WebPPictureInit(pic) == 0 {
		return errWebPPictureInitialize
	}
	defer C.WebPPictureFree(pic)

	pic.use_argb = 1

	pic.width = C.int(img.Bounds().Dx())
	pic.height = C.int(img.Bounds().Dy())

	pic.progress_hook = C.WebPProgressHook(C.golibwebpProgressHook)
	pic.writer = C.WebPWriterFunction(C.golibwebpWriteWebP)

	switch p := img.(type) {
	case *RGBImage:
		C.WebPPictureImportRGB(pic, (*C.uint8_t)(&p.Pix[0]), C.int(p.Stride))
	case *image.RGBA:
		C.WebPPictureImportRGBA(pic, (*C.uint8_t)(&p.Pix[0]), C.int(p.Stride))
	case *image.NRGBA:
		C.WebPPictureImportRGBA(pic, (*C.uint8_t)(&p.Pix[0]), C.int(p.Stride))
	default:
		return errUnsupportedImageType
	}

	if C.WebPEncode(&c.c, pic) == 0 {
		return &EncodeError{encodeErrorCode: EncodeErrorCode(pic.error_code)}
	}

	return
}

// EncodeGray encodes and writes Gray Image data into the writer as WebP.
func EncodeGray(w io.Writer, p *image.Gray, c *Config) (err error) {
	return EncodeGrayWithProgress(w, p, c, nil)
}

// EncodeGrayWithProgress encodes and writes Gray Image data into the writer as WebP.
// This function accepts progress hook function and supports cancellation.
func EncodeGrayWithProgress(w io.Writer, p *image.Gray, c *Config, progressHook ProgressHook) (err error) {
	if err = ValidateConfig(c); err != nil {
		return
	}

	pic := C.calloc_WebPPicture()
	if pic == nil {
		return errWebPPictureAllocate
	}
	defer C.free_WebPPicture(pic)

	makeDestinationManager(w, progressHook, pic)
	defer releaseDestinationManager(pic)

	if C.WebPPictureInit(pic) == 0 {
		return errWebPPictureInitialize
	}
	defer C.WebPPictureFree(pic)

	pic.use_argb = 0
	pic.width = C.int(p.Rect.Dx())
	pic.height = C.int(p.Rect.Dy())
	pic.y_stride = C.int(p.Stride)

	if C.webpEncodeGray(&c.c, pic, (*C.uint8_t)(&p.Pix[0])) == 0 {
		return &EncodeError{encodeErrorCode: EncodeErrorCode(pic.error_code)}
	}

	return
}

// EncodeYUVA encodes and writes YUVA Image data into the writer as WebP.
func EncodeYUVA(w io.Writer, img *YUVAImage, c *Config) (err error) {
	return EncodeYUVAWithProgress(w, img, c, nil)
}

// EncodeYUVAWithProgress encodes and writes YUVA Image data into the writer as WebP.
// This function accepts progress hook function and supports cancellation.
func EncodeYUVAWithProgress(w io.Writer, img *YUVAImage, c *Config, progressHook ProgressHook) (err error) {
	if err = ValidateConfig(c); err != nil {
		return
	}

	pic := C.calloc_WebPPicture()
	if pic == nil {
		return errWebPPictureAllocate
	}
	defer C.free_WebPPicture(pic)

	makeDestinationManager(w, progressHook, pic)
	defer releaseDestinationManager(pic)

	if C.WebPPictureInit(pic) == 0 {
		return errWebPPictureInitialize
	}
	defer C.WebPPictureFree(pic)

	pic.use_argb = 0
	pic.colorspace = C.WebPEncCSP(img.ColorSpace)
	pic.width = C.int(img.Rect.Dx())
	pic.height = C.int(img.Rect.Dy())
	pic.y_stride = C.int(img.YStride)
	pic.uv_stride = C.int(img.CStride)
	var a *C.uint8_t
	y, u, v := (*C.uint8_t)(&img.Y[0]), (*C.uint8_t)(&img.Cb[0]), (*C.uint8_t)(&img.Cr[0])
	if img.ColorSpace == YUV420A {
		pic.a_stride = C.int(img.AStride)
		a = (*C.uint8_t)(&img.A[0])
	}

	if C.webpEncodeYUVA(&c.c, pic, y, u, v, a) == 0 {
		return &EncodeError{encodeErrorCode: EncodeErrorCode(pic.error_code)}
	}
	return
}

func ValidateConfig(c *Config) error {
	if C.WebPValidateConfig(&c.c) == 0 {
		return errInvalidConfiguration
	}
	return nil
}
