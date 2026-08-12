package mesh

import (
	"strings"
	"testing"

	"github.com/adwaiyrandale/simd-raytracer/internal/vec3"
)

func TestLoadTriangle(t *testing.T) {
	src := `
# comment line should be ignored
v 0 0 0
v 1 0 0
v 0 1 0
f 1 2 3
`
	tris, err := Load(strings.NewReader(src))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(tris) != 1 {
		t.Fatalf("got %d triangles, want 1", len(tris))
	}
	want := [3]vec3.Vec3{{X: 0, Y: 0, Z: 0}, {X: 1, Y: 0, Z: 0}, {X: 0, Y: 1, Z: 0}}
	got := [3]vec3.Vec3{tris[0].V0, tris[0].V1, tris[0].V2}
	if got != want {
		t.Errorf("triangle vertices = %v, want %v", got, want)
	}
}

func TestLoadQuadTriangulated(t *testing.T) {
	// A quad face should be fan-triangulated into 2 triangles.
	src := `
v 0 0 0
v 1 0 0
v 1 1 0
v 0 1 0
f 1 2 3 4
`
	tris, err := Load(strings.NewReader(src))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(tris) != 2 {
		t.Fatalf("got %d triangles, want 2", len(tris))
	}
}

func TestLoadFaceWithSlashIndices(t *testing.T) {
	// f can reference vertex/texture/normal as v/vt/vn - only the vertex
	// index matters for this loader.
	src := `
v 0 0 0
v 1 0 0
v 0 1 0
vt 0 0
vn 0 0 1
f 1/1/1 2/1/1 3/1/1
`
	tris, err := Load(strings.NewReader(src))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(tris) != 1 {
		t.Fatalf("got %d triangles, want 1", len(tris))
	}
	want := vec3.Vec3{X: 1, Y: 0, Z: 0}
	if tris[0].V1 != want {
		t.Errorf("V1 = %v, want %v", tris[0].V1, want)
	}
}

func TestLoadInvalidFaceIndex(t *testing.T) {
	src := `
v 0 0 0
f 1 2 3
`
	_, err := Load(strings.NewReader(src))
	if err == nil {
		t.Fatalf("expected error for out-of-range face index")
	}
}
