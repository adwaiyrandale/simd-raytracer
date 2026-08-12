// Package bvh implements a scalar bounding volume hierarchy over
// triangles. This is the phase-2 (correctness) implementation: builds
// use a simple median split on the longest axis rather than SAH, and
// traversal is single-ray, not packet-based. Phase 3 introduces a
// SIMD-traversed, packet-oriented structure on top of these same
// build/traversal ideas.
package bvh

import (
	"sort"

	"github.com/adwaiyrandale/simd-raytracer/internal/ray"
	"github.com/adwaiyrandale/simd-raytracer/internal/scene"
	"github.com/adwaiyrandale/simd-raytracer/internal/vec3"
)

// leafSize is the maximum number of triangles stored in a leaf node
// before the builder splits further.
const leafSize = 4

// Node is a BVH tree node. Leaf nodes have Tris set and Left/Right nil;
// internal nodes have Left/Right set and Tris nil.
type Node struct {
	Bounds      scene.AABB
	Left, Right *Node
	Tris        []scene.Triangle
}

// Hit is the result of a successful BVH traversal.
type Hit struct {
	Point    vec3.Vec3
	T        float32
	Triangle scene.Triangle
}

// Build constructs a BVH over tris. A nil or empty slice produces a
// valid empty tree that always misses.
func Build(tris []scene.Triangle) *Node {
	if len(tris) == 0 {
		return &Node{Tris: nil}
	}
	if len(tris) <= leafSize {
		return &Node{Bounds: boundsOf(tris), Tris: tris}
	}

	axis := longestAxis(tris)
	sorted := append([]scene.Triangle(nil), tris...)
	sort.Slice(sorted, func(i, j int) bool {
		return centroid(sorted[i], axis) < centroid(sorted[j], axis)
	})

	mid := len(sorted) / 2
	left := Build(sorted[:mid])
	right := Build(sorted[mid:])
	return &Node{
		Bounds: left.Bounds.Union(right.Bounds),
		Left:   left,
		Right:  right,
	}
}

// Hit traverses the tree for the closest triangle intersection with r
// in [tMin, tMax].
func (n *Node) Hit(r ray.Ray, tMin, tMax float32) (Hit, bool) {
	if n == nil || (n.Tris == nil && n.Left == nil) {
		return Hit{}, false
	}
	if n.Tris != nil {
		return hitTriangles(n.Tris, r, tMin, tMax)
	}
	if !n.Bounds.Hit(r, tMin, tMax) {
		return Hit{}, false
	}

	closest := tMax
	best := Hit{}
	found := false

	if hit, ok := n.Left.Hit(r, tMin, closest); ok {
		found = true
		closest = hit.T
		best = hit
	}
	if hit, ok := n.Right.Hit(r, tMin, closest); ok {
		found = true
		best = hit
	}
	return best, found
}

func hitTriangles(tris []scene.Triangle, r ray.Ray, tMin, tMax float32) (Hit, bool) {
	closest := tMax
	best := Hit{}
	found := false
	for _, tri := range tris {
		if point, t, ok := tri.Hit(r, tMin, closest); ok {
			found = true
			closest = t
			best = Hit{Point: point, T: t, Triangle: tri}
		}
	}
	return best, found
}

func boundsOf(tris []scene.Triangle) scene.AABB {
	b := tris[0].Bounds()
	for _, tri := range tris[1:] {
		b = b.Union(tri.Bounds())
	}
	return b
}

func centroid(tri scene.Triangle, axis int) float32 {
	switch axis {
	case 0:
		return (tri.V0.X + tri.V1.X + tri.V2.X) / 3
	case 1:
		return (tri.V0.Y + tri.V1.Y + tri.V2.Y) / 3
	default:
		return (tri.V0.Z + tri.V1.Z + tri.V2.Z) / 3
	}
}

func longestAxis(tris []scene.Triangle) int {
	b := boundsOf(tris)
	dx := b.Max.X - b.Min.X
	dy := b.Max.Y - b.Min.Y
	dz := b.Max.Z - b.Min.Z
	if dx > dy && dx > dz {
		return 0
	}
	if dy > dz {
		return 1
	}
	return 2
}
