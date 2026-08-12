package scene

import (
	"github.com/adwaiyrandale/simd-raytracer/internal/ray"
	"github.com/adwaiyrandale/simd-raytracer/internal/vec3"
)

type Triangle struct {
	V0, V1, V2 vec3.Vec3
}

const epsilon = 1e-8

// Hit implements the Möller–Trumbore ray-triangle intersection algorithm.
func (tri Triangle) Hit(r ray.Ray, tMin, tMax float32) (point vec3.Vec3, t float32, ok bool) {
	edge1 := tri.V1.Sub(tri.V0)
	edge2 := tri.V2.Sub(tri.V0)
	pvec := r.Direction.Cross(edge2)
	det := edge1.Dot(pvec)

	if det > -epsilon && det < epsilon {
		return vec3.Vec3{}, 0, false // ray parallel to triangle plane
	}
	invDet := 1 / det

	tvec := r.Origin.Sub(tri.V0)
	u := tvec.Dot(pvec) * invDet
	if u < 0 || u > 1 {
		return vec3.Vec3{}, 0, false
	}

	qvec := tvec.Cross(edge1)
	v := r.Direction.Dot(qvec) * invDet
	if v < 0 || u+v > 1 {
		return vec3.Vec3{}, 0, false
	}

	tHit := edge2.Dot(qvec) * invDet
	if tHit < tMin || tHit > tMax {
		return vec3.Vec3{}, 0, false
	}
	return r.At(tHit), tHit, true
}

// Normal returns the triangle's (unnormalized-input, normalized-output)
// face normal using vertex winding order V0, V1, V2.
func (tri Triangle) Normal() vec3.Vec3 {
	edge1 := tri.V1.Sub(tri.V0)
	edge2 := tri.V2.Sub(tri.V0)
	return edge1.Cross(edge2).Normalize()
}
