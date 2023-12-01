package entities

import (
	"time"

	mathFormulas "github.com/sebastianreh/distance-calculator-api/pkg/math_formulas"
)

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

type (
	TimeRadiusMap map[string]timeRadiusSchedule

	timeRadiusSchedule struct {
		Open   int     `json:"open"`
		Close  int     `json:"close"`
		Radius float64 `json:"radius"`
	}
)

func (request CalculationRequest) FindRestaurantsInRadius(timeRadiusMap TimeRadiusMap,
	restaurantInUserRadius []RestaurantIDLatLng) []string {
	openRestaurants := findOpenRestaurants(request, timeRadiusMap, restaurantInUserRadius)
	inRadius := findRestaurantsWithinDeliveryRadius(openRestaurants, request)

	return inRadius
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
