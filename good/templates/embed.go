package templates

import (
	"embed"
	_ "embed"
	"image"
	"image/png"
)

//go:embed items.json *.png
var Templates embed.FS

func GetEmbedTemplate(iden string) (image.Image, error) {
	f, _ := Templates.Open(iden + ".png")
	return png.Decode(f)
}
