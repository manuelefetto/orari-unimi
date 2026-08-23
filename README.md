# orari-unimi

Applicazione Go per consultare da terminale gli orari delle lezioni UNIMI.

## Fase 1: recupero delle liste

Il package `unimi` usa esclusivamente la libreria standard e permette di
recuperare dal portale pubblico:

- anni accademici;
- corsi di studio e facoltà/scuole;
- docenti;
- insegnamenti.

Tutte le richieste accettano un `context.Context`. Il comando provvisorio della
fase 1 consente di controllare i risultati prima della realizzazione della TUI:

```sh
go run ./cmd/orari-unimi -tipo anni
go run ./cmd/orari-unimi -tipo corsi -anno 2026
go run ./cmd/orari-unimi -tipo docenti -anno 2026
go run ./cmd/orari-unimi -tipo insegnamenti -anno 2026
```

Se `-anno` non viene indicato, viene selezionato automaticamente l'anno più
recente pubblicato dal portale.

Per eseguire i test:

```sh
go test ./...
```
