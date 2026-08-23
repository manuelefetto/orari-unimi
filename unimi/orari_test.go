package unimi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

const rispostaOrariTest = `{
  "celle": [
    {"nome":"Sospensione", "data":"07-12-2026", "ora_inizio":"00:00", "ora_fine":"24:00"},
    {
      "id":"2", "codice_insegnamento":"EC2", "nome_insegnamento":"Reti",
      "codice_docente":"20", "docente":"ROSSI MARIO", "mail_docente":", mario@unimi.it",
      "data":"16-09-2026", "ora_inizio":"14:30", "ora_fine":"16:30",
      "codice_aula":"S#A2", "aula":"Aula 2 [Sede]", "codice_sede":"S",
      "percorso_didattico":"Informatica - 2 anno", "tipo":"Lezione", "Annullato":"1"
    },
    {
      "id":"1", "codice_insegnamento":"EC1", "nome_insegnamento":"Algoritmi",
      "codice_docente":"10", "docente":"BIANCHI LUCA", "data":"15-09-2026",
      "ora_inizio":"10:30", "ora_fine":"12:30", "aula":"Aula 1",
      "tipo":"Laboratorio", "notes":"Portare il PC"
    }
  ]
}`

func TestRecuperaOrariInsegnamentoInteroAnno(t *testing.T) {
	client, richieste := nuovoClientOrariTest(t)
	lezioni, err := client.RecuperaOrariInsegnamento(context.Background(), "2026", "EC1", IntervalloDate{})
	if err != nil {
		t.Fatal(err)
	}
	if len(lezioni) != 2 || lezioni[0].ID != "1" || lezioni[1].ID != "2" {
		t.Fatalf("lezioni inattese: %#v", lezioni)
	}
	if lezioni[1].EmailDocente != "mario@unimi.it" || !lezioni[1].Annullata {
		t.Fatalf("campi lezione inattesi: %#v", lezioni[1])
	}
	form := <-richieste
	if form.Get("include") != "attivita" || form.Get("attivita[]") != "EC1" || form.Get("all_events") != "1" {
		t.Fatalf("parametri inattesi: %v", form)
	}
	if form.Get("date") != "01-08-2026" {
		t.Fatalf("data di riferimento inattesa: %s", form.Get("date"))
	}
}

func TestRecuperaOrariDocenteInIntervallo(t *testing.T) {
	client, richieste := nuovoClientOrariTest(t)
	intervallo, err := NuovoIntervalloDate(
		time.Date(2026, 9, 16, 18, 0, 0, 0, time.Local),
		time.Date(2026, 9, 16, 18, 0, 0, 0, time.Local),
	)
	if err != nil {
		t.Fatal(err)
	}
	lezioni, err := client.RecuperaOrariDocente(context.Background(), "2026", "20", intervallo)
	if err != nil {
		t.Fatal(err)
	}
	if len(lezioni) != 1 || lezioni[0].ID != "2" {
		t.Fatalf("filtro intervallo inatteso: %#v", lezioni)
	}
	form := <-richieste
	if form.Get("include") != "docente" || form.Get("docente") != "20" || form.Get("date") != "16-09-2026" {
		t.Fatalf("parametri inattesi: %v", form)
	}
}

func TestRecuperaOrariCorsoInviaTuttiIPercorsi(t *testing.T) {
	var formOrari url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/combo.php":
			_, _ = w.Write([]byte(`var elenco_corsi = [{"label":"Informatica","valore":"F1X","elenco_anni":[{"label":"1","valore":"F1X-A|1"},{"label":"2","valore":"F1X-B|2"}]}];`))
		case "/grid_call.php":
			_ = r.ParseForm()
			formOrari = r.PostForm
			_, _ = w.Write([]byte(rispostaOrariTest))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	client, _ := NuovoClientConHTTP(server.URL+"/", server.Client())

	lezioni, err := client.RecuperaOrariCorso(context.Background(), "2026", "F1X", IntervalloDate{})
	if err != nil || len(lezioni) != 2 {
		t.Fatalf("lezioni=%#v, errore=%v", lezioni, err)
	}
	percorsi := formOrari["anno2[]"]
	if len(percorsi) != 2 || percorsi[0] != "F1X-A|1" || percorsi[1] != "F1X-B|2" {
		t.Fatalf("percorsi inattesi: %#v", percorsi)
	}
}

func TestIntervalloDateNonValido(t *testing.T) {
	_, err := NuovoIntervalloDate(time.Now(), time.Time{})
	if err == nil || !strings.Contains(err.Error(), "sia la data") {
		t.Fatalf("errore inatteso: %v", err)
	}
	_, err = NuovoIntervalloDate(time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC), time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC))
	if err == nil || !strings.Contains(err.Error(), "precedere") {
		t.Fatalf("errore inatteso: %v", err)
	}
}

func TestRispostaOrariNonValida(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"celle":[{"id":"x","codice_insegnamento":"EC","data":"non-data"}]}`))
	}))
	defer server.Close()
	client, _ := NuovoClientConHTTP(server.URL+"/", server.Client())
	_, err := client.RecuperaOrariDocente(context.Background(), "2026", "1", IntervalloDate{})
	if err == nil || !strings.Contains(err.Error(), "non valida") {
		t.Fatalf("errore inatteso: %v", err)
	}
}

func nuovoClientOrariTest(t *testing.T) (*Client, <-chan url.Values) {
	t.Helper()
	richieste := make(chan url.Values, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/grid_call.php" || r.Method != http.MethodPost {
			t.Fatalf("richiesta inattesa: %s %s", r.Method, r.URL.Path)
		}
		_ = r.ParseForm()
		richieste <- r.PostForm
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(rispostaOrariTest))
	}))
	t.Cleanup(server.Close)
	client, err := NuovoClientConHTTP(server.URL+"/", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	return client, richieste
}
