package compoundsplitting

type NoopSplitter struct{}

func NewNoopSplitter() NoopSplitter {
	return NoopSplitter{}
}

func (n NoopSplitter) Split(words string) ([]string, error) {
	return []string{}, nil
}
