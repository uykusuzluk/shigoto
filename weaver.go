package shigoto

type weaver struct {
	listeners []*listener
	workers   []*worker
	waitnano  int
}

func newWeaver(s *Shigoto, wcount int, queue ...string) (*weaver, error) {
	return nil, nil
}
