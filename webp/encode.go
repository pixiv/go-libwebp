package webp

/*
#cgo LDFLAGS: -lwebp

#include <stdlib.h>
#include <webp/encode.h>

int writeWebP(uint8_t*, size_t, struct WebPPicture*);

static WebPPicture *malloc_WebPPicture(void) {
	return malloc(sizeof(WebPPicture));
}

static void free_WebPPicture(WebPPicture* webpPicture) {
	free(webpPicture);
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

	"github.com/pixiv/go-libjpeg/rgb"
)

// Config specifies WebP encoding configuration.
type Config struct {
	Preset          Preset     // Parameters Preset
	Lossless        bool       // True if use lossless encoding
	Quality         float32    // WebP quality factor, 0-100
	Method          int        // Quality/Speed trade-off factor, 0=faster / 6=slower-better
	TargetSize      int        // Target size of encoded file in bytes
	TargetPSNR      float32    // Target PSNR, takes precedence over TargetSize
	Segments        int        // Maximum number of segments to use, 1..4
	SNSStrength     int        // Strength of spartial noise shaping, 0..100=maximum
	FilterStrength  int        // Strength of filter, 0..100=strength
	FilterSharpness int        // Sharpness of filter, 0..7=sharpness
	FilterType      FilterType // Filtering type
	Pass            int        // Number of entropy-analysis passes, 0..100
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
func EncodeRGBA(w io.Writer, img image.Image, c Config) (err error) {
	webpConfig, err := initConfig(c)
	if err != nil {
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
	case *rgb.Image:
		C.WebPPictureImportRGB(pic, (*C.uint8_t)(&p.Pix[0]), C.int(p.Stride))
	case *image.RGBA:
		C.WebPPictureImportRGBA(pic, (*C.uint8_t)(&p.Pix[0]), C.int(p.Stride))
	case *image.NRGBA:
		C.WebPPictureImportRGBA(pic, (*C.uint8_t)(&p.Pix[0]), C.int(p.Stride))
	default:
		return errors.New("unsupported image type")
	}

	if C.WebPEncode(webpConfig, pic) == 0 {
		return fmt.Errorf("Encoding error: %d", pic.error_code)
	}

	return
}

// EncodeYUVA encodes and writes YUVA Image data into the writer as WebP.
func EncodeYUVA(w io.Writer, img *YUVAImage, c Config) (err error) {
	webpConfig, err := initConfig(c)
	if err != nil {
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

	if C.WebPEncode(webpConfig, pic) == 0 {
		return fmt.Errorf("Encoding error: %d", pic.error_code)
	}

	return
}

// initConfig initializes C.WebPConfig with encoding parameters.
func initConfig(c Config) (config *C.WebPConfig, err error) {
	config = &C.WebPConfig{}
	if C.WebPConfigPreset(config, C.WebPPreset(c.Preset), C.float(c.Quality)) == 0 {
		return nil, errors.New("Could not initialize configuration with preset")
	}
	config.target_size = C.int(c.TargetSize)
	config.target_PSNR = C.float(c.TargetPSNR)
	if c.Lossless {
		config.lossless = C.int(1)
	}

	if C.WebPValidateConfig(config) == 0 {
		return nil, errors.New("Invalid configuration")
	}
	return
}
