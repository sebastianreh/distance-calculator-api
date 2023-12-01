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

	err = r.repository.SetRestaurantGeoData(ctx, restaurants)
	if err != nil {
		return err
	}

	timeRadiusMap := restaurants.CreateTimeRadiusMap()
	err = r.repository.SetTimeRadiusMapData(ctx, timeRadiusMap)
	if err != nil {
		return err
	}

	return nil
}

func (r *calculatorService) CalculateDeliveryRange(ctx context.Context, request entities.CalculationRequest) ([]string, error) {
	var timeRadiusMap entities.TimeRadiusMap
	var restaurantInUserRadius []entities.RestaurantIDLatLng
	const parallelProcesses = 3

	var wg sync.WaitGroup

	errChan := make(chan error, parallelProcesses)
	wg.Add(1)
	go func() {
		defer wg.Done()
		timeRadiusMapData, err := r.repository.GetTimeRadiusMapData(ctx)
		if err != nil {
			errChan <- err
			return
		}
		timeRadiusMap = timeRadiusMapData
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		restaurantInUserRadiusData, err := r.repository.GetRestaurantsInRadius(ctx,
			request.Lat, request.Long, r.config.MaxDeliveryRadius)
		if err != nil {
			errChan <- err
			return
		}
		restaurantInUserRadius = restaurantInUserRadiusData
	}()

	wg.Wait()
	close(errChan)

	for e := range errChan {
		if e != nil {
			return nil, e
		}
	}

	IDs := request.FindRestaurantsInRadius(timeRadiusMap, restaurantInUserRadius)

	return IDs, nil
}
