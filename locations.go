package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/paulmach/orb/geojson"
)

// wellcache stores the wells feature collection in memory
var wellcache struct {
	mux    sync.RWMutex
	expiry time.Time
	Wells  []byte
}

// GetWellLocations is a handler that responds to a request for well locations
func (api *server) GetWellLocations(w http.ResponseWriter, r *http.Request) {
	var jsondata []byte

	wellcache.mux.RLock()
	if time.Now().Before(wellcache.expiry) {
		jsondata = wellcache.Wells
		wellcache.mux.RUnlock()
		log.Println("responding from cache")
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsondata)
		return
	}
	wellcache.mux.RUnlock()

	// cache not valid, retrieve wells from database.
	log.Println("responding from database")

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

	jsondata, err = json.Marshal(fc)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
		return
	}

	// populate cache
	wellcache.mux.Lock()
	wellcache.Wells = jsondata
	wellcache.expiry = time.Now().Add(15 * time.Minute)
	wellcache.mux.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsondata)
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
