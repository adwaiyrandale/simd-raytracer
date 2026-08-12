//go:build goexperiment.simd

package main

import (
	"math/rand"
	"simd"
	"testing"

	"github.com/adwaiyrandale/simd-raytracer/internal/ray"
	"github.com/adwaiyrandale/simd-raytracer/internal/scene"
	"github.com/adwaiyrandale/simd-raytracer/internal/vec3"
)

// benchRayCount is the total number of rays processed by both the
// scalar and SIMD intersection benchmarks below, so their throughput
// numbers are directly comparable regardless of hardware lane width.
const benchRayCount = 1 << 16 // 65536

func benchRays(n int) []ray.Ray {
	rng := rand.New(rand.NewSource(9))
	rays := make([]ray.Ray, n)
	for i := range rays {
		rays[i] = ray.Ray{
			Origin:    vec3.Vec3{X: rng.Float32()*4 - 2, Y: rng.Float32()*4 - 2, Z: rng.Float32() * 2},
			Direction: vec3.Vec3{X: rng.Float32()*2 - 1, Y: rng.Float32()*2 - 1, Z: -1}.Normalize(),
		}
	}
	return rays
}

func BenchmarkScalarSphereIntersect(b *testing.B) {
	sph := scene.Sphere{Center: vec3.Vec3{X: 0, Y: 0, Z: -1}, Radius: 0.5}
	rays := benchRays(benchRayCount)

	b.ResetTimer()
	var sinkT float32
	for i := 0; i < b.N; i++ {
		for _, r := range rays {
			_, t, ok := sph.Hit(r, 0.001, 1000)
			if ok {
				sinkT = t
			}
		}
	}
	_ = sinkT
}

func BenchmarkSIMDSphereIntersect(b *testing.B) {
	_ = simd.LoadFloat32s(make([]float32, 16)) // per-file workaround, see vec3s_test.go
	sph := scene.Sphere{Center: vec3.Vec3{X: 0, Y: 0, Z: -1}, Radius: 0.5}
	rays := benchRays(benchRayCount)
	lanes := laneWidth()

	oxs, oys, ozs := make([]float32, lanes), make([]float32, lanes), make([]float32, lanes)
	dxs, dys, dzs := make([]float32, lanes), make([]float32, lanes), make([]float32, lanes)
	ts := make([]float32, lanes)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for i0 := 0; i0+lanes <= benchRayCount; i0 += lanes {
			for lane := 0; lane < lanes; lane++ {
				r := rays[i0+lane]
				oxs[lane], oys[lane], ozs[lane] = r.Origin.X, r.Origin.Y, r.Origin.Z
				dxs[lane], dys[lane], dzs[lane] = r.Direction.X, r.Direction.Y, r.Direction.Z
			}
			ox, oy, oz := simd.LoadFloat32s(oxs), simd.LoadFloat32s(oys), simd.LoadFloat32s(ozs)
			dx, dy, dz := simd.LoadFloat32s(dxs), simd.LoadFloat32s(dys), simd.LoadFloat32s(dzs)
			t, _, _, _, _, _, _, _ := simdSphereHit(ox, oy, oz, dx, dy, dz, sph, 0.001, 1000)
			t.Store(ts)
		}
	}
}

func BenchmarkScalarTriangleIntersect(b *testing.B) {
	tris := benchCubeTriangles()
	rays := benchRays(benchRayCount)

	b.ResetTimer()
	var sinkT float32
	for i := 0; i < b.N; i++ {
		for _, r := range rays {
			closest := float32(1000)
			for _, tri := range tris {
				if _, t, ok := tri.Hit(r, 0.001, closest); ok {
					closest = t
					sinkT = t
				}
			}
		}
	}
	_ = sinkT
}

func BenchmarkSIMDTriangleIntersect(b *testing.B) {
	_ = simd.LoadFloat32s(make([]float32, 16)) // per-file workaround, see vec3s_test.go
	tris := benchCubeTriangles()
	rays := benchRays(benchRayCount)
	lanes := laneWidth()

	oxs, oys, ozs := make([]float32, lanes), make([]float32, lanes), make([]float32, lanes)
	dxs, dys, dzs := make([]float32, lanes), make([]float32, lanes), make([]float32, lanes)
	ts := make([]float32, lanes)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for i0 := 0; i0+lanes <= benchRayCount; i0 += lanes {
			for lane := 0; lane < lanes; lane++ {
				r := rays[i0+lane]
				oxs[lane], oys[lane], ozs[lane] = r.Origin.X, r.Origin.Y, r.Origin.Z
				dxs[lane], dys[lane], dzs[lane] = r.Direction.X, r.Direction.Y, r.Direction.Z
			}
			ox, oy, oz := simd.LoadFloat32s(oxs), simd.LoadFloat32s(oys), simd.LoadFloat32s(ozs)
			dx, dy, dz := simd.LoadFloat32s(dxs), simd.LoadFloat32s(dys), simd.LoadFloat32s(dzs)

			closestT := simd.BroadcastFloat32s(1000)
			for _, tri := range tris {
				t, _, _, _, _, _, hit := simdTriangleHit(ox, oy, oz, dx, dy, dz, tri, 0.001, 1000)
				closer := hit.And(t.Less(closestT))
				closestT = ifElseFixed(closer, t, closestT)
			}
			closestT.Store(ts)
		}
	}
}

func benchCubeTriangles() []scene.Triangle {
	c := float32(0.5)
	return []scene.Triangle{
		{V0: vec3.Vec3{X: -c, Y: -c, Z: -c - 1}, V1: vec3.Vec3{X: c, Y: -c, Z: -c - 1}, V2: vec3.Vec3{X: c, Y: c, Z: -c - 1}},
		{V0: vec3.Vec3{X: -c, Y: -c, Z: -c - 1}, V1: vec3.Vec3{X: c, Y: c, Z: -c - 1}, V2: vec3.Vec3{X: -c, Y: c, Z: -c - 1}},
	}
}

// Full end-to-end render benchmarks (camera + intersection + shading
// + AA + gamma + PPM writing), same resolution/samples for both, to
// see whether the SIMD intersection win (if any) survives being a
// smaller fraction of the total per-pixel cost once shading/AA/I/O
// are included.
const (
	benchRenderWidth  = 200
	benchRenderHeight = 112
	benchRenderSPP    = 4
)

func BenchmarkScalarFullRender(b *testing.B) {
	origin := vec3.Vec3{}
	lookat := vec3.Vec3{X: 0, Y: 0, Z: -1}
	for i := 0; i < b.N; i++ {
		if err := run("", "/dev/null", benchRenderWidth, benchRenderHeight, benchRenderSPP, origin, lookat); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSIMDFullRender(b *testing.B) {
	origin := vec3.Vec3{}
	lookat := vec3.Vec3{X: 0, Y: 0, Z: -1}
	for i := 0; i < b.N; i++ {
		if err := renderSIMD("", "/dev/null", benchRenderWidth, benchRenderHeight, benchRenderSPP, origin, lookat); err != nil {
			b.Fatal(err)
		}
	}
}
