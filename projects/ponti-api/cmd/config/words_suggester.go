package config

// WordsSuggester define la configuración para el servicio de sugerencia de palabras.
type WordsSuggester struct {
	// Limit es el número máximo de sugerencias a retornar.
	Limit int `envconfig:"WORDS_SUGGESTER_LIMIT" default:"10" validate:"gt=0"`
	// Threshold es el umbral mínimo de similitud para incluir una sugerencia.
	Threshold float64 `envconfig:"WORDS_SUGGESTER_THRESHOLD" default:"0.3" validate:"gte=0,lte=1"`
}
