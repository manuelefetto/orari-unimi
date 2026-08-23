package unimi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRecuperaListe(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/combo.php" || r.URL.Query().Get("sw") != "ec_" {
			t.Fatalf("richiesta inattesa: %s", r.URL.String())
		}
		switch r.URL.Query().Get("page") {
		case "":
			_, _ = w.Write([]byte(`var anni_accademici_ec = {"2025":{"label":"2025/2026","valore":"2025"},"2026":{"label":"2026/2027","valore":"2026"}};`))
		case "corsi":
			_, _ = w.Write([]byte(`var elenco_corsi = [{"label":"Informatica","valore":"F1X","tipo":"Laurea","scuola":"sci"}]; var elenco_scuole = [{"label":"Scienze","valore":"sci"}];`))
		case "docenti":
			_, _ = w.Write([]byte(`var elenco_docenti = [{"label":"ROSSI MARIO","valore":"42"}];`))
		case "attivita":
			_, _ = w.Write([]byte(`var elenco_attivita = [{"label":"Algoritmi [M. ROSSI]","nome_insegnamento":"Algoritmi","valore":"EC1","docente":"M. ROSSI"}];`))
		default:
			t.Fatalf("pagina inattesa: %q", r.URL.Query().Get("page"))
		}
	}))
	defer server.Close()

	client, err := NuovoClientConHTTP(server.URL+"/", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	anni, err := client.RecuperaAnniAccademici(ctx)
	if err != nil || len(anni) != 2 || anni[0].Codice != "2026" {
		t.Fatalf("anni inattesi: %#v, errore: %v", anni, err)
	}
	corsi, err := client.RecuperaCorsi(ctx, "2026")
	if err != nil || len(corsi) != 1 || corsi[0].Codice != "F1X" {
		t.Fatalf("corsi inattesi: %#v, errore: %v", corsi, err)
	}
	facolta, err := client.RecuperaFacolta(ctx, "2026")
	if err != nil || len(facolta) != 1 || facolta[0].Codice != "sci" {
		t.Fatalf("facoltà inattese: %#v, errore: %v", facolta, err)
	}
	docenti, err := client.RecuperaDocenti(ctx, "2026")
	if err != nil || len(docenti) != 1 || docenti[0].Codice != "42" {
		t.Fatalf("docenti inattesi: %#v, errore: %v", docenti, err)
	}
	attivita, err := client.RecuperaInsegnamenti(ctx, "2026")
	if err != nil || len(attivita) != 1 || attivita[0].Nome != "Algoritmi" {
		t.Fatalf("insegnamenti inattesi: %#v, errore: %v", attivita, err)
	}
}

func TestErroriRisposta(t *testing.T) {
	tests := []struct {
		nome, risposta, errore string
	}{
		{"variabile assente", `var altro = [];`, "variabile \"elenco_docenti\" assente"},
		{"json incompleto", `var elenco_docenti = [{`, "incompleto"},
		{"json non valido", `var elenco_docenti = [non-json];`, "decodifica JSON"},
	}
	for _, test := range tests {
		t.Run(test.nome, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(test.risposta))
			}))
			defer server.Close()
			client, _ := NuovoClientConHTTP(server.URL+"/", server.Client())
			_, err := client.RecuperaDocenti(context.Background(), "2026")
			if err == nil || !strings.Contains(err.Error(), test.errore) {
				t.Fatalf("errore ottenuto %v, atteso contenente %q", err, test.errore)
			}
		})
	}
}

func TestAnnoObbligatorio(t *testing.T) {
	client := NuovoClient()
	_, err := client.RecuperaCorsi(context.Background(), " ")
	if err == nil || !strings.Contains(err.Error(), "obbligatorio") {
		t.Fatalf("errore inatteso: %v", err)
	}
}

func TestRiparaTestoDelPortale(t *testing.T) {
	testo := "AbilitÃ  informatiche"
	riparaTesto(&testo)
	if testo != "Abilità informatiche" {
		t.Fatalf("testo non riparato: %q", testo)
	}
}
