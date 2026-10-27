package mux_test

import (
	"bytes"
	"fmt"
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
			got, err := getChunk(data, mux.XMP)
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
			got, err := getChunk(data, mux.ICCP)
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
		if _, err := getChunk(data, mux.ICCP); err == nil {
			t.Errorf("GetChunk(%q): expected error", data)
		}
		if _, err := setChunk(data, mux.ICCP, chunk); err == nil {
			t.Errorf("SetChunk(%q): expected error", data)
		}
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
					out, err := setChunk(src, tt.fourcc, tt.chunk)
					if err != nil {
						t.Fatalf("SetChunk: %v", err)
					}
					got, err := getChunk(out, tt.fourcc)
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
						extra, err := getChunk(out, other)
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
						next, err := setChunk(out, mux.ICCP, chunk)
						if err != nil {
							t.Fatalf("SetChunk: %v", err)
						}
						out = next
					}
					got, err := getChunk(out, mux.ICCP)
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

func TestSetChunkPreservesPixels(t *testing.T) {
	for _, format := range webPFiles {
		t.Run(format.name, func(t *testing.T) {
			src := util.ReadFile(format.file)
			before, err := webp.DecodeNRGBA(src, &webp.DecoderOptions{})
			if err != nil {
				t.Fatalf("DecodeNRGBA(src): %v", err)
			}
			withICC, err := setChunk(src, mux.ICCP, []byte("icc"))
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
				next, err := setChunk(out, chunk.fourcc, chunk.data)
				if err != nil {
					t.Fatalf("SetChunk(%s): %v", chunk.fourcc, err)
				}
				out = next
			}
			for _, chunk := range chunks {
				got, err := getChunk(out, chunk.fourcc)
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

// These helpers deliberately compose the individual bindings only in tests.
func getChunk(data []byte, fourcc mux.FourCC) ([]byte, error) {
	input := mux.NewData(data)
	if input == nil {
		return nil, fmt.Errorf("allocation failed")
	}
	defer mux.DataClear(input)
	d := mux.Demux(input)
	if d == nil {
		return nil, fmt.Errorf("Demux failed")
	}
	defer mux.DemuxDelete(d)
	var iter mux.ChunkIterator
	if !mux.DemuxGetChunk(d, fourcc, 1, &iter) {
		return nil, nil
	}
	defer mux.DemuxReleaseChunkIterator(&iter)
	chunk := iter.Chunk()
	return bytes.Clone(chunk.Bytes()), nil
}

func setChunk(data []byte, fourcc mux.FourCC, chunk []byte) ([]byte, error) {
	input := mux.NewData(data)
	if input == nil {
		return nil, fmt.Errorf("allocation failed")
	}
	defer mux.DataClear(input)
	payload := mux.NewData(chunk)
	if payload == nil {
		return nil, fmt.Errorf("allocation failed")
	}
	defer mux.DataClear(payload)
	m := mux.MuxCreate(input, false)
	if m == nil {
		return nil, fmt.Errorf("MuxCreate failed")
	}
	defer mux.MuxDelete(m)
	if status := mux.MuxSetChunk(m, fourcc, payload, false); status != mux.MuxOK {
		return nil, fmt.Errorf("MuxSetChunk: %s", status)
	}
	var output mux.Data
	if status := mux.MuxAssemble(m, &output); status != mux.MuxOK {
		return nil, fmt.Errorf("MuxAssemble: %s", status)
	}
	defer mux.DataClear(&output)
	return bytes.Clone(output.Bytes()), nil
}

func TestMuxBindings(t *testing.T) {
	for _, format := range webPFiles {
		t.Run(format.name, func(t *testing.T) {
			input := mux.NewData(util.ReadFile(format.file))
			if input == nil {
				t.Fatal("NewData failed")
			}
			m := mux.MuxCreate(input, true)
			mux.DataClear(input)
			if m == nil {
				t.Fatal("MuxCreate failed")
			}
			defer mux.MuxDelete(m)

			for _, cc := range []mux.FourCC{mux.ICCP, mux.EXIF, mux.XMP} {
				payload := mux.NewData([]byte(cc))
				if payload == nil {
					t.Fatal("NewData failed")
				}
				status := mux.MuxSetChunk(m, cc, payload, true)
				mux.DataClear(payload)
				if status != mux.MuxOK {
					t.Fatalf("MuxSetChunk: %s", status)
				}
				var got mux.Data
				if status := mux.MuxGetChunk(m, cc, &got); status != mux.MuxOK {
					t.Fatalf("MuxGetChunk: %s", status)
				}
				if string(got.Bytes()) != string(cc) {
					t.Fatalf("got %q", got.Bytes())
				}
			}
			if status := mux.MuxDeleteChunk(m, mux.EXIF); status != mux.MuxOK {
				t.Fatalf("MuxDeleteChunk: %s", status)
			}
			var missing mux.Data
			if status := mux.MuxGetChunk(m, mux.EXIF, &missing); status != mux.MuxNotFound {
				t.Fatalf("missing chunk: %s", status)
			}
			if status := mux.MuxDeleteChunk(m, mux.EXIF); status != mux.MuxNotFound {
				t.Fatalf("delete missing chunk: %s", status)
			}
			var output mux.Data
			if status := mux.MuxAssemble(m, &output); status != mux.MuxOK {
				t.Fatalf("MuxAssemble: %s", status)
			}
			defer mux.DataClear(&output)
			d := mux.Demux(&output)
			if d == nil {
				t.Fatal("Demux failed")
			}
			defer mux.DemuxDelete(d)
			var iter mux.ChunkIterator
			if mux.DemuxGetChunk(d, mux.EXIF, 1, &iter) {
				t.Fatal("deleted EXIF found")
			}
			if !mux.DemuxGetChunk(d, mux.ICCP, 0, &iter) {
				t.Fatal("ICCP not found")
			}
			defer mux.DemuxReleaseChunkIterator(&iter)
			if iter.ChunkNum() != 1 || iter.NumChunks() != 1 {
				t.Fatal("invalid chunk count")
			}
			chunk := iter.Chunk()
			if string(chunk.Bytes()) != "ICCP" {
				t.Fatalf("got %q", chunk.Bytes())
			}
			if mux.DemuxNextChunk(&iter) || mux.DemuxPrevChunk(&iter) {
				t.Fatal("unexpected adjacent chunk")
			}
		})
	}
}

func TestMuxNewAndDataCopy(t *testing.T) {
	input := mux.NewData(util.ReadFile("checkerboard-lossless.webp"))
	if input == nil {
		t.Fatal("NewData failed")
	}
	defer mux.DataClear(input)
	m := mux.MuxNew()
	if m == nil {
		t.Fatal("MuxNew failed")
	}
	defer mux.MuxDelete(m)
	if status := mux.MuxSetImage(m, input, false); status != mux.MuxOK {
		t.Fatal(status)
	}
	var output mux.Data
	if status := mux.MuxAssemble(m, &output); status != mux.MuxOK {
		t.Fatal(status)
	}
	var copied mux.Data
	if !mux.DataCopy(&output, &copied) {
		t.Fatal("DataCopy failed")
	}
	mux.DataClear(&output)
	defer mux.DataClear(&copied)
	if output.Bytes() != nil {
		t.Fatal("DataClear did not reset output")
	}
	if _, err := webp.GetFeatures(copied.Bytes()); err != nil {
		t.Fatal(err)
	}
}

func TestBindingInvalidArguments(t *testing.T) {
	if mux.MuxCreate(nil, true) != nil || mux.Demux(nil) != nil {
		t.Fatal("nil input accepted")
	}
	var data mux.Data
	var iter mux.ChunkIterator
	if mux.MuxSetChunk(nil, mux.ICCP, &data, true) != mux.MuxInvalidArgument {
		t.Fatal("nil mux accepted")
	}
	for _, cc := range []mux.FourCC{"ICC", "ICCPA"} {
		if mux.MuxSetChunk(nil, cc, &data, true) != mux.MuxInvalidArgument ||
			mux.MuxGetChunk(nil, cc, &data) != mux.MuxInvalidArgument ||
			mux.MuxDeleteChunk(nil, cc) != mux.MuxInvalidArgument ||
			mux.DemuxGetChunk(nil, cc, 1, &iter) {
			t.Fatalf("invalid FourCC accepted: %q", cc)
		}
	}
}
