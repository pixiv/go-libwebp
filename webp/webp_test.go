package webp_test

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"os"
	"reflect"
	"testing"

	"github.com/pixiv/go-libwebp/test/util"
	"github.com/pixiv/go-libwebp/webp"
)

func TestMain(m *testing.M) {
	result := m.Run()
	if webp.GetDestinationManagerMapLen() > 0 {
		fmt.Println("destinationManager leaked")
		result = 2
	}
	os.Exit(result)
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
	data := util.ReadFile("cosmos.webp")
	width, height := webp.GetInfo(data)

	if width != 1024 {
		t.Errorf("Expected width: %d, but got %d", 1024, width)
	}
	if height != 768 {
		t.Errorf("Expected height: %d, but got %d", 768, height)
	}
}

func TestGetFeatures(t *testing.T) {
	data := util.ReadFile("cosmos.webp")
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
		data := util.ReadFile(file)
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
		data := util.ReadFile(file)
		options := &webp.DecoderOptions{}

		_, err := webp.DecodeRGBA(data, options)
		if err != nil {
			t.Errorf("Got Error: %v", err)
			return
		}
	}
}

func TestDecodeNRGBA(t *testing.T) {
	files := []string{
		"cosmos.webp",
		"butterfly.webp",
		"kinkaku.webp",
		"yellow-rose-3.webp",
	}

	for _, file := range files {
		data := util.ReadFile(file)
		options := &webp.DecoderOptions{}

		_, err := webp.DecodeNRGBA(data, options)
		if err != nil {
			t.Errorf("Got Error: %v", err)
			return
		}
	}
}

func TestDecodeRGBAWithCropping(t *testing.T) {
	data := util.ReadFile("cosmos.webp")
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
	data := util.ReadFile("cosmos.webp")
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
	img := util.ReadPNG("yellow-rose-3.png")
	if _, ok := img.(*image.NRGBA); !ok {
		t.Fatalf("image is not NRGBA: %v", reflect.TypeOf(img))
	}

	config, err := webp.ConfigPreset(webp.PresetDefault, 100)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}

	f := util.CreateFile("TestEncodeRGBA.webp")
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

func TestEncodeRGBAWithProgress(t *testing.T) {
	img := util.ReadPNG("yellow-rose-3.png")
	if _, ok := img.(*image.NRGBA); !ok {
		t.Fatalf("image is not NRGBA: %v", reflect.TypeOf(img))
	}

	config, err := webp.ConfigPreset(webp.PresetDefault, 100)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}

	f := util.CreateFile("TestEncodeRGBAWithProgress.webp")
	w := bufio.NewWriter(f)
	defer func() {
		w.Flush()
		f.Close()
	}()

	if err := webp.EncodeRGBAWithProgress(w, img, config, func(i int) bool {
		t.Logf("Progress: %v", i)
		return true
	}); err != nil {
		t.Errorf("Got Error: %v", err)
		return
	}
}

func convertToRGBImage(t *testing.T, origImg image.Image) *webp.RGBImage {
	bounds := origImg.Bounds()
	img := webp.NewRGBImage(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba := origImg.At(x, y)
			r, g, b, _ := rgba.RGBA()
			img.Pix[y*img.Stride+x*3+0] = uint8(r >> 8)
			img.Pix[y*img.Stride+x*3+1] = uint8(g >> 8)
			img.Pix[y*img.Stride+x*3+2] = uint8(b >> 8)
		}
	}

	return img
}

func TestEncodeRGB(t *testing.T) {
	origImg := util.ReadPNG("yellow-rose-3.png")
	img := convertToRGBImage(t, origImg)

	config, err := webp.ConfigPreset(webp.PresetDefault, 100)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}

	f := util.CreateFile("TestEncodeRGB.webp")
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

func TestEncodeRGBWithProgress(t *testing.T) {
	origImg := util.ReadPNG("yellow-rose-3.png")
	img := convertToRGBImage(t, origImg)

	config, err := webp.ConfigPreset(webp.PresetDefault, 100)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}

	f := util.CreateFile("TestEncodeRGBWithProgress.webp")
	w := bufio.NewWriter(f)
	defer func() {
		w.Flush()
		f.Close()
	}()

	if err := webp.EncodeRGBAWithProgress(w, img, config, func(i int) bool {
		t.Logf("Progress: %v", i)
		return true
	}); err != nil {
		t.Errorf("Got Error: %v", err)
		return
	}
}

func TestEncodeYUVA(t *testing.T) {
	data := util.ReadFile("cosmos.webp")
	options := &webp.DecoderOptions{}

	img, err := webp.DecodeYUVA(data, options)
	if err != nil {
		t.Errorf("Got Error: %v in decoding", err)
		return
	}

	f := util.CreateFile("TestEncodeYUVA.webp")
	w := bufio.NewWriter(f)
	defer func() {
		w.Flush()
		f.Close()
	}()

	config, err := webp.ConfigPreset(webp.PresetDefault, 100)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}

	if err := webp.EncodeYUVA(w, img, config); err != nil {
		t.Errorf("Got Error: %v", err)
		return
	}
}

func TestEncodeYUVAWithProgress(t *testing.T) {
	data := util.ReadFile("cosmos.webp")
	options := &webp.DecoderOptions{}

	img, err := webp.DecodeYUVA(data, options)
	if err != nil {
		t.Errorf("Got Error: %v in decoding", err)
		return
	}

	f := util.CreateFile("TestEncodeYUVA.webp")
	w := bufio.NewWriter(f)
	defer func() {
		w.Flush()
		f.Close()
	}()

	config, err := webp.ConfigPreset(webp.PresetDefault, 100)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}

	if err := webp.EncodeYUVAWithProgress(w, img, config, func(i int) bool {
		t.Logf("Progress: %v", i)
		return true
	}); err != nil {
		t.Errorf("Got Error: %v", err)
		return
	}
}

func TestEncodeGray(t *testing.T) {
	p := image.NewGray(image.Rect(0, 0, 1, 10))
	for i := 0; i < 10; i++ {
		p.SetGray(0, i, color.Gray{uint8(float32(i) / 10 * 255)})
	}

	f := util.CreateFile("TestEncodeGray.webp")
	w := bufio.NewWriter(f)
	defer func() {
		w.Flush()
		f.Close()
	}()

	config, err := webp.ConfigPreset(webp.PresetDefault, 100)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}

	if err := webp.EncodeGray(w, p, config); err != nil {
		t.Errorf("Got Error: %v", err)
		return
	}
}

func TestEncodeGrayWithProgress(t *testing.T) {
	p := image.NewGray(image.Rect(0, 0, 1, 10))
	for i := 0; i < 10; i++ {
		p.SetGray(0, i, color.Gray{uint8(float32(i) / 10 * 255)})
	}

	f := util.CreateFile("TestEncodeGray.webp")
	w := bufio.NewWriter(f)
	defer func() {
		w.Flush()
		f.Close()
	}()

	config, err := webp.ConfigPreset(webp.PresetDefault, 100)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}

	if err := webp.EncodeGrayWithProgress(w, p, config, func(i int) bool {
		t.Logf("Progress: %v", i)
		return true
	}); err != nil {
		t.Errorf("Got Error: %v", err)
		return
	}
}
