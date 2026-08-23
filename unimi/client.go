// Package unimi accede alle liste pubbliche del portale degli orari UNIMI.
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
	"strings"
	"time"
	"unicode/utf8"
)

const (
	indirizzoPredefinito = "https://orari-be.divsi.unimi.it/AgendaWeb/Orario/"
	dimensioneMassima    = 64 << 20
)

// Client esegue le richieste verso il portale pubblico degli orari.
type Client struct {
	indirizzo *url.URL
	http      *http.Client
}

// NuovoClient crea un client con l'indirizzo ufficiale e un timeout ragionevole.
func NuovoClient() *Client {
	client, err := NuovoClientConHTTP(indirizzoPredefinito, &http.Client{Timeout: 30 * time.Second})
	if err != nil {
		panic(err) // l'indirizzo costante è sempre valido
	}
	return client
}

// NuovoClientConHTTP permette di configurare indirizzo e client HTTP.
// È utile anche per usare un server locale nei test.
func NuovoClientConHTTP(indirizzo string, clientHTTP *http.Client) (*Client, error) {
	parsed, err := url.Parse(indirizzo)
	if err != nil {
		return nil, fmt.Errorf("indirizzo del portale non valido: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("l'indirizzo del portale deve essere assoluto")
	}
	if clientHTTP == nil {
		return nil, errors.New("il client HTTP non può essere nil")
	}
	return &Client{indirizzo: parsed, http: clientHTTP}, nil
}

// RecuperaAnniAccademici restituisce gli anni pubblicati, dal più recente.
func (c *Client) RecuperaAnniAccademici(ctx context.Context) ([]AnnoAccademico, error) {
	var valori map[string]AnnoAccademico
	if err := c.recuperaVariabile(ctx, "1", "", "anni_accademici_ec", &valori); err != nil {
		return nil, err
	}

	anni := make([]AnnoAccademico, 0, len(valori))
	for _, anno := range valori {
		riparaTesto(&anno.Nome)
		anni = append(anni, anno)
	}
	sort.Slice(anni, func(i, j int) bool { return anni[i].Codice > anni[j].Codice })
	return anni, nil
}

// RecuperaCorsi restituisce tutti i corsi di studio dell'anno indicato.
func (c *Client) RecuperaCorsi(ctx context.Context, anno string) ([]CorsoDiStudio, error) {
	var corsi []CorsoDiStudio
	if err := c.recuperaVariabile(ctx, anno, "corsi", "elenco_corsi", &corsi); err != nil {
		return nil, err
	}
	for i := range corsi {
		riparaTesto(&corsi[i].Nome)
		riparaTesto(&corsi[i].Tipo)
		for j := range corsi[i].Periodi {
			riparaTesto(&corsi[i].Periodi[j].Nome)
		}
	}
	return corsi, nil
}

// RecuperaFacolta restituisce le strutture usate per filtrare i corsi.
func (c *Client) RecuperaFacolta(ctx context.Context, anno string) ([]Facolta, error) {
	var facolta []Facolta
	if err := c.recuperaVariabile(ctx, anno, "corsi", "elenco_scuole", &facolta); err != nil {
		return nil, err
	}
	for i := range facolta {
		riparaTesto(&facolta[i].Nome)
	}
	return facolta, nil
}

// RecuperaDocenti restituisce tutti i docenti dell'anno indicato.
func (c *Client) RecuperaDocenti(ctx context.Context, anno string) ([]Docente, error) {
	var docenti []Docente
	if err := c.recuperaVariabile(ctx, anno, "docenti", "elenco_docenti", &docenti); err != nil {
		return nil, err
	}
	for i := range docenti {
		riparaTesto(&docenti[i].Nome)
	}
	return docenti, nil
}

// RecuperaInsegnamenti restituisce tutte le attività didattiche dell'anno indicato.
func (c *Client) RecuperaInsegnamenti(ctx context.Context, anno string) ([]Insegnamento, error) {
	var insegnamenti []Insegnamento
	if err := c.recuperaVariabile(ctx, anno, "attivita", "elenco_attivita", &insegnamenti); err != nil {
		return nil, err
	}
	for i := range insegnamenti {
		riparaTesto(&insegnamenti[i].Nome)
		riparaTesto(&insegnamenti[i].Descrizione)
		riparaTesto(&insegnamenti[i].Docenti)
	}
	return insegnamenti, nil
}

func (c *Client) recuperaVariabile(ctx context.Context, anno, pagina, variabile string, destinazione any) error {
	if strings.TrimSpace(anno) == "" {
		return errors.New("l'anno accademico è obbligatorio")
	}

	endpoint := c.indirizzo.ResolveReference(&url.URL{Path: "combo.php"})
	query := endpoint.Query()
	query.Set("sw", "ec_")
	query.Set("aa", anno)
	if pagina != "" {
		query.Set("page", pagina)
	}
	endpoint.RawQuery = query.Encode()

	richiesta, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return fmt.Errorf("creazione richiesta %s: %w", variabile, err)
	}
	richiesta.Header.Set("Accept", "application/javascript, application/json;q=0.9")
	richiesta.Header.Set("User-Agent", "orari-unimi/1.0")

	risposta, err := c.http.Do(richiesta)
	if err != nil {
		return fmt.Errorf("recupero %s: %w", variabile, err)
	}
	defer risposta.Body.Close()
	if risposta.StatusCode < 200 || risposta.StatusCode >= 300 {
		return fmt.Errorf("recupero %s: il portale ha risposto %s", variabile, risposta.Status)
	}

	lettore := io.LimitReader(risposta.Body, dimensioneMassima+1)
	corpo, err := io.ReadAll(lettore)
	if err != nil {
		return fmt.Errorf("lettura %s: %w", variabile, err)
	}
	if len(corpo) > dimensioneMassima {
		return fmt.Errorf("lettura %s: risposta più grande di %d byte", variabile, dimensioneMassima)
	}

	dati, err := estraiVariabile(corpo, variabile)
	if err != nil {
		return fmt.Errorf("decodifica %s: %w", variabile, err)
	}
	if err := json.Unmarshal(dati, destinazione); err != nil {
		return fmt.Errorf("decodifica JSON di %s: %w", variabile, err)
	}
	return nil
}

// estraiVariabile isola il JSON dalla forma "var nome = <json>;" usata dal portale.
func estraiVariabile(corpo []byte, nome string) ([]byte, error) {
	testo := string(corpo)
	indice := strings.Index(testo, "var "+nome)
	if indice < 0 {
		return nil, fmt.Errorf("variabile %q assente nella risposta", nome)
	}
	resto := testo[indice+len("var "+nome):]
	uguale := strings.IndexByte(resto, '=')
	if uguale < 0 {
		return nil, fmt.Errorf("assegnazione di %q non valida", nome)
	}
	resto = strings.TrimSpace(resto[uguale+1:])
	if len(resto) == 0 || (resto[0] != '[' && resto[0] != '{') {
		return nil, fmt.Errorf("valore di %q non valido", nome)
	}

	aperta, chiusa := resto[0], byte('}')
	if aperta == '[' {
		chiusa = ']'
	}
	profondita := 0
	inString, escape := false, false
	for i := 0; i < len(resto); i++ {
		carattere := resto[i]
		if inString {
			if escape {
				escape = false
			} else if carattere == '\\' {
				escape = true
			} else if carattere == '"' {
				inString = false
			}
			continue
		}
		if carattere == '"' {
			inString = true
		} else if carattere == aperta {
			profondita++
		} else if carattere == chiusa {
			profondita--
			if profondita == 0 {
				return []byte(resto[:i+1]), nil
			}
		}
	}
	return nil, fmt.Errorf("valore di %q incompleto", nome)
}

func riparaTesto(testo *string) {
	if !strings.ContainsAny(*testo, "ÃÂ") {
		return
	}
	byteLatin1 := make([]byte, 0, len(*testo))
	for _, carattere := range *testo {
		if carattere > 255 {
			return
		}
		byteLatin1 = append(byteLatin1, byte(carattere))
	}
	if utf8.Valid(byteLatin1) {
		*testo = string(byteLatin1)
	}
}
