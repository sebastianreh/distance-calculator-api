package calculator

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/sebastianreh/distance-calculator-api/internal/config"
	"github.com/sebastianreh/distance-calculator-api/internal/entities"
	"github.com/sebastianreh/distance-calculator-api/pkg/logger"
	"github.com/sebastianreh/distance-calculator-api/pkg/redis"
	str "github.com/sebastianreh/distance-calculator-api/pkg/strings"
)

const (
	repositoryName = "calculator.repository"

	coordinatesLatKey     = "lat_keys"
	coordinatesLongKey    = "long_keys"
	timeRadiusKeysKey     = "time_radius_map_keys"
	timeRadiusKeyTemplate = "time_radius_map:"

	LongDataSelector = "long"
	LatDataSelector  = "lat"

	InactiveTimeTTL = time.Duration(6)*time.Hour + time.Duration(30)*time.Minute
)

type CalculatorRepository interface {
	SetCoordinatesIDData(ctx context.Context, coordinatesChunks map[string][]entities.CoordinateIDData, coordinateType string) error
	GetCoordinateIDData(ctx context.Context, coordinate string) ([]entities.CoordinateIDData, error)
	SetTimeRadiusMapData(ctx context.Context, timeRadiusMapChunks map[string]entities.TimeRadiusMap) error
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

func (r *calculatorRepository) SetCoordinatesIDData(ctx context.Context,
	coordinatesChunks map[string][]entities.CoordinateIDData, coordinateType string) error {
	var coordinateKey string

	switch coordinateType {
	case LongDataSelector:
		coordinateKey = coordinatesLongKey
	case LatDataSelector:
		coordinateKey = coordinatesLatKey
	}

	keys := make([]string, 0)
	for k, v := range coordinatesChunks {
		keys = append(keys, k)
		coordinatesDataBytes, _ := r.json.Marshal(v)
		err := r.redis.Set(ctx, k, string(coordinatesDataBytes), InactiveTimeTTL)
		if err != nil {
			r.logs.Error(str.ErrorConcat(err, repositoryName, "SetCoordinatesIDData"))
			return err
		}
	}

	sort.Strings(keys)
	keysBytes, _ := r.json.Marshal(keys)
	err := r.redis.Set(ctx, fmt.Sprintf(coordinateKey), string(keysBytes), InactiveTimeTTL)
	if err != nil {
		r.logs.Error(str.ErrorConcat(err, repositoryName, "SetCoordinatesIDData"))
		return err
	}

	return nil
}

func (r *calculatorRepository) SetTimeRadiusMapData(ctx context.Context,
	timeRadiusMapChunks map[string]entities.TimeRadiusMap) error {
	keys := make([]string, 0)
	for k, v := range timeRadiusMapChunks {
		keys = append(keys, k)
		timeRadiusMapBytes, _ := r.json.Marshal(v)
		err := r.redis.Set(ctx, k, string(timeRadiusMapBytes), InactiveTimeTTL)
		if err != nil {
			r.logs.Error(str.ErrorConcat(err, repositoryName, "SetTimeRadiusMapData"))
			return err
		}
	}

	sort.Strings(keys)
	keysBytes, _ := r.json.Marshal(keys)
	err := r.redis.Set(ctx, timeRadiusKeysKey, string(keysBytes), InactiveTimeTTL)
	if err != nil {
		r.logs.Error(str.ErrorConcat(err, repositoryName, "SetCoordinatesIDData"))
		return err
	}

	return nil
}

func (r *calculatorRepository) GetCoordinateIDData(ctx context.Context, coordinateType string) ([]entities.CoordinateIDData, error) {
	var finalCoordinateIDData []entities.CoordinateIDData
	var coordinatesChunkKeys []string
	var key string

	switch coordinateType {
	case LongDataSelector:
		key = coordinatesLongKey
	case LatDataSelector:
		key = coordinatesLatKey
	}

	coordinatesChunkKeysString, err := r.redis.Get(ctx, key)
	if err != nil {
		r.logs.Error(str.ErrorConcat(err, repositoryName, "GetCoordinatesData"))
		return finalCoordinateIDData, err
	}
	_ = r.json.Unmarshal([]byte(coordinatesChunkKeysString), &coordinatesChunkKeys)

	dataChan := make(chan []entities.CoordinateIDData, len(coordinatesChunkKeys))
	errChan := make(chan error, len(coordinatesChunkKeys))

	var wg sync.WaitGroup
	wg.Add(len(coordinatesChunkKeys))

	for _, coordinatesChunkKey := range coordinatesChunkKeys {
		go func(chunkKey string) {
			defer wg.Done()
			var coordinateIDDataSlice []entities.CoordinateIDData
			coordinateIDDataSliceString, err := r.redis.Get(ctx, chunkKey)
			if err != nil {
				errChan <- err
				return
			}
			_ = r.json.Unmarshal([]byte(coordinateIDDataSliceString), &coordinateIDDataSlice)
			dataChan <- coordinateIDDataSlice
		}(coordinatesChunkKey)
	}

	wg.Wait()
	close(dataChan)
	close(errChan)

	for chunkData := range dataChan {
		finalCoordinateIDData = append(finalCoordinateIDData, chunkData...)
	}

	for err := range errChan {
		if err != nil {
			r.logs.Error(str.ErrorConcat(err, repositoryName, "GetCoordinatesData"))
			return nil, err
		}
	}

	sort.Slice(finalCoordinateIDData, func(i, j int) bool {
		return finalCoordinateIDData[i].Coordinate < finalCoordinateIDData[j].Coordinate
	})

	return finalCoordinateIDData, nil
}

func (r *calculatorRepository) GetTimeRadiusMapData(ctx context.Context) (entities.TimeRadiusMap, error) {
	var timeRadiusMapFinal entities.TimeRadiusMap
	var timeRadiusMapKeys []string

	timeRadiusChunksKeysString, err := r.redis.Get(ctx, timeRadiusKeysKey)
	if err != nil {
		r.logs.Error(str.ErrorConcat(err, repositoryName, "GetCoordinatesData"))
		return timeRadiusMapFinal, nil
	}
	_ = r.json.Unmarshal([]byte(timeRadiusChunksKeysString), &timeRadiusMapKeys)

	for _, timeRadiusMapKey := range timeRadiusMapKeys {
		var timeRadiusMap entities.TimeRadiusMap
		timeRadiusMapString, err := r.redis.Get(ctx, timeRadiusMapKey)
		if err != nil {
			r.logs.Error(str.ErrorConcat(err, repositoryName, "GetCoordinatesData"))
			return timeRadiusMapFinal, err
		}
		_ = r.json.Unmarshal([]byte(timeRadiusMapString), &timeRadiusMap)
		timeRadiusMapFinal = unifyMaps(timeRadiusMapFinal, timeRadiusMap)
	}

	return timeRadiusMapFinal, nil
}

func unifyMaps(maps ...entities.TimeRadiusMap) entities.TimeRadiusMap {
	result := make(entities.TimeRadiusMap)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}
