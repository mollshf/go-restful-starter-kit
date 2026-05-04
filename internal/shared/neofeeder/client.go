package neofeeder

import (
	"errors"
)

var ErrAbcNil = errors.New("INI ERRROR NIL CUI")

type Semantic struct {
	Nama        string
	Rumah       string
	AlamatRumah string `json:"alamat_rumah"`
}
