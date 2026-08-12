package scene

import (
	"testing"

	"github.com/adwaiyrandale/simd-raytracer/internal/ray"
	"github.com/adwaiyrandale/simd-raytracer/internal/vec3"
)

func TestAABBHit(t *testing.T) {
	box := AABB{Min: vec3.Vec3{X: -1, Y: -1, Z: -1}, Max: vec3.Vec3{X: 1, Y: 1, Z: 1}}
	r := ray.Ray{Origin: vec3.Vec3{X: 0, Y: 0, Z: 5}, Direction: vec3.Vec3{X: 0, Y: 0, Z: -1}}
	if !box.Hit(r, 0.001, 1000) {
		t.Errorf("expected hit")
	}
}

func TestAABBMiss(t *testing.T) {
	box := AABB{Min: vec3.Vec3{X: -1, Y: -1, Z: -1}, Max: vec3.Vec3{X: 1, Y: 1, Z: 1}}
	r := ray.Ray{Origin: vec3.Vec3{X: 5, Y: 5, Z: 5}, Direction: vec3.Vec3{X: 0, Y: 0, Z: -1}}
	if box.Hit(r, 0.001, 1000) {
		t.Errorf("expected miss")
	}
}

func TestAABBHitOutOfTRange(t *testing.T) {
	box := AABB{Min: vec3.Vec3{X: -1, Y: -1, Z: -1}, Max: vec3.Vec3{X: 1, Y: 1, Z: 1}}
	r := ray.Ray{Origin: vec3.Vec3{X: 0, Y: 0, Z: 5}, Direction: vec3.Vec3{X: 0, Y: 0, Z: -1}}
	// The box is hit at t=4, so a tMax of 1 should miss.
	if box.Hit(r, 0.001, 1) {
		t.Errorf("expected miss (box hit beyond tMax)")
	}
}

func TestAABBHitDegenerateFlatBox(t *testing.T) {
	// A box with zero thickness along one axis (e.g. a bounding box over
	// coplanar triangles) must still register a hit when a ray passes
	// exactly through that plane, not reject it as an empty interval.
	box := AABB{Min: vec3.Vec3{X: -5, Y: -5, Z: -10}, Max: vec3.Vec3{X: 5, Y: 5, Z: -10}}
	r := ray.Ray{Origin: vec3.Vec3{X: 0, Y: 0, Z: 0}, Direction: vec3.Vec3{X: 0.1, Y: -0.3, Z: -0.9}.Normalize()}
	if !box.Hit(r, 0.001, 1000) {
		t.Errorf("expected hit on degenerate flat box")
	}
}

func TestAABBUnion(t *testing.T) {
	a := AABB{Min: vec3.Vec3{X: -1, Y: 0, Z: 0}, Max: vec3.Vec3{X: 0, Y: 1, Z: 1}}
	b := AABB{Min: vec3.Vec3{X: 0, Y: -1, Z: -1}, Max: vec3.Vec3{X: 1, Y: 0, Z: 0}}
	got := a.Union(b)
	want := AABB{Min: vec3.Vec3{X: -1, Y: -1, Z: -1}, Max: vec3.Vec3{X: 1, Y: 1, Z: 1}}
	if got != want {
		t.Errorf("Union() = %v, want %v", got, want)
	}
}

func TestTriangleBounds(t *testing.T) {
	tri := Triangle{
		V0: vec3.Vec3{X: -1, Y: -1, Z: -5},
		V1: vec3.Vec3{X: 1, Y: -1, Z: -5},
		V2: vec3.Vec3{X: 0, Y: 1, Z: -5},
	}
	got := tri.Bounds()
	want := AABB{Min: vec3.Vec3{X: -1, Y: -1, Z: -5}, Max: vec3.Vec3{X: 1, Y: 1, Z: -5}}
	if got != want {
		t.Errorf("Bounds() = %v, want %v", got, want)
	}
}
