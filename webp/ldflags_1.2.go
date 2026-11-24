//go:build !1.3

package webp

// #cgo linux CFLAGS: -I/usr/local/include
// #cgo linux LDFLAGS: -L/usr/local/lib
// #cgo darwin CFLAGS: -I/opt/homebrew/include
// #cgo darwin LDFLAGS: -L/opt/homebrew/lib
// #cgo LDFLAGS: -lwebp -lm
import "C"
