package bvh

import (
	"math/rand"
	"testing"

	"github.com/adwaiyrandale/simd-raytracer/internal/ray"
	"github.com/adwaiyrandale/simd-raytracer/internal/scene"
	"github.com/adwaiyrandale/simd-raytracer/internal/vec3"
)

func threeDisjointTriangles() []scene.Triangle {
	return []scene.Triangle{
		{ // nearest, z=-2
			V0: vec3.Vec3{X: -1, Y: -1, Z: -2}, V1: vec3.Vec3{X: 1, Y: -1, Z: -2}, V2: vec3.Vec3{X: 0, Y: 1, Z: -2},
		},
		{ // middle, z=-5
			V0: vec3.Vec3{X: -1, Y: -1, Z: -5}, V1: vec3.Vec3{X: 1, Y: -1, Z: -5}, V2: vec3.Vec3{X: 0, Y: 1, Z: -5},
		},
		{ // farthest, z=-8
			V0: vec3.Vec3{X: -1, Y: -1, Z: -8}, V1: vec3.Vec3{X: 1, Y: -1, Z: -8}, V2: vec3.Vec3{X: 0, Y: 1, Z: -8},
		},
	}
}

func TestBVHHitReturnsClosest(t *testing.T) {
	tris := threeDisjointTriangles()
	root := Build(tris)
	r := ray.Ray{Origin: vec3.Vec3{X: 0, Y: -0.3, Z: 0}, Direction: vec3.Vec3{X: 0, Y: 0, Z: -1}}

	hit, ok := root.Hit(r, 0.001, 1000)
	if !ok {
		t.Fatalf("expected hit")
	}
	wantZ := float32(-2) // nearest triangle
	if hit.Point.Z != wantZ {
		t.Errorf("hit.Point.Z = %v, want %v (should hit nearest triangle, not just any)", hit.Point.Z, wantZ)
	}
}

func TestBVHMiss(t *testing.T) {
	tris := threeDisjointTriangles()
	root := Build(tris)
	r := ray.Ray{Origin: vec3.Vec3{X: 50, Y: 50, Z: 0}, Direction: vec3.Vec3{X: 0, Y: 0, Z: -1}}

	_, ok := root.Hit(r, 0.001, 1000)
	if ok {
		t.Errorf("expected miss")
	}
}

func TestBVHEmptyTriangles(t *testing.T) {
	root := Build(nil)
	r := ray.Ray{Origin: vec3.Vec3{X: 0, Y: 0, Z: 0}, Direction: vec3.Vec3{X: 0, Y: 0, Z: -1}}

	_, ok := root.Hit(r, 0.001, 1000)
	if ok {
		t.Errorf("expected miss on empty BVH")
	}
}

// TestBVHMatchesBruteForce builds a BVH over a grid of many small
// triangles and fires random rays at it, checking every result against
// a brute-force linear scan over the same triangles. This is the
// scalar correctness oracle referenced in the design doc.
func TestBVHMatchesBruteForce(t *testing.T) {
	var tris []scene.Triangle
	for gx := -5; gx <= 5; gx++ {
		for gy := -5; gy <= 5; gy++ {
			x := float32(gx)
			y := float32(gy)
			tris = append(tris, scene.Triangle{
				V0: vec3.Vec3{X: x, Y: y, Z: -10},
				V1: vec3.Vec3{X: x + 0.8, Y: y, Z: -10},
				V2: vec3.Vec3{X: x, Y: y + 0.8, Z: -10},
			})
		}
	}
	root := Build(tris)

	rng := rand.New(rand.NewSource(42))
	for i := 0; i < 200; i++ {
		r := ray.Ray{
			Origin:    vec3.Vec3{X: 0, Y: 0, Z: 0},
			Direction: vec3.Vec3{X: rng.Float32()*12 - 6, Y: rng.Float32()*12 - 6, Z: -10}.Normalize(),
		}

		bvhHit, bvhOK := root.Hit(r, 0.001, 1000)
		_, bruteT, bruteOK := bruteForceHit(tris, r, 0.001, 1000)

		if bvhOK != bruteOK {
			t.Fatalf("ray %d: bvh hit=%v, brute force hit=%v", i, bvhOK, bruteOK)
		}
		if !bvhOK {
			continue
		}
		const eps = 1e-4
		if abs32(bvhHit.T-bruteT) > eps {
			t.Errorf("ray %d: bvh t=%v, brute force t=%v", i, bvhHit.T, bruteT)
		}
	}
}

func bruteForceHit(tris []scene.Triangle, r ray.Ray, tMin, tMax float32) (point vec3.Vec3, t float32, ok bool) {
	closestT := tMax
	found := false
	var closestPoint vec3.Vec3
	for _, tri := range tris {
		p, tt, hitOK := tri.Hit(r, tMin, closestT)
		if hitOK {
			found = true
			closestT = tt
			closestPoint = p
		}
	}
	return closestPoint, closestT, found
}

func abs32(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}
