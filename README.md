# Distance Calculator API

## Overview

The Distance Calculator API is designed to provide real-time information about nearby restaurant availability based on the user's location. It processes user location data (latitude and longitude) and returns a list of restaurant IDs that are capable of delivering orders to the user's location at the time of the query.

---
## Key Features

- **Real-Time Availability**: Responds to user location queries in real time.
- **Location-Based Filtering**: Uses the user's latitude and longitude to find nearby restaurants.
- **Delivery Range Consideration**: Takes into account each restaurant's delivery radius and operational hours.
- **List of Deliverable Restaurants**: Returns a list of restaurant IDs that can deliver to the user's current location.

---
## Data Source

The information about the restaurants is stored in a CSV file available at a specified URL. The CSV file is refreshed every 6 hours, there is a cronjob that refreshes the data source at this interval. The CSV file includes the following columns:

- Restaurant ID
- Latitude and Longitude of the Restaurant
- Delivery Radius
- Operational Hours
- Restaurant Raiting

---
## Endpoint Description

- `/calculate`: Accepts `GET` requests with parameters `lat` (latitude) and `long` (longitude) to calculate and return a list of restaurant IDs available for delivery to the specified location.
- `/preprocess`: A `POST` request endpoint that processes the CSV file to update the list of restaurants in the system.

---
## Usage

1. **Preprocessing Data**: First, the `/preprocess` endpoint should be called to process and load restaurant data from the CSV file.
2. **Querying Available Restaurants**: Users can then make requests to the `/calculate` endpoint with their latitude and longitude to receive a list of available restaurant IDs.
---

## Distance Calculator Api Setup Guide

This README provides a step-by-step guide to set up and run the distance-calculator-api project.

### Prerequisites

Before starting, ensure you have the following installed:

- Docker and Docker Compose
- Go (v1.19)

### Step-by-Step Guide

1. **Build the docker images**:

   This step will create the docker images both for the http service and cronjob.
   
   Run: `make build-server-image && make build-cron-image`
2. **Start Redis, Http Server and CronJob on Docker-compose**:

   This step will run docker-compose.yml file, starting redis, the http server and cronjob

   Run: `make start-compose`
3. **Shut Down the Services (Optional)**:
   
   To shut down the project and it's dependencies, use this command: 
   
   Run: `make down-compose`

---
## Example

### Request:
```http
GET /calculate/restaurants?lat=40.7128&long=74.0060
```

### Response:
```json
{
  "restaurant_ids": ["id1", "id2", "id3", ...]
}
```

---
### Benchmarks:

You can find benchmarks made with Postman in the directory `files/benchmarks`

---
### Test

In order to run the server tests, use this **command**: `make test`