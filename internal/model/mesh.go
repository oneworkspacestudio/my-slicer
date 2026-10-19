package model

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

const MaxTriangles = 1500000

type Vec [3]float64
type Triangle [3]Vec
type Mesh struct {
	Triangles []Triangle `json:"-"`
	Min       Vec        `json:"min"`
	Max       Vec        `json:"max"`
}
type Transform struct {
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Z     float64 `json:"z"`
	RX    float64 `json:"rx"`
	RY    float64 `json:"ry"`
	RZ    float64 `json:"rz"`
	Scale float64 `json:"scale"`
}

func ValidNumber(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) && math.Abs(x) < 1e6 }
func (t Transform) Valid() bool {
	for _, x := range []float64{t.X, t.Y, t.Z, t.RX, t.RY, t.RZ, t.Scale} {
		if !ValidNumber(x) {
			return false
		}
	}
	return t.Scale >= 0.001 && t.Scale <= 1000
}
func (t Transform) Apply(v Vec) Vec {
	x, y, z := v[0]*t.Scale, v[1]*t.Scale, v[2]*t.Scale
	rx, ry, rz := t.RX*math.Pi/180, t.RY*math.Pi/180, t.RZ*math.Pi/180
	y, z = y*math.Cos(rx)-z*math.Sin(rx), y*math.Sin(rx)+z*math.Cos(rx)
	x, z = x*math.Cos(ry)+z*math.Sin(ry), -x*math.Sin(ry)+z*math.Cos(ry)
	x, y = x*math.Cos(rz)-y*math.Sin(rz), x*math.Sin(rz)+y*math.Cos(rz)
	return Vec{x + t.X, y + t.Y, z + t.Z}
}
func Bounds(triangles []Triangle) (Vec, Vec, error) {
	lo := Vec{math.Inf(1), math.Inf(1), math.Inf(1)}
	hi := Vec{math.Inf(-1), math.Inf(-1), math.Inf(-1)}
	if len(triangles) == 0 || len(triangles) > MaxTriangles {
		return lo, hi, errors.New("Model must contain 1–1,500,000 triangles")
	}
	for _, tri := range triangles {
		for _, v := range tri {
			for n, x := range v {
				if !ValidNumber(x) {
					return lo, hi, errors.New("Model has an invalid coordinate")
				}
				lo[n] = min(lo[n], x)
				hi[n] = max(hi[n], x)
			}
		}
	}
	return lo, hi, nil
}
func ParseSTL(b []byte) (Mesh, error) {
	var m Mesh
	if len(b) > 100<<20 {
		return m, errors.New("Model upload exceeds 100 MiB")
	}
	if len(b) >= 84 {
		count := int(binary.LittleEndian.Uint32(b[80:84]))
		if count > 0 && count <= MaxTriangles && 84+50*count == len(b) {
			m.Triangles = make([]Triangle, count)
			for i := range m.Triangles {
				offset := 84 + i*50 + 12
				for j := 0; j < 3; j++ {
					for k := 0; k < 3; k++ {
						m.Triangles[i][j][k] = float64(math.Float32frombits(binary.LittleEndian.Uint32(b[offset : offset+4])))
						offset += 4
					}
				}
			}
			var e error
			m.Min, m.Max, e = Bounds(m.Triangles)
			return m, e
		}
	}
	if len(b) > 40<<20 || !bytes.HasPrefix(bytes.TrimSpace(b), []byte("solid")) {
		return m, errors.New("Choose a valid binary or ASCII STL file")
	}
	words := strings.Fields(string(b))
	var vertices []Vec
	for i := 0; i < len(words); i++ {
		if words[i] != "vertex" {
			continue
		}
		if i+3 >= len(words) {
			return m, errors.New("Incomplete STL vertex")
		}
		var v Vec
		for n := 0; n < 3; n++ {
			x, e := strconv.ParseFloat(words[i+n+1], 64)
			if e != nil {
				return m, errors.New("Invalid STL vertex")
			}
			v[n] = x
		}
		vertices = append(vertices, v)
		if len(vertices) > MaxTriangles*3 {
			return m, errors.New("STL triangle limit exceeded")
		}
		i += 3
	}
	if len(vertices) == 0 || len(vertices)%3 != 0 {
		return m, errors.New("STL does not contain complete triangles")
	}
	m.Triangles = make([]Triangle, len(vertices)/3)
	for i := range m.Triangles {
		copy(m.Triangles[i][:], vertices[i*3:i*3+3])
	}
	var e error
	m.Min, m.Max, e = Bounds(m.Triangles)
	return m, e
}
func (m Mesh) Centred() Mesh {
	out := m
	out.Triangles = make([]Triangle, len(m.Triangles))
	cx, cy := (m.Min[0]+m.Max[0])/2, (m.Min[1]+m.Max[1])/2
	for i, tri := range m.Triangles {
		for j, v := range tri {
			out.Triangles[i][j] = Vec{v[0] - cx, v[1] - cy, v[2] - m.Min[2]}
		}
	}
	out.Min, out.Max, _ = Bounds(out.Triangles)
	return out
}
func (m Mesh) Binary(previewLimit int) []byte {
	stride := 1
	if previewLimit > 0 && len(m.Triangles) > previewLimit {
		stride = (len(m.Triangles) + previewLimit - 1) / previewLimit
	}
	count := (len(m.Triangles) + stride - 1) / stride
	b := make([]byte, 84+count*50)
	copy(b, "My Slicer validated millimeter triangle mesh")
	binary.LittleEndian.PutUint32(b[80:84], uint32(count))
	n := 0
	for i := 0; i < len(m.Triangles); i += stride {
		offset := 84 + n*50 + 12
		for _, v := range m.Triangles[i] {
			for _, x := range v {
				binary.LittleEndian.PutUint32(b[offset:offset+4], math.Float32bits(float32(x)))
				offset += 4
			}
		}
		n++
	}
	return b
}
func InPolygon(x, y float64, p [][2]float64) bool {
	inside := false
	for i, a := range p {
		b := p[(i+1)%len(p)]
		dx, dy := b[0]-a[0], b[1]-a[1]
		cross := (x-a[0])*dy - (y-a[1])*dx
		if math.Abs(cross) < 0.001 && x >= min(a[0], b[0])-0.001 && x <= max(a[0], b[0])+0.001 && y >= min(a[1], b[1])-0.001 && y <= max(a[1], b[1])+0.001 {
			return true
		}
		if (a[1] > y) != (b[1] > y) && x < (b[0]-a[0])*(y-a[1])/(b[1]-a[1])+a[0] {
			inside = !inside
		}
	}
	return inside
}
func (m Mesh) ValidatePlate(t Transform, polygon, exclusion [][2]float64, height float64) error {
	if !t.Valid() {
		return errors.New("Check object position, rotation and scale")
	}
	for _, tri := range m.Triangles {
		for _, v := range tri {
			p := t.Apply(v)
			if p[2] < -0.01 || p[2] > height+0.01 || !InPolygon(p[0], p[1], polygon) {
				return errors.New("An object extends outside this printer's precise printable area or height; move, scale or choose another printer")
			}
			if len(exclusion) >= 3 && InPolygon(p[0], p[1], exclusion) {
				return errors.New("An object overlaps the machine profile's excluded bed area")
			}
		}
	}
	return nil
}
func Cube(size float64) Mesh {
	v := []Vec{{0, 0, 0}, {size, 0, 0}, {size, size, 0}, {0, size, 0}, {0, 0, size}, {size, 0, size}, {size, size, size}, {0, size, size}}
	idx := [][3]int{{0, 2, 1}, {0, 3, 2}, {4, 5, 6}, {4, 6, 7}, {0, 1, 5}, {0, 5, 4}, {1, 2, 6}, {1, 6, 5}, {2, 3, 7}, {2, 7, 6}, {3, 0, 4}, {3, 4, 7}}
	m := Mesh{}
	for _, t := range idx {
		m.Triangles = append(m.Triangles, Triangle{v[t[0]], v[t[1]], v[t[2]]})
	}
	m.Min, m.Max, _ = Bounds(m.Triangles)
	return m
}
func Number(x float64) string { return strconv.FormatFloat(x, 'f', 6, 64) }
func Matrix(t Transform) string {
	zero := t.Apply(Vec{})
	a, b, c := t.Apply(Vec{1, 0, 0}), t.Apply(Vec{0, 1, 0}), t.Apply(Vec{0, 0, 1})
	values := []float64{a[0] - zero[0], a[1] - zero[1], a[2] - zero[2], b[0] - zero[0], b[1] - zero[1], b[2] - zero[2], c[0] - zero[0], c[1] - zero[1], c[2] - zero[2], zero[0], zero[1], zero[2]}
	var s []string
	for _, x := range values {
		s = append(s, Number(x))
	}
	return strings.Join(s, " ")
}
func (m Mesh) String() string { return fmt.Sprintf("%d triangles", len(m.Triangles)) }
