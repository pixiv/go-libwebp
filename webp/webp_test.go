package webp_test

import (
	"bufio"
	"fmt"
	"image"
	"os"
	"testing"

	"github.com/harukasan/go-libwebp/test/util"
	"github.com/harukasan/go-libwebp/webp"
	"image/color"
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

	config := webp.Config{
		Preset:  webp.PresetDefault,
		Quality: 100,
		Method:  6,
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

func TestEncodeRGB(t *testing.T) {
	img := util.ReadPNG("yellow-rose-3.png")

	config := webp.Config{
		Preset:  webp.PresetDefault,
		Quality: 100,
		Method:  6,
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

func TestImageInterface(t *testing.T) {
	rect := image.Rect(0, 0, 100, 100)
	img := webp.NewRGBImage(rect)

	if got := img.ColorModel(); got != webp.ColorModel {
		t.Errorf("ColorModel() should return rgb.ColorModel, got: %v", got)
	}

	if got := img.Bounds(); got != rect {
		t.Errorf("Bounds() should return %v, got: %v", rect, got)
	}

	black := color.RGBA{0x00, 0x00, 0x00, 0xFF}
	if got := img.At(0, 0); got != black {
		t.Errorf("At(0, 0) should return %v, got: %v", black, got)
	}

	blank := color.RGBA{}
	if got := img.At(-1, -1); got != blank {
		t.Errorf("At(0, 0) should return %v, got: %v", blank, got)
	}
}

func TestConvertFromRGBA(t *testing.T) {
	rgba := color.RGBA{0x11, 0x22, 0x33, 0xFF}
	expect := webp.RGB{0x11, 0x22, 0x33}
	if got := webp.ColorModel.Convert(rgba); got != expect {
		t.Errorf("got: %v, expect: %v", got, expect)
	}
}

func TestConvertFromRGB(t *testing.T) {
	c := webp.RGB{0x11, 0x22, 0x33}
	if got := webp.ColorModel.Convert(c); got != c {
		t.Errorf("got: %v, expect: %v", got, c)
	}
}

func TestColorRGBA(t *testing.T) {
	c := webp.RGB{0x11, 0x22, 0x33}
	r, g, b, a := uint32(0x1111), uint32(0x2222), uint32(0x3333), uint32(0xFFFF)

	gotR, gotG, gotB, gotA := c.RGBA()
	if gotR != r {
		t.Errorf("got R: %v, expect R: %v", gotR, r)
	}
	if gotG != g {
		t.Errorf("got G: %v, expect G: %v", gotG, g)
	}
	if gotB != b {
		t.Errorf("got B: %v, expect B: %v", gotB, b)
	}
	if gotA != a {
		t.Errorf("got A: %v, expect A: %v", gotA, a)
	}
}
