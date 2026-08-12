package vec3

import "testing"

func TestAdd(t *testing.T) {
	a := Vec3{1, 2, 3}
	b := Vec3{4, 5, 6}
	got := a.Add(b)
	want := Vec3{5, 7, 9}
	if got != want {
		t.Errorf("Add() = %v, want %v", got, want)
	}
}

func TestSub(t *testing.T) {
	a := Vec3{5, 7, 9}
	b := Vec3{4, 5, 6}
	got := a.Sub(b)
	want := Vec3{1, 2, 3}
	if got != want {
		t.Errorf("Sub() = %v, want %v", got, want)
	}
}

func TestScale(t *testing.T) {
	a := Vec3{1, 2, 3}
	got := a.Scale(2)
	want := Vec3{2, 4, 6}
	if got != want {
		t.Errorf("Scale() = %v, want %v", got, want)
	}
}

func TestDot(t *testing.T) {
	a := Vec3{1, 2, 3}
	b := Vec3{4, 5, 6}
	got := a.Dot(b)
	want := float32(32) // 1*4 + 2*5 + 3*6
	if got != want {
		t.Errorf("Dot() = %v, want %v", got, want)
	}
}

func TestCross(t *testing.T) {
	a := Vec3{1, 0, 0}
	b := Vec3{0, 1, 0}
	got := a.Cross(b)
	want := Vec3{0, 0, 1}
	if got != want {
		t.Errorf("Cross() = %v, want %v", got, want)
	}
}

func TestLength(t *testing.T) {
	a := Vec3{3, 4, 0}
	got := a.Length()
	want := float32(5)
	if got != want {
		t.Errorf("Length() = %v, want %v", got, want)
	}
}

func TestNormalize(t *testing.T) {
	a := Vec3{3, 4, 0}
	got := a.Normalize()
	want := Vec3{0.6, 0.8, 0}
	const eps = 1e-6
	if abs32(got.X-want.X) > eps || abs32(got.Y-want.Y) > eps || abs32(got.Z-want.Z) > eps {
		t.Errorf("Normalize() = %v, want %v", got, want)
	}
}

func abs32(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}
