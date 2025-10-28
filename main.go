package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"GeoRunner/quadtree"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var worldBoundary = quadtree.Boundary{
	X:      0,
	Y:      0,
	Width:  180,
	Height: 90,
}

var tree *quadtree.QuadTree

const (
	numDrivers    = 10000
	moveInterval  = 2 * time.Second
	searchRadiusX = 20.0
	searchRadiusY = 20.0
)

func simulateDriver(driverID string, seed int64) {

	rng := rand.New(rand.NewSource(time.Now().UnixNano() + seed))

	time.Sleep(time.Duration(rng.Intn(5000)) * time.Millisecond)

	currentPoint := &quadtree.Point{
		X:    (rng.Float64() * 360) - 180,
		Y:    (rng.Float64() * 180) - 90,
		Data: driverID,
	}

	tree.Insert(currentPoint)

	for {

		time.Sleep(moveInterval)

		tree.Remove(currentPoint)

		newLon := currentPoint.X + (rng.Float64()-0.5)*0.1
		newLat := currentPoint.Y + (rng.Float64()-0.5)*0.1

		if newLon > 180 {
			newLon = -180
		}
		if newLon < -180 {
			newLon = 180
		}
		if newLat > 90 {
			newLat = -90
		}
		if newLat < -90 {
			newLat = 90
		}

		newPoint := &quadtree.Point{
			X:    newLon,
			Y:    newLat,
			Data: driverID,
		}

		tree.Insert(newPoint)

		currentPoint = newPoint
	}
}

func handleFindNearby(c *gin.Context) {

	latStr := c.Query("lat")
	lonStr := c.Query("lon")

	lat, errLat := strconv.ParseFloat(latStr, 64)
	lon, errLon := strconv.ParseFloat(lonStr, 64)

	if errLat != nil || errLon != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parametri 'lat' e 'lon' non validi o mancanti"})
		return
	}

	searchArea := &quadtree.Boundary{
		X:      lon,
		Y:      lat,
		Width:  searchRadiusX,
		Height: searchRadiusY,
	}

	foundPoints := tree.Query(searchArea)

	type DriverResponse struct {
		ID  string  `json:"id"`
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	}

	results := make([]DriverResponse, 0, len(foundPoints))
	for _, p := range foundPoints {

		if id, ok := p.Data.(string); ok {
			results = append(results, DriverResponse{
				ID:  id,
				Lat: p.Y,
				Lon: p.X,
			})
		}
	}

	c.JSON(http.StatusOK, results)
}

func main() {

	tree = quadtree.NewQuadTree(worldBoundary, 4)

	log.Printf("Starting simulation with %d driver...", numDrivers)
	for i := 0; i < numDrivers; i++ {
		driverID := fmt.Sprintf("driver-%d", i)
		go simulateDriver(driverID, int64(i))
	}
	log.Println("Simulation started in the background.")

	r := gin.Default()

	r.Use(cors.Default())

	r.GET("/find-nearby", handleFindNearby)

	log.Println("API server listening on http://localhost:8080")
	r.Run(":8080")
}
