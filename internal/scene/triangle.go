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

// Barycentric returns the barycentric weights (u, v, w) of point p with
// respect to (V0, V1, V2), where p = u*V0 + v*V1 + w*V2 and
// u+v+w == 1. p is assumed to lie in the triangle's plane (e.g. a
// point returned by Hit); no in-triangle check is performed.
func (tri Triangle) Barycentric(p vec3.Vec3) (u, v, w float32) {
	e0 := tri.V1.Sub(tri.V0)
	e1 := tri.V2.Sub(tri.V0)
	e2 := p.Sub(tri.V0)

	d00 := e0.Dot(e0)
	d01 := e0.Dot(e1)
	d11 := e1.Dot(e1)
	d20 := e2.Dot(e0)
	d21 := e2.Dot(e1)
	denom := d00*d11 - d01*d01

	v = (d11*d20 - d01*d21) / denom
	w = (d00*d21 - d01*d20) / denom
	u = 1 - v - w
	return u, v, w
}

// Normal returns the triangle's (unnormalized-input, normalized-output)
// face normal using vertex winding order V0, V1, V2.
func (tri Triangle) Normal() vec3.Vec3 {
	edge1 := tri.V1.Sub(tri.V0)
	edge2 := tri.V2.Sub(tri.V0)
	return edge1.Cross(edge2).Normalize()
}
