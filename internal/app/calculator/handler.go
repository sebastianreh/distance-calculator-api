package calculator

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sebastianreh/distance-calculator-api/internal/config"
	"github.com/sebastianreh/distance-calculator-api/internal/entities"
	"github.com/sebastianreh/distance-calculator-api/pkg/logger"
	str "github.com/sebastianreh/distance-calculator-api/pkg/strings"
)

const handlerName = "calculation.handler"

type CalculatorHandler interface {
	Calculate(ctx echo.Context) error
	PreprocessRestaurants(ctx echo.Context) error
}

type calculatorHandler struct {
	config  config.Config
	service CalculatorService
	logs    logger.Logger
}

func NewCalculatorHandler(cfg config.Config, service CalculatorService, logs logger.Logger) CalculatorHandler {
	return &calculatorHandler{
		config:  cfg,
		service: service,
		logs:    logs,
	}
}

func (h *calculatorHandler) Calculate(ctx echo.Context) error {
	request := new(entities.CalculationRequest)
	if err := ctx.Bind(request); err != nil {
		h.logs.Error(str.ErrorConcat(err, handlerName, "CalculateDeliveryRange"))
		ctx.Error(err)
		return nil
	}
	request.Now = time.Now()

	response, err := h.service.CalculateDeliveryRange(ctx.Request().Context(), *request)
	if err != nil {
		ctx.Error(err)
		return nil
	}

	return ctx.JSON(http.StatusOK, response)
}

func (h *calculatorHandler) PreprocessRestaurants(ctx echo.Context) error {
	err := h.service.PreprocessRestaurants(ctx.Request().Context())
	if err != nil {
		ctx.Error(err)
		return nil
	}
	h.logs.Info("Finish pre processing CSV data", fmt.Sprintf("%s.%s", handlerName, "CalculateDeliveryRange"))

	return ctx.NoContent(http.StatusOK)
}
