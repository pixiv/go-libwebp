package webp_test

import (
	"bufio"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/harukasan/go-libwebp/webp"
)

//
// Utitlities of input/output example images
//

// GetExFilePath returns the path of specified example file.
func GetExFilePath(name string) string {
	return filepath.Join(os.Getenv("GOPATH"), "src/github.com/harukasan/go-libwebp/examples/images", name)
}

// OpenExFile opens specified example file
func OpenExFile(name string) (io io.Reader) {
	io, err := os.Open(GetExFilePath(name))
	if err != nil {
		panic(err)
	}
	return
}

// ReadExFile reads and returns data bytes of specified example file.
func ReadExFile(name string) (data []byte) {
	data, err := ioutil.ReadFile(GetExFilePath(name))
	if err != nil {
		panic(err)
	}
	return
}

// CreateExOutFile creates output file into example/out directory.
func CreateExOutFile(name string) (file *os.File) {
	// Make output directory
	dir := filepath.Join(os.Getenv("GOPATH"), "src/github.com/harukasan/go-libwebp/examples/out")
	if err := os.Mkdir(dir, 0755); err != nil && os.IsNotExist(err) {
		panic(err)
	}

	// Create example output file
	file, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		panic(err)
	}
	return
}

//
// Decode
//

// Test Get Decoder Version
func TestGetDecoderVersion(t *testing.T) {
	v := webp.GetDecoderVersion()
	if v < 0 {
		t.Errorf("GetDecoderVersion should returns positive version number, got %v\n", v)
	}
}

func TestGetInfo(t *testing.T) {
	data := ReadExFile("cosmos.webp")
	width, height := webp.GetInfo(data)

	if width != 1024 {
		t.Errorf("Expected width: %d, but got %d", 1024, width)
	}
	if height != 768 {
		t.Errorf("Expected height: %d, but got %d", 768, height)
	}
}

func TestGetFeatures(t *testing.T) {
	data := ReadExFile("cosmos.webp")
	f, err := webp.GetFeatures(data)
	if err != nil {
		t.Errorf("Got Error: %v", err)
		return
	}
	if got := f.Width; got != 1024 {
		t.Errorf("Expected Width: %v, but got %v", 1024, got)
	}
	if got := f.Height; got != 768 {
		t.Errorf("Expected Width: %v, but got %v", 768, got)
	}
	if got := f.HasAlpha; got != false {
		t.Errorf("Expected HasAlpha: %v, but got %v", false, got)
	}
	if got := f.HasAnimation; got != false {
		t.Errorf("Expected HasAlpha: %v, but got %v", false, got)
	}
	if got := f.Format; got != 1 {
		t.Errorf("Expected Format: %v, but got %v", 1, got)
	}
}

func TestDecodeYUV(t *testing.T) {
	files := []string{
		"cosmos.webp",
		"butterfly.webp",
		"kinkaku.webp",
		"yellow-rose-3.webp",
	}

	for _, file := range files {
		data := ReadExFile(file)
		options := &webp.DecoderOptions{}

		_, err := webp.DecodeYUVA(data, options)
		if err != nil {
			t.Errorf("Got Error: %v", err)
			return
		}
	}
}

func TestDecodeRGBA(t *testing.T) {
	files := []string{
		"cosmos.webp",
		"butterfly.webp",
		"kinkaku.webp",
		"yellow-rose-3.webp",
	}

	for _, file := range files {
		data := ReadExFile(file)
		options := &webp.DecoderOptions{}

		_, err := webp.DecodeRGBA(data, options)
		if err != nil {
			t.Errorf("Got Error: %v", err)
			return
		}
	}
}

func TestDecodeRGBAWithCropping(t *testing.T) {
	data := ReadExFile("cosmos.webp")
	crop := image.Rect(100, 100, 300, 200)

	options := &webp.DecoderOptions{
		Crop: crop,
	}

	img, err := webp.DecodeRGBA(data, options)
	if err != nil {
		t.Errorf("Got Error: %v", err)
		return
	}
	if img.Rect.Dx() != crop.Dx() || img.Rect.Dy() != crop.Dy() {
		t.Errorf("Decoded image should cropped to %v, but got %v", crop, img.Rect)
	}
}

func TestDecodeRGBAWithScaling(t *testing.T) {
	data := ReadExFile("cosmos.webp")
	scale := image.Rect(0, 0, 640, 480)

	options := &webp.DecoderOptions{
		Scale: scale,
	}

	img, err := webp.DecodeRGBA(data, options)
	if err != nil {
		t.Errorf("Got Error: %v", err)
		return
	}
	if img.Rect.Dx() != scale.Dx() || img.Rect.Dy() != scale.Dy() {
		t.Errorf("Decoded image should scaled to %v, but got %v", scale, img.Rect)
	}
}

//
// Encoding
//

func TestEncodeRGBA(t *testing.T) {
	img, _ := png.Decode(OpenExFile("yellow-rose-3.png"))

	config := webp.Config{
		Preset:  webp.PresetDefault,
		Quality: 100,
		Method:  6,
	}

	f := CreateExOutFile("TestEncodeRGBA.webp")
	w := bufio.NewWriter(f)
	defer func() {
		w.Flush()
		f.Close()
	}()

	if err := webp.EncodeRGBA(w, img, config); err != nil {
		t.Errorf("Got Error: %v", err)
		return
	}
}

func TestEncodeYUVA(t *testing.T) {
	data := ReadExFile("cosmos.webp")
	options := &webp.DecoderOptions{}

	img, err := webp.DecodeYUVA(data, options)
	if err != nil {
		t.Errorf("Got Error: %v in decoding", err)
		return
	}

	f := CreateExOutFile("TestEncodeYUVA.webp")
	w := bufio.NewWriter(f)
	defer func() {
		w.Flush()
		f.Close()
	}()

	config := webp.Config{
		Preset:  webp.PresetDefault,
		Quality: 100,
		Method:  6,
	}

	if err := webp.EncodeYUVA(w, img, config); err != nil {
		t.Errorf("Got Error: %v", err)
		return
	}
}
