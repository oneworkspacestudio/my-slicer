package model

// A read-only geometry view. The original archive remains the slicing source.
import (
	"encoding/xml"
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"
)

type previewRef struct {
	ID        string `xml:"objectid,attr"`
	Path      string `xml:"path,attr"`
	Transform string `xml:"transform,attr"`
	Printable string `xml:"printable,attr"`
}
type previewDocument struct {
	Unit    string `xml:"unit,attr"`
	Objects []struct {
		ID       string `xml:"id,attr"`
		Vertices []struct {
			X float64 `xml:"x,attr"`
			Y float64 `xml:"y,attr"`
			Z float64 `xml:"z,attr"`
		} `xml:"mesh>vertices>vertex"`
		Faces []struct {
			A int `xml:"v1,attr"`
			B int `xml:"v2,attr"`
			C int `xml:"v3,attr"`
		} `xml:"mesh>triangles>triangle"`
		Components []previewRef `xml:"components>component"`
	} `xml:"resources>object"`
	Items []previewRef `xml:"build>item"`
}
type previewMetadata struct {
	Key   string `xml:"key,attr"`
	Value string `xml:"value,attr"`
}
type previewConfig struct {
	Plates []struct {
		Meta      []previewMetadata `xml:"metadata"`
		Instances []struct {
			Meta []previewMetadata `xml:"metadata"`
		} `xml:"model_instance"`
	} `xml:"plate"`
}

type affine [12]float64

func parseAffine(s string) (affine, error) {
	m := affine{1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0}
	if strings.TrimSpace(s) == "" {
		return m, nil
	}
	f := strings.Fields(s)
	if len(f) != 12 {
		return m, errors.New("The saved model position could not be read")
	}
	for i, v := range f {
		x, e := strconv.ParseFloat(v, 64)
		if e != nil || !ValidNumber(x) {
			return m, errors.New("The saved model position is invalid")
		}
		m[i] = x
	}
	return m, nil
}
func (m affine) apply(v Vec) Vec {
	return Vec{v[0]*m[0] + v[1]*m[3] + v[2]*m[6] + m[9], v[0]*m[1] + v[1]*m[4] + v[2]*m[7] + m[10], v[0]*m[2] + v[1]*m[5] + v[2]*m[8] + m[11]}
}

// PreviewProject follows core/production component references, build transforms
// and Bambu plate membership. No archive members are extracted or executed.
func PreviewProject(filename string, plate int) (Mesh, error) {
	var out Mesh
	z, files, e := OpenArchive(filename, 512<<20)
	if e != nil {
		return out, e
	}
	defer z.Close()
	docs := map[string]*previewDocument{}
	load := func(name string) (*previewDocument, error) {
		if d := docs[name]; d != nil {
			return d, nil
		}
		b, e := ReadMember(files[name], 128<<20)
		if e != nil {
			return nil, fmt.Errorf("A model part is missing or unreadable: %s", name)
		}
		if e = SafeXML(b); e != nil {
			return nil, e
		}
		if e = validateReferences(b); e != nil {
			return nil, e
		}
		var d previewDocument
		if e = xml.Unmarshal(b, &d); e != nil {
			return nil, e
		}
		if len(d.Objects) > 10000 {
			return nil, errors.New("Too many model parts to preview")
		}
		docs[name] = &d
		return &d, nil
	}
	root, e := load("3D/3dmodel.model")
	if e != nil {
		return out, e
	}
	allowed := map[string]bool{}
	filtered := false
	if f := files["Metadata/model_settings.config"]; f != nil {
		b, e := ReadMember(f, 32<<20)
		if e != nil {
			return out, e
		}
		if e = SafeXML(b); e != nil {
			return out, e
		}
		var cfg previewConfig
		if e = xml.Unmarshal(b, &cfg); e != nil {
			return out, e
		}
		for i, p := range cfg.Plates {
			index := i + 1
			for _, m := range p.Meta {
				if m.Key == "plater_id" {
					index, _ = strconv.Atoi(m.Value)
				}
			}
			if index != plate {
				continue
			}
			filtered = true
			for _, instance := range p.Instances {
				id, n := "", "0"
				for _, m := range instance.Meta {
					if m.Key == "object_id" {
						id = m.Value
					}
					if m.Key == "instance_id" {
						n = m.Value
					}
				}
				allowed[id+":"+n] = true
			}
		}
		if len(cfg.Plates) > 0 && !filtered {
			return out, errors.New("Choose an existing imported plate")
		}
	} else if plate != 1 {
		return out, errors.New("Choose an existing imported plate")
	}
	visiting := map[string]bool{}
	nodes := 0
	var visit func(string, previewRef, []affine, int) error
	visit = func(name string, r previewRef, parents []affine, depth int) error {
		nodes++
		if depth > 32 || nodes > 10000 {
			return errors.New("Model assembly is too complex to preview")
		}
		if r.Path != "" {
			if strings.HasPrefix(r.Path, "/") {
				name = strings.TrimPrefix(r.Path, "/")
			} else {
				name = path.Join(path.Dir(name), r.Path)
			}
		}
		key := name + ":" + r.ID
		if visiting[key] {
			return errors.New("Model parts contain a circular reference")
		}
		visiting[key] = true
		defer delete(visiting, key)
		d, e := load(name)
		if e != nil {
			return e
		}
		if d.Unit != "" && root.Unit != "" && d.Unit != root.Unit {
			return errors.New("Model parts use different measurement units; re-export using millimeters")
		}
		m, e := parseAffine(r.Transform)
		if e != nil {
			return e
		}
		parents = append(append([]affine{}, parents...), m)
		for _, o := range d.Objects {
			if o.ID != r.ID {
				continue
			}
			if len(out.Triangles)+len(o.Faces) > MaxTriangles {
				return errors.New("Model preview triangle limit exceeded")
			}
			for _, f := range o.Faces {
				var tri Triangle
				for j, k := range []int{f.A, f.B, f.C} {
					if k < 0 || k >= len(o.Vertices) {
						return errors.New("A model face refers to a missing point")
					}
					v := o.Vertices[k]
					p := Vec{v.X, v.Y, v.Z}
					for n := len(parents) - 1; n >= 0; n-- {
						p = parents[n].apply(p)
					}
					tri[j] = p
				}
				out.Triangles = append(out.Triangles, tri)
			}
			for _, c := range o.Components {
				if e = visit(name, c, parents, depth+1); e != nil {
					return e
				}
			}
			return nil
		}
		return errors.New("A model part is missing from the project")
	}
	// 3MF units apply to vertex and translation coordinates throughout a model.
	factor := 1.0
	switch root.Unit {
	case "", "millimeter":
	case "micron":
		factor = .001
	case "centimeter":
		factor = 10
	case "meter":
		factor = 1000
	case "inch":
		factor = 25.4
	case "foot":
		factor = 304.8
	default:
		return out, errors.New("The model uses unsupported measurement units")
	}
	counts := map[string]int{}
	for _, r := range root.Items {
		n := counts[r.ID]
		counts[r.ID]++
		if r.Printable == "0" || filtered && !allowed[r.ID+":"+strconv.Itoa(n)] {
			continue
		}
		if e = visit("3D/3dmodel.model", r, nil, 0); e != nil {
			return out, e
		}
	}
	if len(out.Triangles) == 0 {
		return out, errors.New("This plate has no printable model. Choose another imported plate")
	}
	for i := range out.Triangles {
		for j := range out.Triangles[i] {
			for k := range out.Triangles[i][j] {
				out.Triangles[i][j][k] *= factor
			}
		}
	}
	out.Min, out.Max, e = Bounds(out.Triangles)
	return out, e
}
