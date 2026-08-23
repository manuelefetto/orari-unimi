package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"orari-unimi/unimi"
)

func main() {
	tipo := flag.String("tipo", "anni", "lista da recuperare: anni, corsi, facolta, docenti o insegnamenti")
	anno := flag.String("anno", "", "codice dell'anno accademico, ad esempio 2026; se omesso usa il più recente")
	flag.Parse()

	client := unimi.NuovoClient()
	ctx := context.Background()
	if *tipo == "anni" {
		anni, err := client.RecuperaAnniAccademici(ctx)
		terminaSeErrore(err)
		for _, valore := range anni {
			fmt.Printf("%s\t%s\n", valore.Codice, valore.Nome)
		}
		return
	}

	annoSelezionato := *anno
	if annoSelezionato == "" {
		anni, err := client.RecuperaAnniAccademici(ctx)
		terminaSeErrore(err)
		if len(anni) == 0 {
			log.Fatal("il portale non ha restituito anni accademici")
		}
		annoSelezionato = anni[0].Codice
	}

	switch *tipo {
	case "corsi":
		valori, err := client.RecuperaCorsi(ctx, annoSelezionato)
		terminaSeErrore(err)
		for _, valore := range valori {
			fmt.Printf("%s\t%s\n", valore.Codice, valore.Nome)
		}
	case "facolta":
		valori, err := client.RecuperaFacolta(ctx, annoSelezionato)
		terminaSeErrore(err)
		for _, valore := range valori {
			fmt.Printf("%s\t%s\n", valore.Codice, valore.Nome)
		}
	case "docenti":
		valori, err := client.RecuperaDocenti(ctx, annoSelezionato)
		terminaSeErrore(err)
		for _, valore := range valori {
			fmt.Printf("%s\t%s\n", valore.Codice, valore.Nome)
		}
	case "insegnamenti":
		valori, err := client.RecuperaInsegnamenti(ctx, annoSelezionato)
		terminaSeErrore(err)
		for _, valore := range valori {
			fmt.Printf("%s\t%s\n", valore.Codice, valore.Descrizione)
		}
	default:
		fmt.Fprintf(os.Stderr, "tipo %q non valido\n", *tipo)
		flag.Usage()
		os.Exit(2)
	}
}

func terminaSeErrore(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
