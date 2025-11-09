package domain

type WeightedDocument interface {
	ContainField(field string) bool
}

func ContainFields[D WeightedDocument](doc D, fields []string) bool {
	for _, field := range fields {
		if !doc.ContainField(field) {
			return false
		}
	}
	return true
}
