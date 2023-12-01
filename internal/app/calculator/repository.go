package calculator

import (
	"context"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/sebastianreh/distance-calculator-api/internal/config"
	"github.com/sebastianreh/distance-calculator-api/internal/entities"
	"github.com/sebastianreh/distance-calculator-api/pkg/logger"
	"github.com/sebastianreh/distance-calculator-api/pkg/redis"
	str "github.com/sebastianreh/distance-calculator-api/pkg/strings"
)

const (
	repositoryName     = "calculator.repository"
	restaurantsListKey = "restaurants:raw"
	coordinatesLatKey  = "coordinates:lat"
	coordinatesLongKey = "coordinates:long"
	timeRadiusMapKey   = "restaurants:time_radius_map"

	longDataSelector = "longDataSelector"
	latDataSelector  = "lataData"

	InactiveTimeTTL = time.Duration(6)*time.Hour + time.Duration(30)*time.Minute
)

type CalculatorRepository interface {
	SetRestaurantList(ctx context.Context, restaurants entities.Restaurants) error
	GetRestaurantList(ctx context.Context) (entities.Restaurants, error)
	SetPreprocessData(ctx context.Context, coordinatesData entities.CoordinatesData, timeRadiusMap entities.TimeRadiusMap) error
	GetCoordinateIDData(ctx context.Context, coordinate string) ([]entities.CoordinateIDData, error)
	GetTimeRadiusMapData(ctx context.Context) (entities.TimeRadiusMap, error)
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

func (r *calculatorRepository) SetRestaurantList(ctx context.Context, restaurants entities.Restaurants) error {
	restaurantBytes, _ := r.json.Marshal(restaurants)

	err := r.redis.Set(ctx, restaurantsListKey, string(restaurantBytes), InactiveTimeTTL)
	if err != nil {
		r.logs.Error(str.ErrorConcat(err, repositoryName, "SetRestaurantList"))
		return err
	}

	return nil
}

func (r *calculatorRepository) GetRestaurantList(ctx context.Context) (entities.Restaurants, error) {
	var restaurants entities.Restaurants
	resString, err := r.redis.Get(ctx, restaurantsListKey)
	if err != nil {
		r.logs.Error(str.ErrorConcat(err, repositoryName, "GetRestaurantList"))
		return restaurants, err
	}

	err = r.json.Unmarshal([]byte(resString), &restaurants)
	if err != nil {
		r.logs.Error(str.ErrorConcat(err, repositoryName, "GetRestaurantList"))
		return restaurants, err
	}

	return restaurants, nil
}

func (r *calculatorRepository) SetPreprocessData(ctx context.Context, coordinatesData entities.CoordinatesData,
	timeRadiusMap entities.TimeRadiusMap) error {
	coordinatesLatDataBytes, _ := r.json.Marshal(coordinatesData.LatData)
	coordinatesLongDataBytes, _ := r.json.Marshal(coordinatesData.LongData)
	timeRadiusMapBytes, _ := r.json.Marshal(timeRadiusMap)

	err := r.redis.Set(ctx, coordinatesLatKey, string(coordinatesLatDataBytes), InactiveTimeTTL)
	if err != nil {
		r.logs.Error(str.ErrorConcat(err, repositoryName, "SetPreprocessData"))
		return err
	}

	err = r.redis.Set(ctx, coordinatesLongKey, string(coordinatesLongDataBytes), InactiveTimeTTL)
	if err != nil {
		r.logs.Error(str.ErrorConcat(err, repositoryName, "SetPreprocessData"))
		return err
	}

	err = r.redis.Set(ctx, timeRadiusMapKey, string(timeRadiusMapBytes), InactiveTimeTTL)
	if err != nil {
		r.logs.Error(str.ErrorConcat(err, repositoryName, "SetPreprocessData"))
		return err
	}

	return nil
}

func (r *calculatorRepository) GetCoordinateIDData(ctx context.Context, coordinate string) ([]entities.CoordinateIDData, error) {
	var coordinateIDData []entities.CoordinateIDData
	var key string

	switch coordinate {
	case longDataSelector:
		key = coordinatesLongKey
	case latDataSelector:
		key = coordinatesLatKey
	}

	coordinatesLatDataString, err := r.redis.Get(ctx, key)
	if err != nil {
		r.logs.Error(str.ErrorConcat(err, repositoryName, "GetCoordinatesData"))
		return coordinateIDData, nil
	}
	_ = r.json.Unmarshal([]byte(coordinatesLatDataString), &coordinateIDData)

	return coordinateIDData, nil
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
