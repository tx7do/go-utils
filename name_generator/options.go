package name_generator

type DictionaryType string

type Dictionary []string
type DictionaryMap map[DictionaryType]Dictionary

type CombinedDictionaryType []DictionaryType
