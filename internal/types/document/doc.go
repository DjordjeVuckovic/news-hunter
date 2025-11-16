package document

type Document interface {
	ContainsField(field string) bool
}

func ContainsFields[D Document](doc D, fields []string) bool {
	for _, field := range fields {
		if !doc.ContainsField(field) {
			return false
		}
	}
	return true
}
