package geometry

import "github.com/go-gl/mathgl/mgl32"

// Defines an identifiable line
type IdLine struct {
	Id   int64
	Line [2]mgl32.Vec2
}

func NewIdLine(id int64, line [2]mgl32.Vec2) IdLine {
	return IdLine{
		Id:   id,
		Line: line}
}
