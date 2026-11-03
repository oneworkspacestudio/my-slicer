package model

import "testing"

func TestDenseClosedSurfaceSurvivesBinaryRoundtrip(t *testing.T) {
	m := Cube(10)
	mid := func(a, b Vec) Vec { return Vec{(a[0] + b[0]) / 2, (a[1] + b[1]) / 2, (a[2] + b[2]) / 2} }
	for level := 0; level < 7; level++ {
		var next []Triangle
		for _, tri := range m.Triangles {
			a, b, c := tri[0], tri[1], tri[2]
			ab, bc, ca := mid(a, b), mid(b, c), mid(c, a)
			next = append(next, Triangle{a, ab, ca}, Triangle{ab, b, bc}, Triangle{ca, bc, c}, Triangle{ab, bc, ca})
		}
		m.Triangles = next
	}
	if len(m.Triangles) <= 100000 {
		t.Fatal("fixture must exceed the old cutoff")
	}
	got, e := ParseSTL(m.Binary())
	if e != nil {
		t.Fatal(e)
	}
	if len(got.Triangles) != len(m.Triangles) {
		t.Fatalf("lost surface triangles: %d vs %d", len(got.Triangles), len(m.Triangles))
	}
	type edge [2]Vec
	edges := map[edge]int{}
	less := func(a, b Vec) bool {
		for i := 0; i < 3; i++ {
			if a[i] != b[i] {
				return a[i] < b[i]
			}
		}
		return false
	}
	for i, tri := range got.Triangles {
		if tri != m.Triangles[i] {
			t.Fatalf("surface geometry changed at facet %d", i)
		}
		for j := 0; j < 3; j++ {
			a, b := tri[j], tri[(j+1)%3]
			if less(b, a) {
				a, b = b, a
			}
			edges[edge{a, b}]++
		}
	}
	for _, count := range edges {
		if count != 2 {
			t.Fatal("closed input developed a hole or non-manifold edge", count)
		}
	}
}
