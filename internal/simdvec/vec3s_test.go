//go:build goexperiment.simd

package simdvec

import (
	"math/rand"
	"simd"
	"testing"

	"github.com/adwaiyrandale/simd-raytracer/internal/vec3"
)

// simdTouch works around a confirmed go1.27rc1 compiler limitation
// (see the doc comment atop vec3s.go): any function that uses Vec3s
// (even only indirectly, through this package's own LoadVec3s/Add/Dot)
// must itself directly call a simd.* function, or compilation fails
// with "undefined: bridge". A package-level var achieving the same
// thing triggers a different, worse failure (an actual internal
// compiler error) - so this must be called inline, per function.
func simdTouch() {
	_ = simd.LoadFloat32s(make([]float32, 16))
}

func TestWidthIsPositive(t *testing.T) {
	if Width() <= 0 {
		t.Fatalf("Width() = %d, want > 0", Width())
	}
}

// randComponents returns n random scalar Vec3s and their split
// component slices, for feeding both the scalar and SIMD paths from
// the same data.
func randComponents(rng *rand.Rand, n int) (vecs []vec3.Vec3, xs, ys, zs []float32) {
	vecs = make([]vec3.Vec3, n)
	xs = make([]float32, n)
	ys = make([]float32, n)
	zs = make([]float32, n)
	for i := range vecs {
		vecs[i] = vec3.Vec3{X: rng.Float32()*20 - 10, Y: rng.Float32()*20 - 10, Z: rng.Float32()*20 - 10}
		xs[i], ys[i], zs[i] = vecs[i].X, vecs[i].Y, vecs[i].Z
	}
	return vecs, xs, ys, zs
}

func TestAddMatchesScalar(t *testing.T) {
	simdTouch()
	width := Width()
	rng := rand.New(rand.NewSource(1))
	aVecs, aX, aY, aZ := randComponents(rng, width)
	bVecs, bX, bY, bZ := randComponents(rng, width)

	simdSum := LoadVec3s(aX, aY, aZ).Add(LoadVec3s(bX, bY, bZ))
	outX, outY, outZ := make([]float32, width), make([]float32, width), make([]float32, width)
	simdSum.Store(outX, outY, outZ)

	const eps = 1e-4
	for i := 0; i < width; i++ {
		want := aVecs[i].Add(bVecs[i])
		if abs32(outX[i]-want.X) > eps || abs32(outY[i]-want.Y) > eps || abs32(outZ[i]-want.Z) > eps {
			t.Errorf("lane %d: SIMD Add = (%v,%v,%v), want %v", i, outX[i], outY[i], outZ[i], want)
		}
	}
}

func TestDotMatchesScalar(t *testing.T) {
	simdTouch()
	width := Width()
	rng := rand.New(rand.NewSource(2))
	aVecs, aX, aY, aZ := randComponents(rng, width)
	bVecs, bX, bY, bZ := randComponents(rng, width)

	simdDot := LoadVec3s(aX, aY, aZ).Dot(LoadVec3s(bX, bY, bZ))
	out := make([]float32, width)
	simdDot.Store(out)

	const eps = 1e-3
	for i := 0; i < width; i++ {
		want := aVecs[i].Dot(bVecs[i])
		if abs32(out[i]-want) > eps {
			t.Errorf("lane %d: SIMD Dot = %v, want %v", i, out[i], want)
		}
	}
}

func abs32(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}
