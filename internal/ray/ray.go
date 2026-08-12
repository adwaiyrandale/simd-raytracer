// Package ray implements the scalar ray type for the reference
// (non-SIMD) path tracer.
package ray

import "github.com/adwaiy/simd-raytracer/internal/vec3"

type Ray struct {
	Origin    vec3.Vec3
	Direction vec3.Vec3
}

// At returns the point along the ray at parameter t.
func (r Ray) At(t float32) vec3.Vec3 {
	return r.Origin.Add(r.Direction.Scale(t))
}
