package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/Zmahl/chirpy/internal/db"
	"github.com/joho/godotenv"
)

type apiConfig struct {
	fileServerHits int
	DB             *db.DB
	SecretString   string
	PolkaKey       string
}

func main() {
	// Load environment variables from a .env file
	godotenv.Load()

	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()
	if *dbg {
		os.Remove("database.json")
	}

	db, err := db.NewDB("database.json")
	if err != nil {
		log.Fatal(err)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	polkaKey := os.Getenv("POLKA_KEY")

	config := &apiConfig{
		fileServerHits: 0,
		DB:             db,
		SecretString:   jwtSecret,
		PolkaKey:       polkaKey,
	}

	mux := http.NewServeMux()
	fileHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	mux.Handle("/app/*", config.middlewareMetricInc(fileHandler))
	mux.HandleFunc("/api/reset", config.reset)
	mux.HandleFunc("GET /api/healthz", checkHealth)
	mux.HandleFunc("GET /admin/metrics", config.getMetrics)
	mux.HandleFunc("POST /api/chirps", config.postChirp)
	mux.HandleFunc("GET /api/chirps", config.getChirps)
	mux.HandleFunc("GET /api/chirps/{id}", config.getSingleChirp)
	mux.HandleFunc("POST /api/users", config.createUser)
	mux.HandleFunc("POST /api/login", config.userLogin)
	mux.HandleFunc("PUT /api/users", config.updateUser)
	mux.HandleFunc("POST /api/refresh", config.refreshJWT)
	mux.HandleFunc("POST /api/revoke", config.revokeJWT)
	mux.HandleFunc("DELETE /api/chirps/{id}", config.deleteChirp)
	mux.HandleFunc("POST /api/polka/webhooks", config.upgradeUser)

	// Struct that describes server configuration
	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	server.ListenAndServe()
}

func (cfg *apiConfig) middlewareMetricInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits += 1
		next.ServeHTTP(w, r)
	})
}
