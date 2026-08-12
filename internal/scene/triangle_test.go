package scene

import (
	"testing"

	"github.com/adwaiyrandale/simd-raytracer/internal/ray"
	"github.com/adwaiyrandale/simd-raytracer/internal/vec3"
)

func testTriangle() Triangle {
	return Triangle{
		V0: vec3.Vec3{X: -1, Y: -1, Z: -5},
		V1: vec3.Vec3{X: 1, Y: -1, Z: -5},
		V2: vec3.Vec3{X: 0, Y: 1, Z: -5},
	}
}

func TestTriangleHit(t *testing.T) {
	tri := testTriangle()
	r := ray.Ray{
		Origin:    vec3.Vec3{X: 0, Y: -0.3, Z: 0},
		Direction: vec3.Vec3{X: 0, Y: 0, Z: -1},
	}
	point, tt, ok := tri.Hit(r, 0.001, 1000)
	if !ok {
		t.Fatalf("expected hit")
	}
	if tt <= 0 {
		t.Errorf("expected positive t, got %v", tt)
	}
	wantPoint := vec3.Vec3{X: 0, Y: -0.3, Z: -5}
	const eps = 1e-5
	if abs32(point.X-wantPoint.X) > eps || abs32(point.Y-wantPoint.Y) > eps || abs32(point.Z-wantPoint.Z) > eps {
		t.Errorf("hit point = %v, want %v", point, wantPoint)
	}
}

func TestTriangleMissOutsideEdges(t *testing.T) {
	tri := testTriangle()
	r := ray.Ray{
		Origin:    vec3.Vec3{X: 5, Y: 5, Z: 0},
		Direction: vec3.Vec3{X: 0, Y: 0, Z: -1},
	}
	_, _, ok := tri.Hit(r, 0.001, 1000)
	if ok {
		t.Errorf("expected miss (ray outside triangle edges)")
	}
}

func TestTriangleMissParallel(t *testing.T) {
	tri := testTriangle()
	r := ray.Ray{
		Origin:    vec3.Vec3{X: 0, Y: 0, Z: 0},
		Direction: vec3.Vec3{X: 1, Y: 0, Z: 0},
	}
	_, _, ok := tri.Hit(r, 0.001, 1000)
	if ok {
		t.Errorf("expected miss (ray parallel to triangle plane)")
	}
}

func TestTriangleNormal(t *testing.T) {
	tri := testTriangle()
	n := tri.Normal()
	// Triangle lies in the z=-5 plane facing the +z direction (toward the
	// ray origins used in these tests).
	want := vec3.Vec3{X: 0, Y: 0, Z: 1}
	const eps = 1e-5
	if abs32(n.X-want.X) > eps || abs32(n.Y-want.Y) > eps || abs32(n.Z-want.Z) > eps {
		t.Errorf("Normal() = %v, want %v", n, want)
	}
}
