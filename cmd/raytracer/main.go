// Command raytracer renders a scene to a PPM file. With no flags it
// renders a single shaded sphere (phase-1 scaffold). With -obj it
// loads a Wavefront OBJ mesh and renders it via the scalar BVH
// (phase-2 scaffold) - both exist to prove their respective pipelines
// end-to-end with scalar code before any SIMD work begins.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/adwaiyrandale/simd-raytracer/internal/bvh"
	"github.com/adwaiyrandale/simd-raytracer/internal/mesh"
	"github.com/adwaiyrandale/simd-raytracer/internal/ray"
	"github.com/adwaiyrandale/simd-raytracer/internal/scene"
	"github.com/adwaiyrandale/simd-raytracer/internal/vec3"
)

const (
	width  = 400
	height = 225
)

func main() {
	objPath := flag.String("obj", "", "path to a Wavefront OBJ file to render (renders a demo sphere if empty)")
	outPath := flag.String("out", "out.ppm", "output PPM path")
	flag.Parse()

	if err := run(*objPath, *outPath); err != nil {
		log.Fatal(err)
	}
}

func run(objPath, outPath string) error {
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	light := vec3.Vec3{X: -1, Y: 1, Z: 1}.Normalize()

	var shadeFn func(r ray.Ray) vec3.Vec3
	if objPath == "" {
		sphere := scene.Sphere{Center: vec3.Vec3{X: 0, Y: 0, Z: -1}, Radius: 0.5}
		shadeFn = func(r ray.Ray) vec3.Vec3 { return shadeSphere(r, sphere, light) }
	} else {
		root, err := loadMeshBVH(objPath, vec3.Vec3{X: 0, Y: 0, Z: -3})
		if err != nil {
			return fmt.Errorf("loading %s: %w", objPath, err)
		}
		shadeFn = func(r ray.Ray) vec3.Vec3 { return shadeMesh(r, root, light) }
	}

	return renderPPM(w, shadeFn)
}

// loadMeshBVH loads an OBJ file, translates its vertices by offset (so
// demo meshes authored around the origin land in front of the camera),
// and builds a BVH over the result.
func loadMeshBVH(path string, offset vec3.Vec3) (*bvh.Node, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tris, err := mesh.Load(f)
	if err != nil {
		return nil, err
	}
	for i, tri := range tris {
		tris[i] = scene.Triangle{
			V0: tri.V0.Add(offset),
			V1: tri.V1.Add(offset),
			V2: tri.V2.Add(offset),
		}
	}
	return bvh.Build(tris), nil
}

func renderPPM(w *bufio.Writer, shadeFn func(r ray.Ray) vec3.Vec3) error {
	const aspect = float32(width) / float32(height)
	const viewportHeight = 2.0
	const viewportWidth = viewportHeight * aspect
	const focalLength = 1.0

	origin := vec3.Vec3{}
	horizontal := vec3.Vec3{X: viewportWidth}
	vertical := vec3.Vec3{Y: viewportHeight}
	lowerLeft := origin.
		Sub(horizontal.Scale(0.5)).
		Sub(vertical.Scale(0.5)).
		Sub(vec3.Vec3{Z: focalLength})

	fmt.Fprintf(w, "P3\n%d %d\n255\n", width, height)
	for j := height - 1; j >= 0; j-- {
		for i := 0; i < width; i++ {
			u := float32(i) / float32(width-1)
			v := float32(j) / float32(height-1)
			dir := lowerLeft.
				Add(horizontal.Scale(u)).
				Add(vertical.Scale(v)).
				Sub(origin)
			r := ray.Ray{Origin: origin, Direction: dir}
			writeColor(w, shadeFn(r))
		}
	}
	return w.Flush()
}

func shadeSphere(r ray.Ray, s scene.Sphere, light vec3.Vec3) vec3.Vec3 {
	if point, _, ok := s.Hit(r, 0.001, 1000); ok {
		n := s.Normal(point)
		return litColor(n, light, vec3.Vec3{X: 1, Y: 0.3, Z: 0.3})
	}
	return skyColor(r)
}

func shadeMesh(r ray.Ray, root *bvh.Node, light vec3.Vec3) vec3.Vec3 {
	if hit, ok := root.Hit(r, 0.001, 1000); ok {
		n := hit.Triangle.Normal()
		return litColor(n, light, vec3.Vec3{X: 0.3, Y: 0.6, Z: 1})
	}
	return skyColor(r)
}

// litColor applies simple Lambertian shading. It shades both faces of
// a triangle the same way (abs of the dot product) since demo OBJ
// assets aren't guaranteed consistently wound.
func litColor(normal, light, baseColor vec3.Vec3) vec3.Vec3 {
	intensity := normal.Dot(light)
	if intensity < 0 {
		intensity = -intensity
	}
	return baseColor.Scale(intensity)
}

func skyColor(r ray.Ray) vec3.Vec3 {
	unitDir := r.Direction.Normalize()
	t := 0.5 * (unitDir.Y + 1)
	white := vec3.Vec3{X: 1, Y: 1, Z: 1}
	skyBlue := vec3.Vec3{X: 0.5, Y: 0.7, Z: 1}
	return white.Scale(1 - t).Add(skyBlue.Scale(t))
}

func writeColor(w *bufio.Writer, c vec3.Vec3) {
	r := clamp(c.X)
	g := clamp(c.Y)
	b := clamp(c.Z)
	fmt.Fprintf(w, "%d %d %d\n", int(255*r), int(255*g), int(255*b))
}

func clamp(x float32) float32 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}
