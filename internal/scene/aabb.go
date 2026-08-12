package scene

import (
	"github.com/adwaiyrandale/simd-raytracer/internal/ray"
	"github.com/adwaiyrandale/simd-raytracer/internal/vec3"
)

// AABB is an axis-aligned bounding box.
type AABB struct {
	Min, Max vec3.Vec3
}

// Hit tests the box against r using the slab method, for t in
// [tMin, tMax]. It reports whether the ray intersects the box at all -
// callers use it as a cheap reject test before testing box contents.
func (b AABB) Hit(r ray.Ray, tMin, tMax float32) bool {
	for axis := 0; axis < 3; axis++ {
		origin, dir, lo, hi := axisComponents(axis, r, b)
		invD := 1 / dir
		t0 := (lo - origin) * invD
		t1 := (hi - origin) * invD
		if invD < 0 {
			t0, t1 = t1, t0
		}
		if t0 > tMin {
			tMin = t0
		}
		if t1 < tMax {
			tMax = t1
		}
		if tMax <= tMin {
			return false
		}
	}
	return true
}

func axisComponents(axis int, r ray.Ray, b AABB) (origin, dir, lo, hi float32) {
	switch axis {
	case 0:
		return r.Origin.X, r.Direction.X, b.Min.X, b.Max.X
	case 1:
		return r.Origin.Y, r.Direction.Y, b.Min.Y, b.Max.Y
	default:
		return r.Origin.Z, r.Direction.Z, b.Min.Z, b.Max.Z
	}
}

// Union returns the smallest AABB containing both a and b.
func (a AABB) Union(b AABB) AABB {
	return AABB{
		Min: vec3.Vec3{
			X: minF(a.Min.X, b.Min.X),
			Y: minF(a.Min.Y, b.Min.Y),
			Z: minF(a.Min.Z, b.Min.Z),
		},
		Max: vec3.Vec3{
			X: maxF(a.Max.X, b.Max.X),
			Y: maxF(a.Max.Y, b.Max.Y),
			Z: maxF(a.Max.Z, b.Max.Z),
		},
	}
}

// Bounds returns the AABB of the triangle.
func (tri Triangle) Bounds() AABB {
	return AABB{
		Min: vec3.Vec3{
			X: minF(tri.V0.X, minF(tri.V1.X, tri.V2.X)),
			Y: minF(tri.V0.Y, minF(tri.V1.Y, tri.V2.Y)),
			Z: minF(tri.V0.Z, minF(tri.V1.Z, tri.V2.Z)),
		},
		Max: vec3.Vec3{
			X: maxF(tri.V0.X, maxF(tri.V1.X, tri.V2.X)),
			Y: maxF(tri.V0.Y, maxF(tri.V1.Y, tri.V2.Y)),
			Z: maxF(tri.V0.Z, maxF(tri.V1.Z, tri.V2.Z)),
		},
	}
}

func minF(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func maxF(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
