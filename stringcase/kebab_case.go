package stringcase

func KebabCase(s string) string {
	return delimiterCase(s, '-', false)
}

func UpperKebabCase(s string) string {
	return delimiterCase(s, '-', true)
}
