package tui

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// InsegnamentoSalvato contiene i dati minimi necessari per ritrovare un
// singolo insegnamento.
type InsegnamentoSalvato struct {
	Anno   string `json:"anno"`
	Codice string `json:"codice"`
	Nome   string `json:"nome"`
}

// ArchivioInsegnamenti salva la sezione "I miei orari" in un file JSON.
type ArchivioInsegnamenti struct {
	percorso string
}

func NuovoArchivioInsegnamenti(percorso string) (*ArchivioInsegnamenti, error) {
	if percorso == "" {
		return nil, errors.New("il percorso dell'archivio è obbligatorio")
	}
	return &ArchivioInsegnamenti{percorso: percorso}, nil
}

func (a *ArchivioInsegnamenti) Elenca() ([]InsegnamentoSalvato, error) {
	dati, err := os.ReadFile(a.percorso)
	if errors.Is(err, os.ErrNotExist) {
		return []InsegnamentoSalvato{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("lettura insegnamenti salvati: %w", err)
	}
	if len(dati) == 0 {
		return []InsegnamentoSalvato{}, nil
	}

	var insegnamenti []InsegnamentoSalvato
	if err := json.Unmarshal(dati, &insegnamenti); err != nil {
		return nil, fmt.Errorf("decodifica insegnamenti salvati: %w", err)
	}
	return insegnamenti, nil
}

func (a *ArchivioInsegnamenti) Aggiungi(insegnamento InsegnamentoSalvato) (bool, error) {
	if insegnamento.Anno == "" || insegnamento.Codice == "" || insegnamento.Nome == "" {
		return false, errors.New("insegnamento da salvare non valido")
	}
	insegnamenti, err := a.Elenca()
	if err != nil {
		return false, err
	}
	for _, salvato := range insegnamenti {
		if salvato.Anno == insegnamento.Anno && salvato.Codice == insegnamento.Codice {
			return false, nil
		}
	}
	insegnamenti = append(insegnamenti, insegnamento)
	return true, a.salva(insegnamenti)
}

func (a *ArchivioInsegnamenti) Rimuovi(indice int) error {
	insegnamenti, err := a.Elenca()
	if err != nil {
		return err
	}
	if indice < 0 || indice >= len(insegnamenti) {
		return errors.New("insegnamento da rimuovere non valido")
	}
	insegnamenti = append(insegnamenti[:indice], insegnamenti[indice+1:]...)
	return a.salva(insegnamenti)
}

func (a *ArchivioInsegnamenti) Pulisci() error {
	return a.salva([]InsegnamentoSalvato{})
}

func (a *ArchivioInsegnamenti) salva(insegnamenti []InsegnamentoSalvato) error {
	if err := os.MkdirAll(filepath.Dir(a.percorso), 0o700); err != nil {
		return fmt.Errorf("creazione cartella dati: %w", err)
	}
	dati, err := json.MarshalIndent(insegnamenti, "", "  ")
	if err != nil {
		return fmt.Errorf("codifica insegnamenti salvati: %w", err)
	}
	dati = append(dati, '\n')
	if err := os.WriteFile(a.percorso, dati, 0o600); err != nil {
		return fmt.Errorf("salvataggio insegnamenti: %w", err)
	}
	return nil
}
