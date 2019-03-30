package enedis

const (
	// Source constant for worker message
	Source = `enedis`
)

// Consumption describes consumption response
type Consumption struct {
	Graphe Graphe
}

// Graphe describes graphical data point
type Graphe struct {
	Data []Value
}

// Value describes data point
type Value struct {
	Valeur float64
	Ordre  int
}
