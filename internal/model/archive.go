package model

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

type Object struct {
	Mesh      Mesh
	Transform Transform
	Material  int
}

func WriteProject(filename string, objects []Object) error {
	if len(objects) == 0 || len(objects) > 100 {
		return errors.New("Choose 1–100 model instances")
	}
	var body strings.Builder
	body.WriteString(`<?xml version="1.0" encoding="UTF-8"?><model unit="millimeter" xml:lang="en-US" xmlns="http://schemas.microsoft.com/3dmanufacturing/core/2015/02"><resources>`)
	for i, obj := range objects {
		fmt.Fprintf(&body, `<object id="%d" type="model"><mesh><vertices>`, i+1)
		for _, tri := range obj.Mesh.Triangles {
			for _, v := range tri {
				fmt.Fprintf(&body, `<vertex x="%s" y="%s" z="%s"/>`, Number(v[0]), Number(v[1]), Number(v[2]))
			}
		}
		body.WriteString(`</vertices><triangles>`)
		for n := range obj.Mesh.Triangles {
			fmt.Fprintf(&body, `<triangle v1="%d" v2="%d" v3="%d"/>`, n*3, n*3+1, n*3+2)
		}
		body.WriteString(`</triangles></mesh></object>`)
	}
	body.WriteString(`</resources><build>`)
	for i, obj := range objects {
		fmt.Fprintf(&body, `<item objectid="%d" transform="%s" printable="1"/>`, i+1, Matrix(obj.Transform))
	}
	body.WriteString(`</build></model>`)
	f, e := os.OpenFile(filename, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	z := zip.NewWriter(f)
	// Explicit plate membership is needed for the native CLI to export a
	// project that can subsequently be imported with a selected plate.
	var plate strings.Builder
	plate.WriteString(`<?xml version="1.0"?><config><plate><metadata key="plater_id" value="1"/>`)
	for i := range objects {
		fmt.Fprintf(&plate, `<model_instance><metadata key="object_id" value="%d"/><metadata key="instance_id" value="0"/><metadata key="identify_id" value="%d"/></model_instance>`, i+1, i+1)
	}
	plate.WriteString(`</plate></config>`)
	members := map[string]string{"[Content_Types].xml": `<?xml version="1.0"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="model" ContentType="application/vnd.ms-package.3dmanufacturing-3dmodel+xml"/></Types>`, "_rels/.rels": `<?xml version="1.0"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rel0" Type="http://schemas.microsoft.com/3dmanufacturing/2013/01/3dmodel" Target="/3D/3dmodel.model"/></Relationships>`, "3D/3dmodel.model": body.String()}
	members["Metadata/model_settings.config"] = plate.String()
	for name, data := range members {
		w, e := z.Create(name)
		if e != nil {
			f.Close()
			return e
		}
		if _, e = io.WriteString(w, data); e != nil {
			f.Close()
			return e
		}
	}
	if e = z.Close(); e != nil {
		f.Close()
		return e
	}
	return f.Close()
}
func OpenArchive(filename string, maxBytes uint64) (*zip.ReadCloser, map[string]*zip.File, error) {
	z, e := zip.OpenReader(filename)
	if e != nil {
		return nil, nil, errors.New("This is not a readable 3MF archive")
	}
	fail := func(msg string) (*zip.ReadCloser, map[string]*zip.File, error) {
		z.Close()
		return nil, nil, errors.New(msg)
	}
	if len(z.File) > 5000 {
		return fail("3MF has too many entries")
	}
	var size uint64
	files := map[string]*zip.File{}
	seen := map[string]bool{}
	for _, f := range z.File {
		name := strings.ReplaceAll(f.Name, "\\", "/")
		if name != f.Name || strings.HasPrefix(name, "/") || strings.HasPrefix(name, "../") || name == ".." || strings.Contains(name, "/../") || strings.Contains(name, ":") || path.Clean(name) != strings.TrimSuffix(name, "/") || f.Mode()&os.ModeSymlink != 0 {
			return fail("3MF contains an unsafe archive path")
		}
		if seen[strings.ToLower(strings.TrimSuffix(name, "/"))] {
			return fail("3MF contains duplicate entries")
		}
		size += f.UncompressedSize64
		if size > maxBytes || f.UncompressedSize64 > 256<<20 {
			return fail("3MF decompressed size exceeds the supported limit")
		}
		files[name] = f
		seen[strings.ToLower(strings.TrimSuffix(name, "/"))] = true
	}
	return z, files, nil
}
func ReadMember(f *zip.File, maxBytes int64) ([]byte, error) {
	if f == nil {
		return nil, os.ErrNotExist
	}
	if f.UncompressedSize64 > uint64(maxBytes) {
		return nil, errors.New("Archive member exceeds the supported size")
	}
	r, e := f.Open()
	if e != nil {
		return nil, e
	}
	defer r.Close()
	b, e := io.ReadAll(io.LimitReader(r, maxBytes+1))
	if len(b) > int(maxBytes) {
		return nil, errors.New("Archive member exceeds the supported size")
	}
	return b, e
}
func SafeXML(b []byte) error {
	if bytes.Contains(bytes.ToUpper(b), []byte("<!DOCTYPE")) || bytes.Contains(bytes.ToUpper(b), []byte("<!ENTITY")) {
		return errors.New("XML entities and document types are not supported")
	}
	d := xml.NewDecoder(bytes.NewReader(b))
	depth := 0
	for {
		t, e := d.Token()
		if e == io.EOF {
			return nil
		}
		if e != nil {
			return errors.New("3MF contains invalid XML")
		}
		switch t.(type) {
		case xml.StartElement:
			depth++
			if depth > 64 {
				return errors.New("3MF XML nesting limit exceeded")
			}
		case xml.EndElement:
			depth--
		}
	}
}
