package camera

import (
	"testing"

	"github.com/adwaiyrandale/simd-raytracer/internal/vec3"
)

func abs32(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}

func TestCenterRayPointsAtLookAt(t *testing.T) {
	lookfrom := vec3.Vec3{X: 0, Y: 0, Z: 0}
	lookat := vec3.Vec3{X: 0, Y: 0, Z: -1}
	vup := vec3.Vec3{X: 0, Y: 1, Z: 0}
	c := New(lookfrom, lookat, vup, 90, 1)

	r := c.Ray(0.5, 0.5)
	got := r.Direction.Normalize()
	want := lookat.Sub(lookfrom).Normalize()
	const eps = 1e-4
	if abs32(got.X-want.X) > eps || abs32(got.Y-want.Y) > eps || abs32(got.Z-want.Z) > eps {
		t.Errorf("center ray direction = %v, want %v", got, want)
	}
}

func TestCenterRayPointsAtLookAtOffAxis(t *testing.T) {
	lookfrom := vec3.Vec3{X: 2, Y: 2, Z: 2}
	lookat := vec3.Vec3{X: 0, Y: 0, Z: 0}
	vup := vec3.Vec3{X: 0, Y: 1, Z: 0}
	c := New(lookfrom, lookat, vup, 60, 16.0/9.0)

	r := c.Ray(0.5, 0.5)
	got := r.Direction.Normalize()
	want := lookat.Sub(lookfrom).Normalize()
	const eps = 1e-4
	if abs32(got.X-want.X) > eps || abs32(got.Y-want.Y) > eps || abs32(got.Z-want.Z) > eps {
		t.Errorf("center ray direction = %v, want %v", got, want)
	}
}

func TestRayOriginIsLookfrom(t *testing.T) {
	lookfrom := vec3.Vec3{X: 1, Y: 2, Z: 3}
	c := New(lookfrom, vec3.Vec3{X: 0, Y: 0, Z: 0}, vec3.Vec3{X: 0, Y: 1, Z: 0}, 45, 1.5)
	r := c.Ray(0.2, 0.8)
	if r.Origin != lookfrom {
		t.Errorf("Ray origin = %v, want %v", r.Origin, lookfrom)
	}
}

func TestCornersDivergeFromCenter(t *testing.T) {
	// Sanity check the viewport actually has extent: corner rays should
	// differ from the center ray, not degenerate to the same direction.
	c := New(vec3.Vec3{X: 0, Y: 0, Z: 0}, vec3.Vec3{X: 0, Y: 0, Z: -1}, vec3.Vec3{X: 0, Y: 1, Z: 0}, 90, 1)
	center := c.Ray(0.5, 0.5).Direction
	corner := c.Ray(0, 0).Direction
	if center == corner {
		t.Errorf("corner ray should differ from center ray")
	}
}
