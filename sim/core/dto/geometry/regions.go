package geometry

import (
	"common/commonmath"
)

// Defines an identifiable region
type IdRegion struct {
	Id     int64
	Region commonMath.Region
}

func NewIdRegion(id int64, region commonMath.Region) IdRegion {
	return IdRegion{
		Id:     id,
		Region: region}
}
