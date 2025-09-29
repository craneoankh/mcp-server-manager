package services

// contains checks if a slice contains a specific item
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// addUnique adds an item to a slice if it's not already present
func addUnique(slice []string, item string) []string {
	if contains(slice, item) {
		return slice
	}
	return append(slice, item)
}

// removeItem removes an item from a slice
func removeItem(slice []string, item string) []string {
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}