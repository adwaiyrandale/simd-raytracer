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

// If a future toolchain fixes the IfElse branch-swap bug documented
// on ifElseFixed, these tests will start failing (the workaround
// would then double-invert) - that's the point: they're the signal
// to remove the workaround, not just a one-time check.

func randRay(rng *rand.Rand) ray.Ray {
	return ray.Ray{
		Origin:    vec3.Vec3{X: rng.Float32()*4 - 2, Y: rng.Float32()*4 - 2, Z: rng.Float32()*4 - 2},
		Direction: vec3.Vec3{X: rng.Float32()*2 - 1, Y: rng.Float32()*2 - 1, Z: rng.Float32()*2 - 1}.Normalize(),
	}
}

func TestSIMDSphereHitMatchesScalar(t *testing.T) {
	lanes := laneWidth()
	sph := scene.Sphere{Center: vec3.Vec3{X: 0, Y: 0, Z: -1}, Radius: 0.5}
	rng := rand.New(rand.NewSource(3))

	rays := make([]ray.Ray, lanes)
	oxs, oys, ozs := make([]float32, lanes), make([]float32, lanes), make([]float32, lanes)
	dxs, dys, dzs := make([]float32, lanes), make([]float32, lanes), make([]float32, lanes)
	for i := range rays {
		rays[i] = randRay(rng)
		oxs[i], oys[i], ozs[i] = rays[i].Origin.X, rays[i].Origin.Y, rays[i].Origin.Z
		dxs[i], dys[i], dzs[i] = rays[i].Direction.X, rays[i].Direction.Y, rays[i].Direction.Z
	}

	ox, oy, oz := simd.LoadFloat32s(oxs), simd.LoadFloat32s(oys), simd.LoadFloat32s(ozs)
	dx, dy, dz := simd.LoadFloat32s(dxs), simd.LoadFloat32s(dys), simd.LoadFloat32s(dzs)
	simdT, _, _, _, nx, ny, nz, hit := simdSphereHit(ox, oy, oz, dx, dy, dz, sph, 0.001, 1000)

	ts := make([]float32, lanes)
	simdT.Store(ts)
	nxs, nys, nzs := make([]float32, lanes), make([]float32, lanes), make([]float32, lanes)
	nx.Store(nxs)
	ny.Store(nys)
	nz.Store(nzs)
	hitFlags := make([]float32, lanes)
	maskToFlags(hit, hitFlags)

	for i := 0; i < lanes; i++ {
		_, scalarT, scalarOK := sph.Hit(rays[i], 0.001, 1000)
		simdOK := hitFlags[i] != 0
		if simdOK != scalarOK {
			t.Fatalf("lane %d: SIMD hit=%v, scalar hit=%v (ray %+v)", i, simdOK, scalarOK, rays[i])
		}
		if !scalarOK {
			continue
		}
		const eps = 1e-3
		if abs32SIMD(ts[i]-scalarT) > eps {
			t.Errorf("lane %d: SIMD t=%v, scalar t=%v", i, ts[i], scalarT)
		}
		wantN := sph.Normal(rays[i].At(scalarT))
		if abs32SIMD(nxs[i]-wantN.X) > eps || abs32SIMD(nys[i]-wantN.Y) > eps || abs32SIMD(nzs[i]-wantN.Z) > eps {
			t.Errorf("lane %d: SIMD normal=(%v,%v,%v), scalar normal=%v", i, nxs[i], nys[i], nzs[i], wantN)
		}
	}
}

func TestSIMDTriangleHitMatchesScalar(t *testing.T) {
	lanes := laneWidth()
	tri := scene.Triangle{
		V0: vec3.Vec3{X: -1, Y: -1, Z: -2},
		V1: vec3.Vec3{X: 1, Y: -1, Z: -2},
		V2: vec3.Vec3{X: 0, Y: 1, Z: -2},
	}
	rng := rand.New(rand.NewSource(4))

	rays := make([]ray.Ray, lanes)
	oxs, oys, ozs := make([]float32, lanes), make([]float32, lanes), make([]float32, lanes)
	dxs, dys, dzs := make([]float32, lanes), make([]float32, lanes), make([]float32, lanes)
	for i := range rays {
		rays[i] = ray.Ray{
			Origin:    vec3.Vec3{X: rng.Float32()*4 - 2, Y: rng.Float32()*4 - 2, Z: 0},
			Direction: vec3.Vec3{X: 0, Y: 0, Z: -1},
		}
		oxs[i], oys[i], ozs[i] = rays[i].Origin.X, rays[i].Origin.Y, rays[i].Origin.Z
		dxs[i], dys[i], dzs[i] = rays[i].Direction.X, rays[i].Direction.Y, rays[i].Direction.Z
	}

	ox, oy, oz := simd.LoadFloat32s(oxs), simd.LoadFloat32s(oys), simd.LoadFloat32s(ozs)
	dx, dy, dz := simd.LoadFloat32s(dxs), simd.LoadFloat32s(dys), simd.LoadFloat32s(dzs)
	simdT, _, _, _, _, _, hit := simdTriangleHit(ox, oy, oz, dx, dy, dz, tri, 0.001, 1000)

	ts := make([]float32, lanes)
	simdT.Store(ts)
	hitFlags := make([]float32, lanes)
	maskToFlags(hit, hitFlags)

	for i := 0; i < lanes; i++ {
		_, scalarT, scalarOK := tri.Hit(rays[i], 0.001, 1000)
		simdOK := hitFlags[i] != 0
		if simdOK != scalarOK {
			t.Fatalf("lane %d: SIMD hit=%v, scalar hit=%v (ray %+v)", i, simdOK, scalarOK, rays[i])
		}
		if !scalarOK {
			continue
		}
		const eps = 1e-3
		if abs32SIMD(ts[i]-scalarT) > eps {
			t.Errorf("lane %d: SIMD t=%v, scalar t=%v", i, ts[i], scalarT)
		}
	}
}

func abs32SIMD(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}
