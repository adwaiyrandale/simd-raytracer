// Package mesh loads triangle meshes from Wavefront OBJ files.
package mesh

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/adwaiyrandale/simd-raytracer/internal/scene"
	"github.com/adwaiyrandale/simd-raytracer/internal/vec3"
)

// Load parses a Wavefront OBJ stream into triangles. Only "v" (vertex)
// and "f" (face) lines are interpreted; texture/normal data on face
// lines (v/vt/vn) is accepted but ignored. Faces with more than 3
// vertices are fan-triangulated around the first vertex.
func Load(r io.Reader) ([]scene.Triangle, error) {
	var verts []vec3.Vec3
	var tris []scene.Triangle

	scanner := bufio.NewScanner(r)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		switch fields[0] {
		case "v":
			v, err := parseVertex(fields[1:])
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", lineNo, err)
			}
			verts = append(verts, v)
		case "f":
			faceTris, err := parseFace(fields[1:], verts)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", lineNo, err)
			}
			tris = append(tris, faceTris...)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return tris, nil
}

func parseVertex(fields []string) (vec3.Vec3, error) {
	if len(fields) < 3 {
		return vec3.Vec3{}, fmt.Errorf("vertex line needs 3 coordinates, got %d", len(fields))
	}
	x, err := strconv.ParseFloat(fields[0], 32)
	if err != nil {
		return vec3.Vec3{}, fmt.Errorf("invalid x: %w", err)
	}
	y, err := strconv.ParseFloat(fields[1], 32)
	if err != nil {
		return vec3.Vec3{}, fmt.Errorf("invalid y: %w", err)
	}
	z, err := strconv.ParseFloat(fields[2], 32)
	if err != nil {
		return vec3.Vec3{}, fmt.Errorf("invalid z: %w", err)
	}
	return vec3.Vec3{X: float32(x), Y: float32(y), Z: float32(z)}, nil
}

func parseFace(fields []string, verts []vec3.Vec3) ([]scene.Triangle, error) {
	if len(fields) < 3 {
		return nil, fmt.Errorf("face line needs at least 3 vertices, got %d", len(fields))
	}
	idx := make([]int, len(fields))
	for i, f := range fields {
		vi, err := parseFaceVertexIndex(f)
		if err != nil {
			return nil, err
		}
		if vi < 1 || vi > len(verts) {
			return nil, fmt.Errorf("face vertex index %d out of range (have %d vertices)", vi, len(verts))
		}
		idx[i] = vi - 1 // OBJ indices are 1-based
	}

	// Fan-triangulate around the first vertex.
	tris := make([]scene.Triangle, 0, len(idx)-2)
	for i := 1; i < len(idx)-1; i++ {
		tris = append(tris, scene.Triangle{
			V0: verts[idx[0]],
			V1: verts[idx[i]],
			V2: verts[idx[i+1]],
		})
	}
	return tris, nil
}

// parseFaceVertexIndex extracts the vertex index from a face field that
// may be "v", "v/vt", "v/vt/vn", or "v//vn".
func parseFaceVertexIndex(field string) (int, error) {
	vStr := field
	if i := strings.IndexByte(field, '/'); i >= 0 {
		vStr = field[:i]
	}
	v, err := strconv.Atoi(vStr)
	if err != nil {
		return 0, fmt.Errorf("invalid face vertex index %q: %w", field, err)
	}
	return v, nil
}
