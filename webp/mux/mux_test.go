package mux_test

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"testing"

	"github.com/pixiv/go-libwebp/test/util"
	"github.com/pixiv/go-libwebp/webp"
	"github.com/pixiv/go-libwebp/webp/mux"
)

func TestGetICCProfileMissing(t *testing.T) {
	data := util.ReadFile("cosmos.webp")
	got, err := mux.GetICCProfile(data)
	if err != nil {
		t.Fatalf("GetICCProfile: %v", err)
	}
	if got != nil {
		t.Errorf("GetICCProfile: got %d bytes, want nil", len(got))
	}
}

func TestGetICCProfileEmpty(t *testing.T) {
	if _, err := mux.GetICCProfile(nil); err == nil {
		t.Fatal("GetICCProfile(nil): expected error")
	}
	if _, err := mux.GetICCProfile([]byte{}); err == nil {
		t.Fatal("GetICCProfile(empty): expected error")
	}
}

func TestGetICCProfileInvalid(t *testing.T) {
	if _, err := mux.GetICCProfile([]byte("not a webp")); err == nil {
		t.Fatal("GetICCProfile: expected error for invalid bitstream")
	}
}

func TestSetGetICCProfile(t *testing.T) {
	src := encodeTestWebP(t, false)
	icc := []byte("ICC PROFILE") // odd length exercises RIFF padding

	withICC, err := mux.SetICCProfile(src, icc)
	if err != nil {
		t.Fatalf("SetICCProfile: %v", err)
	}
	got, err := mux.GetICCProfile(withICC)
	if err != nil {
		t.Fatalf("GetICCProfile: %v", err)
	}
	if !bytes.Equal(got, icc) {
		t.Errorf("extracted ICC = %q, want %q", got, icc)
	}
}

func TestSetICCProfileReplacesExisting(t *testing.T) {
	src := encodeTestWebP(t, false)
	withOld, err := mux.SetICCProfile(src, []byte("old-icc-profile-data"))
	if err != nil {
		t.Fatalf("SetICCProfile(old): %v", err)
	}

	newICC := []byte("new-icc")
	withNew, err := mux.SetICCProfile(withOld, newICC)
	if err != nil {
		t.Fatalf("SetICCProfile(new): %v", err)
	}
	got, err := mux.GetICCProfile(withNew)
	if err != nil {
		t.Fatalf("GetICCProfile: %v", err)
	}
	if !bytes.Equal(got, newICC) {
		t.Errorf("extracted ICC = %q, want %q", got, newICC)
	}
}

func TestSetICCProfileEmpty(t *testing.T) {
	src := encodeTestWebP(t, false)
	out, err := mux.SetICCProfile(src, nil)
	if err != nil {
		t.Fatalf("SetICCProfile: %v", err)
	}
	if !bytes.Equal(out, src) {
		t.Error("empty ICC should leave bitstream unchanged")
	}
}

func TestSetICCProfilePreservesPixels(t *testing.T) {
	src := encodeTestWebP(t, false)
	before, err := webp.DecodeNRGBA(src, &webp.DecoderOptions{})
	if err != nil {
		t.Fatalf("DecodeNRGBA(src): %v", err)
	}

	withICC, err := mux.SetICCProfile(src, []byte("test-icc-profile"))
	if err != nil {
		t.Fatalf("SetICCProfile: %v", err)
	}
	after, err := webp.DecodeNRGBA(withICC, &webp.DecoderOptions{})
	if err != nil {
		t.Fatalf("DecodeNRGBA(withICC): %v", err)
	}
	if !bytes.Equal(before.Pix, after.Pix) {
		t.Error("ICC is metadata-only; pixels must not change")
	}
}

func TestSetICCProfileLosslessAlpha(t *testing.T) {
	src := encodeTestWebP(t, true)
	features, err := webp.GetFeatures(src)
	if err != nil {
		t.Fatalf("GetFeatures(src): %v", err)
	}
	if !features.HasAlpha {
		t.Fatal("test image should have alpha")
	}

	withICC, err := mux.SetICCProfile(src, []byte("alpha-icc"))
	if err != nil {
		t.Fatalf("SetICCProfile: %v", err)
	}
	got, err := mux.GetICCProfile(withICC)
	if err != nil {
		t.Fatalf("GetICCProfile: %v", err)
	}
	if !bytes.Equal(got, []byte("alpha-icc")) {
		t.Errorf("extracted ICC = %q, want %q", got, "alpha-icc")
	}

	features, err = webp.GetFeatures(withICC)
	if err != nil {
		t.Fatalf("GetFeatures(withICC): %v", err)
	}
	if !features.HasAlpha {
		t.Error("ALPHA_FLAG must remain set after inserting ICCP")
	}
}

func TestGetICCProfileDespiteMissingAlphaFlag(t *testing.T) {
	src := encodeTestWebP(t, true)
	withICC, err := mux.SetICCProfile(src, []byte("alpha-icc"))
	if err != nil {
		t.Fatalf("SetICCProfile: %v", err)
	}

	broken := clearVP8XAlphaFlag(t, withICC)
	got, err := mux.GetICCProfile(broken)
	if err != nil {
		t.Fatalf("GetICCProfile: %v", err)
	}
	if !bytes.Equal(got, []byte("alpha-icc")) {
		t.Errorf("extracted ICC = %q, want %q", got, "alpha-icc")
	}
}

func encodeTestWebP(t *testing.T, alpha bool) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			a := uint8(255)
			if alpha && (x+y)%2 == 0 {
				a = 128
			}
			img.SetNRGBA(x, y, color.NRGBA{R: 200, G: 40, B: 80, A: a})
		}
	}

	var config *webp.Config
	var err error
	if alpha {
		config, err = webp.ConfigLosslessPreset(6)
	} else {
		config, err = webp.ConfigPreset(webp.PresetDefault, 100)
	}
	if err != nil {
		t.Fatalf("config: %v", err)
	}

	buf := &bytes.Buffer{}
	if err := webp.EncodeRGBA(buf, img, config); err != nil {
		t.Fatalf("EncodeRGBA: %v", err)
	}
	return buf.Bytes()
}

func clearVP8XAlphaFlag(t *testing.T, data []byte) []byte {
	t.Helper()
	const (
		riffHeaderSize  = 12
		chunkHeaderSize = 8
		vp8xFlagAlpha   = 0x10
	)
	if len(data) < riffHeaderSize+chunkHeaderSize+1 {
		t.Fatalf("bitstream too short")
	}
	if string(data[riffHeaderSize:riffHeaderSize+4]) != "VP8X" {
		t.Fatalf("first chunk = %q, want VP8X", data[riffHeaderSize:riffHeaderSize+4])
	}
	payloadSize := binary.LittleEndian.Uint32(data[riffHeaderSize+4 : riffHeaderSize+8])
	if payloadSize < 1 {
		t.Fatal("VP8X payload too small")
	}
	out := append([]byte(nil), data...)
	out[riffHeaderSize+chunkHeaderSize] &^= vp8xFlagAlpha
	return out
}
