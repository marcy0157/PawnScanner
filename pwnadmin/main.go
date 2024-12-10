package main

import (
	"context"
	"extract/extractor"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
)

var (
	mongoClient   *mongo.Client
	dbName        string
	adminUsername string
	adminPassword string
)

func main() {
	// Leggi le credenziali admin dalle variabili d'ambiente
	adminUsername = os.Getenv("ADMIN_USERNAME")
	adminPassword = os.Getenv("ADMIN_PASSWORD")
	if adminUsername == "" || adminPassword == "" {
		log.Fatal("Le credenziali admin (ADMIN_USERNAME e ADMIN_PASSWORD) devono essere definite nelle variabili d'ambiente")
	}

	// Configura la connessione a MongoDB
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("La variabile d'ambiente MONGODB_URI non è impostata")
	}
	dbName = os.Getenv("MONGODB_DBNAME")
	if dbName == "" {
		dbName = "extract"
	}

	clientOptions := options.Client().ApplyURI(mongoURI)
	var err error
	mongoClient, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatalf("Errore nella connessione a MongoDB: %v", err)
	}

	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			log.Fatalf("Errore durante la disconnessione da MongoDB: %v", err)
		}
	}()

	// Configura gli handler HTTP
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/", authMiddleware(indexHandler))
	http.HandleFunc("/upload", authMiddleware(uploadHandler))

	fmt.Println("Il server è in esecuzione sulla porta 8081...")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

// Renderizza un template HTML
func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	t, err := template.ParseFiles(fmt.Sprintf("templates/%s.html", tmpl))
	if err != nil {
		http.Error(w, "Errore nel caricamento del template", http.StatusInternalServerError)
		log.Printf("Errore nel caricamento del template %s: %v", tmpl, err)
		return
	}
	t.Execute(w, data)
}

// Handler per il login
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		renderTemplate(w, "login", nil)
		return
	}

	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		if username == adminUsername && password == adminPassword {
			http.SetCookie(w, &http.Cookie{
				Name:    "session_token",
				Value:   "authenticated",
				Expires: time.Now().Add(1 * time.Hour),
			})
			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			http.Error(w, "Credenziali non valide", http.StatusUnauthorized)
		}
		return
	}

	http.Error(w, "Metodo non consentito", http.StatusMethodNotAllowed)
}

// Middleware per autenticazione
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil || cookie.Value != "authenticated" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	}
}

// Handler per la pagina principale
func indexHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index", nil)
}

// Handler per l'upload

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Metodo non consentito", http.StatusMethodNotAllowed)
		return
	}

	breachName := r.FormValue("breachName")
	if breachName == "" {
		http.Error(w, "Il nome del breach è richiesto", http.StatusBadRequest)
		return
	}
	collectionName := "breaches" // Nome fisso della collezione

	// Analizza il form multipart
	err := r.ParseMultipartForm(0)
	if err != nil {
		http.Error(w, "Errore durante l'analisi dei dati del form", http.StatusInternalServerError)
		log.Printf("Errore durante l'analisi dei dati del form: %v", err)
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		http.Error(w, "Nessun file caricato", http.StatusBadRequest)
		log.Println("Nessun file caricato.")
		return
	}

	var filePaths []string
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Errore nell'apertura del file caricato", http.StatusInternalServerError)
			log.Printf("Errore nell'apertura del file caricato %s: %v", fileHeader.Filename, err)
			continue
		}
		defer file.Close()

		// Salva il file caricato in una directory temporanea
		tempDir := os.TempDir()
		// Mantieni la struttura delle sottocartelle
		tempFilePath := filepath.Join(tempDir, fileHeader.Filename)
		os.MkdirAll(filepath.Dir(tempFilePath), os.ModePerm)
		tempFile, err := os.Create(tempFilePath)
		if err != nil {
			http.Error(w, "Errore nella creazione del file temporaneo", http.StatusInternalServerError)
			log.Printf("Errore nella creazione del file temporaneo %s: %v", tempFilePath, err)
			continue
		}
		defer tempFile.Close()

		_, err = tempFile.ReadFrom(file)
		if err != nil {
			http.Error(w, "Errore nel salvataggio del file caricato", http.StatusInternalServerError)
			log.Printf("Errore nel salvataggio del file caricato %s: %v", tempFilePath, err)
			continue
		}

		// Verifica la dimensione del file
		fileInfo, err := tempFile.Stat()
		if err != nil {
			log.Printf("Errore nel recupero delle informazioni del file %s: %v", tempFilePath, err)
			continue
		}
		if fileInfo.Size() == 0 {
			log.Printf("Il file %s è vuoto e verrà ignorato.", tempFilePath)
			continue
		}

		log.Printf("File caricato salvato in: %s", tempFilePath)
		filePaths = append(filePaths, tempFilePath)
	}

	ctx := context.Background()

	// Estrae le email dai file e le carica nel database con progressione
	totalFiles := len(filePaths)
	var filesProcessed int32 = 0

	// Processa i file uno alla volta
	for _, filePath := range filePaths {
		log.Printf("Inizio estrazione email dal file: %s", filePath)
		emails, err := extractor.ExtractEmailsFromFile(filePath)
		if err != nil {
			log.Printf("Errore durante l'estrazione dal file %s: %v", filePath, err)
			continue
		}

		log.Printf("Numero di email estratte dal file %s: %d", filePath, len(emails))

		if len(emails) > 0 {
			// Carica le email nel database
			err = uploadEmailsToMongo(ctx, mongoClient, dbName, collectionName, emails, breachName)
			if err != nil {
				log.Printf("Errore durante il caricamento delle email in MongoDB dal file %s: %v", filePath, err)
				continue
			}
			log.Printf("Email dal file %s caricate con successo.", filePath)
		} else {
			log.Printf("Nessuna email valida trovata nel file %s.", filePath)
		}

		// Aggiorna il conteggio dei file processati
		atomic.AddInt32(&filesProcessed, 1)
	}

	// Calcola la percentuale di avanzamento
	progress := (float64(filesProcessed) / float64(totalFiles)) * 100

	// Mostra il risultato all'utente
	fmt.Fprintf(w, "Caricamento completato! Percentuale di avanzamento: %.2f%%", progress)
	log.Printf("Processamento completato. Files processati: %d su %d", filesProcessed, totalFiles)
}

func uploadEmailsToMongo(ctx context.Context, client *mongo.Client, dbName, collectionName string, emails []string, breachName string) error {
	collection := client.Database(dbName).Collection(collectionName)

	// Rimuovi le email duplicate
	emailSet := make(map[string]struct{})
	for _, email := range emails {
		emailSet[email] = struct{}{}
	}
	uniqueEmails := make([]string, 0, len(emailSet))
	for email := range emailSet {
		uniqueEmails = append(uniqueEmails, email)
	}

	log.Printf("Numero di email uniche da processare: %d", len(uniqueEmails))

	var models []mongo.WriteModel
	for _, email := range uniqueEmails {
		// Crea un modello di aggiornamento con upsert
		filter := bson.M{"email": email}
		update := bson.M{
			"$addToSet": bson.M{
				"breaches": breachName,
			},
		}
		model := mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(true)
		models = append(models, model)
	}

	if len(models) > 0 {
		batchSize := 950 // Imposta il batchSize
		totalModels := len(models)
		for i := 0; i < totalModels; i += batchSize {
			end := i + batchSize
			if end > totalModels {
				end = totalModels
			}

			batch := models[i:end]
			log.Printf("Esecuzione di un batch di %d operazioni di upsert (da %d a %d).", len(batch), i, end)
			result, err := collection.BulkWrite(ctx, batch)
			if err != nil {
				log.Printf("Errore durante l'operazione BulkWrite: %v", err)
				return err
			}
			log.Printf("Risultati del BulkWrite: MatchedCount=%d, ModifiedCount=%d, UpsertedCount=%d", result.MatchedCount, result.ModifiedCount, result.UpsertedCount)
		}
		log.Printf("Operazioni di upsert completate per %d email nella collezione %s.", len(uniqueEmails), collectionName)

		// Conta il numero di documenti nella collezione
		count, err := collection.CountDocuments(ctx, bson.M{})
		if err != nil {
			log.Printf("Errore durante il conteggio dei documenti nella collezione %s: %v", collectionName, err)
		} else {
			log.Printf("La collezione %s contiene ora %d documenti.", collectionName, count)
		}
	} else {
		log.Println("Nessuna email da inserire nel database.")
	}

	return nil
}
