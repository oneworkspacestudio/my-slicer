package model

import (
	"bufio"
	"errors"
	"io"
	"math"
	"strconv"
	"strings"
)

type Segment [7]float64
type Layer struct {
	Number   int       `json:"number"`
	Z        float64   `json:"z"`
	Segments []Segment `json:"segments"`
}
type Toolpath struct {
	Layers         []Layer  `json:"layers"`
	Features       []string `json:"features"`
	Extrusions     int      `json:"extrusions"`
	MotionSegments int      `json:"motionSegments"`
}

// Reads actual G-code coordinates and extrusion modes. It never synthesizes a
// decorative layer stack or estimates time/material from triangle geometry.
func ParseGCode(r io.Reader) (Toolpath, error) {
	out := Toolpath{Layers: []Layer{{Number: 0, Segments: []Segment{}}}}
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 65536), 1<<20)
	var pos Vec
	var extruder float64
	relativeXYZ, relativeE := false, false
	layer := 0
	feature := 0
	featureMap := map[string]int{}
	kind := func(s string) int {
		if n, ok := featureMap[s]; ok {
			return n
		}
		n := len(out.Features)
		featureMap[s] = n
		out.Features = append(out.Features, s)
		return n
	}
	feature = kind("Preparation")
	add := func(a, b Vec, extruding bool) error {
		out.MotionSegments++
		if !extruding {
			return nil
		}
		if len(out.Layers[layer].Segments) >= 2000000 || out.Extrusions >= 3000000 {
			return errors.New("Toolpath exceeds the supported preview complexity; reduce the model or split the plate")
		}
		out.Layers[layer].Segments = append(out.Layers[layer].Segments, Segment{a[0], a[1], a[2], b[0], b[1], b[2], float64(feature)})
		if layer > 0 {
			out.Extrusions++
		}
		return nil
	}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "; layer num/total_layer_count:") {
			s := strings.TrimSpace(strings.TrimPrefix(line, "; layer num/total_layer_count:"))
			n, e := strconv.Atoi(strings.Split(s, "/")[0])
			if e != nil || n < 1 || n > 20000 {
				return out, errors.New("Invalid native G-code layer marker")
			}
			layer = len(out.Layers)
			out.Layers = append(out.Layers, Layer{Number: n, Z: pos[2], Segments: []Segment{}})
			continue
		}
		if strings.HasPrefix(line, "; FEATURE:") {
			feature = kind(strings.TrimSpace(strings.TrimPrefix(line, "; FEATURE:")))
			continue
		}
		if strings.HasPrefix(line, ";TYPE:") {
			feature = kind(strings.TrimSpace(strings.TrimPrefix(line, ";TYPE:")))
			continue
		}
		if strings.HasPrefix(line, "; Z_HEIGHT:") {
			z, e := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line, "; Z_HEIGHT:")), 64)
			if e == nil && ValidNumber(z) {
				out.Layers[layer].Z = z
			}
			continue
		}
		line = strings.SplitN(line, ";", 2)[0]
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}
		command := strings.ToUpper(parts[0])
		values := map[byte]float64{}
		for _, part := range parts[1:] {
			if len(part) < 2 {
				continue
			}
			k := part[0]
			if !strings.ContainsRune("XYZEFIJKR", rune(k)) {
				continue
			}
			v, e := strconv.ParseFloat(part[1:], 64)
			if e != nil || !ValidNumber(v) {
				return out, errors.New("Native toolpath has invalid numeric coordinates")
			}
			values[k] = v
		}
		switch command {
		case "G90":
			relativeXYZ = false
			continue
		case "G91":
			relativeXYZ = true
			continue
		case "M82":
			relativeE = false
			continue
		case "M83":
			relativeE = true
			continue
		case "G92":
			for n, k := range []byte{'X', 'Y', 'Z'} {
				if v, ok := values[k]; ok {
					pos[n] = v
				}
			}
			if v, ok := values['E']; ok {
				extruder = v
			}
			continue
		case "G0", "G00", "G1", "G01", "G2", "G02", "G3", "G03":
		default:
			continue
		}
		next := pos
		for n, k := range []byte{'X', 'Y', 'Z'} {
			if v, ok := values[k]; ok {
				if relativeXYZ {
					next[n] += v
				} else {
					next[n] = v
				}
			}
		}
		deltaE := 0.0
		if v, ok := values['E']; ok {
			if relativeE {
				deltaE = v
				extruder += v
			} else {
				deltaE = v - extruder
				extruder = v
			}
		}
		distance := math.Hypot(next[0]-pos[0], next[1]-pos[1])
		extruding := deltaE > 0.000001 && (distance > 0.00001 || math.Abs(next[2]-pos[2]) > 0.00001)
		if command == "G2" || command == "G02" || command == "G3" || command == "G03" {
			if deltaE > 0.000001 {
				cx, okI := values['I']
				cy, okJ := values['J']
				if !okI && !okJ {
					return out, errors.New("Unsupported arc center in native preview")
				}
				cx += pos[0]
				cy += pos[1]
				radius := math.Hypot(pos[0]-cx, pos[1]-cy)
				a0, a1 := math.Atan2(pos[1]-cy, pos[0]-cx), math.Atan2(next[1]-cy, next[0]-cx)
				sweep := a1 - a0
				clockwise := command == "G2" || command == "G02"
				if clockwise && sweep >= 0 {
					sweep -= 2 * math.Pi
				}
				if !clockwise && sweep <= 0 {
					sweep += 2 * math.Pi
				}
				steps := max(1, min(10000, int(math.Ceil(math.Abs(sweep)*radius/.2))))
				last := pos
				for n := 1; n <= steps; n++ {
					t := float64(n) / float64(steps)
					a := a0 + sweep*t
					v := Vec{cx + radius*math.Cos(a), cy + radius*math.Sin(a), pos[2] + (next[2]-pos[2])*t}
					if n == steps {
						v = next
					}
					if e := add(last, v, true); e != nil {
						return out, e
					}
					last = v
				}
			}
		} else {
			if e := add(pos, next, extruding); e != nil {
				return out, e
			}
		}
		pos = next
	}
	if e := scanner.Err(); e != nil {
		return out, e
	}
	if out.Extrusions == 0 {
		return out, errors.New("The native output has no non-empty printable extrusion paths")
	}
	return out, nil
}
