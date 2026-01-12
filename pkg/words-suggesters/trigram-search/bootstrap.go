package pkgsuggester

func Bootstrap(opts ...Option) (*WordsSuggester, error) {
	cfg, err := newConfig(opts...)
	if err != nil {
		return nil, err
	}
	return newSuggester(cfg), nil
}
