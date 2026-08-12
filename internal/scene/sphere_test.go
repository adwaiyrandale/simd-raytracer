package scene

import (
	"testing"

	"github.com/adwaiyrandale/simd-raytracer/internal/ray"
	"github.com/adwaiyrandale/simd-raytracer/internal/vec3"
)

func TestSphereHit(t *testing.T) {
	s := Sphere{Center: vec3.Vec3{X: 0, Y: 0, Z: -5}, Radius: 1}
	r := ray.Ray{
		Origin:    vec3.Vec3{X: 0, Y: 0, Z: 0},
		Direction: vec3.Vec3{X: 0, Y: 0, Z: -1},
	}
	hit, t0, ok := s.Hit(r, 0.001, 1000)
	if !ok {
		t.Fatalf("expected hit")
	}
	if t0 <= 0 {
		t.Errorf("expected positive t, got %v", t0)
	}
	wantPoint := vec3.Vec3{X: 0, Y: 0, Z: -4}
	if hit != wantPoint {
		t.Errorf("hit point = %v, want %v", hit, wantPoint)
	}
}

func TestSphereMiss(t *testing.T) {
	s := Sphere{Center: vec3.Vec3{X: 5, Y: 5, Z: -5}, Radius: 1}
	r := ray.Ray{
		Origin:    vec3.Vec3{X: 0, Y: 0, Z: 0},
		Direction: vec3.Vec3{X: 0, Y: 0, Z: -1},
	}
	_, _, ok := s.Hit(r, 0.001, 1000)
	if ok {
		t.Errorf("expected miss")
	}
}
