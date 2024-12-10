package checker

import (
	"context"
	"pwnscanner/pkg/database"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/prometheus/client_golang/prometheus"
)

// Definizione delle metriche Prometheus
var (
	totalRequests = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "total_requests",
			Help: "Numero totale di richieste al checker.",
		},
	)
	cacheHits = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_hits",
			Help: "Numero di richieste soddisfatte dalla cache.",
		},
	)
	cacheMisses = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_misses",
			Help: "Numero di richieste che non hanno trovato risultati nella cache.",
		},
	)
	responseTimes = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "response_times",
			Help:    "Distribuzione dei tempi di risposta.",
			Buckets: prometheus.DefBuckets,
		},
	)
)

// init registra le metriche Prometheus al momento dell'avvio.
func init() {
	prometheus.MustRegister(totalRequests, cacheHits, cacheMisses, responseTimes)
}

// Checker gestisce le query al database e la cache in memoria.
type Checker struct {
	db    database.Database
	cache *lru.Cache[string, []string]
	mu    sync.Mutex
}

// NewChecker crea un nuovo Checker con una cache LRU.
// Accetta un'istanza del database e la dimensione massima della cache in MB.
func NewChecker(db database.Database, cacheSizeMB int) (*Checker, error) {
	cacheSize := (cacheSizeMB * 1024 * 1024) / 1024 // Calcola il numero massimo di elementi nella cache
	cache, err := lru.New[string, []string](cacheSize)
	if err != nil {
		return nil, err
	}

	return &Checker{
		db:    db,
		cache: cache,
	}, nil
}

// FindEmailInBreaches cerca un'email nel database e utilizza la cache.
// Se l'email è presente nella cache, restituisce il risultato senza accedere al database.
// Aggiorna le metriche Prometheus per registrare le richieste, hit/miss della cache e i tempi di risposta.
func (c *Checker) FindEmailInBreaches(ctx context.Context, email string) ([]string, error) {
	totalRequests.Inc() // Incrementa il numero totale di richieste

	start := time.Now() // Inizia il timer per misurare il tempo di risposta

	// Verifica se l'email è già presente nella cache
	c.mu.Lock()
	if breaches, found := c.cache.Get(email); found {
		cacheHits.Inc() // Incrementa il contatore delle cache hit
		c.mu.Unlock()
		responseTimes.Observe(time.Since(start).Seconds()) // Registra il tempo di risposta
		return breaches, nil
	}
	cacheMisses.Inc() // Incrementa il contatore delle cache miss
	c.mu.Unlock()

	// Cerca l'email nel database
	breaches, err := c.db.FindEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	// Aggiungi il risultato alla cache
	c.mu.Lock()
	c.cache.Add(email, breaches)
	c.mu.Unlock()

	// Registra il tempo di risposta
	responseTimes.Observe(time.Since(start).Seconds())
	return breaches, nil
}
