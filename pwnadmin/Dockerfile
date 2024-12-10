# Utilizza un'immagine base di Go per compilare l'applicazione
FROM golang:1.23.3 AS builder

# Imposta la directory di lavoro all'interno del container
WORKDIR /app

# Copia i file go.mod e go.sum
COPY go.mod go.sum ./

# Scarica le dipendenze
RUN go mod download

# Copia l'intero progetto nel container
COPY . .

# Compila l'applicazione
RUN go build -o main .

# Fase finale: immagine leggera per eseguire l'applicazione
FROM debian:bookworm-slim

# Installazione di CA Certificates per HTTPS
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

# Imposta la directory di lavoro
WORKDIR /app

# Copia il file binario compilato e le risorse necessarie dalla fase di build
COPY --from=builder /app/main .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/extractor ./extractor

# Esponi la porta su cui il server è in esecuzione
EXPOSE 8081

# Comando per avviare l'applicazione
CMD ["./main"]
