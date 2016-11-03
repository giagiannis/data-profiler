package core

// DatasetsIntersection function is used to calculate the intersection
// of two datasets and returns the tuples that belong to it.
func DatasetsIntersection(a, b *Dataset) []DatasetTuple {
	a.ReadFromFile()
	b.ReadFromFile()
	dict := make(map[string]bool)
	for _, dt := range a.Data() {
		dict[dt.Serialize()] = true
	}
	result := make([]DatasetTuple, 0)
	for _, dt := range b.Data() {
		ok := dict[dt.Serialize()]
		if ok {
			result = append(result, dt)
		}
	}

	return result
}

// DatasetsUnion function is used to calculate the union of two datasets
// and returns the tuples that belong to it.
func DatasetsUnion(a, b *Dataset) []DatasetTuple {
	a.ReadFromFile()
	b.ReadFromFile()
	dict := make(map[string]bool)
	for _, dt := range a.Data() {
		dict[dt.Serialize()] = true
	}
	for _, dt := range b.Data() {
		dict[dt.Serialize()] = true
	}
	result := make([]DatasetTuple, 0)
	for k, _ := range dict {
		t := new(DatasetTuple)
		t.Deserialize(k)
		result = append(result, *t)
	}

	return result
}
