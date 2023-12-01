package rest

import (
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sebastianreh/distance-calculator-api/internal/config"
	"github.com/sebastianreh/distance-calculator-api/pkg/logger"

	str "github.com/sebastianreh/distance-calculator-api/pkg/strings"
)

const (
	url                         = "https://s3.amazonaws.com/test.jampp.com/dmarasca/takehome.csv"
	GetRestaurantsCSVMethodName = "GetRestaurantsCSV"
	apiClientName               = "stooq_client"
)

type (
	S3Client interface {
		GetRestaurantsCSV() ([]byte, error)
	}

	s3Client struct {
		configs    config.Config
		logs       logger.Logger
		restClient *resty.Client
	}
)

func NewS3Client(logs logger.Logger, restClient *resty.Client) S3Client {
	return &s3Client{
		logs:       logs,
		restClient: restClient,
	}
}

func (client *s3Client) GetRestaurantsCSV() ([]byte, error) {
	var res []byte
	var err error
	var resp *resty.Response

	req := client.restClient.R()
	resp, err = req.Get(url)
	if err != nil {
		client.logs.Error(str.ErrorConcat(err, apiClientName, GetRestaurantsCSVMethodName))
		return res, err
	}

	if !resp.IsSuccess() {
		err = fmt.Errorf("error getting restaurants https status code: %d, body %s",
			resp.StatusCode(), string(resp.Body()))
		client.logs.Error(str.ErrorConcat(err, apiClientName, GetRestaurantsCSVMethodName))
		return res, err
	}

	return resp.Body(), nil
}
