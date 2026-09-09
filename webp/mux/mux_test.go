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

func TestInvalidBitstream(t *testing.T) {
	chunk := []byte("x")
	for _, data := range [][]byte{nil, {}, []byte("not a webp")} {
		if _, err := mux.GetChunk(data, mux.ICCP); err == nil {
			t.Errorf("GetChunk(%q): expected error", data)
		}
		if _, err := mux.SetChunk(data, mux.ICCP, chunk); err == nil {
			t.Errorf("SetChunk(%q): expected error", data)
		}
	}
}

func TestInvalidFourCC(t *testing.T) {
	data := encodeTestWebP(t, false)
	for _, fourcc := range []mux.FourCC{"ICC", "ICCPA"} {
		if _, err := mux.GetChunk(data, fourcc); err == nil {
			t.Errorf("GetChunk(%q): expected error", fourcc)
		}
		if _, err := mux.SetChunk(data, fourcc, []byte("x")); err == nil {
			t.Errorf("SetChunk(%q): expected error", fourcc)
		}
	}
}

func TestSetGetChunk(t *testing.T) {
	src := encodeTestWebP(t, false)
	tests := []struct {
		fourcc mux.FourCC
		chunk  []byte
	}{
		{mux.ICCP, []byte("ICC PROFILE")}, // odd length exercises RIFF padding
		{mux.EXIF, []byte("Exif\x00\x00MM")},
		{mux.XMP, []byte("<x:xmpmeta/>")},
	}
	for _, tt := range tests {
		t.Run(string(tt.fourcc), func(t *testing.T) {
			out, err := mux.SetChunk(src, tt.fourcc, tt.chunk)
			if err != nil {
				t.Fatalf("SetChunk: %v", err)
			}
			got, err := mux.GetChunk(out, tt.fourcc)
			if err != nil {
				t.Fatalf("GetChunk: %v", err)
			}
			if !bytes.Equal(got, tt.chunk) {
				t.Errorf("got %q, want %q", got, tt.chunk)
			}
			for _, other := range tests {
				if other.fourcc == tt.fourcc {
					continue
				}
				extra, err := mux.GetChunk(out, other.fourcc)
				if err != nil {
					t.Fatalf("GetChunk(%s): %v", other.fourcc, err)
				}
				if extra != nil {
					t.Errorf("SetChunk(%s) invented a %q chunk", tt.fourcc, other.fourcc)
				}
			}
		})
	}
}

func TestSetChunkReplacesExisting(t *testing.T) {
	src := encodeTestWebP(t, false)
	withOld, err := mux.SetChunk(src, mux.ICCP, []byte("old"))
	if err != nil {
		t.Fatalf("SetChunk(old): %v", err)
	}
	withNew, err := mux.SetChunk(withOld, mux.ICCP, []byte("new"))
	if err != nil {
		t.Fatalf("SetChunk(new): %v", err)
	}
	got, err := mux.GetChunk(withNew, mux.ICCP)
	if err != nil {
		t.Fatalf("GetChunk: %v", err)
	}
	if !bytes.Equal(got, []byte("new")) {
		t.Errorf("got %q, want %q", got, "new")
	}
}

func TestSetChunkNil(t *testing.T) {
	src := encodeTestWebP(t, false)
	if _, err := mux.SetChunk(src, mux.ICCP, nil); err == nil {
		t.Fatal("SetChunk: expected error for nil chunk")
	}
}

func TestSetChunkEmpty(t *testing.T) {
	src := encodeTestWebP(t, false)
	out, err := mux.SetChunk(src, mux.ICCP, []byte{})
	if err != nil {
		t.Fatalf("SetChunk: %v", err)
	}
	got, err := mux.GetChunk(out, mux.ICCP)
	if err != nil {
		t.Fatalf("GetChunk: %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Errorf("got %#v, want empty non-nil slice", got)
	}

	withOld, err := mux.SetChunk(src, mux.ICCP, []byte("old"))
	if err != nil {
		t.Fatalf("SetChunk(old): %v", err)
	}
	replaced, err := mux.SetChunk(withOld, mux.ICCP, []byte{})
	if err != nil {
		t.Fatalf("SetChunk(replace empty): %v", err)
	}
	got, err = mux.GetChunk(replaced, mux.ICCP)
	if err != nil {
		t.Fatalf("GetChunk: %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Errorf("replaced %#v, want empty non-nil slice", got)
	}
}

func TestSetChunkPreservesPixels(t *testing.T) {
	src := encodeTestWebP(t, false)
	before, err := webp.DecodeNRGBA(src, &webp.DecoderOptions{})
	if err != nil {
		t.Fatalf("DecodeNRGBA(src): %v", err)
	}
	withICC, err := mux.SetChunk(src, mux.ICCP, []byte("icc"))
	if err != nil {
		t.Fatalf("SetChunk: %v", err)
	}
	after, err := webp.DecodeNRGBA(withICC, &webp.DecoderOptions{})
	if err != nil {
		t.Fatalf("DecodeNRGBA: %v", err)
	}
	if !bytes.Equal(before.Pix, after.Pix) {
		t.Error("metadata chunks must not change pixels")
	}
}

func TestSetChunkKeepsAlphaFlag(t *testing.T) {
	src := encodeTestWebP(t, true)
	features, err := webp.GetFeatures(src)
	if err != nil {
		t.Fatalf("GetFeatures(src): %v", err)
	}
	if !features.HasAlpha {
		t.Fatal("test image should have alpha")
	}
	withICC, err := mux.SetChunk(src, mux.ICCP, []byte("icc"))
	if err != nil {
		t.Fatalf("SetChunk: %v", err)
	}
	features, err = webp.GetFeatures(withICC)
	if err != nil {
		t.Fatalf("GetFeatures: %v", err)
	}
	if !features.HasAlpha {
		t.Error("ALPHA_FLAG must remain set after inserting ICCP")
	}
}

func TestGetChunkDespiteMissingAlphaFlag(t *testing.T) {
	src := encodeTestWebP(t, true)
	withICC, err := mux.SetChunk(src, mux.ICCP, []byte("icc"))
	if err != nil {
		t.Fatalf("SetChunk: %v", err)
	}
	got, err := mux.GetChunk(clearVP8XAlphaFlag(t, withICC), mux.ICCP)
	if err != nil {
		t.Fatalf("GetChunk: %v", err)
	}
	if !bytes.Equal(got, []byte("icc")) {
		t.Errorf("got %q, want %q", got, "icc")
	}
}

func TestSetMultipleChunks(t *testing.T) {
	src := encodeTestWebP(t, false)
	chunks := []struct {
		fourcc mux.FourCC
		data   []byte
	}{
		{mux.ICCP, []byte("icc-bytes")},
		{mux.EXIF, []byte("exif-bytes")},
		{mux.XMP, []byte("xmp-bytes")},
	}

	out := src
	for _, tt := range chunks {
		next, err := mux.SetChunk(out, tt.fourcc, tt.data)
		if err != nil {
			t.Fatalf("SetChunk(%s): %v", tt.fourcc, err)
		}
		out = next
	}
	for _, tt := range chunks {
		got, err := mux.GetChunk(out, tt.fourcc)
		if err != nil {
			t.Fatalf("GetChunk(%s): %v", tt.fourcc, err)
		}
		if !bytes.Equal(got, tt.data) {
			t.Errorf("%s = %q, want %q", tt.fourcc, got, tt.data)
		}
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
