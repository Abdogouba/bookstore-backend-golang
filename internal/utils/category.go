package utils

var AllowedCategories = map[string]bool{

	"Programming": true,
	"Fiction":     true,
	"Science":     true,
	"History":     true,
	"Business":    true,
	"Religion":    true,
	"Biography":   true,
	"Children":    true,
	"Education":   true,
}

// IsValidCategory checks whether
// the category is supported.
func IsValidCategory(
	category string,
) bool {

	return AllowedCategories[category]
}