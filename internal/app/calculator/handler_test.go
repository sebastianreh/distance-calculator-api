package calculator_test

import (
	"errors"
	"fmt"
	"github.com/sebastianreh/distance-calculator-api/internal/app/calculator"
	"github.com/sebastianreh/distance-calculator-api/internal/config"
	"github.com/sebastianreh/distance-calculator-api/internal/entities"
	"github.com/sebastianreh/distance-calculator-api/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sebastianreh/distance-calculator-api/cmd/httpserver"
	"github.com/sebastianreh/distance-calculator-api/internal/container"
	"github.com/sebastianreh/distance-calculator-api/pkg/logger"
)

var latLongParams = []string{"lat", "long"}

func setup(method, target string, body *strings.Reader) (echo.Context, *httptest.ResponseRecorder) {
	mockServer := httpserver.NewServer(container.Dependencies{})

	request := httptest.NewRequest(method, target, body)
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	w := httptest.NewRecorder()
	mockServer.Server.HTTPErrorHandler = httpserver.HTTPErrorHandler
	ctx := mockServer.NewServerContext(request, w)

	return ctx, w
}

func setPathAndParams(ctx echo.Context, names, values []string, path string) {
	ctx.SetPath(path)
	ctx.SetParamNames(names...)
	ctx.SetParamValues(values...)
}

func Test_CalculatorHandler_Calculate(t *testing.T) {
	logs := logger.NewLogger()
	cfg := config.NewConfig()

	t.Run("successful calculation", func(t *testing.T) {
		serviceMock := mocks.NewCalculatorServiceMock()
		request := entities.CalculationRequest{
			Lat:  50.0534197,
			Long: 8.6705214,
			Now:  time.Now(),
		}

		response := entities.CalculationResponse{RestaurantIDs: []string{"1", "2", "3"}}

		ctx, recorder := setup(http.MethodPost, "/calculate", strings.NewReader(""))

		latLongValues := []string{fmt.Sprintf("%f", request.Lat), fmt.Sprintf("%f", request.Long)}
		setPathAndParams(ctx, latLongParams, latLongValues, "/calculate?lat:lat&long:long")
		serviceMock.On("CalculateDeliveryRange", ctx.Request().Context(),
			mock.AnythingOfType("entities.CalculationRequest")).Return(response, nil)

		handler := calculator.NewCalculatorHandler(cfg, serviceMock, logs)
		err := handler.Calculate(ctx)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})

	t.Run("binding error", func(t *testing.T) {
		serviceMock := mocks.NewCalculatorServiceMock()

		body := strings.NewReader("invalid json")
		ctx, recorder := setup(http.MethodPost, "/calculate", body)

		setPathAndParams(ctx, latLongParams, []string{}, "/calculate?lat:lat&long:long")

		handler := calculator.NewCalculatorHandler(cfg, serviceMock, logs)
		err := handler.Calculate(ctx)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("service error", func(t *testing.T) {
		serviceMock := mocks.NewCalculatorServiceMock()
		request := entities.CalculationRequest{
			Lat:  50.0534197,
			Long: 8.6705214,
			Now:  time.Now(),
		}

		expectedError := errors.New("service error")

		ctx, recorder := setup(http.MethodPost, "/calculate", strings.NewReader(""))

		latLongValues := []string{fmt.Sprintf("%f", request.Lat), fmt.Sprintf("%f", request.Long)}
		setPathAndParams(ctx, latLongParams, latLongValues, "/calculate?lat:lat&long:long")
		serviceMock.On("CalculateDeliveryRange", ctx.Request().Context(),
			mock.AnythingOfType("entities.CalculationRequest")).Return(entities.CalculationResponse{}, expectedError)

		handler := calculator.NewCalculatorHandler(cfg, serviceMock, logs)
		err := handler.Calculate(ctx)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, recorder.Code) //
	})
}

func Test_CalculatorHandler_PreprocessRestaurants(t *testing.T) {
	logs := logger.NewLogger()
	cfg := config.NewConfig()

	t.Run("successful preprocessing", func(t *testing.T) {
		serviceMock := mocks.NewCalculatorServiceMock()

		ctx, recorder := setup(http.MethodPost, "/preprocess", strings.NewReader(""))

		serviceMock.On("PreprocessRestaurants", ctx.Request().Context()).Return(nil)

		handler := calculator.NewCalculatorHandler(cfg, serviceMock, logs)
		err := handler.PreprocessRestaurants(ctx)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})

	t.Run("service error", func(t *testing.T) {
		serviceMock := mocks.NewCalculatorServiceMock()

		expectedError := errors.New("preprocessing error")

		ctx, recorder := setup(http.MethodPost, "/preprocess", strings.NewReader(""))

		serviceMock.On("PreprocessRestaurants", ctx.Request().Context()).Return(expectedError)

		handler := calculator.NewCalculatorHandler(cfg, serviceMock, logs)
		err := handler.PreprocessRestaurants(ctx)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})
}
