package caser

import "strings"

type Caser string

func (c Caser) PascalCase() string {
	return string(c)
}

// Parts splits the PascalCase string into individual capitalized words
func (c Caser) Parts() []string {
	parts := make([]string, 0)
	// start at the first character
	start := 0
	// iterate over each character
	for i := 1; i < len(c); i++ {
		// if the character is uppercase
		if 'A' <= c[i] && c[i] <= 'Z' {
			// add the part to the list
			parts = append(parts, string(c)[start:i])
			// start the next part
			start = i
		}
	}
	// add the last part to the list
	parts = append(parts, string(c)[start:])

	return parts
}

func (c Caser) PascalSnakeCase() string {
	// split at uppercase letters
	// join the parts together with underscores
	return strings.Join(c.Parts(), "_")
}

func (c Caser) CamelCase() string {
	str := c.PascalCase()
	return strings.ToLower(str[:1]) + str[1:]
}

func (c Caser) ScreamingSnakeCase() string {
	return strings.ToUpper(c.PascalSnakeCase())
}

func (c Caser) SnakeCase() string {
	return strings.ToLower(c.PascalSnakeCase())
}

func (c Caser) KebabCase() string {
	return strings.ReplaceAll(c.PascalSnakeCase(), "_", "-")
}

func capitalize(s string) string {
	return strings.ToUpper(s[:1]) + s[1:]
}

func ToPascalCase(s string) string {
	// split at spaces, underscores, and hyphens
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == ' ' || r == '_' || r == '-'
	})
	// capitalize the first letter of each part
	for i, p := range parts {
		parts[i] = capitalize(p)
	}
	// join the parts together
	return strings.Join(parts, "")
}

func New(s string) Caser {
	// convert the string to PascalCase for default string representation
	return Caser(ToPascalCase(s))
}
