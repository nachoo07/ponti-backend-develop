package defs

type Service interface {
	ReadImports(filePath string) ([]string, error)
}

type Config interface {
	GetAnalyzePath() string
	SetAnalyzePath(string)
	Validate() error
}
