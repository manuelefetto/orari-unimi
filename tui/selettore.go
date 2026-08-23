package tui

import (
	"errors"
	"fmt"

	manucli "github.com/manuelefetto/ManuCli"
)

// Selettore astrae una scelta interattiva per rendere la TUI verificabile nei test.
type Selettore interface {
	Scegli(titolo string, opzioni []string) (int, error)
}

// SelettoreManuCLI usa le select a frecce della libreria ManuCli.
type SelettoreManuCLI struct{}

func NuovoSelettoreManuCLI() *SelettoreManuCLI {
	return &SelettoreManuCLI{}
}

func (s *SelettoreManuCLI) Scegli(titolo string, opzioni []string) (int, error) {
	if len(opzioni) == 0 {
		return -1, errors.New("la select non contiene opzioni")
	}
	fmt.Println()
	fmt.Println(titolo)
	selezionata := manucli.GenerateSelect(opzioni).Select()
	for i, opzione := range opzioni {
		if opzione == selezionata {
			return i, nil
		}
	}
	return -1, errors.New("ManuCli ha restituito un'opzione sconosciuta")
}
