package webp_test

import (
	"fmt"
	"testing"

	"github.com/tidbyt/go-libwebp/test/util"
	"github.com/tidbyt/go-libwebp/webp"
)

func TestDecodeAnimationInfo(t *testing.T) {
	data := util.ReadFile("weather-anim.webp")

	dec, err := webp.NewAnimationDecoder(data)
	if err != nil {
		t.Fatalf("initializing decoder: %v", err)
	}
	defer dec.Close()

	info, err := dec.GetInfo()
	if err != nil {
		t.Fatalf("getting animatiion info: %v", err)
	}

	if got := info.CanvasWidth; got != 64 {
		t.Errorf("Expected CanvasWidth: %v, but got %v", 64, got)
	}
	if got := info.CanvasHeight; got != 32 {
		t.Errorf("Expected CanvasHeight: %v, but got %v", 32, got)
	}
	if got := info.LoopCount; got != 0 {
		t.Errorf("Expected LoopCount: %v, but got %v", 0, got)
	}
	if got := info.FrameCount; got != 18 {
		t.Errorf("Expected FrameCount: %v, but got %v", 18, got)
	}
}

func TestDecodeAnimation(t *testing.T) {
	data := util.ReadFile("weather-anim.webp")

	dec, err := webp.NewAnimationDecoder(data)
	if err != nil {
		t.Fatalf("initializing decoder: %v", err)
	}
	defer dec.Close()

	anim, err := dec.Decode()
	if err != nil {
		t.Fatalf("error decoding: %v", err)
	}

	if got := len(anim.Image); got != 18 {
		t.Errorf("Expected len(Image): %v, but got %v", 18, got)
	}
	if got := len(anim.Timestamp); got != 18 {
		t.Errorf("Expected len(Timestamp): %v, but got %v", 18, got)
	}

	for i := 0; i < 18; i++ {
		frame := util.ReadPNG(fmt.Sprintf("weather-anim-frames/%02d.png", i))
		got := anim.Image[i]

		for y := 0; y < got.Bounds().Dy(); y++ {
			for x := 0; x < got.Bounds().Dx(); x++ {
				if got.At(x, y) != frame.At(x, y) {
					t.Fatalf(
						"Expected frame %d (%d,%d): %v, but got %v",
						i, x, y,
						frame.At(x, y),
						got.At(x, y),
					)
				}
			}
		}
	}
}
