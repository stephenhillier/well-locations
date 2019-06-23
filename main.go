package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/namsral/flag"
)

// server represents the server environment (db and router)
type server struct {
	router    chi.Router
	datastore DB
	config    config
}

// config holds server/database/auth service configuration
type config struct {
	// authCert         PEMCert // defined in auth.go
	authAudience     string
	authIssuer       string
	authJWKSEndpoint string
	dbuser           string
	dbpass           string
	dbname           string
	dbhost           string
	dbport           string
	dbdriver         string
	dbsslmode        string
	defaultPageLimit int
	maxPageLimit     int
	authGroupClaim   string
	authRoleClaim    string
}

func main() {

	// set config parameters
	// the flag library grabs values either from command line args, env variables, or the default specified here
	// see github.com/namsral/flag
	conf := config{}
	flag.StringVar(&conf.dbdriver, "dbdriver", "postgres", "database driver")
	flag.StringVar(&conf.dbuser, "dbuser", "gwells", "database username")
	flag.StringVar(&conf.dbpass, "dbpass", "", "database password")
	flag.StringVar(&conf.dbname, "dbname", "gwells", "database name")
	flag.StringVar(&conf.dbhost, "dbhost", "127.0.0.1", "database service host")
	flag.StringVar(&conf.dbport, "dbport", "5432", "database service port")
	flag.StringVar(&conf.dbsslmode, "dbsslmode", "disable", "database ssl mode")
	flag.StringVar(&conf.authAudience, "auth_audience", "", "authentication service audience claim")
	flag.StringVar(&conf.authIssuer, "auth_issuer", "", "authentication service issuer claim")
	flag.StringVar(&conf.authJWKSEndpoint, "jwks_endpoint", "/.well-known/jwks.json", "authentication JWKS endpoint")
	flag.Parse()

	api := &server{}
	api.config = conf

	api.config.defaultPageLimit = 10
	api.config.maxPageLimit = 100

	// // get new certificate when server initially starts
	// // see auth.go
	// cert, err := api.getCert(nil)
	// if err != nil {
	// 	log.Panic(err)
	// }

	// api.config.authCert = cert

	// api.config.authGroupClaim = api.config.authAudience + "/claims/authorization/groups"
	// api.config.authRoleClaim = api.config.authAudience + "/claims/authorization/roles"

	api.config.dbuser = os.Getenv("DBUSER")
	api.config.dbpass = os.Getenv("DBPASS")

	// create db connection and router and use them to create a new "Server" instance
	db, err := api.NewDB()
	if err != nil {
		log.Panic(err)
	}

	api.datastore = DB{db}

	router := chi.NewRouter()

	// CORS settings
	cors := cors.New(cors.Options{
		// AllowedOrigins: []string{"https://foo.com"},
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})

	router.Use(cors.Handler)

	// register middleware
	router.Use(middleware.Logger)

	// register routes from routes.go
	api.router = api.appRoutes(router)

	h := http.Server{Addr: ":8000", Handler: api.router}

	log.Printf("Starting HTTP server on port 8000.\n")
	log.Printf("Press CTRL+C to stop.")
	go func() {
		if err := h.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Server is listening; Wait here for interrupt signal...
	<-stop
	log.Println("Shutting down...")
	h.Shutdown(context.Background())
	log.Println("Server stopped.")
}

// health is a simple health check handler that returns HTTP 200 OK.
func (api *server) health(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// appRoutes registers application routes and returns a chi router.
func (api *server) appRoutes(r chi.Router) chi.Router {
	r.Route("/api/v1", func(r chi.Router) {
		// Public routes
		r.Group(func(r chi.Router) {
			// server health check
			r.Get("/health", api.health)
			r.Get("/locations", api.WellLocations)
		})
	})
	return r
}
