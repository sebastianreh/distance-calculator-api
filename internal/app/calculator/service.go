package calculator

import (
	"context"
	"strconv"
	"sync"

	"github.com/sebastianreh/distance-calculator-api/internal/config"
	"github.com/sebastianreh/distance-calculator-api/internal/entities"
	customCsv "github.com/sebastianreh/distance-calculator-api/pkg/csv"
	"github.com/sebastianreh/distance-calculator-api/pkg/logger"
	"github.com/sebastianreh/distance-calculator-api/pkg/rest"
	str "github.com/sebastianreh/distance-calculator-api/pkg/strings"
)

const (
	serviceName = "calculator.service"
	chunkSize   = 100000
)

type CalculatorService interface {
	CalculateDeliveryRange(ctx context.Context, request entities.CalculationRequest) ([]string, error)
	PreprocessRestaurants(ctx context.Context) error
}

type calculatorService struct {
	config     config.Config
	repository CalculatorRepository
	restClient rest.S3Client
	logs       logger.Logger
}

func NewCalculatorService(cfg config.Config, repository CalculatorRepository, restClient rest.S3Client,
	logs logger.Logger) CalculatorService {
	return &calculatorService{
		config:     cfg,
		repository: repository,
		restClient: restClient,
		logs:       logs,
	}
}

func (r *calculatorService) PreprocessRestaurants(ctx context.Context) error {
	restaurantRecordsBytes, err := r.restClient.GetRestaurantsCSV()
	if err != nil {
		return err
	}

	restaurantRecords, err := customCsv.CsvBytesToRecords(restaurantRecordsBytes)
	if err != nil {
		r.logs.Error(str.ErrorConcat(err, serviceName, "PreprocessRestaurants"))
		return err
	}

	restaurants, err := entities.MapRecordsToRestaurants(restaurantRecords, r.logs)
	if err != nil {
		r.logs.Error(str.ErrorConcat(err, serviceName, "PreprocessRestaurants"))
		return err
	}

	coordinatesData, timeRadiusMap := restaurants.PreprocessData()

	latChunks := processCoordinatesChunks(coordinatesData.LatData, LatDataSelector)
	longChunks := processCoordinatesChunks(coordinatesData.LongData, LongDataSelector)
	timeRadiusMapChunks := processTimeRadiusMapChunks(timeRadiusMap)

	err = r.repository.SetCoordinatesIDData(ctx, latChunks, LatDataSelector)
	if err != nil {
		return err
	}

	err = r.repository.SetCoordinatesIDData(ctx, longChunks, LongDataSelector)
	if err != nil {
		return err
	}

	err = r.repository.SetTimeRadiusMapData(ctx, timeRadiusMapChunks)
	if err != nil {
		return err
	}

	return nil
}

func processCoordinatesChunks(coordinateIDData []entities.CoordinateIDData, coordinateType string) map[string][]entities.CoordinateIDData {
	coordinatesChunks := make(map[string][]entities.CoordinateIDData)

	for i := 0; i < len(coordinateIDData); i += chunkSize {
		end := i + chunkSize
		if end > len(coordinateIDData) {
			end = len(coordinateIDData)
		}

		key := "coordinates:" + coordinateType + strconv.Itoa(i/chunkSize)
		coordinatesChunks[key] = append(coordinatesChunks[key], coordinateIDData[i:end]...)

		if end == len(coordinateIDData) {
			break
		}
	}

	return coordinatesChunks
}

func processTimeRadiusMapChunks(trMap entities.TimeRadiusMap) map[string]entities.TimeRadiusMap {
	chunkedMaps := make(map[string]entities.TimeRadiusMap)
	counter := 0
	chunkCounter := 0

	for key, value := range trMap {
		if counter%chunkSize == 0 {
			chunkedMaps[timeRadiusKeyTemplate+strconv.Itoa(chunkCounter)] = make(entities.TimeRadiusMap)
			chunkCounter++
		}

		currentChunkKey := timeRadiusKeyTemplate + strconv.Itoa(chunkCounter-1)
		if chunkedMaps[currentChunkKey] == nil {
			chunkedMaps[currentChunkKey] = make(entities.TimeRadiusMap)
		}

		chunkedMaps[currentChunkKey][key] = value

		counter++
	}

	return chunkedMaps
}

func (r *calculatorService) CalculateDeliveryRange(ctx context.Context, request entities.CalculationRequest) ([]string, error) {
	var timeRadiusMap entities.TimeRadiusMap
	var coordinatesData entities.CoordinatesData
	var err error
	const parallelProcesses = 3

	var wg sync.WaitGroup
	wg.Add(parallelProcesses)

	errChan := make(chan error, parallelProcesses)

	go func() {
		defer wg.Done()
		latData, e := r.repository.GetCoordinateIDData(ctx, LatDataSelector)
		if e != nil {
			errChan <- e
			return
		}
		coordinatesData.LatData = latData
	}()

	go func() {
		defer wg.Done()
		longData, e := r.repository.GetCoordinateIDData(ctx, LongDataSelector)
		if e != nil {
			errChan <- e
			return
		}
		coordinatesData.LongData = longData
	}()

	go func() {
		defer wg.Done()
		timeRadiusMapData, e := r.repository.GetTimeRadiusMapData(ctx)
		if e != nil {
			errChan <- e
			return
		}
		timeRadiusMap = timeRadiusMapData
	}()

	wg.Wait()
	close(errChan)

	for e := range errChan {
		if e != nil {
			return nil, e
		}
	}

	IDsInRadius := request.FindRestaurantsInRadius(coordinatesData, timeRadiusMap)

	return IDsInRadius, err
}
