package utils

func GenerateSlug(name string) string {
	// Simple slug generation: convert to lowercase and replace spaces with hyphens
	slug := ""
	for _, ch := range name {
		if ch >= 'A' && ch <= 'Z' {
			slug += string(ch + 32) // Convert to lowercase
		} else if ch == ' ' || ch == '_' {
			slug += "-"
		} else {
			slug += string(ch)
		}
	}
	return slug
}
