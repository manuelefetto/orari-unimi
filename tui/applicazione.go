package tui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"orari-unimi/unimi"
)

const numeroSuggerimenti = 10

type sorgenteUNIMI interface {
	RecuperaAnniAccademici(context.Context) ([]unimi.AnnoAccademico, error)
	RecuperaCorsi(context.Context, string) ([]unimi.CorsoDiStudio, error)
	RecuperaDocenti(context.Context, string) ([]unimi.Docente, error)
	RecuperaInsegnamenti(context.Context, string) ([]unimi.Insegnamento, error)
}

// Applicazione gestisce la navigazione testuale e mantiene in memoria le liste
// già scaricate durante la sessione.
type Applicazione struct {
	lettore   LettoreTesto
	uscita    io.Writer
	sorgente  sorgenteUNIMI
	archivio  *ArchivioCorsi
	selettore Selettore
	anno      unimi.AnnoAccademico
	corsi     []unimi.CorsoDiStudio
	docenti   []unimi.Docente
	attivita  []unimi.Insegnamento
}

func NuovaApplicazione(lettore LettoreTesto, uscita io.Writer, sorgente sorgenteUNIMI, archivio *ArchivioCorsi, selettore Selettore) (*Applicazione, error) {
	if lettore == nil || uscita == nil || sorgente == nil || archivio == nil || selettore == nil {
		return nil, errors.New("dipendenze dell'applicazione non valide")
	}
	return &Applicazione{
		lettore:   lettore,
		uscita:    uscita,
		sorgente:  sorgente,
		archivio:  archivio,
		selettore: selettore,
	}, nil
}

// Esegui mostra il menu principale finché l'utente non sceglie di uscire.
func (a *Applicazione) Esegui(ctx context.Context) error {
	if err := a.caricaAnno(ctx); err != nil {
		return err
	}

	for {
		fmt.Fprintf(a.uscita, "\nOrari UNIMI — anno accademico %s\n", a.anno.Nome)
		scelta, err := a.selettore.Scegli("Scegli come consultare gli orari", []string{
			"Per corso",
			"Per docente",
			"Per insegnamento",
			"I miei orari",
			"Esci",
		})
		if err != nil {
			return err
		}
		switch scelta {
		case 0:
			a.menuPerCorso(ctx)
		case 1:
			a.menuPerDocente(ctx)
		case 2:
			a.menuPerInsegnamento(ctx)
		case 3:
			if err := a.menuMieiOrari(ctx); err != nil {
				fmt.Fprintf(a.uscita, "Errore: %v\n", err)
			}
		case 4:
			fmt.Fprintln(a.uscita, "A presto!")
			return nil
		}
	}
}

func (a *Applicazione) menuPerCorso(ctx context.Context) {
	corso, selezionato, err := a.selezionaCorso(ctx)
	if err != nil {
		fmt.Fprintf(a.uscita, "Errore: %v\n", err)
		return
	}
	if selezionato {
		fmt.Fprintf(a.uscita, "Corso selezionato: %s (%s). Gli orari saranno disponibili nella fase 3.\n", corso.Nome, corso.Codice)
	}
}

func (a *Applicazione) menuPerDocente(ctx context.Context) {
	if err := a.caricaDocenti(ctx); err != nil {
		fmt.Fprintf(a.uscita, "Errore: %v\n", err)
		return
	}
	voci := make([]voceRicerca, len(a.docenti))
	for i, docente := range a.docenti {
		voci[i] = voceRicerca{Codice: docente.Codice, Nome: docente.Nome}
	}
	voce, selezionata, err := a.selezionaConRicerca("Cerca docente", voci)
	if err != nil {
		fmt.Fprintf(a.uscita, "Errore: %v\n", err)
		return
	}
	if selezionata {
		fmt.Fprintf(a.uscita, "Docente selezionato: %s (%s). Gli orari saranno disponibili nella fase 3.\n", voce.Nome, voce.Codice)
	}
}

func (a *Applicazione) menuPerInsegnamento(ctx context.Context) {
	if err := a.caricaInsegnamenti(ctx); err != nil {
		fmt.Fprintf(a.uscita, "Errore: %v\n", err)
		return
	}
	voci := make([]voceRicerca, len(a.attivita))
	for i, insegnamento := range a.attivita {
		nome := insegnamento.Descrizione
		if strings.TrimSpace(nome) == "" {
			nome = insegnamento.Nome
		}
		voci[i] = voceRicerca{Codice: insegnamento.Codice, Nome: nome}
	}
	voce, selezionata, err := a.selezionaConRicerca("Cerca insegnamento", voci)
	if err != nil {
		fmt.Fprintf(a.uscita, "Errore: %v\n", err)
		return
	}
	if selezionata {
		fmt.Fprintf(a.uscita, "Insegnamento selezionato: %s (%s). Gli orari saranno disponibili nella fase 3.\n", voce.Nome, voce.Codice)
	}
}

func (a *Applicazione) menuMieiOrari(ctx context.Context) error {
	for {
		scelta, err := a.selettore.Scegli("I miei orari", []string{
			"Gli orari dei corsi selezionati",
			"Aggiungi corso",
			"Rimuovi corso",
			"Pulisci corsi",
			"Torna al menu principale",
		})
		if err != nil {
			return err
		}
		switch scelta {
		case 0:
			if err := a.mostraCorsiSalvati(); err != nil {
				return err
			}
		case 1:
			if err := a.aggiungiCorso(ctx); err != nil {
				return err
			}
		case 2:
			if err := a.rimuoviCorso(); err != nil {
				return err
			}
		case 3:
			if err := a.pulisciCorsi(); err != nil {
				return err
			}
		case 4:
			return nil
		}
	}
}

func (a *Applicazione) selezionaCorso(ctx context.Context) (unimi.CorsoDiStudio, bool, error) {
	if err := a.caricaCorsi(ctx); err != nil {
		return unimi.CorsoDiStudio{}, false, err
	}
	voci := make([]voceRicerca, len(a.corsi))
	perCodice := make(map[string]unimi.CorsoDiStudio, len(a.corsi))
	for i, corso := range a.corsi {
		voci[i] = voceRicerca{Codice: corso.Codice, Nome: corso.Nome}
		perCodice[corso.Codice] = corso
	}
	voce, selezionata, err := a.selezionaConRicerca("Cerca corso", voci)
	if err != nil || !selezionata {
		return unimi.CorsoDiStudio{}, false, err
	}
	return perCodice[voce.Codice], true, nil
}

func (a *Applicazione) selezionaConRicerca(titolo string, voci []voceRicerca) (voceRicerca, bool, error) {
	ricerca := ""
	for {
		if ricerca == "" {
			var err error
			ricerca, err = a.leggi(titolo + " (INVIO per tornare): ")
			if err != nil {
				return voceRicerca{}, false, err
			}
			if strings.TrimSpace(ricerca) == "" {
				return voceRicerca{}, false, nil
			}
		}

		risultati := filtraVoci(voci, ricerca, numeroSuggerimenti)
		if len(risultati) == 0 {
			fmt.Fprintln(a.uscita, "Nessun risultato. Prova con un'altra ricerca.")
			ricerca = ""
			continue
		}
		fmt.Fprintf(a.uscita, "Suggerimenti per %q:\n", ricerca)
		opzioni := make([]string, 0, len(risultati)+2)
		for _, voce := range risultati {
			opzioni = append(opzioni, fmt.Sprintf("%s [%s]", voce.Nome, voce.Codice))
		}
		opzioni = append(opzioni, "Nuova ricerca", "Indietro")
		scelta, err := a.selettore.Scegli(fmt.Sprintf("Suggerimenti per %q", ricerca), opzioni)
		if err != nil {
			return voceRicerca{}, false, err
		}
		if scelta == len(risultati)+1 {
			return voceRicerca{}, false, nil
		}
		if scelta == len(risultati) {
			ricerca = ""
			continue
		}
		return risultati[scelta], true, nil
	}
}

func (a *Applicazione) aggiungiCorso(ctx context.Context) error {
	corso, selezionato, err := a.selezionaCorso(ctx)
	if err != nil || !selezionato {
		return err
	}
	aggiunto, err := a.archivio.Aggiungi(CorsoSalvato{Anno: a.anno.Codice, Codice: corso.Codice, Nome: corso.Nome})
	if err != nil {
		return err
	}
	if aggiunto {
		fmt.Fprintf(a.uscita, "%s aggiunto ai tuoi corsi.\n", corso.Nome)
	} else {
		fmt.Fprintln(a.uscita, "Il corso è già presente nei tuoi corsi.")
	}
	return nil
}

func (a *Applicazione) mostraCorsiSalvati() error {
	corsi, err := a.archivio.Elenca()
	if err != nil {
		return err
	}
	if len(corsi) == 0 {
		fmt.Fprintln(a.uscita, "Non hai ancora selezionato corsi.")
		return nil
	}
	for i, corso := range corsi {
		fmt.Fprintf(a.uscita, "%d) %s [%s, a.a. %s]\n", i+1, corso.Nome, corso.Codice, corso.Anno)
	}
	buf, err := a.leggi("Gli orari saranno disponibili nella fase 3. Premi INVIO per continuare: ")
	_ = buf
	return err
}

func (a *Applicazione) rimuoviCorso() error {
	corsi, err := a.archivio.Elenca()
	if err != nil {
		return err
	}
	if len(corsi) == 0 {
		fmt.Fprintln(a.uscita, "Non ci sono corsi da rimuovere.")
		return nil
	}
	opzioni := make([]string, 0, len(corsi)+1)
	for _, corso := range corsi {
		opzioni = append(opzioni, fmt.Sprintf("%s [%s]", corso.Nome, corso.Codice))
	}
	opzioni = append(opzioni, "Annulla")
	scelta, err := a.selettore.Scegli("Scegli il corso da rimuovere", opzioni)
	if err != nil {
		return err
	}
	if scelta == len(corsi) {
		return nil
	}
	if err := a.archivio.Rimuovi(scelta); err != nil {
		return err
	}
	fmt.Fprintf(a.uscita, "%s rimosso dai tuoi corsi.\n", corsi[scelta].Nome)
	return nil
}

func (a *Applicazione) pulisciCorsi() error {
	corsi, err := a.archivio.Elenca()
	if err != nil {
		return err
	}
	if len(corsi) == 0 {
		fmt.Fprintln(a.uscita, "L'elenco è già vuoto.")
		return nil
	}
	conferma, err := a.selettore.Scegli("Eliminare tutti i corsi?", []string{"No, annulla", "Sì, elimina tutto"})
	if err != nil {
		return err
	}
	if conferma != 1 {
		fmt.Fprintln(a.uscita, "Operazione annullata.")
		return nil
	}
	if err := a.archivio.Pulisci(); err != nil {
		return err
	}
	fmt.Fprintln(a.uscita, "Tutti i corsi sono stati rimossi.")
	return nil
}

func (a *Applicazione) caricaAnno(ctx context.Context) error {
	anni, err := a.sorgente.RecuperaAnniAccademici(ctx)
	if err != nil {
		return fmt.Errorf("recupero anni accademici: %w", err)
	}
	if len(anni) == 0 {
		return errors.New("il portale non ha restituito anni accademici")
	}
	a.anno = anni[0]
	return nil
}

func (a *Applicazione) caricaCorsi(ctx context.Context) error {
	if a.corsi != nil {
		return nil
	}
	corsi, err := a.sorgente.RecuperaCorsi(ctx, a.anno.Codice)
	if err != nil {
		return fmt.Errorf("recupero corsi: %w", err)
	}
	a.corsi = corsi
	return nil
}

func (a *Applicazione) caricaDocenti(ctx context.Context) error {
	if a.docenti != nil {
		return nil
	}
	docenti, err := a.sorgente.RecuperaDocenti(ctx, a.anno.Codice)
	if err != nil {
		return fmt.Errorf("recupero docenti: %w", err)
	}
	a.docenti = docenti
	return nil
}

func (a *Applicazione) caricaInsegnamenti(ctx context.Context) error {
	if a.attivita != nil {
		return nil
	}
	attivita, err := a.sorgente.RecuperaInsegnamenti(ctx, a.anno.Codice)
	if err != nil {
		return fmt.Errorf("recupero insegnamenti: %w", err)
	}
	a.attivita = attivita
	return nil
}

func (a *Applicazione) leggi(messaggio string) (string, error) {
	return a.lettore.Leggi(messaggio)
}
