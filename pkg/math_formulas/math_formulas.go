package math_formulas

import "math"

const (
	earthRadiusKm = 6371
	kmsForDegree  = 111.0
	degrees180    = 180
	two           = 2
)

func Haversine(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := (lat2 - lat1) * math.Pi / degrees180
	dLon := (lon2 - lon1) * math.Pi / degrees180

	a := math.Sin(dLat/two)*math.Sin(dLat/two) +
		math.Cos(lat1*math.Pi/degrees180)*math.Cos(lat2*math.Pi/degrees180)*math.Sin(dLon/two)*math.Sin(dLon/two)
	c := two * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusKm * c
}

func KmToDegrees(km float64) float64 {
	return km / kmsForDegree
}
