#!/bin/bash

BASE_URL="http://httpserver:8000"

if curl -X POST "${BASE_URL}/distance-calculator-api/calculate/preprocess"; then
    echo "$(date): POST request to ${BASE_URL}/distance-calculator-api/calculate/preprocess succeeded."
else
    echo "$(date): POST request to ${BASE_URL}/distance-calculator-api/calculate/preprocess failed."
fi