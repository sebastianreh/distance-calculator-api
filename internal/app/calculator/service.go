package calculator

import (
	"context"
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
	err = r.repository.SetPreprocessData(ctx, coordinatesData, timeRadiusMap)
	if err != nil {
		return err
	}

	return nil
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
		latData, e := r.repository.GetCoordinateIDData(ctx, latDataSelector)
		if e != nil {
			errChan <- e
			return
		}
		coordinatesData.LatData = latData
	}()

	go func() {
		defer wg.Done()
		longData, e := r.repository.GetCoordinateIDData(ctx, longDataSelector)
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
