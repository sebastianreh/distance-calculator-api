package calculator

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/sebastianreh/distance-calculator-api/internal/config"
	"github.com/sebastianreh/distance-calculator-api/internal/entities"
	"github.com/sebastianreh/distance-calculator-api/pkg/logger"
	"github.com/sebastianreh/distance-calculator-api/pkg/redis"
	str "github.com/sebastianreh/distance-calculator-api/pkg/strings"
)

const (
	repositoryName         = "calculator.repository"
	timeRadiusMapKey       = "restaurants:time_radius_map"
	restaurantsGeoDataKey  = "restaurants:geodata"
	InactiveTimeTTL        = time.Duration(12)*time.Hour + time.Duration(30)*time.Minute
	invalidLatLongRedisErr = "ERR invalid longitude,latitude pair"
	keySeparator           = "-"
	keyParts               = 4
	bitSize                = 64
)

type CalculatorRepository interface {
	SetTimeRadiusMapData(ctx context.Context, timeRadiusMap entities.TimeRadiusMap) error
	GetTimeRadiusMapData(ctx context.Context) (entities.TimeRadiusMap, error)
	SetRestaurantGeoData(ctx context.Context, restaurants entities.Restaurants) error
	GetRestaurantsInRadius(ctx context.Context, lat, long, radius float64) ([]entities.RestaurantIDLatLng, error)
}

type calculatorRepository struct {
	config config.Config
	redis  redis.Redis
	logs   logger.Logger
	json   jsoniter.API
}

func NewCalculatorRepository(cfg config.Config, rds redis.Redis, logs logger.Logger) CalculatorRepository {
	return &calculatorRepository{
		config: cfg,
		redis:  rds,
		logs:   logs,
		json:   jsoniter.ConfigCompatibleWithStandardLibrary,
	}
}

func (r *calculatorRepository) SetTimeRadiusMapData(ctx context.Context, timeRadiusMap entities.TimeRadiusMap) error {
	timeRadiusMapBytes, _ := r.json.Marshal(timeRadiusMap)

	err := r.redis.Set(ctx, timeRadiusMapKey, string(timeRadiusMapBytes), InactiveTimeTTL)
	if err != nil {
		r.logs.Error(str.ErrorConcat(err, repositoryName, "SetTimeRadiusMapData"))
		return err
	}

	return nil
}

func (r *calculatorRepository) GetTimeRadiusMapData(ctx context.Context) (entities.TimeRadiusMap, error) {
	var timeRadiusMap entities.TimeRadiusMap
	timeRadiusMapString, err := r.redis.Get(ctx, timeRadiusMapKey)
	if err != nil {
		r.logs.Error(str.ErrorConcat(err, repositoryName, "GetCoordinatesData"))
		return timeRadiusMap, err
	}

	_ = r.json.Unmarshal([]byte(timeRadiusMapString), &timeRadiusMap)

	return timeRadiusMap, nil
}

func (r *calculatorRepository) SetRestaurantGeoData(ctx context.Context, restaurants entities.Restaurants) error {
	for _, restaurant := range restaurants {
		err := r.redis.GeoAdd(ctx, restaurantsGeoDataKey, restaurant.ID, restaurant.Lat, restaurant.Long, restaurant.Radius)
		if err != nil {
			r.logs.Error(str.ErrorConcat(err, repositoryName, "SetRestaurantGeoData"))
			return err
		}
	}
	return nil
}

func (r *calculatorRepository) GetRestaurantsInRadius(ctx context.Context, lat, long,
	radius float64) ([]entities.RestaurantIDLatLng, error) {
	var restaurants []entities.RestaurantIDLatLng
	rawRestaurantsStrings, err := r.redis.GeoSearch(ctx, restaurantsGeoDataKey, lat, long, radius)
	if err != nil {
		if strings.Contains(err.Error(), invalidLatLongRedisErr) {
			return restaurants, nil
		}
		r.logs.Error(str.ErrorConcat(err, repositoryName, "GetRestaurantsInRadius"))
		return restaurants, err
	}

	for _, restaurantString := range rawRestaurantsStrings {
		restaurant, err := rawRestaurantStringToData(restaurantString)
		if err != nil {
			r.logs.Error(str.ErrorConcat(err, repositoryName, "GetRestaurantsInRadius"))
			return restaurants, err
		}
		restaurants = append(restaurants, restaurant)
	}

	return restaurants, nil
}

func rawRestaurantStringToData(restaurantString string) (entities.RestaurantIDLatLng, error) {
	parts := strings.Split(restaurantString, keySeparator)
	var restaurant entities.RestaurantIDLatLng

	if len(parts) != keyParts {
		err := errors.New("error: Input string does not have exactly 4 parts")
		return restaurant, err
	}

	id := parts[0]
	lat, err := strconv.ParseFloat(parts[1], bitSize)
	if err != nil {
		err = errors.New(fmt.Sprintf("error converting %s to float64: %v", parts[1], err))
		return restaurant, err
	}

	long, err := strconv.ParseFloat(parts[2], bitSize)
	if err != nil {
		err = errors.New(fmt.Sprintf("error converting %s to float64: %v", parts[2], err))
		return restaurant, err
	}

	radius, err := strconv.ParseFloat(parts[3], bitSize)
	if err != nil {
		err = errors.New(fmt.Sprintf("Error converting %s to float64: %v\n", parts[3], err))
		return restaurant, err
	}

	restaurant = entities.RestaurantIDLatLng{
		ID:             id,
		Lat:            lat,
		Long:           long,
		DeliveryRadius: radius,
	}

	return restaurant, nil
}
