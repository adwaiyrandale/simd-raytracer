//go:build goexperiment.simd

// This file lives in package main (not a separate internal/ package)
// for a specific reason: a confirmed go1.27rc1 compiler bug makes a
// struct with simd.Float32s fields fail `go build`/`go run` whenever
// it crosses a package-archive boundary (i.e. gets imported by
// another package), even though it works fine within a single
// self-contained build. See internal/simdvec/vec3s.go for the full
// writeup and reproduction. Keeping the packet type and all its users
// in this one package sidesteps the bug entirely.
package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"simd"

	"github.com/adwaiyrandale/simd-raytracer/internal/camera"
	"github.com/adwaiyrandale/simd-raytracer/internal/ray"
	"github.com/adwaiyrandale/simd-raytracer/internal/scene"
	"github.com/adwaiyrandale/simd-raytracer/internal/vec3"
)

// laneWidth returns this hardware's simd.Float32s vector width (e.g.
// 4 for 128-bit NEON/SSE). 16 covers the widest case in use today
// (AVX-512).
func laneWidth() int {
	var probe [16]float32
	return simd.LoadFloat32s(probe[:]).Len()
}

// renderSIMD renders the same two demo scenes as run() (a sphere, or
// an OBJ mesh via -obj) but with ray-object intersection done as SIMD
// packets instead of one ray at a time. Primary ray generation and
// final per-pixel shading stay scalar (they're cheap and not the
// point of this exercise); intersection - the actual hot path in any
// ray tracer - is where the SIMD work happens.
//
// The mesh path is brute-force over all triangles (no SIMD BVH yet -
// that's future phase-3 work) since a 12-triangle test cube doesn't
// need acceleration to prove the packet intersection math is correct
// and fast.
func renderSIMD(objPath, outPath string, width, height, samplesPerPixel int, lookfrom, lookat vec3.Vec3) error {
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	light := vec3.Vec3{X: -1, Y: 1, Z: 1}.Normalize()
	cam := camera.New(lookfrom, lookat, vec3.Vec3{X: 0, Y: 1, Z: 0}, 90, float32(width)/float32(height))

	var tris []scene.Triangle
	var sphere scene.Sphere
	useMesh := objPath != ""
	if useMesh {
		tris, err = loadMeshTriangles(objPath, vec3.Vec3{X: 0, Y: 0, Z: -3})
		if err != nil {
			return fmt.Errorf("loading %s: %w", objPath, err)
		}
	} else {
		sphere = scene.Sphere{Center: vec3.Vec3{X: 0, Y: 0, Z: -1}, Radius: 0.5}
	}

	lanes := laneWidth()
	rng := rand.New(rand.NewSource(1))
	fmt.Fprintf(w, "P3\n%d %d\n255\n", width, height)

	sums := make([]vec3.Vec3, width)
	rayBuf := make([]ray.Ray, lanes)
	oxs, oys, ozs := make([]float32, lanes), make([]float32, lanes), make([]float32, lanes)
	dxs, dys, dzs := make([]float32, lanes), make([]float32, lanes), make([]float32, lanes)
	colorBuf := make([]vec3.Vec3, lanes)
	scratch := newSIMDScratch(lanes)

	for j := height - 1; j >= 0; j-- {
		for i := range sums {
			sums[i] = vec3.Vec3{}
		}
		for s := 0; s < samplesPerPixel; s++ {
			for i0 := 0; i0 < width; i0 += lanes {
				n := lanes
				if i0+n > width {
					n = width - i0
				}
				for lane := 0; lane < lanes; lane++ {
					col := i0 + lane
					if col >= width {
						col = width - 1 // pad lanes past the edge: harmless repeated work
					}
					u := (float32(col) + rng.Float32()) / float32(width)
					v := (float32(j) + rng.Float32()) / float32(height)
					r := cam.Ray(u, v)
					rayBuf[lane] = r
					oxs[lane], oys[lane], ozs[lane] = r.Origin.X, r.Origin.Y, r.Origin.Z
					dxs[lane], dys[lane], dzs[lane] = r.Direction.X, r.Direction.Y, r.Direction.Z
				}

				ox, oy, oz := simd.LoadFloat32s(oxs), simd.LoadFloat32s(oys), simd.LoadFloat32s(ozs)
				dx, dy, dz := simd.LoadFloat32s(dxs), simd.LoadFloat32s(dys), simd.LoadFloat32s(dzs)

				if useMesh {
					simdShadeMesh(ox, oy, oz, dx, dy, dz, tris, rayBuf, light, lanes, colorBuf, scratch)
				} else {
					simdShadeSphere(ox, oy, oz, dx, dy, dz, sphere, rayBuf, light, lanes, colorBuf, scratch)
				}

				for lane := 0; lane < n; lane++ {
					sums[i0+lane] = sums[i0+lane].Add(colorBuf[lane])
				}
			}
		}
		for i := 0; i < width; i++ {
			writeColor(w, sums[i].Scale(1/float32(samplesPerPixel)))
		}
	}
	return w.Flush()
}

const simdEpsilon = 1e-8

// ifElseFixed is a workaround for a confirmed go1.27rc1 compiler bug
// on this arm64 backend: Float32s.IfElse's branches are swapped in
// codegen (x.IfElse(mask, y) actually evaluates to "y where mask is
// true, x where false" - the exact opposite of its documented
// contract). Reproduced and isolated outside this repo with a minimal
// program before assuming it was a bug in this code, not upstream:
// comparisons (GreaterEqual/Less/etc) and And/Or were verified
// correct in isolation; only IfElse's branch selection is swapped.
// This function swaps the receiver/argument to compensate, so callers
// can use it with IfElse's documented (correct) semantics and this is
// the only place the workaround needs to live.
func ifElseFixed(mask simd.Mask32s, whenTrue, whenFalse simd.Float32s) simd.Float32s {
	return whenFalse.IfElse(mask, whenTrue)
}

// simdSphereHit intersects a packet of rays against one sphere,
// mirroring scene.Sphere.Hit's branchless nearest-root selection but
// vectorized across lanes with masks instead of per-ray branches.
func simdSphereHit(ox, oy, oz, dx, dy, dz simd.Float32s, sph scene.Sphere, tMin, tMax float32) (t, px, py, pz, nx, ny, nz simd.Float32s, hit simd.Mask32s) {
	cx, cy, cz := simd.BroadcastFloat32s(sph.Center.X), simd.BroadcastFloat32s(sph.Center.Y), simd.BroadcastFloat32s(sph.Center.Z)
	ocx, ocy, ocz := ox.Sub(cx), oy.Sub(cy), oz.Sub(cz)

	a := dx.Mul(dx).Add(dy.Mul(dy)).Add(dz.Mul(dz))
	halfB := ocx.Mul(dx).Add(ocy.Mul(dy)).Add(ocz.Mul(dz))
	rr := simd.BroadcastFloat32s(sph.Radius * sph.Radius)
	c := ocx.Mul(ocx).Add(ocy.Mul(ocy)).Add(ocz.Mul(ocz)).Sub(rr)

	disc := halfB.Mul(halfB).Sub(a.Mul(c))
	zero := simd.BroadcastFloat32s(0)
	discPos := disc.GreaterEqual(zero)
	sqrtD := disc.Max(zero).Sqrt() // clamp before Sqrt: inactive (disc<0) lanes would NaN otherwise

	tMinB, tMaxB := simd.BroadcastFloat32s(tMin), simd.BroadcastFloat32s(tMax)

	root1 := halfB.Neg().Sub(sqrtD).Div(a)
	root1Valid := root1.GreaterEqual(tMinB).And(root1.LessEqual(tMaxB)).And(discPos)

	root2 := halfB.Neg().Add(sqrtD).Div(a)
	root2Valid := root2.GreaterEqual(tMinB).And(root2.LessEqual(tMaxB)).And(discPos)

	t = ifElseFixed(root1Valid, root1, root2)
	hit = root1Valid.Or(root2Valid)

	px = ox.Add(dx.Mul(t))
	py = oy.Add(dy.Mul(t))
	pz = oz.Add(dz.Mul(t))

	radiusB := simd.BroadcastFloat32s(sph.Radius)
	nx = px.Sub(cx).Div(radiusB)
	ny = py.Sub(cy).Div(radiusB)
	nz = pz.Sub(cz).Div(radiusB)
	return t, px, py, pz, nx, ny, nz, hit
}

// simdTriangleHit intersects a packet of rays against one triangle
// via vectorized Moller-Trumbore, mirroring scene.Triangle.Hit.
func simdTriangleHit(ox, oy, oz, dx, dy, dz simd.Float32s, tri scene.Triangle, tMin, tMax float32) (t, px, py, pz, baryU, baryV simd.Float32s, hit simd.Mask32s) {
	edge1 := tri.V1.Sub(tri.V0)
	edge2 := tri.V2.Sub(tri.V0)
	e1x, e1y, e1z := simd.BroadcastFloat32s(edge1.X), simd.BroadcastFloat32s(edge1.Y), simd.BroadcastFloat32s(edge1.Z)
	e2x, e2y, e2z := simd.BroadcastFloat32s(edge2.X), simd.BroadcastFloat32s(edge2.Y), simd.BroadcastFloat32s(edge2.Z)

	// pvec = D x edge2
	pvx := dy.Mul(e2z).Sub(dz.Mul(e2y))
	pvy := dz.Mul(e2x).Sub(dx.Mul(e2z))
	pvz := dx.Mul(e2y).Sub(dy.Mul(e2x))

	det := e1x.Mul(pvx).Add(e1y.Mul(pvy)).Add(e1z.Mul(pvz))
	nonParallel := det.Abs().GreaterEqual(simd.BroadcastFloat32s(simdEpsilon))
	invDet := simd.BroadcastFloat32s(1).Div(det)

	v0x, v0y, v0z := simd.BroadcastFloat32s(tri.V0.X), simd.BroadcastFloat32s(tri.V0.Y), simd.BroadcastFloat32s(tri.V0.Z)
	tvx, tvy, tvz := ox.Sub(v0x), oy.Sub(v0y), oz.Sub(v0z)

	u := tvx.Mul(pvx).Add(tvy.Mul(pvy)).Add(tvz.Mul(pvz)).Mul(invDet)
	zero, one := simd.BroadcastFloat32s(0), simd.BroadcastFloat32s(1)
	uValid := u.GreaterEqual(zero).And(u.LessEqual(one))

	qvx := tvy.Mul(e1z).Sub(tvz.Mul(e1y))
	qvy := tvz.Mul(e1x).Sub(tvx.Mul(e1z))
	qvz := tvx.Mul(e1y).Sub(tvy.Mul(e1x))

	v := dx.Mul(qvx).Add(dy.Mul(qvy)).Add(dz.Mul(qvz)).Mul(invDet)
	uv := u.Add(v)
	vValid := v.GreaterEqual(zero).And(uv.LessEqual(one))

	tHit := e2x.Mul(qvx).Add(e2y.Mul(qvy)).Add(e2z.Mul(qvz)).Mul(invDet)
	tMinB, tMaxB := simd.BroadcastFloat32s(tMin), simd.BroadcastFloat32s(tMax)
	tValid := tHit.GreaterEqual(tMinB).And(tHit.LessEqual(tMaxB))

	hit = nonParallel.And(uValid).And(vValid).And(tValid)
	px = ox.Add(dx.Mul(tHit))
	py = oy.Add(dy.Mul(tHit))
	pz = oz.Add(dz.Mul(tHit))
	return tHit, px, py, pz, u, v, hit
}

// simdScratch holds all the per-packet working buffers used by
// simdShadeSphere/simdShadeMesh, allocated once by newSIMDScratch and
// reused across every packet in a render. Without this, allocating
// fresh slices inside the innermost per-triangle/per-packet loop
// dominates render time far more than any SIMD math benefit -
// measured via `go test -bench -benchmem` before this fix: ~1MB and
// ~22k allocations per single mesh render, vs the scalar path's ~9KB
// and 6 allocations. See bench_test.go's commit message for the full
// before/after numbers.
type simdScratch struct {
	flagsBuf      []float32
	nxs, nys, nzs []float32
	us, vs        []float32
	closestTriIdx []int
}

func newSIMDScratch(lanes int) *simdScratch {
	return &simdScratch{
		flagsBuf:      make([]float32, lanes),
		nxs:           make([]float32, lanes),
		nys:           make([]float32, lanes),
		nzs:           make([]float32, lanes),
		us:            make([]float32, lanes),
		vs:            make([]float32, lanes),
		closestTriIdx: make([]int, lanes),
	}
}

// maskToFlags converts a Mask32s to 1s (true) and 0s (false) into the
// caller-provided out buffer (len must be >= lanes), one per lane.
// Mask32s.ToInt32s is declared in go1.27rc1's stub surface but not
// actually implemented for this arm64 backend yet ("Mask32x4 has no
// field or method ToInt32s" at compile time) - this works around it
// using only IfElse/Store, both confirmed working.
func maskToFlags(mask simd.Mask32s, out []float32) {
	flags := ifElseFixed(mask, simd.BroadcastFloat32s(1), simd.BroadcastFloat32s(0))
	flags.Store(out)
}

// simdShadeSphere writes lanes shaded colors into out (len must be >=
// lanes). scratch's buffers are reused, not allocated, per call.
func simdShadeSphere(ox, oy, oz, dx, dy, dz simd.Float32s, sph scene.Sphere, rays []ray.Ray, light vec3.Vec3, lanes int, out []vec3.Vec3, scratch *simdScratch) {
	_, _, _, _, nx, ny, nz, hit := simdSphereHit(ox, oy, oz, dx, dy, dz, sph, 0.001, 1000)

	nx.Store(scratch.nxs)
	ny.Store(scratch.nys)
	nz.Store(scratch.nzs)
	maskToFlags(hit, scratch.flagsBuf)

	for i := 0; i < lanes; i++ {
		if scratch.flagsBuf[i] != 0 {
			n := vec3.Vec3{X: scratch.nxs[i], Y: scratch.nys[i], Z: scratch.nzs[i]}
			out[i] = litColor(n, rays[i].Direction, light, vec3.Vec3{X: 1, Y: 0.3, Z: 0.3})
		} else {
			out[i] = skyColor(rays[i])
		}
	}
}

// simdShadeMesh writes lanes shaded colors into out (len must be >=
// lanes). scratch's buffers are reused, not allocated, per call -
// this includes the per-triangle loop below, which previously
// allocated a fresh []float32 on every single triangle iteration.
func simdShadeMesh(ox, oy, oz, dx, dy, dz simd.Float32s, tris []scene.Triangle, rays []ray.Ray, light vec3.Vec3, lanes int, out []vec3.Vec3, scratch *simdScratch) {
	tMaxInit := float32(1000)
	closestT := simd.BroadcastFloat32s(tMaxInit)
	var closestU, closestV simd.Float32s
	anyHit := simd.BroadcastFloat32s(0).GreaterEqual(simd.BroadcastFloat32s(1))
	for i := range scratch.closestTriIdx {
		scratch.closestTriIdx[i] = -1
	}

	for triIdx, tri := range tris {
		t, _, _, _, u, v, hit := simdTriangleHit(ox, oy, oz, dx, dy, dz, tri, 0.001, tMaxInit)
		closer := hit.And(t.Less(closestT))

		closestU = ifElseFixed(closer, u, closestU)
		closestV = ifElseFixed(closer, v, closestV)
		closestT = ifElseFixed(closer, t, closestT)
		anyHit = anyHit.Or(closer)

		maskToFlags(closer, scratch.flagsBuf)
		for lane := 0; lane < lanes; lane++ {
			if scratch.flagsBuf[lane] != 0 {
				scratch.closestTriIdx[lane] = triIdx
			}
		}
	}

	closestU.Store(scratch.us)
	closestV.Store(scratch.vs)
	maskToFlags(anyHit, scratch.flagsBuf)

	for i := 0; i < lanes; i++ {
		if scratch.flagsBuf[i] == 0 {
			out[i] = skyColor(rays[i])
			continue
		}
		triIdx := scratch.closestTriIdx[i]
		bw := scratch.us[i]
		bv := scratch.vs[i]
		bu := 1 - bw - bv
		if minOf3(bu, bw, bv) < edgeBarycentricThreshold {
			out[i] = vec3.Vec3{} // black edge line, matches scalar shadeMesh
			continue
		}
		n := tris[triIdx].Normal()
		out[i] = litColor(n, rays[i].Direction, light, vec3.Vec3{X: 0.3, Y: 0.6, Z: 1})
	}
}
