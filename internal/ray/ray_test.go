package ray

import (
	"testing"

	"github.com/adwaiy/simd-raytracer/internal/vec3"
)

func TestAt(t *testing.T) {
	r := Ray{
		Origin:    vec3.Vec3{X: 1, Y: 1, Z: 1},
		Direction: vec3.Vec3{X: 1, Y: 0, Z: 0},
	}
	got := r.At(3)
	want := vec3.Vec3{X: 4, Y: 1, Z: 1}
	if got != want {
		t.Errorf("At(3) = %v, want %v", got, want)
	}
}
