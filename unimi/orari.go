package unimi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const formatoDataPortale = "02-01-2006"

// IntervalloDate limita le lezioni alle date Dal e Al, estremi inclusi.
// Il valore zero indica l'intero anno accademico.
type IntervalloDate struct {
	Dal time.Time
	Al  time.Time
}

// NuovoIntervalloDate crea e valida un intervallo inclusivo.
func NuovoIntervalloDate(dal, al time.Time) (IntervalloDate, error) {
	intervallo := IntervalloDate{Dal: dal, Al: al}
	if err := intervallo.valida(); err != nil {
		return IntervalloDate{}, err
	}
	return intervallo, nil
}

// Lezione rappresenta un singolo evento didattico restituito da Agenda Web.
type Lezione struct {
	ID                 string
	CodiceInsegnamento string
	Insegnamento       string
	CodiceDocente      string
	Docente            string
	EmailDocente       string
	Data               time.Time
	OraInizio          string
	OraFine            string
	CodiceAula         string
	Aula               string
	CodiceSede         string
	PercorsoDidattico  string
	Tipo               string
	Note               string
	Annullata          bool
}

type rispostaOrari struct {
	Celle []cellaPortale `json:"celle"`
}

type cellaPortale struct {
	ID                 string `json:"id"`
	CodiceInsegnamento string `json:"codice_insegnamento"`
	NomeInsegnamento   string `json:"nome_insegnamento"`
	CodiceDocente      string `json:"codice_docente"`
	Docente            string `json:"docente"`
	EmailDocente       string `json:"mail_docente"`
	Data               string `json:"data"`
	OraInizio          string `json:"ora_inizio"`
	OraFine            string `json:"ora_fine"`
	CodiceAula         string `json:"codice_aula"`
	Aula               string `json:"aula"`
	CodiceSede         string `json:"codice_sede"`
	PercorsoDidattico  string `json:"percorso_didattico"`
	Tipo               string `json:"tipo"`
	Note               string `json:"notes"`
	NoteSettimanali    string `json:"NoteSettimanali"`
	NoteEasyRoom       string `json:"note_easyroom"`
	Annullato          string `json:"Annullato"`
}

// RecuperaOrariCorso restituisce le lezioni di tutti gli anni e percorsi del
// corso. Un intervallo vuoto restituisce l'intero anno accademico.
func (c *Client) RecuperaOrariCorso(ctx context.Context, anno, codiceCorso string, intervallo IntervalloDate) ([]Lezione, error) {
	if strings.TrimSpace(codiceCorso) == "" {
		return nil, errors.New("il codice del corso è obbligatorio")
	}
	corsi, err := c.RecuperaCorsi(ctx, anno)
	if err != nil {
		return nil, fmt.Errorf("recupero percorsi del corso: %w", err)
	}
	var percorsi []string
	for _, corso := range corsi {
		if corso.Codice != codiceCorso {
			continue
		}
		for _, percorso := range corso.Percorsi {
			if percorso.Codice != "" {
				percorsi = append(percorsi, percorso.Codice)
			}
		}
		break
	}
	if len(percorsi) == 0 {
		return nil, fmt.Errorf("il corso %q non ha percorsi pubblicati per l'anno %s", codiceCorso, anno)
	}

	parametri := url.Values{
		"include": {"corso"},
		"corso":   {codiceCorso},
		"anno2[]": percorsi,
	}
	return c.recuperaOrari(ctx, anno, parametri, intervallo)
}

// RecuperaOrariDocente restituisce tutte le lezioni del docente. Un intervallo
// vuoto restituisce l'intero anno accademico.
func (c *Client) RecuperaOrariDocente(ctx context.Context, anno, codiceDocente string, intervallo IntervalloDate) ([]Lezione, error) {
	if strings.TrimSpace(codiceDocente) == "" {
		return nil, errors.New("il codice del docente è obbligatorio")
	}
	parametri := url.Values{"include": {"docente"}, "docente": {codiceDocente}}
	return c.recuperaOrari(ctx, anno, parametri, intervallo)
}

// RecuperaOrariInsegnamento restituisce tutte le lezioni dell'insegnamento. Un
// intervallo vuoto restituisce l'intero anno accademico.
func (c *Client) RecuperaOrariInsegnamento(ctx context.Context, anno, codiceInsegnamento string, intervallo IntervalloDate) ([]Lezione, error) {
	if strings.TrimSpace(codiceInsegnamento) == "" {
		return nil, errors.New("il codice dell'insegnamento è obbligatorio")
	}
	parametri := url.Values{"include": {"attivita"}, "attivita[]": {codiceInsegnamento}}
	return c.recuperaOrari(ctx, anno, parametri, intervallo)
}

func (c *Client) recuperaOrari(ctx context.Context, anno string, parametri url.Values, intervallo IntervalloDate) ([]Lezione, error) {
	if strings.TrimSpace(anno) == "" {
		return nil, errors.New("l'anno accademico è obbligatorio")
	}
	if err := intervallo.valida(); err != nil {
		return nil, err
	}
	annoNumerico, err := strconv.Atoi(anno)
	if err != nil {
		return nil, fmt.Errorf("anno accademico %q non valido", anno)
	}

	dataRiferimento := time.Date(annoNumerico, time.August, 1, 0, 0, 0, 0, time.UTC)
	if !intervallo.Dal.IsZero() {
		dataRiferimento = intervallo.Dal
	}
	parametri.Set("view", "easycourse")
	parametri.Set("_lang", "it")
	parametri.Set("anno", anno)
	parametri.Set("date", dataRiferimento.Format(formatoDataPortale))
	parametri.Set("all_events", "1")

	endpoint := c.indirizzo.ResolveReference(&url.URL{Path: "grid_call.php"})
	richiesta, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), strings.NewReader(parametri.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creazione richiesta orari: %w", err)
	}
	richiesta.Header.Set("Accept", "application/json")
	richiesta.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	richiesta.Header.Set("User-Agent", "orari-unimi/1.0")

	risposta, err := c.http.Do(richiesta)
	if err != nil {
		return nil, fmt.Errorf("recupero orari: %w", err)
	}
	defer risposta.Body.Close()
	if risposta.StatusCode < 200 || risposta.StatusCode >= 300 {
		return nil, fmt.Errorf("recupero orari: il portale ha risposto %s", risposta.Status)
	}
	corpo, err := io.ReadAll(io.LimitReader(risposta.Body, dimensioneMassima+1))
	if err != nil {
		return nil, fmt.Errorf("lettura orari: %w", err)
	}
	if len(corpo) > dimensioneMassima {
		return nil, fmt.Errorf("lettura orari: risposta più grande di %d byte", dimensioneMassima)
	}

	var rispostaPortale rispostaOrari
	if err := json.Unmarshal(corpo, &rispostaPortale); err != nil {
		return nil, fmt.Errorf("decodifica orari: %w", err)
	}
	lezioni := make([]Lezione, 0, len(rispostaPortale.Celle))
	for _, cella := range rispostaPortale.Celle {
		lezione, valida, err := cella.lezione()
		if err != nil {
			return nil, err
		}
		if valida && intervallo.contiene(lezione.Data) {
			lezioni = append(lezioni, lezione)
		}
	}
	sort.SliceStable(lezioni, func(i, j int) bool {
		if !lezioni[i].Data.Equal(lezioni[j].Data) {
			return lezioni[i].Data.Before(lezioni[j].Data)
		}
		if lezioni[i].OraInizio != lezioni[j].OraInizio {
			return lezioni[i].OraInizio < lezioni[j].OraInizio
		}
		return lezioni[i].Insegnamento < lezioni[j].Insegnamento
	})
	return lezioni, nil
}

func (c cellaPortale) lezione() (Lezione, bool, error) {
	if strings.TrimSpace(c.CodiceInsegnamento) == "" {
		return Lezione{}, false, nil // chiusura, festività o altra cella non didattica
	}
	data, err := time.Parse(formatoDataPortale, c.Data)
	if err != nil {
		return Lezione{}, false, fmt.Errorf("data %q della lezione %q non valida: %w", c.Data, c.ID, err)
	}
	note := unisciTesti(c.Note, c.NoteSettimanali, c.NoteEasyRoom)
	lezione := Lezione{
		ID:                 c.ID,
		CodiceInsegnamento: c.CodiceInsegnamento,
		Insegnamento:       c.NomeInsegnamento,
		CodiceDocente:      c.CodiceDocente,
		Docente:            c.Docente,
		EmailDocente:       strings.TrimSpace(strings.Trim(strings.TrimSpace(c.EmailDocente), ",")),
		Data:               data,
		OraInizio:          c.OraInizio,
		OraFine:            c.OraFine,
		CodiceAula:         c.CodiceAula,
		Aula:               c.Aula,
		CodiceSede:         c.CodiceSede,
		PercorsoDidattico:  c.PercorsoDidattico,
		Tipo:               c.Tipo,
		Note:               note,
		Annullata:          c.Annullato == "1",
	}
	riparaTesto(&lezione.Insegnamento)
	riparaTesto(&lezione.Docente)
	riparaTesto(&lezione.Aula)
	riparaTesto(&lezione.PercorsoDidattico)
	riparaTesto(&lezione.Tipo)
	riparaTesto(&lezione.Note)
	return lezione, true, nil
}

func (i IntervalloDate) valida() error {
	if i.Dal.IsZero() && i.Al.IsZero() {
		return nil
	}
	if i.Dal.IsZero() || i.Al.IsZero() {
		return errors.New("l'intervallo deve contenere sia la data iniziale sia quella finale")
	}
	if dataSenzaOra(i.Al).Before(dataSenzaOra(i.Dal)) {
		return errors.New("la data finale non può precedere quella iniziale")
	}
	return nil
}

func (i IntervalloDate) contiene(data time.Time) bool {
	if i.Dal.IsZero() && i.Al.IsZero() {
		return true
	}
	data = dataSenzaOra(data)
	return !data.Before(dataSenzaOra(i.Dal)) && !data.After(dataSenzaOra(i.Al))
}

func dataSenzaOra(data time.Time) time.Time {
	return time.Date(data.Year(), data.Month(), data.Day(), 0, 0, 0, 0, time.UTC)
}

func unisciTesti(testi ...string) string {
	parti := make([]string, 0, len(testi))
	for _, testo := range testi {
		testo = strings.TrimSpace(testo)
		if testo != "" {
			parti = append(parti, testo)
		}
	}
	return strings.Join(parti, " — ")
}
