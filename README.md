# simd-raytracer

A Monte Carlo path tracer in Go, using SIMD ray-packet tracing against a
BVH4 over triangle meshes. Design doc:
`~/dev/project-docs/go-simd-raytracer-DESIGN.md`.

## Build phases

1. Scalar reference path tracer (spheres only) — done.
2. OBJ loading + scalar BVH + scalar triangle intersection — done.
3. SIMD ray-packet infrastructure vs scalar oracle — sphere and
   brute-force triangle intersection done and wired into the renderer
   (`-simd` flag); SIMD BVH traversal not yet done (mesh path is
   brute-force over all triangles, fine for a 12-triangle test cube).
4. Full SIMD-integrated path tracer with a SIMD BVH, benchmarked
   against scalar for real - not started.
5. Stretch: wider hardware vectors, NEE/MIS, multithreading.

## Renders

`renders/` has example output at 1920x1080, 32 samples/pixel:
`sphere_scalar.png` / `sphere_simd.png` and `cube_scalar.png` /
`cube_simd.png`. The scalar and SIMD paths produce visually identical
images (same silhouette, shading, antialiasing) - correctness is also
covered by `TestSIMDSphereHitMatchesScalar` /
`TestSIMDTriangleHitMatchesScalar` in `cmd/raytracer`, which check the
SIMD intersection math against the scalar implementations on random
rays. Reproduce with:

```sh
go run ./cmd/raytracer -width 1920 -height 1080 -spp 32 -out sphere_scalar.ppm
GOEXPERIMENT=simd $(go env GOPATH)/bin/go1.27rc1 run ./cmd/raytracer -simd -width 1920 -height 1080 -spp 32 -out sphere_simd.ppm
```
(swap in `-obj testdata/cube.obj -camX 2.5 -camY 2 -camZ 1.5 -lookX 0 -lookY 0 -lookZ -3` for the cube)

## Toolchain

SIMD code (`internal/simdvec`, `cmd/raytracer/simd_render.go`)
requires Go 1.27rc1 with `GOEXPERIMENT=simd` (the stable `go` on PATH
here is too old, and Go 1.26's `simd/archsimd` is amd64-only - useless
on Apple Silicon). Install once:

```sh
go install golang.org/dl/go1.27rc1@latest
$(go env GOPATH)/bin/go1.27rc1 download
```

Then build/run/test with `GOEXPERIMENT=simd $(go env GOPATH)/bin/go1.27rc1 ...`:

```sh
# render (real image output, works fine via `run`/`build` since
# everything SIMD-related lives in package main - see simd_render.go's
# doc comment for why)
GOEXPERIMENT=simd $(go env GOPATH)/bin/go1.27rc1 run ./cmd/raytracer -simd -out out.ppm

# internal/simdvec specifically needs `test`, not `build` - see its
# doc comment for the confirmed compiler bug behind that
GOEXPERIMENT=simd $(go env GOPATH)/bin/go1.27rc1 test ./internal/simdvec/... -v -bench=.
```

**Known go1.27rc1 compiler bugs hit and worked around while building
this** (all reproduced in isolation outside this repo before assuming
they weren't bugs in this code):
- A struct with `simd.Float32s` fields fails `go build` as a
  standalone importable package ("undefined: bridge"), though it
  builds fine as part of a self-contained `go test`/`go run`/`go
  build` of a single package. This is *why* all the SIMD render code
  lives directly in `cmd/raytracer` (`package main`) instead of a
  separate `internal/` package - crossing that package-archive
  boundary is what triggers the bug.
- Every file touching such a type needs its own direct inline
  `simd.*` call, or compilation fails the same way (a package-level
  var trigger a different, worse failure: an actual internal compiler
  error).
- `Mask32s.ToInt32s` is declared in the portable stub surface but not
  implemented for this arm64 backend yet.
- **The big one**: `Float32s.IfElse`'s branches are swapped in codegen
  on this arm64 backend - `x.IfElse(mask, y)` evaluates to "y where
  mask is true, x where false", the *opposite* of its documented
  contract. Comparisons (`GreaterEqual`/`Less`/etc.) and `Mask32s.And`/
  `Or` were verified correct in isolation; only `IfElse` is affected.
  Worked around via `ifElseFixed()` in `simd_render.go`, which swaps
  the receiver/argument to compensate. If a future toolchain fixes
  this upstream, `TestSIMDSphereHitMatchesScalar` and
  `TestSIMDTriangleHitMatchesScalar` will start failing (the
  workaround would then double-invert) - that's the intended signal
  to remove it.

Everything else (`go build ./...`, `go test ./...`, no flags, any
Go >= 1.21) continues to work exactly as before - SIMD-only code is
excluded from those by `//go:build goexperiment.simd` tags, so scalar
development is unaffected.

## Layout

- `internal/vec3` — 3D vector math (scalar).
- `internal/ray` — ray type.
- `internal/camera` — camera / primary ray generation.
- `internal/scene` — scene description: spheres, triangles, AABBs.
- `internal/mesh` — Wavefront OBJ loading.
- `internal/bvh` — scalar bounding volume hierarchy.
- `internal/simdvec` — SIMD vector-math benchmarking scaffold (SoA
  `Vec3s`, benchmarks vs scalar). Not a renderer - see its doc comment
  for toolchain caveats.
- `cmd/raytracer` — CLI entry point, both the scalar renderer (`run`,
  `main.go`) and the SIMD renderer (`renderSIMD`,
  `simd_render.go`/`simd_stub.go`), selected via `-simd`.
