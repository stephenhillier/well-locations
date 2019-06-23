package main

import (
	"log"
	"net/http"

	"github.com/go-chi/render"
	"github.com/paulmach/orb/geojson"
)

// GetWellLocations is a handler that responds to a request for well locations
func (api *server) GetWellLocations(w http.ResponseWriter, r *http.Request) {
	fc := geojson.NewFeatureCollection()
	points, err := api.datastore.AllWellLocations()
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
		return
	}

	for _, pt := range points {
		fc.Append(geojson.NewFeature(pt))
	}
	render.JSON(w, r, fc)
}

// AllWellLocations retrieves all well points and returns them in a slice
func (db *DB) AllWellLocations() ([]*PointLocation, error) {
	query := `
		SELECT ST_AsBinary(geom)
		FROM well
	`

	points := []*PointLocation{}

	err := db.Select(&points, query)
	return points, err
}
