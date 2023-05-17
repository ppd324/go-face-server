package models

import (
	"github.com/Kagami/go-face"
	"image"
)

type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Face struct {
	Id         int          `uri:"id"`
	Name       string       `json:"name"`
	File       string       `json:"file"`
	Shapes     []Point      `json:"shapes"`
	Descriptor [128]float32 `json:"descriptors"`
}

func NewFace(id int, name string, file string, point []image.Point, des face.Descriptor) *Face {
	pointArr := make([]Point, 0)
	for _, point := range point {
		pointArr = append(pointArr, Point{point.X, point.Y})
	}
	return &Face{
		Id:         id,
		Name:       name,
		File:       file,
		Shapes:     pointArr,
		Descriptor: des,
	}
}
