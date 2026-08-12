//go:build goexperiment.simd

// Package simdvec is the phase-3 SIMD scaffold: a batched, SoA Vec3
// type built on Go's experimental simd package, kept separate from
// the scalar internal/vec3 package used by phases 1-2 so the two can
// be benchmarked against each other.
//
// Build/test constraints (see ~/dev/project-docs/go-simd-raytracer-DESIGN.md
// "Toolchain verification" for the full story):
//   - Requires GOEXPERIMENT=simd and the Go 1.27rc1 toolchain
//     (~/sdk/go1.27rc1) - the default `go` on PATH here is 1.24.4 and
//     cannot build this package at all.
//   - Use `go test`, not `go build`, against this package. A confirmed
//     go1.27rc1 toolchain limitation: a struct with simd.Float32s
//     fields compiles fine as part of a self-contained test binary
//     (go test) but fails ("undefined: bridge") when compiled as a
//     standalone importable archive (go build ./internal/simdvec, or
//     any other package importing Vec3s). Reproduced in isolation
//     2026-08-12 outside this repo, so it's a toolchain bug, not a
//     bug in this code. Consequence: this package is not yet wired
//     into cmd/raytracer - it exists for correctness tests and
//     benchmarks only, until either the bug is fixed in a later RC or
//     the design is reworked around free functions instead of a
//     struct (which do cross package boundaries fine, confirmed by
//     the same reproduction).
package simdvec

import "simd"

// Vec3s is a lane-width batch of Vec3 values in struct-of-arrays
// form: X holds the X component of every lane, etc.
type Vec3s struct {
	X, Y, Z simd.Float32s
}

// Width reports the number of float32 lanes in a Vec3s on this
// hardware (e.g. 4 for 128-bit NEON/SSE, 8 for AVX2, 16 for AVX-512).
func Width() int {
	var probe [16]float32
	return simd.LoadFloat32s(probe[:]).Len()
}

// LoadVec3s loads one Vec3s from three component slices, each of
// which must have at least Width() elements.
func LoadVec3s(xs, ys, zs []float32) Vec3s {
	return Vec3s{
		X: simd.LoadFloat32s(xs),
		Y: simd.LoadFloat32s(ys),
		Z: simd.LoadFloat32s(zs),
	}
}

// Store writes each component lane out to its slice, which must have
// at least Width() elements.
func (a Vec3s) Store(xs, ys, zs []float32) {
	a.X.Store(xs)
	a.Y.Store(ys)
	a.Z.Store(zs)
}

func (a Vec3s) Add(b Vec3s) Vec3s {
	return Vec3s{a.X.Add(b.X), a.Y.Add(b.Y), a.Z.Add(b.Z)}
}

func (a Vec3s) Sub(b Vec3s) Vec3s {
	return Vec3s{a.X.Sub(b.X), a.Y.Sub(b.Y), a.Z.Sub(b.Z)}
}

// Dot returns the per-lane dot product as a lane-width batch of
// scalars, not a Vec3s.
func (a Vec3s) Dot(b Vec3s) simd.Float32s {
	return a.X.Mul(b.X).Add(a.Y.Mul(b.Y)).Add(a.Z.Mul(b.Z))
}
