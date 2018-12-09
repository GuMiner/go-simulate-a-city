package vehicledto

// Defines a vehicle update traveling along a road line.
type VehicleUpdate struct {
	Id            int64
	RoadId        int64
	TravelLength  float32 // Negative -- high to low. Positive -- low to high.
	VehicleLength float32
}
