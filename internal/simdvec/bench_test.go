//go:build goexperiment.simd

package simdvec

import (
	"math/rand"
	"simd"
	"testing"

	"github.com/adwaiyrandale/simd-raytracer/internal/vec3"
)

// See simdTouch's doc comment in vec3s_test.go: the go1.27rc1
// workaround must be repeated per-file (a direct simd.* reference
// textually present in this file), not just per-package - calling
// simdTouch() from elsewhere doesn't satisfy it.
func init() {
	_ = simd.LoadFloat32s(make([]float32, 16))
}

// benchN is the batch size used by both the scalar and SIMD
// benchmarks below, so their throughput numbers are directly
// comparable. It's a multiple of every lane width this hardware
// generation is likely to have (4, 8, or 16).
//
// Measured result (2026-08-12, Apple M4, go1.27rc1, NEON width 4):
// SIMD Add/Dot are ~1.6x SLOWER than scalar here, even though
// `go build -gcflags=-S` confirms real VFADD/VFMUL NEON instructions
// are emitted (verified by disassembling Vec3s@simd128.Add - not a
// software-emulated fallback). The likely cause: Go's SIMD
// multiversioning compiles each function 3x (generic/simd0/simd128)
// with a CPU-feature dispatch check, and that per-call dispatch
// overhead dominates when the actual work is only 3 elementwise
// float ops on a 4-wide vector. This is a real, useful result: it
// confirms the design doc's approach (batch whole ray-packet/BVH
// operations per call, not one Vec3 at a time) rather than
// invalidating SIMD as a strategy - naive per-Vec3 replacement of
// scalar code is not where the win is.
const benchN = 1 << 16 // 65536

func benchScalarData(n int) []vec3.Vec3 {
	rng := rand.New(rand.NewSource(7))
	vecs := make([]vec3.Vec3, n)
	for i := range vecs {
		vecs[i] = vec3.Vec3{X: rng.Float32(), Y: rng.Float32(), Z: rng.Float32()}
	}
	return vecs
}

func benchSIMDData(n int) (xs, ys, zs []float32) {
	rng := rand.New(rand.NewSource(7)) // same seed as benchScalarData - same values
	xs, ys, zs = make([]float32, n), make([]float32, n), make([]float32, n)
	for i := 0; i < n; i++ {
		xs[i], ys[i], zs[i] = rng.Float32(), rng.Float32(), rng.Float32()
	}
	return xs, ys, zs
}

func BenchmarkScalarAdd(b *testing.B) {
	a := benchScalarData(benchN)
	c := benchScalarData(benchN)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := range a {
			_ = a[j].Add(c[j])
		}
	}
}

func BenchmarkSIMDAdd(b *testing.B) {
	simdTouch()
	width := Width()
	ax, ay, az := benchSIMDData(benchN)
	bx, by, bz := benchSIMDData(benchN)
	outX, outY, outZ := make([]float32, width), make([]float32, width), make([]float32, width)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j+width <= benchN; j += width {
			sum := LoadVec3s(ax[j:], ay[j:], az[j:]).Add(LoadVec3s(bx[j:], by[j:], bz[j:]))
			sum.Store(outX, outY, outZ)
		}
	}
}

func BenchmarkScalarDot(b *testing.B) {
	a := benchScalarData(benchN)
	c := benchScalarData(benchN)
	b.ResetTimer()
	var sink float32
	for i := 0; i < b.N; i++ {
		for j := range a {
			sink = a[j].Dot(c[j])
		}
	}
	_ = sink
}

func BenchmarkSIMDDot(b *testing.B) {
	simdTouch()
	width := Width()
	ax, ay, az := benchSIMDData(benchN)
	bx, by, bz := benchSIMDData(benchN)
	out := make([]float32, width)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j+width <= benchN; j += width {
			dot := LoadVec3s(ax[j:], ay[j:], az[j:]).Dot(LoadVec3s(bx[j:], by[j:], bz[j:]))
			dot.Store(out)
		}
	}
}
