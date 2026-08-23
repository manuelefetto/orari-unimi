package tui

import (
	"regexp"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"orari-unimi/unimi"
)

var sequenzaColoreANSI = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func TestRenderCalendarioRispettaDimensioniFinestra(t *testing.T) {
	settimana := time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC)
	lezioni := []unimi.Lezione{
		{ID: "1", Data: settimana, OraInizio: "08:30", OraFine: "10:30", Insegnamento: "Programmazione", Aula: "Aula Alfa"},
		{ID: "2", Data: settimana, OraInizio: "10:30", OraFine: "12:30", Insegnamento: "Algoritmi e strutture dati", Aula: "Aula Beta"},
		{ID: "3", Data: settimana, OraInizio: "14:30", OraFine: "16:30", Insegnamento: "Basi di dati", Aula: "Aula Gamma"},
		{ID: "4", Data: settimana.AddDate(0, 0, 2), OraInizio: "09:30", OraFine: "11:30", Insegnamento: "Reti", Annullata: true},
	}

	verificaDimensioniRender(t, renderCalendario("Informatica", lezioni, settimana, 140, 16), 140, 16)
	verificaDimensioniRender(t, renderCalendario("Informatica", lezioni, settimana, 80, 12), 80, 12)
}

func TestRenderCalendarioUsaLaGrigliaInUnTerminaleStandard(t *testing.T) {
	settimana := time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC)
	lezioni := []unimi.Lezione{{
		ID: "1", Data: settimana, OraInizio: "08:30", OraFine: "10:30", Insegnamento: "Programmazione", Aula: "Aula Alfa",
	}}

	griglia := renderCalendario("Informatica", lezioni, settimana, 80, 24)
	if !strings.Contains(griglia, "+----------+") || !strings.Contains(griglia, "Lun 14/09") || !strings.Contains(griglia, "Dom 20/09") {
		t.Fatalf("griglia settimanale inattesa:\n%s", griglia)
	}
	if !strings.Contains(griglia, "08:30") || !strings.Contains(griglia, "|Lun 14/09 ") {
		t.Fatalf("le lezioni non sono disposte nelle celle del calendario:\n%s", griglia)
	}
	compatto := renderCalendario("Informatica", lezioni, settimana, 60, 18)
	if strings.Contains(compatto, "+----------+") || !strings.Contains(compatto, "Lun 14/09 |") || !strings.Contains(compatto, "Dom 20/09 | -") {
		t.Fatalf("agenda compatta inattesa:\n%s", compatto)
	}
}

func TestSettimanaInizialeSceglieLaLezionePiuVicina(t *testing.T) {
	lezioni := []unimi.Lezione{
		{Data: time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC)},
		{Data: time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)},
	}
	riferimento := time.Date(2026, 8, 23, 15, 0, 0, 0, time.Local)
	ottenuta := settimanaIniziale(lezioni, riferimento)
	attesa := time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC)
	if !ottenuta.Equal(attesa) {
		t.Fatalf("settimana iniziale = %s, attesa %s", ottenuta, attesa)
	}
}

func TestOrdinaEDeduplicaLezioni(t *testing.T) {
	giorno := time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC)
	lezioni := []unimi.Lezione{
		{ID: "2", Data: giorno, OraInizio: "10:30"},
		{ID: "1", Data: giorno, OraInizio: "08:30"},
		{ID: "1", Data: giorno, OraInizio: "08:30"},
	}
	ottenute := ordinaEDeduplica(lezioni)
	if len(ottenute) != 2 || ottenute[0].ID != "1" || ottenute[1].ID != "2" {
		t.Fatalf("lezioni ordinate inattese: %#v", ottenute)
	}
}

func TestNomeInsegnamentoVaACapoEdHaUnColoreStabile(t *testing.T) {
	righe := mandaACapo("Ingegneria del software", 10)
	attese := []string{"Ingegneria", "del", "software"}
	if strings.Join(righe, "|") != strings.Join(attese, "|") {
		t.Fatalf("a capo inatteso: %#v", righe)
	}

	primo := coloreInsegnamento("Ingegneria del software")
	secondo := coloreInsegnamento("Ingegneria del software")
	if primo == 0 || primo != secondo {
		t.Fatalf("colore non stabile: %d, %d", primo, secondo)
	}
}

func verificaDimensioniRender(t *testing.T, render string, larghezza, altezza int) {
	t.Helper()
	righe := strings.Split(strings.TrimSuffix(render, "\n"), "\n")
	if len(righe) > altezza {
		t.Fatalf("il render usa %d righe, massimo %d:\n%s", len(righe), altezza, render)
	}
	for indice, riga := range righe {
		rigaVisibile := sequenzaColoreANSI.ReplaceAllString(riga, "")
		if utf8.RuneCountInString(rigaVisibile) > larghezza {
			t.Fatalf("riga %d larga %d caratteri, massimo %d: %q", indice+1, utf8.RuneCountInString(rigaVisibile), larghezza, rigaVisibile)
		}
	}
}
