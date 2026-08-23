package tui

import (
	"path/filepath"
	"testing"
)

func TestArchivioInsegnamenti(t *testing.T) {
	archivio, err := NuovoArchivioInsegnamenti(filepath.Join(t.TempDir(), "dati", "insegnamenti.json"))
	if err != nil {
		t.Fatal(err)
	}
	insegnamento := InsegnamentoSalvato{Anno: "2026", Codice: "ECF1X-1", Nome: "Algoritmi"}

	aggiunto, err := archivio.Aggiungi(insegnamento)
	if err != nil || !aggiunto {
		t.Fatalf("prima aggiunta: aggiunto=%v, errore=%v", aggiunto, err)
	}
	aggiunto, err = archivio.Aggiungi(insegnamento)
	if err != nil || aggiunto {
		t.Fatalf("duplicato: aggiunto=%v, errore=%v", aggiunto, err)
	}
	insegnamenti, err := archivio.Elenca()
	if err != nil || len(insegnamenti) != 1 || insegnamenti[0] != insegnamento {
		t.Fatalf("insegnamenti inattesi: %#v, errore=%v", insegnamenti, err)
	}
	if err := archivio.Rimuovi(0); err != nil {
		t.Fatal(err)
	}
	insegnamenti, err = archivio.Elenca()
	if err != nil || len(insegnamenti) != 0 {
		t.Fatalf("archivio non vuoto: %#v, errore=%v", insegnamenti, err)
	}
}

func TestArchivioPulisci(t *testing.T) {
	archivio, _ := NuovoArchivioInsegnamenti(filepath.Join(t.TempDir(), "insegnamenti.json"))
	_, _ = archivio.Aggiungi(InsegnamentoSalvato{Anno: "2026", Codice: "A", Nome: "A"})
	if err := archivio.Pulisci(); err != nil {
		t.Fatal(err)
	}
	insegnamenti, err := archivio.Elenca()
	if err != nil || len(insegnamenti) != 0 {
		t.Fatalf("archivio non pulito: %#v, errore=%v", insegnamenti, err)
	}
}
