package geometry

import "github.com/go-gl/mathgl/mgl32"

// Defines an identifiable point
type IdPoint struct {
	Id    int64
	Point mgl32.Vec2
}

func NewIdPoint(id int64, point mgl32.Vec2) IdPoint {
	return IdPoint{
		Id:    id,
		Point: point}
}
