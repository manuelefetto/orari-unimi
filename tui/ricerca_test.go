package tui

import "testing"

func TestFiltraVociIgnoraMaiuscoleEAccenti(t *testing.T) {
	voci := []voceRicerca{
		{Codice: "A01", Nome: "Scienze politiche"},
		{Codice: "F1X", Nome: "Informatica"},
		{Codice: "C42", Nome: "Abilità informatiche"},
	}
	risultati := filtraVoci(voci, "abilita inform", 10)
	if len(risultati) != 1 || risultati[0].Codice != "C42" {
		t.Fatalf("risultati inattesi: %#v", risultati)
	}

	risultati = filtraVoci(voci, "f1x", 10)
	if len(risultati) != 1 || risultati[0].Nome != "Informatica" {
		t.Fatalf("ricerca per codice inattesa: %#v", risultati)
	}
}

func TestFiltraVociRispettaLimite(t *testing.T) {
	voci := []voceRicerca{{Nome: "Corso A"}, {Nome: "Corso B"}, {Nome: "Corso C"}}
	if risultati := filtraVoci(voci, "corso", 2); len(risultati) != 2 {
		t.Fatalf("numero risultati inatteso: %d", len(risultati))
	}
}
