//go:build !cgo

package webp

// This package requires cgo: build with CGO_ENABLED=1 and libwebp installed.
var _ = cgo_is_required_to_build_this_package
