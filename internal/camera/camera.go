// Package camera implements a perspective camera for primary ray
// generation.
package camera

import (
	"math"

	"github.com/adwaiyrandale/simd-raytracer/internal/ray"
	"github.com/adwaiyrandale/simd-raytracer/internal/vec3"
)

type Camera struct {
	origin          vec3.Vec3
	lowerLeftCorner vec3.Vec3
	horizontal      vec3.Vec3
	vertical        vec3.Vec3
}

// New builds a camera positioned at lookfrom, aimed at lookat, with vup
// defining the roll (which direction is "up"). vfovDeg is the vertical
// field of view in degrees, aspect is width/height.
func New(lookfrom, lookat, vup vec3.Vec3, vfovDeg, aspect float32) Camera {
	theta := vfovDeg * math.Pi / 180
	h := float32(math.Tan(float64(theta) / 2))
	viewportHeight := 2 * h
	viewportWidth := aspect * viewportHeight

	w := lookfrom.Sub(lookat).Normalize()
	u := vup.Cross(w).Normalize()
	v := w.Cross(u)

	horizontal := u.Scale(viewportWidth)
	vertical := v.Scale(viewportHeight)
	lowerLeftCorner := lookfrom.
		Sub(horizontal.Scale(0.5)).
		Sub(vertical.Scale(0.5)).
		Sub(w)

	return Camera{
		origin:          lookfrom,
		lowerLeftCorner: lowerLeftCorner,
		horizontal:      horizontal,
		vertical:        vertical,
	}
}

// Ray returns the primary ray through viewport coordinates (s, t), each
// in [0, 1].
func (c Camera) Ray(s, t float32) ray.Ray {
	dir := c.lowerLeftCorner.
		Add(c.horizontal.Scale(s)).
		Add(c.vertical.Scale(t)).
		Sub(c.origin)
	return ray.Ray{Origin: c.origin, Direction: dir}
}
