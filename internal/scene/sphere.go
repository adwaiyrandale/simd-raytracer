// Package scene implements scalar scene geometry for the reference
// (non-SIMD) path tracer.
package scene

import (
	"math"

	"github.com/adwaiy/simd-raytracer/internal/ray"
	"github.com/adwaiy/simd-raytracer/internal/vec3"
)

type Sphere struct {
	Center vec3.Vec3
	Radius float32
}

// Hit tests r against the sphere for parameter t in [tMin, tMax].
// Returns the hit point and t on success.
func (s Sphere) Hit(r ray.Ray, tMin, tMax float32) (point vec3.Vec3, t float32, ok bool) {
	oc := r.Origin.Sub(s.Center)
	a := r.Direction.Dot(r.Direction)
	b := 2 * oc.Dot(r.Direction)
	c := oc.Dot(oc) - s.Radius*s.Radius
	discriminant := b*b - 4*a*c
	if discriminant < 0 {
		return vec3.Vec3{}, 0, false
	}
	sqrtD := float32(math.Sqrt(float64(discriminant)))

	root := (-b - sqrtD) / (2 * a)
	if root < tMin || root > tMax {
		root = (-b + sqrtD) / (2 * a)
		if root < tMin || root > tMax {
			return vec3.Vec3{}, 0, false
		}
	}
	return r.At(root), root, true
}

// Normal returns the outward unit normal at point p on the sphere.
func (s Sphere) Normal(p vec3.Vec3) vec3.Vec3 {
	return p.Sub(s.Center).Normalize()
}
