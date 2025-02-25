package pptx

import "image"

type Image struct {
	Raw    image.Image
	Name   string
	Format string
}
