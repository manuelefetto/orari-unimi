package unimi

// AnnoAccademico identifica un anno pubblicato nel portale degli orari.
type AnnoAccademico struct {
	Codice string `json:"valore"`
	Nome   string `json:"label"`
}

// PeriodoDidattico descrive uno dei periodi in cui viene pubblicato l'orario.
type PeriodoDidattico struct {
	Codice string `json:"valore"`
	Nome   string `json:"label"`
}

// CorsoDiStudio è una voce selezionabile nella ricerca per corso.
type CorsoDiStudio struct {
	Codice   string              `json:"valore"`
	Nome     string              `json:"label"`
	Tipo     string              `json:"tipo"`
	Facolta  string              `json:"scuola"`
	Periodi  []PeriodoDidattico  `json:"pub_periodi"`
	Percorsi []PercorsoDidattico `json:"elenco_anni"`
}

// PercorsoDidattico identifica un anno, curriculum o gruppo di un corso.
type PercorsoDidattico struct {
	Codice string `json:"valore"`
	Nome   string `json:"label"`
}

// Facolta è una facoltà, scuola o altra struttura didattica del portale.
type Facolta struct {
	Codice string `json:"valore"`
	Nome   string `json:"label"`
}

// Docente è una voce selezionabile nella ricerca per docente.
type Docente struct {
	Codice string `json:"valore"`
	Nome   string `json:"label"`
}

// Insegnamento è una voce selezionabile nella ricerca per insegnamento.
type Insegnamento struct {
	Codice      string `json:"valore"`
	Nome        string `json:"nome_insegnamento"`
	Descrizione string `json:"label"`
	Docenti     string `json:"docente"`
	CodiceUnimi string `json:"codice_combinato"`
}
