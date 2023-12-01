package entities

import (
	"sync"
	"time"

	mathFormulas "github.com/sebastianreh/distance-calculator-api/pkg/math_formulas"
)

const MaxSearchRadius = 8

type CalculationRequest struct {
	Now  time.Time
	Lat  float64 `query:"lat"`
	Long float64 `query:"long"`
}

type RestaurantIDLatLng struct {
	ID             string  `json:"id"`
	Lat            float64 `json:"lat"`
	Long           float64 `json:"long"`
	DeliveryRadius float64 `json:"radius"`
}

func (request CalculationRequest) FindRestaurantsInRadius(data CoordinatesData, timeRadiusMap TimeRadiusMap) []string {
	IDs := make([]string, 0)
	restaurantsInSquareArea := findRestaurantsInSquareArea(data, request)
	if len(restaurantsInSquareArea) < 1 {
		return IDs
	}

	openRestaurants := findOpenRestaurants(request, timeRadiusMap, restaurantsInSquareArea)

	inRadius := findRestaurantsWithinDeliveryRadius(openRestaurants, request)

	return inRadius
}

func findRestaurantsInSquareArea(coordinatesData CoordinatesData, request CalculationRequest) []RestaurantIDLatLng {
	var wg sync.WaitGroup
	var possibleLatList, possibleLongList []CoordinateIDData
	var mu sync.Mutex
	restaurantMap := make(map[string]float64)

	wg.Add(1)
	go func() {
		defer wg.Done()
		possibleLatList = removeOutsideRange(coordinatesData.LatData, request.Lat, MaxSearchRadius)
		mu.Lock()
		for _, latCoord := range possibleLatList {
			restaurantMap[latCoord.ID] = latCoord.Coordinate
		}
		mu.Unlock()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		possibleLongList = removeOutsideRange(coordinatesData.LongData, request.Long, MaxSearchRadius)
	}()

	wg.Wait()

	var matchingRestaurants []RestaurantIDLatLng
	for _, longCoord := range possibleLongList {
		mu.Lock()
		if lat, exists := restaurantMap[longCoord.ID]; exists {
			matchingRestaurants = append(matchingRestaurants, RestaurantIDLatLng{
				ID:   longCoord.ID,
				Lat:  lat,
				Long: longCoord.Coordinate,
			})
		}
		mu.Unlock()
	}

	return matchingRestaurants
}

func removeOutsideRange(coordinateData []CoordinateIDData, targetLatitude, kmDistance float64) []CoordinateIDData {
	lowerIndex, upperIndex := findIndexRangeWithinDistance(coordinateData, targetLatitude, kmDistance)

	if lowerIndex == -1 || upperIndex == -1 {
		// No points within the range, clear the slice
		coordinateData = coordinateData[:0]
		return coordinateData
	}

	// Keep only the elements within the range
	return coordinateData[lowerIndex:upperIndex]
}

func findIndexRangeWithinDistance(coordinateData []CoordinateIDData, targetCoordinate, kmDistance float64) (lower int,
	upper int) {
	lower = lowerIndexBound(coordinateData, targetCoordinate, kmDistance)
	upper = upperIndexBound(coordinateData, targetCoordinate, kmDistance)

	if lower < 0 || upper > len(coordinateData) || lower >= upper {
		// No points within distance kmDistance
		return -1, -1
	}

	return lower, upper
}

func lowerIndexBound(data []CoordinateIDData, targetCoordinate, kmDistance float64) int {
	degreeDistance := mathFormulas.KmToDegrees(kmDistance)
	lowerBound := targetCoordinate - degreeDistance

	low, high := 0, len(data)
	for low < high {
		mid := low + (high-low)/2
		if data[mid].Coordinate < lowerBound {
			low = mid + 1
		} else {
			high = mid
		}
	}
	return low
}

func upperIndexBound(data []CoordinateIDData, targetCoordinate, kmDistance float64) int {
	degreeDistance := mathFormulas.KmToDegrees(kmDistance)
	upperBound := targetCoordinate + degreeDistance

	low, high := 0, len(data)
	for low < high {
		mid := low + (high-low)/2
		if data[mid].Coordinate <= upperBound {
			low = mid + 1
		} else {
			high = mid
		}
	}
	return low
}

func findOpenRestaurants(request CalculationRequest, timeRadiusMap TimeRadiusMap,
	restaurants []RestaurantIDLatLng) []RestaurantIDLatLng {
	var openRestaurants []RestaurantIDLatLng
	for _, restaurant := range restaurants {
		timeRadius, ok := timeRadiusMap[restaurant.ID]
		if ok {
			currentTime := request.Now.Hour()*100 + request.Now.Minute()
			// validates if is still Open after midnight
			if timeRadius.Open > timeRadius.Close {
				if currentTime >= timeRadius.Open && currentTime < timeRadius.Close {
					restaurant.DeliveryRadius = timeRadius.Radius
					openRestaurants = append(openRestaurants, restaurant)
				}
			} else {
				if currentTime >= timeRadius.Open || currentTime < timeRadius.Close {
					restaurant.DeliveryRadius = timeRadius.Radius
					openRestaurants = append(openRestaurants, restaurant)
				}
			}
		}
	}

	return openRestaurants
}

func findRestaurantsWithinDeliveryRadius(restaurants []RestaurantIDLatLng, request CalculationRequest) []string {
	var withinDeliveryRadius []string

	for _, restaurant := range restaurants {
		distance := mathFormulas.Haversine(request.Lat, request.Long, restaurant.Lat, restaurant.Long)
		if distance <= restaurant.DeliveryRadius {
			withinDeliveryRadius = append(withinDeliveryRadius, restaurant.ID)
		}
	}

	return withinDeliveryRadius
}
