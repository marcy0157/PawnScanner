package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"pwnscanner/pkg/database"
	"strconv"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	httpSwagger "github.com/swaggo/http-swagger"
	"pwnscanner/pkg/checker"

	"pwnscanner/pkg/utils"

	_ "pwnscanner/docs" // Importa i file generati da Swagger
)

var db database.Database

func main() {
	// Carica la configurazione dalle variabili d'ambiente
	if err := loadConfig(); err != nil {
		fmt.Printf("Errore nel caricamento della configurazione: %v\n", err)
		os.Exit(1)
	}

	// Configura il logger
	setupLogger()
	log.Info().Msg("Avvio di PwnScannerFront...")

	// Inizializza il contesto
	ctx := context.Background()

	// Inizializza il database
	var err error
	dbType := os.Getenv("DB_TYPE")
	if dbType == "" {
		dbType = "mongodb"
	}

	switch dbType {
	case "mongodb":
		dbHost := os.Getenv("DB_HOST")
		dbPortStr := os.Getenv("DB_PORT")
		dbUser := os.Getenv("DB_USERNAME")
		dbPassword := os.Getenv("DB_PASSWORD")
		dbName := os.Getenv("DB_NAME")

		if dbHost == "" || dbPortStr == "" || dbUser == "" || dbPassword == "" || dbName == "" {
			log.Fatal().Msg("Variabili d'ambiente del database mancanti")
		}

		dbPort, err := strconv.Atoi(dbPortStr)
		if err != nil {
			log.Fatal().Err(err).Msg("Porta del database non valida")
		}

		db, err = database.NewMongoDB(
			ctx,
			dbHost,
			dbPort,
			dbUser,
			dbPassword,
			dbName,
			"breaches",
		)
	default:
		log.Fatal().Msg("Tipo di database non supportato")
	}
	if err != nil {
		log.Fatal().Err(err).Msg("Errore durante l'inizializzazione del database")
	}
	defer db.Close()

	// Inizializza il Checker
	cacheSizeMBStr := os.Getenv("CACHE_SIZE_MB")
	if cacheSizeMBStr == "" {
		cacheSizeMBStr = "100" // Valore di default
	}
	cacheSizeMB, err := strconv.Atoi(cacheSizeMBStr)
	if err != nil {
		log.Fatal().Err(err).Msg("Cache size non valida")
	}

	log.Info().Msgf("Inizializzazione del Checker con cache di %d MB...", cacheSizeMB)
	c, err := checker.NewChecker(db, cacheSizeMB)
	if err != nil {
		log.Fatal().Err(err).Msg("Errore durante l'inizializzazione del Checker")
	}
	log.Info().Msg("Checker inizializzato con successo.")

	// Configura e avvia gli endpoint
	http.Handle("/metrics", promhttp.Handler()) // Endpoint Prometheus
	http.Handle("/check-email", authMiddleware(http.HandlerFunc(handleCheckEmail(c))))
	http.Handle("/breaches", authMiddleware(http.HandlerFunc(handleGetBreaches(db))))
	http.Handle("/swagger/", httpSwagger.WrapHandler) // Endpoint Swagger

	// Servire file statici
	fs := http.FileServer(http.Dir("./web"))
	http.Handle("/", fs)

	log.Info().Msg("Endpoint REST esposti: /check-email, /breaches, /metrics, /swagger/")
	log.Info().Msg("File statici serviti su /")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal().Err(err).Msg("Errore durante l'avvio del server HTTP")
	}
}

// loadConfig carica le variabili d'ambiente
func loadConfig() error {
	// In questo caso, non abbiamo nulla da caricare
	return nil
}

// setupLogger configura il logger globale
func setupLogger() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	levelStr := os.Getenv("LOG_LEVEL")
	if levelStr == "" {
		levelStr = "info" // Livello di default
	}
	logLevel, err := zerolog.ParseLevel(levelStr)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}
	log.Logger = zerolog.New(os.Stdout).Level(logLevel).With().Timestamp().Logger()
}

// @Summary Verifica un'email nei breach
// @Description Cerca se un'email è presente in uno o più breach
// @Tags Email
// @Accept json
// @Produce json
// @Param email body string true "Email da verificare"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /check-email [post]
func handleCheckEmail(c *checker.Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			utils.WriteError(w, http.StatusMethodNotAllowed, "Metodo non supportato")
			return
		}

		var req struct {
			Email string `json:"email"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteError(w, http.StatusBadRequest, "Richiesta non valida")
			return
		}

		breaches, err := c.FindEmailInBreaches(context.Background(), req.Email)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "Errore interno del server")
			return
		}

		if len(breaches) == 0 {
			utils.WriteError(w, http.StatusNotFound, "Nessun breach trovato per questa email")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(struct {
			Email    string   `json:"email"`
			Breaches []string `json:"breaches"`
		}{
			Email:    req.Email,
			Breaches: breaches,
		})
	}
}

// @Summary Ottiene tutti i breach disponibili
// @Description Restituisce un elenco di tutti i breach registrati nel sistema
// @Tags Breach
// @Accept json
// @Produce json
// @Success 200 {array} string
// @Failure 500 {object} utils.ErrorResponse
// @Router /breaches [get]
func handleGetBreaches(db database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			utils.WriteError(w, http.StatusMethodNotAllowed, "Metodo non supportato")
			return
		}

		breaches, err := db.GetAllBreaches(context.Background())
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "Errore nel recupero dei breach")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(breaches)
	}
}

// authMiddleware protegge gli endpoint con autenticazione basata su token
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token != "Bearer YOUR_SECRET_TOKEN" {
			utils.WriteError(w, http.StatusUnauthorized, "Accesso non autorizzato")
			return
		}
		next.ServeHTTP(w, r)
	})
}
