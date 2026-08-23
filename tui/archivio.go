package tui

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// CorsoSalvato contiene i dati minimi necessari per ritrovare un corso.
type CorsoSalvato struct {
	Anno   string `json:"anno"`
	Codice string `json:"codice"`
	Nome   string `json:"nome"`
}

// ArchivioCorsi salva la sezione "I miei orari" in un semplice file JSON.
type ArchivioCorsi struct {
	percorso string
}

func NuovoArchivioCorsi(percorso string) (*ArchivioCorsi, error) {
	if percorso == "" {
		return nil, errors.New("il percorso dell'archivio è obbligatorio")
	}
	return &ArchivioCorsi{percorso: percorso}, nil
}

func (a *ArchivioCorsi) Elenca() ([]CorsoSalvato, error) {
	dati, err := os.ReadFile(a.percorso)
	if errors.Is(err, os.ErrNotExist) {
		return []CorsoSalvato{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("lettura corsi salvati: %w", err)
	}
	if len(dati) == 0 {
		return []CorsoSalvato{}, nil
	}

	var corsi []CorsoSalvato
	if err := json.Unmarshal(dati, &corsi); err != nil {
		return nil, fmt.Errorf("decodifica corsi salvati: %w", err)
	}
	return corsi, nil
}

func (a *ArchivioCorsi) Aggiungi(corso CorsoSalvato) (bool, error) {
	corsi, err := a.Elenca()
	if err != nil {
		return false, err
	}
	for _, salvato := range corsi {
		if salvato.Anno == corso.Anno && salvato.Codice == corso.Codice {
			return false, nil
		}
	}
	corsi = append(corsi, corso)
	return true, a.salva(corsi)
}

func (a *ArchivioCorsi) Rimuovi(indice int) error {
	corsi, err := a.Elenca()
	if err != nil {
		return err
	}
	if indice < 0 || indice >= len(corsi) {
		return errors.New("corso da rimuovere non valido")
	}
	corsi = append(corsi[:indice], corsi[indice+1:]...)
	return a.salva(corsi)
}

func (a *ArchivioCorsi) Pulisci() error {
	return a.salva([]CorsoSalvato{})
}

func (a *ArchivioCorsi) salva(corsi []CorsoSalvato) error {
	if err := os.MkdirAll(filepath.Dir(a.percorso), 0o700); err != nil {
		return fmt.Errorf("creazione cartella dati: %w", err)
	}
	dati, err := json.MarshalIndent(corsi, "", "  ")
	if err != nil {
		return fmt.Errorf("codifica corsi salvati: %w", err)
	}
	dati = append(dati, '\n')
	if err := os.WriteFile(a.percorso, dati, 0o600); err != nil {
		return fmt.Errorf("salvataggio corsi: %w", err)
	}
	return nil
}
