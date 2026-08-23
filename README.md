# orari-unimi

Applicazione Go per consultare da terminale gli orari delle lezioni UNIMI.

## Avvio

```sh
go run ./cmd/orari-unimi
```

I menu e le scelte interattive usano
[ManuCli](https://github.com/manuelefetto/ManuCli): usa le frecce verticali per
spostarti e `Invio` per confermare. È possibile cercare corsi, docenti e singoli
insegnamenti attraverso suggerimenti ottenuti dal portale UNIMI, oltre a gestire
i corsi salvati nella sezione "I miei orari".
La ricerca non distingue tra maiuscole, minuscole e vocali accentate.

La selezione di un corso, docente o insegnamento recupera gli orari effettivi
pubblicati da UNIMI e ne mostra un riepilogo. Il package `unimi` espone inoltre
le funzioni `RecuperaOrariCorso`, `RecuperaOrariDocente` e
`RecuperaOrariInsegnamento`: passando un `IntervalloDate` vuoto si ottiene
l'intero anno accademico, mentre un intervallo valorizzato filtra entrambe le
date incluse.

I corsi personali vengono salvati in `orari-unimi/corsi.json` all'interno della
cartella di configurazione dell'utente. Il percorso può essere personalizzato
impostando la variabile d'ambiente `ORARI_UNIMI_FILE`.

## Recupero delle liste

- anni accademici;
- corsi di studio e facoltà/scuole;
- docenti;
- insegnamenti.

## Comandi diagnostici per l'ottenimento delle liste

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
