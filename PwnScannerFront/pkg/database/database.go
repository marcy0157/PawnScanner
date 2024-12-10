package database

import "context"

// Database è l'interfaccia per astrarre le operazioni sul database
// @Description Interfaccia che definisce le operazioni principali del database
type Database interface {
	// FindEmail cerca un'email nei breach
	FindEmail(ctx context.Context, email string) ([]string, error)

	// GetAllBreaches restituisce tutti i breach unici
	GetAllBreaches(ctx context.Context) ([]string, error)

	// Close chiude la connessione al database
	Close() error
}
