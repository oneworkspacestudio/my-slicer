//go:build windows

package printer

import (
	"errors"
	"golang.org/x/sys/windows"
	"unsafe"
)

func Protect(code string) ([]byte, error) {
	if len(code) < 4 || len(code) > 64 {
		return nil, errors.New("Enter the printer's LAN access code")
	}
	data := []byte(code)
	in := windows.DataBlob{Size: uint32(len(data)), Data: &data[0]}
	var out windows.DataBlob
	if e := windows.CryptProtectData(&in, nil, nil, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &out); e != nil {
		return nil, e
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))
	b := unsafe.Slice(out.Data, out.Size)
	return append([]byte(nil), b...), nil
}
func Unprotect(cipher []byte) (string, error) {
	if len(cipher) == 0 {
		return "", errors.New("Pair this printer with its LAN access code")
	}
	in := windows.DataBlob{Size: uint32(len(cipher)), Data: &cipher[0]}
	var out windows.DataBlob
	if e := windows.CryptUnprotectData(&in, nil, nil, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &out); e != nil {
		return "", errors.New("Re-enter the LAN access code on this Windows account after moving or restoring the app")
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))
	return string(unsafe.Slice(out.Data, out.Size)), nil
}
