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

Phases 3+ require Go 1.27rc1 with `GOEXPERIMENT=simd` (the stable `go`
on PATH here is too old). Install once:

```sh
go install golang.org/dl/go1.27rc1@latest
$(go env GOPATH)/bin/go1.27rc1 download
```

Then build/run phase 3+ code with:

```sh
GOEXPERIMENT=simd $(go env GOPATH)/bin/go1.27rc1 build ./...
```

Phases 1-2 (scalar) build with any Go >= 1.21, no experiment flag needed.

## Layout

- `internal/vec3` — 3D vector math (scalar).
- `internal/ray` — ray type.
- `internal/camera` — camera / primary ray generation.
- `internal/scene` — scene description, spheres, materials.
- `internal/scalar` — scalar (non-SIMD) path tracer, phase 1-2 baseline
  and the correctness oracle for later SIMD phases.
- `cmd/raytracer` — CLI entry point.
