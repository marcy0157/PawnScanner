# PwnScanner
## Descrizione
Questo progetto fornisce un sistema per la gestione e la consultazione di dati relativi a data breach. È composto da due programmi principali:

1. **PwnScanner**: 
   - Un'interfaccia frontend che consente agli utenti di verificare se un'email è stata coinvolta in un data breach.
   - Si connette a un database MongoDB per interrogare i dati.
   - Mostra dove è stato rilevato il breach (es. Facebook, Twitter, ecc.).

2. **PwnAdmin**:
   - Un tool di amministrazione che consente di caricare i file dei data breach nel database MongoDB.
   - Utile per la manutenzione e l'aggiornamento dei dati.

---

## Struttura del Progetto

![](?raw=true)
---

## Come usare il progetto

### Prerequisiti
- Docker e Docker Compose installati.
- (Facoltativo) Un'istanza MongoDB configurata se si utilizza `composeNOMongo.yml`.

### Configurazione
1. Clona la repo:
   git clone <url-della-repo>
   cd <nome-della-repo>

2. Configura i file:
   - `composeMongo.yml`: Assicurati che le configurazioni corrispondano al tuo ambiente locale.
   - `composeNOMongo.yml`: Modifica i dati di connessione al database MongoDB esterno.
   - `config.yaml`: Specifica l'URL del database MongoDB (locale o remoto).

---

### Avvio dei container

#### Con MongoDB incluso
Usa il file `composeMongo.yml`:
   docker-compose -f composeMongo.yml up

#### Con MongoDB esterno
Usa il file `composeNOMongo.yml`:
   docker-compose -f composeNOMongo.yml up

#### Fermare i container
Per fermare i container, usa:
   docker-compose down

---

## Funzionalità principali

### PwnScanner (Frontend)
- Verifica se un'email è stata coinvolta in un data breach.
- Mostra i dettagli di ogni breach (es. servizio coinvolto).

### PwnAdmin (Admin Tool)
- Carica file di breach nel database MongoDB.
- Funzionalità per gestire i dati caricati.

---

## Note aggiuntive
- **Port**: Verifica le porte esposte nei file Docker Compose e assicurati che non siano già in uso.
- **Database**: Se utilizzi `composeNOMongo.yml`, assicurati che il database MongoDB sia correttamente configurato e accessibile.
