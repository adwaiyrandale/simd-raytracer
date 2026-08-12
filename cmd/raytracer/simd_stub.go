//go:build !goexperiment.simd

package main

import (
	"errors"

	"github.com/adwaiyrandale/simd-raytracer/internal/vec3"
)

func renderSIMD(objPath, outPath string, width, height, samplesPerPixel int, lookfrom, lookat vec3.Vec3) error {
	return errors.New("SIMD rendering requires building with GOEXPERIMENT=simd and the go1.27rc1 toolchain; see README")
}
