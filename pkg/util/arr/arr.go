package arr

type StrDataSet map[string]struct{}

func NewStrDataSet(items []string) StrDataSet {
	set := StrDataSet{}
	for _, item := range items {
		set[item] = struct{}{}
	}
	return set
}

func (s StrDataSet) Has(item string) bool {
	_, ok := s[item]
	return ok
}

func (s StrDataSet) HasOne(items []string) bool {
	for _, item := range items {
		if s.Has(item) {
			return true
		}
	}
	return false
}
