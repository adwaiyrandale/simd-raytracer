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

func TestTriangleBarycentricAtVertices(t *testing.T) {
	tri := testTriangle()
	cases := []struct {
		name  string
		point vec3.Vec3
		wantU float32
		wantV float32
		wantW float32
	}{
		{"V0", tri.V0, 1, 0, 0},
		{"V1", tri.V1, 0, 1, 0},
		{"V2", tri.V2, 0, 0, 1},
	}
	const eps = 1e-4
	for _, c := range cases {
		u, v, w := tri.Barycentric(c.point)
		if abs32(u-c.wantU) > eps || abs32(v-c.wantV) > eps || abs32(w-c.wantW) > eps {
			t.Errorf("%s: Barycentric() = (%v, %v, %v), want (%v, %v, %v)", c.name, u, v, w, c.wantU, c.wantV, c.wantW)
		}
	}
}

func TestTriangleBarycentricAtCentroid(t *testing.T) {
	tri := testTriangle()
	centroid := tri.V0.Add(tri.V1).Add(tri.V2).Scale(1.0 / 3)
	u, v, w := tri.Barycentric(centroid)
	const eps = 1e-4
	want := float32(1.0 / 3)
	if abs32(u-want) > eps || abs32(v-want) > eps || abs32(w-want) > eps {
		t.Errorf("centroid: Barycentric() = (%v, %v, %v), want all ~%v", u, v, w, want)
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
