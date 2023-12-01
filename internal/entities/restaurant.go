package entities

import (
	"errors"
	"strconv"

	"github.com/sebastianreh/distance-calculator-api/pkg/logger"
	str "github.com/sebastianreh/distance-calculator-api/pkg/strings"
)

const (
	badlyFormattedErr = "CSV badly formatted"
	bitSize           = 64
	minRecords        = 2
	minRecordSize     = 7
)

type Restaurant struct {
	ID     string  `json:"id"`
	Lat    float64 `json:"Lat"`
	Long   float64 `json:"Long"`
	Radius float64 `json:"Radius"`
	Open   int     `json:"Open"`
	Close  int     `json:"Close"`
	Rating float64 `json:"Rating"`
}

type Restaurants []Restaurant

func MapRecordsToRestaurants(records [][]string, logs logger.Logger) (Restaurants, error) {
	var restaurants Restaurants
	if len(records) <= minRecords {
		return restaurants, errors.New("records len is less than 2")
	}
	for i, record := range records {
		if i == 0 {
			err := checkCsvFormat(record)
			if err != nil {
				return restaurants, err
			}
			continue
		}

		restaurant, err := processRestaurantRecord(record)
		if err != nil {
			logs.Warn(str.ErrorConcat(err, "entities.restaurant", "MapRecordsToRestaurants"))
			continue
		}

		restaurants = append(restaurants, restaurant)
	}

	return restaurants, nil
}

func checkCsvFormat(record []string) error {
	if len(record) != minRecordSize {
		return errors.New(badlyFormattedErr)
	}

	if record[0] != "id" || record[1] != "latitude" || record[2] != "longitude" ||
		record[3] != "availability_radius" || record[4] != "open_hour" || record[5] != "close_hour" ||
		record[6] != "rating" {
		return errors.New(badlyFormattedErr)
	}

	return nil
}

func processRestaurantRecord(record []string) (Restaurant, error) {
	if len(record) < minRecordSize {
		return Restaurant{}, errors.New(badlyFormattedErr)
	}

	lat, err := strconv.ParseFloat(record[1], bitSize)
	if err != nil {
		return Restaurant{}, errors.New(badlyFormattedErr)
	}

	long, err := strconv.ParseFloat(record[2], bitSize)
	if err != nil {
		return Restaurant{}, errors.New(badlyFormattedErr)
	}

	radius, err := strconv.ParseFloat(record[3], bitSize)
	if err != nil {
		return Restaurant{}, errors.New(badlyFormattedErr)
	}

	openHour, err := str.TimeToInt(record[4])
	if err != nil {
		return Restaurant{}, errors.New(badlyFormattedErr)
	}

	closeHour, err := str.TimeToInt(record[5])
	if err != nil {
		return Restaurant{}, errors.New(badlyFormattedErr)
	}

	rating, err := strconv.ParseFloat(record[6], bitSize)
	if err != nil {
		return Restaurant{}, errors.New(badlyFormattedErr)
	}

	restaurant := Restaurant{
		ID:     record[0],
		Lat:    lat,
		Long:   long,
		Radius: radius,
		Open:   openHour,
		Close:  closeHour,
		Rating: rating,
	}

	return restaurant, nil
}

func (r Restaurants) CreateTimeRadiusMap() TimeRadiusMap {
	timeScheduleMap := make(TimeRadiusMap)
	for _, rest := range r {
		timeScheduleMap[rest.ID] = timeRadiusSchedule{
			Open:   rest.Open,
			Close:  rest.Close,
			Radius: rest.Radius,
		}
	}
	return timeScheduleMap
}
