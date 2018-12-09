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

// An identifiable line with identifiers for the start and end
type IdOnlyLine struct {
	Id    int64
	Start int64
	End   int64
}

func NewIdOnlyLine(id, start, end int64) IdOnlyLine {
	return IdOnlyLine{Id: id, Start: start, End: end}
}
