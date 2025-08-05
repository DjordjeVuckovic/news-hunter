package stringsutil

func RemoveEmptyStrings(slice []string) []string {
	var result []string

	for _, s := range slice {
		if s != "" {
			result = append(result, s)
		}
	}

	return result
}
