package tui

import (
	"path/filepath"
	"testing"
)

func TestArchivioCorsi(t *testing.T) {
	archivio, err := NuovoArchivioCorsi(filepath.Join(t.TempDir(), "dati", "corsi.json"))
	if err != nil {
		t.Fatal(err)
	}
	corso := CorsoSalvato{Anno: "2026", Codice: "F1X", Nome: "Informatica"}

	aggiunto, err := archivio.Aggiungi(corso)
	if err != nil || !aggiunto {
		t.Fatalf("prima aggiunta: aggiunto=%v, errore=%v", aggiunto, err)
	}
	aggiunto, err = archivio.Aggiungi(corso)
	if err != nil || aggiunto {
		t.Fatalf("duplicato: aggiunto=%v, errore=%v", aggiunto, err)
	}
	corsi, err := archivio.Elenca()
	if err != nil || len(corsi) != 1 || corsi[0] != corso {
		t.Fatalf("corsi inattesi: %#v, errore=%v", corsi, err)
	}
	if err := archivio.Rimuovi(0); err != nil {
		t.Fatal(err)
	}
	corsi, err = archivio.Elenca()
	if err != nil || len(corsi) != 0 {
		t.Fatalf("archivio non vuoto: %#v, errore=%v", corsi, err)
	}
}

func TestArchivioPulisci(t *testing.T) {
	archivio, _ := NuovoArchivioCorsi(filepath.Join(t.TempDir(), "corsi.json"))
	_, _ = archivio.Aggiungi(CorsoSalvato{Anno: "2026", Codice: "A", Nome: "A"})
	if err := archivio.Pulisci(); err != nil {
		t.Fatal(err)
	}
	corsi, err := archivio.Elenca()
	if err != nil || len(corsi) != 0 {
		t.Fatalf("archivio non pulito: %#v, errore=%v", corsi, err)
	}
}
