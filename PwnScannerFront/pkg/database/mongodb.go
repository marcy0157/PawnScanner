package database

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB rappresenta l'implementazione del database per MongoDB.
type MongoDB struct {
	client     *mongo.Client
	collection *mongo.Collection
}

// NewMongoDB crea una nuova connessione a MongoDB.
// Accetta parametri come host, porta, credenziali di autenticazione, nome del database e della collezione.
func NewMongoDB(ctx context.Context, host string, port int, username, password, dbName, collectionName string) (Database, error) {
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d", username, password, host, port)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("errore durante la connessione a MongoDB: %w", err)
	}

	collection := client.Database(dbName).Collection(collectionName)
	return &MongoDB{
		client:     client,
		collection: collection,
	}, nil
}

// FindEmail cerca un'email nei breach.
// Restituisce un elenco di breach associati all'email specificata.
func (db *MongoDB) FindEmail(ctx context.Context, email string) ([]string, error) {
	var result struct {
		Breaches []string `bson:"breaches"`
	}

	filter := bson.M{"email": email}
	err := db.collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return result.Breaches, nil
}

// Close chiude la connessione a MongoDB.
// Deve essere chiamata per liberare le risorse.
func (db *MongoDB) Close() error {
	return db.client.Disconnect(context.Background())
}

// GetAllBreaches restituisce un elenco di tutti i breach senza duplicati.
// Esegue una query sulla collezione per raccogliere tutti i breach.
func (db *MongoDB) GetAllBreaches(ctx context.Context) ([]string, error) {
	var breaches []string
	cursor, err := db.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var doc struct {
			Breaches []string `bson:"breaches"`
		}
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		breaches = append(breaches, doc.Breaches...)
	}

	// Rimuovi duplicati
	uniqueBreachSet := make(map[string]struct{})
	for _, breach := range breaches {
		uniqueBreachSet[breach] = struct{}{}
	}

	var uniqueBreaches []string
	for breach := range uniqueBreachSet {
		uniqueBreaches = append(uniqueBreaches, breach)
	}

	return uniqueBreaches, nil
}
