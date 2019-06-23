package main

import (
	"log"
	"net/http"

	"github.com/go-chi/render"
	"github.com/paulmach/orb/geojson"
)

func (api *server) WellLocations(w http.ResponseWriter, r *http.Request) {
	fc := geojson.NewFeatureCollection()
	points, err := api.datastore.GetWellLocations()
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

// GetWellLocations retrieves all well points and returns them in a slice
func (db *DB) GetWellLocations() ([]*PointLocation, error) {
	query := `
		SELECT well_tag_number, ST_AsBinary(geom)
		FROM well
	`

	points := []*PointLocation{}

	err := db.Select(&points, query)
	return points, err
}
