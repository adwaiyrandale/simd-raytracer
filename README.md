# simd-raytracer

A Monte Carlo path tracer in Go, using SIMD ray-packet tracing against a
BVH4 over triangle meshes. Design doc:
`~/dev/project-docs/go-simd-raytracer-DESIGN.md`.

## Build phases

1. Scalar reference path tracer (spheres only) — in progress.
2. OBJ loading + scalar BVH + scalar triangle intersection.
3. SIMD ray-packet infrastructure vs scalar oracle.
4. SIMD-integrated path tracer, benchmarked against scalar.
5. Stretch: wider hardware vectors, NEE/MIS, multithreading.

## Toolchain

Phase 3+ SIMD code (`internal/simdvec`) requires Go 1.27rc1 with
`GOEXPERIMENT=simd` (the stable `go` on PATH here is too old, and
Go 1.26's `simd/archsimd` is amd64-only - useless on Apple Silicon).
Install once:

```sh
go install golang.org/dl/go1.27rc1@latest
$(go env GOPATH)/bin/go1.27rc1 download
```

Then test SIMD code with (note: `test`, not `build` - see the doc
comment in `internal/simdvec/vec3s.go` for why `go build` on this
package hits a confirmed go1.27rc1 compiler bug):

```sh
GOEXPERIMENT=simd $(go env GOPATH)/bin/go1.27rc1 test ./internal/simdvec/... -v -bench=.
```

Everything else (`go build ./...`, `go test ./...`, no flags, any
Go >= 1.21) continues to work exactly as before - the SIMD package is
excluded from those by its own `//go:build goexperiment.simd` tag, so
scalar development is unaffected.

## Layout

- `internal/vec3` — 3D vector math (scalar).
- `internal/ray` — ray type.
- `internal/camera` — camera / primary ray generation.
- `internal/scene` — scene description: spheres, triangles, AABBs.
- `internal/mesh` — Wavefront OBJ loading.
- `internal/bvh` — scalar bounding volume hierarchy.
- `internal/simdvec` — phase-3 SIMD scaffold (SoA `Vec3s`, benchmarks
  vs the scalar path). Not yet wired into the renderer - see its doc
  comment for the toolchain caveats above.
- `cmd/raytracer` — CLI entry point.
