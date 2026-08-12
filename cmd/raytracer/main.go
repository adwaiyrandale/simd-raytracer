// Command raytracer renders a single sphere to a PPM file. This is the
// phase-1 scaffold: prove the vec3/ray/scene pipeline end-to-end with
// scalar code before any SIMD work begins.
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/adwaiyrandale/simd-raytracer/internal/ray"
	"github.com/adwaiyrandale/simd-raytracer/internal/scene"
	"github.com/adwaiyrandale/simd-raytracer/internal/vec3"
)

const (
	width  = 400
	height = 225
)

func main() {
	if err := run("out.ppm"); err != nil {
		log.Fatal(err)
	}
}

func run(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	sphere := scene.Sphere{Center: vec3.Vec3{X: 0, Y: 0, Z: -1}, Radius: 0.5}
	light := vec3.Vec3{X: -1, Y: 1, Z: 1}.Normalize()

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

			col := shade(r, sphere, light)
			writeColor(w, col)
		}
	}
	return w.Flush()
}

func shade(r ray.Ray, s scene.Sphere, light vec3.Vec3) vec3.Vec3 {
	if point, _, ok := s.Hit(r, 0.001, 1000); ok {
		n := s.Normal(point)
		intensity := n.Dot(light)
		if intensity < 0 {
			intensity = 0
		}
		return vec3.Vec3{X: 1, Y: 0.3, Z: 0.3}.Scale(intensity)
	}
	// Background: sky gradient.
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
