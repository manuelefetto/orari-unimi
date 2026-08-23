package tui

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"orari-unimi/unimi"
)

type sorgenteFinta struct {
	anni         []unimi.AnnoAccademico
	corsi        []unimi.CorsoDiStudio
	docenti      []unimi.Docente
	insegnamenti []unimi.Insegnamento
}

func (s *sorgenteFinta) RecuperaAnniAccademici(context.Context) ([]unimi.AnnoAccademico, error) {
	return s.anni, nil
}

func (s *sorgenteFinta) RecuperaCorsi(context.Context, string) ([]unimi.CorsoDiStudio, error) {
	return s.corsi, nil
}

func (s *sorgenteFinta) RecuperaDocenti(context.Context, string) ([]unimi.Docente, error) {
	return s.docenti, nil
}

func (s *sorgenteFinta) RecuperaInsegnamenti(context.Context, string) ([]unimi.Insegnamento, error) {
	return s.insegnamenti, nil
}

type selettoreFinto struct {
	scelte    []int
	posizione int
}

type lettoreFinto struct {
	risposte  []string
	posizione int
}

func (l *lettoreFinto) Leggi(string) (string, error) {
	risposta := l.risposte[l.posizione]
	l.posizione++
	return risposta, nil
}

func (s *selettoreFinto) Scegli(_ string, opzioni []string) (int, error) {
	scelta := s.scelte[s.posizione]
	s.posizione++
	if scelta < 0 || scelta >= len(opzioni) {
		panic("scelta del test fuori dalle opzioni")
	}
	return scelta, nil
}

func TestApplicazioneSelezionaCorso(t *testing.T) {
	sorgente := &sorgenteFinta{
		anni:  []unimi.AnnoAccademico{{Codice: "2026", Nome: "2026/2027"}},
		corsi: []unimi.CorsoDiStudio{{Codice: "F1X", Nome: "Informatica"}},
	}
	archivio, _ := NuovoArchivioCorsi(filepath.Join(t.TempDir(), "corsi.json"))
	lettore := &lettoreFinto{risposte: []string{"info"}}
	var uscita bytes.Buffer
	selettore := &selettoreFinto{scelte: []int{0, 0, 4}}
	app, err := NuovaApplicazione(lettore, &uscita, sorgente, archivio, selettore)
	if err != nil {
		t.Fatal(err)
	}
	if err := app.Esegui(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(uscita.String(), "Corso selezionato: Informatica") {
		t.Fatalf("output inatteso:\n%s", uscita.String())
	}
}

func TestApplicazioneGestisceMieiCorsi(t *testing.T) {
	sorgente := &sorgenteFinta{
		anni:  []unimi.AnnoAccademico{{Codice: "2026", Nome: "2026/2027"}},
		corsi: []unimi.CorsoDiStudio{{Codice: "F1X", Nome: "Informatica"}},
	}
	archivio, _ := NuovoArchivioCorsi(filepath.Join(t.TempDir(), "corsi.json"))
	// I miei orari -> aggiungi -> cerca -> seleziona -> pulisci -> conferma -> indietro -> esci.
	lettore := &lettoreFinto{risposte: []string{"info"}}
	var uscita bytes.Buffer
	selettore := &selettoreFinto{scelte: []int{3, 1, 0, 3, 1, 4, 4}}
	app, _ := NuovaApplicazione(lettore, &uscita, sorgente, archivio, selettore)
	if err := app.Esegui(context.Background()); err != nil {
		t.Fatal(err)
	}
	corsi, err := archivio.Elenca()
	if err != nil || len(corsi) != 0 {
		t.Fatalf("archivio inatteso: %#v, errore=%v", corsi, err)
	}
	if !strings.Contains(uscita.String(), "Tutti i corsi sono stati rimossi") {
		t.Fatalf("output inatteso:\n%s", uscita.String())
	}
}

func TestApplicazioneSelezionaInsegnamento(t *testing.T) {
	sorgente := &sorgenteFinta{
		anni: []unimi.AnnoAccademico{{Codice: "2026", Nome: "2026/2027"}},
		insegnamenti: []unimi.Insegnamento{{
			Codice: "ECF1X-1", Descrizione: "Algoritmi [M. ROSSI]",
		}},
	}
	archivio, _ := NuovoArchivioCorsi(filepath.Join(t.TempDir(), "corsi.json"))
	selettore := &selettoreFinto{scelte: []int{2, 0, 4}}
	var uscita bytes.Buffer
	lettore := &lettoreFinto{risposte: []string{"algoritmi"}}
	app, _ := NuovaApplicazione(lettore, &uscita, sorgente, archivio, selettore)
	if err := app.Esegui(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(uscita.String(), "Insegnamento selezionato: Algoritmi") {
		t.Fatalf("output inatteso:\n%s", uscita.String())
	}
}
