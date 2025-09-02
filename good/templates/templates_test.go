package templates

import (
	"bytes"
	"image"
	_ "image/png"
	"testing"
)

func TestTemapltes(t *testing.T) {
	reader := bytes.NewReader(Resources[0].StaticContent)
	_, _, err := image.Decode(reader)
	if err != nil {
		panic(err)
	}
}
