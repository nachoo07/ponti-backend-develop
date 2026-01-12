package pkgsuggester

type Suggestion struct {
	ID   int64  `json:"id"`
	Text string `json:"text"`
}

type Logger interface {
	Debug(msg string)
	Error(msg string, err error)
}

type noopLogger struct{}

func (noopLogger) Debug(_ string)          {}
func (noopLogger) Error(_ string, _ error) {}
