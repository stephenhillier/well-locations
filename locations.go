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

// Well represents a well with a well tag number and a location
type Well struct {
	WTN      int64         `db:"well_tag_number"`
	Location PointLocation `db:"geom"`
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
		new := geojson.NewFeature(pt.Location)
		new.Properties["n"] = pt.WTN
		fc.Append(new)
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
func (db *DB) AllWellLocations() ([]*Well, error) {
	query := `
		SELECT well_tag_number, ST_AsBinary(geom) AS geom
		FROM well
	`

	points := []*Well{}

	err := db.Select(&points, query)
	return points, err
}
