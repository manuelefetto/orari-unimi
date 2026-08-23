package tui

import (
	"sort"
	"strings"
)

type voceRicerca struct {
	Codice string
	Nome   string
}

func filtraVoci(voci []voceRicerca, ricerca string, limite int) []voceRicerca {
	parole := strings.Fields(normalizza(ricerca))
	if len(parole) == 0 || limite <= 0 {
		return nil
	}

	type risultato struct {
		voce      voceRicerca
		posizione int
	}
	risultati := make([]risultato, 0)
	for _, voce := range voci {
		testo := normalizza(voce.Codice + " " + voce.Nome)
		posizione := len(testo)
		presente := true
		for _, parola := range parole {
			indice := strings.Index(testo, parola)
			if indice < 0 {
				presente = false
				break
			}
			if indice < posizione {
				posizione = indice
			}
		}
		if presente {
			risultati = append(risultati, risultato{voce: voce, posizione: posizione})
		}
	}

	sort.SliceStable(risultati, func(i, j int) bool {
		if risultati[i].posizione != risultati[j].posizione {
			return risultati[i].posizione < risultati[j].posizione
		}
		return normalizza(risultati[i].voce.Nome) < normalizza(risultati[j].voce.Nome)
	})
	if len(risultati) > limite {
		risultati = risultati[:limite]
	}

	vociFiltrate := make([]voceRicerca, len(risultati))
	for i := range risultati {
		vociFiltrate[i] = risultati[i].voce
	}
	return vociFiltrate
}

func normalizza(testo string) string {
	testo = strings.ToLower(strings.TrimSpace(testo))
	sostituzioni := strings.NewReplacer(
		"à", "a", "á", "a", "è", "e", "é", "e", "ì", "i", "í", "i",
		"ò", "o", "ó", "o", "ù", "u", "ú", "u",
	)
	return sostituzioni.Replace(testo)
}
