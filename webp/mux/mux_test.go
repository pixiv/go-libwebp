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
	tests := []struct {
		name   string
		fourcc mux.FourCC
	}{
		{"fourCC is too short", mux.FourCC("ICC")},
		{"fourCC is too long", mux.FourCC("ICCPA")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := mux.GetChunk(data, tt.fourcc); err == nil {
				t.Errorf("GetChunk(%q): expected error", tt.fourcc)
			}
			if _, err := mux.SetChunk(data, tt.fourcc, []byte("x")); err == nil {
				t.Errorf("SetChunk(%q): expected error", tt.fourcc)
			}
		})
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
			for _, other := range []mux.FourCC{mux.ICCP, mux.EXIF, mux.XMP} {
				if other == tt.fourcc {
					continue
				}
				extra, err := mux.GetChunk(out, other)
				if err != nil {
					t.Fatalf("GetChunk(%s): %v", other, err)
				}
				if extra != nil {
					t.Errorf("SetChunk(%s) invented a %q chunk", tt.fourcc, other)
				}
			}
		})
	}
}

func TestSetChunkReplacesExisting(t *testing.T) {
	src := encodeTestWebP(t, false)
	tests := []struct {
		name string
		sets [][]byte
		want []byte
	}{
		{"replace existing", [][]byte{[]byte("old"), []byte("new")}, []byte("new")},
		{"empty", [][]byte{[]byte{}}, []byte{}},
		{"replace with empty", [][]byte{[]byte("old"), []byte{}}, []byte{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := src
			for _, chunk := range tt.sets {
				next, err := mux.SetChunk(out, mux.ICCP, chunk)
				if err != nil {
					t.Fatalf("SetChunk: %v", err)
				}
				out = next
			}
			got, err := mux.GetChunk(out, mux.ICCP)
			if err != nil {
				t.Fatalf("GetChunk: %v", err)
			}
			if len(tt.want) == 0 {
				if got == nil || len(got) != 0 {
					t.Errorf("got %#v, want empty non-nil slice", got)
				}
				return
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSetChunkNil(t *testing.T) {
	src := encodeTestWebP(t, false)
	if _, err := mux.SetChunk(src, mux.ICCP, nil); err == nil {
		t.Fatal("SetChunk: expected error for nil chunk")
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
	for _, chunk := range chunks {
		next, err := mux.SetChunk(out, chunk.fourcc, chunk.data)
		if err != nil {
			t.Fatalf("SetChunk(%s): %v", chunk.fourcc, err)
		}
		out = next
	}
	for _, chunk := range chunks {
		got, err := mux.GetChunk(out, chunk.fourcc)
		if err != nil {
			t.Fatalf("GetChunk(%s): %v", chunk.fourcc, err)
		}
		if !bytes.Equal(got, chunk.data) {
			t.Errorf("%s = %q, want %q", chunk.fourcc, got, chunk.data)
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
