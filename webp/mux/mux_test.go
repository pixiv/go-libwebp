package mux_test

import (
	"bytes"
	"testing"

	"github.com/pixiv/go-libwebp/test/util"
	"github.com/pixiv/go-libwebp/webp"
	"github.com/pixiv/go-libwebp/webp/mux"
)

var webPFiles = []struct {
	name   string
	file   string
	format int
}{
	{"lossy", "checkerboard-lossy.webp", 1},
	{"lossless", "checkerboard-lossless.webp", 2},
}

func TestGetChunkExisting(t *testing.T) {
	want := util.ReadFile("checkerboard.xmp")
	for _, tt := range webPFiles {
		t.Run(tt.name, func(t *testing.T) {
			data := util.ReadFile(tt.file)
			features, err := webp.GetFeatures(data)
			if err != nil {
				t.Fatalf("GetFeatures: %v", err)
			}
			if features.Format != tt.format {
				t.Fatalf("format = %d, want %d", features.Format, tt.format)
			}
			got, err := mux.GetChunk(data, mux.XMP)
			if err != nil {
				t.Fatalf("GetChunk: %v", err)
			}
			if !bytes.Equal(got, want) {
				t.Errorf("got %q, want %q", got, want)
			}
		})
	}
}

func TestGetChunkMissing(t *testing.T) {
	for _, tt := range webPFiles {
		t.Run(tt.name, func(t *testing.T) {
			data := util.ReadFile(tt.file)
			got, err := mux.GetChunk(data, mux.ICCP)
			if err != nil {
				t.Fatalf("GetChunk: %v", err)
			}
			if got != nil {
				t.Errorf("GetChunk: got %d bytes, want nil", len(got))
			}
		})
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

func TestReadFourCC(t *testing.T) {
	for _, format := range webPFiles {
		t.Run(format.name, func(t *testing.T) {
			data := util.ReadFile(format.file)
			tests := []struct {
				name       string
				fourcc     mux.FourCC
				wantGetErr bool
			}{
				{"fourCC is too short", "ICC", true},
				{"fourCC is too long", "ICCPA", true},
				{"VP8 without trailing space", "VP8", true},
				{"VP8 image chunk", "VP8 ", false},
				{"VP8L image chunk", "VP8L", false},
				{"ALPH image chunk", "ALPH", false},
			}
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					got, err := mux.GetChunk(data, tt.fourcc)
					if (err != nil) != tt.wantGetErr {
						t.Errorf("GetChunk(%q): error = %v, want error = %t", tt.fourcc, err, tt.wantGetErr)
					}
					if got != nil {
						t.Errorf("GetChunk(%q): got %d bytes, want nil", tt.fourcc, len(got))
					}
					if _, err := mux.SetChunk(data, tt.fourcc, []byte("x")); err == nil {
						t.Errorf("SetChunk(%q): expected error", tt.fourcc)
					}
				})
			}
		})
	}
}

func TestSetGetChunk(t *testing.T) {
	existingXMP := util.ReadFile("checkerboard.xmp")
	for _, format := range webPFiles {
		t.Run(format.name, func(t *testing.T) {
			src := util.ReadFile(format.file)
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
						if other == mux.XMP {
							if !bytes.Equal(extra, existingXMP) {
								t.Errorf("SetChunk(%s) changed the existing XMP chunk", tt.fourcc)
							}
						} else if extra != nil {
							t.Errorf("SetChunk(%s) invented a %q chunk", tt.fourcc, other)
						}
					}
				})
			}
		})
	}
}

func TestSetChunkReplacesExisting(t *testing.T) {
	for _, format := range webPFiles {
		t.Run(format.name, func(t *testing.T) {
			src := util.ReadFile(format.file)
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
		})
	}
}

func TestSetChunkNil(t *testing.T) {
	src := util.ReadFile("checkerboard-lossy.webp")
	if _, err := mux.SetChunk(src, mux.ICCP, nil); err == nil {
		t.Fatal("SetChunk: expected error for nil chunk")
	}
}

func TestSetChunkPreservesPixels(t *testing.T) {
	for _, format := range webPFiles {
		t.Run(format.name, func(t *testing.T) {
			src := util.ReadFile(format.file)
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
		})
	}
}

func TestSetMultipleChunks(t *testing.T) {
	for _, format := range webPFiles {
		t.Run(format.name, func(t *testing.T) {
			src := util.ReadFile(format.file)
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
			for _, target := range chunks {
				t.Run(string(target.fourcc), func(t *testing.T) {
					tests := []struct {
						name string
						data []byte
					}{
						{"replace", []byte("replacement")},
						{"replace with empty", []byte{}},
					}
					for _, tt := range tests {
						t.Run(tt.name, func(t *testing.T) {
							updated, err := mux.SetChunk(out, target.fourcc, tt.data)
							if err != nil {
								t.Fatalf("SetChunk(%s): %v", target.fourcc, err)
							}
							for _, chunk := range chunks {
								want := chunk.data
								if chunk.fourcc == target.fourcc {
									want = tt.data
								}
								got, err := mux.GetChunk(updated, chunk.fourcc)
								if err != nil {
									t.Fatalf("GetChunk(%s): %v", chunk.fourcc, err)
								}
								if got == nil || !bytes.Equal(got, want) {
									t.Errorf("after SetChunk(%s), %s = %#v, want %#v", target.fourcc, chunk.fourcc, got, want)
								}
							}
						})
					}
				})
			}
		})
	}
}
