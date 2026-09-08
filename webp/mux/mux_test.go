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

func TestGetChunkMissing(t *testing.T) {
	data := util.ReadFile("cosmos.webp")
	got, err := mux.GetChunk(data, mux.ICCP)
	if err != nil {
		t.Fatalf("GetChunk: %v", err)
	}
	if got != nil {
		t.Errorf("GetChunk: got %d bytes, want nil", len(got))
	}
}

func TestGetChunkEmpty(t *testing.T) {
	if _, err := mux.GetChunk(nil, mux.ICCP); err == nil {
		t.Fatal("GetChunk(nil): expected error")
	}
	if _, err := mux.GetChunk([]byte{}, mux.ICCP); err == nil {
		t.Fatal("GetChunk(empty): expected error")
	}
}

func TestGetChunkInvalid(t *testing.T) {
	if _, err := mux.GetChunk([]byte("not a webp"), mux.ICCP); err == nil {
		t.Fatal("GetChunk: expected error for invalid bitstream")
	}
}

func TestGetChunkInvalidFourCC(t *testing.T) {
	data := encodeTestWebP(t, false)
	if _, err := mux.GetChunk(data, "ICC"); err == nil {
		t.Fatal("GetChunk: expected error for short fourcc")
	}
	if _, err := mux.SetChunk(data, "ICC", []byte("x")); err == nil {
		t.Fatal("SetChunk: expected error for short fourcc")
	}
}

func TestSetGetChunkICCP(t *testing.T) {
	src := encodeTestWebP(t, false)
	icc := []byte("ICC PROFILE") // odd length exercises RIFF padding

	withICC, err := mux.SetChunk(src, mux.ICCP, icc)
	if err != nil {
		t.Fatalf("SetChunk: %v", err)
	}
	got, err := mux.GetChunk(withICC, mux.ICCP)
	if err != nil {
		t.Fatalf("GetChunk: %v", err)
	}
	if !bytes.Equal(got, icc) {
		t.Errorf("extracted ICCP = %q, want %q", got, icc)
	}
}

func TestSetGetChunkEXIF(t *testing.T) {
	src := encodeTestWebP(t, false)
	exif := []byte("Exif\x00\x00MM")

	withExif, err := mux.SetChunk(src, mux.EXIF, exif)
	if err != nil {
		t.Fatalf("SetChunk(EXIF): %v", err)
	}
	got, err := mux.GetChunk(withExif, mux.EXIF)
	if err != nil {
		t.Fatalf("GetChunk(EXIF): %v", err)
	}
	if !bytes.Equal(got, exif) {
		t.Errorf("extracted EXIF = %q, want %q", got, exif)
	}
	icc, err := mux.GetChunk(withExif, mux.ICCP)
	if err != nil {
		t.Fatalf("GetChunk(ICCP): %v", err)
	}
	if icc != nil {
		t.Errorf("EXIF insert must not invent an ICCP chunk")
	}
}

func TestSetChunkReplacesExisting(t *testing.T) {
	src := encodeTestWebP(t, false)
	withOld, err := mux.SetChunk(src, mux.ICCP, []byte("old-icc-profile-data"))
	if err != nil {
		t.Fatalf("SetChunk(old): %v", err)
	}

	newICC := []byte("new-icc")
	withNew, err := mux.SetChunk(withOld, mux.ICCP, newICC)
	if err != nil {
		t.Fatalf("SetChunk(new): %v", err)
	}
	got, err := mux.GetChunk(withNew, mux.ICCP)
	if err != nil {
		t.Fatalf("GetChunk: %v", err)
	}
	if !bytes.Equal(got, newICC) {
		t.Errorf("extracted ICCP = %q, want %q", got, newICC)
	}
}

func TestSetChunkEmpty(t *testing.T) {
	src := encodeTestWebP(t, false)
	if _, err := mux.SetChunk(src, mux.ICCP, nil); err == nil {
		t.Fatal("SetChunk: expected error for empty chunk")
	}
}

func TestSetChunkPreservesPixels(t *testing.T) {
	src := encodeTestWebP(t, false)
	before, err := webp.DecodeNRGBA(src, &webp.DecoderOptions{})
	if err != nil {
		t.Fatalf("DecodeNRGBA(src): %v", err)
	}

	withICC, err := mux.SetChunk(src, mux.ICCP, []byte("test-icc-profile"))
	if err != nil {
		t.Fatalf("SetChunk: %v", err)
	}
	after, err := webp.DecodeNRGBA(withICC, &webp.DecoderOptions{})
	if err != nil {
		t.Fatalf("DecodeNRGBA(withICC): %v", err)
	}
	if !bytes.Equal(before.Pix, after.Pix) {
		t.Error("metadata chunks must not change pixels")
	}
}

func TestSetChunkLosslessAlpha(t *testing.T) {
	src := encodeTestWebP(t, true)
	features, err := webp.GetFeatures(src)
	if err != nil {
		t.Fatalf("GetFeatures(src): %v", err)
	}
	if !features.HasAlpha {
		t.Fatal("test image should have alpha")
	}

	withICC, err := mux.SetChunk(src, mux.ICCP, []byte("alpha-icc"))
	if err != nil {
		t.Fatalf("SetChunk: %v", err)
	}
	got, err := mux.GetChunk(withICC, mux.ICCP)
	if err != nil {
		t.Fatalf("GetChunk: %v", err)
	}
	if !bytes.Equal(got, []byte("alpha-icc")) {
		t.Errorf("extracted ICCP = %q, want %q", got, "alpha-icc")
	}

	features, err = webp.GetFeatures(withICC)
	if err != nil {
		t.Fatalf("GetFeatures(withICC): %v", err)
	}
	if !features.HasAlpha {
		t.Error("ALPHA_FLAG must remain set after inserting ICCP")
	}
}

func TestGetChunkDespiteMissingAlphaFlag(t *testing.T) {
	src := encodeTestWebP(t, true)
	withICC, err := mux.SetChunk(src, mux.ICCP, []byte("alpha-icc"))
	if err != nil {
		t.Fatalf("SetChunk: %v", err)
	}

	broken := clearVP8XAlphaFlag(t, withICC)
	got, err := mux.GetChunk(broken, mux.ICCP)
	if err != nil {
		t.Fatalf("GetChunk: %v", err)
	}
	if !bytes.Equal(got, []byte("alpha-icc")) {
		t.Errorf("extracted ICCP = %q, want %q", got, "alpha-icc")
	}
}

func TestSetChunkICCPAndEXIF(t *testing.T) {
	src := encodeTestWebP(t, false)
	icc := []byte("icc-bytes")
	exif := []byte("exif-bytes")

	withICC, err := mux.SetChunk(src, mux.ICCP, icc)
	if err != nil {
		t.Fatalf("SetChunk(ICCP): %v", err)
	}
	withBoth, err := mux.SetChunk(withICC, mux.EXIF, exif)
	if err != nil {
		t.Fatalf("SetChunk(EXIF): %v", err)
	}

	gotICC, err := mux.GetChunk(withBoth, mux.ICCP)
	if err != nil {
		t.Fatalf("GetChunk(ICCP): %v", err)
	}
	gotExif, err := mux.GetChunk(withBoth, mux.EXIF)
	if err != nil {
		t.Fatalf("GetChunk(EXIF): %v", err)
	}
	if !bytes.Equal(gotICC, icc) {
		t.Errorf("ICCP = %q, want %q", gotICC, icc)
	}
	if !bytes.Equal(gotExif, exif) {
		t.Errorf("EXIF = %q, want %q", gotExif, exif)
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
