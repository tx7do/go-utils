package name_generator

const (
	DictionaryTypeAdjective DictionaryType = "adjective"
	DictionaryTypeGoods     DictionaryType = "goods"
	DictionaryTypeName      DictionaryType = "name"
	DictionaryTypePrefix    DictionaryType = "prefix"
	DictionaryTypeRole      DictionaryType = "role"
	DictionaryTypeVerb      DictionaryType = "verb"
	DictionaryTypeSensitive DictionaryType = "sensitive"

	DictionaryTypeSingleSurnames         DictionaryType = "single_surnames"
	DictionaryTypeCompoundSurnames       DictionaryType = "compound_surnames"
	DictionaryTypeChineseFirstNameFemale DictionaryType = "chinese_first_name_female"
	DictionaryTypeChineseFirstNameMale   DictionaryType = "chinese_first_name_male"

	DictionaryTypeEnglishFirstNameFemale DictionaryType = "english_first_name_female"
	DictionaryTypeEnglishFirstNameMale   DictionaryType = "english_first_name_male"
	DictionaryTypeEnglishLastName        DictionaryType = "english_last_name"

	DictionaryTypeJapaneseName     DictionaryType = "japanese_name"
	DictionaryTypeJapaneseSurnames DictionaryType = "japanese_surnames"
	DictionaryTypeJapaneseLastName DictionaryType = "japanese_last_name"
)

var Scheme1 = CombinedDictionaryType{
	DictionaryTypePrefix,
	DictionaryTypeName,
	DictionaryTypeVerb,
}

var Scheme2 = CombinedDictionaryType{
	DictionaryTypePrefix,
	DictionaryTypeRole,
	DictionaryTypeVerb,
}

var Scheme3 = CombinedDictionaryType{
	DictionaryTypePrefix,
	DictionaryTypeName,
	DictionaryTypeAdjective,
}

var Scheme4 = CombinedDictionaryType{
	DictionaryTypePrefix,
	DictionaryTypeVerb,
	DictionaryTypeRole,
}

var Scheme5 = CombinedDictionaryType{
	DictionaryTypePrefix,
	DictionaryTypeVerb,
	DictionaryTypeName,
}

var Scheme6 = CombinedDictionaryType{
	DictionaryTypeName,
	DictionaryTypePrefix,
	DictionaryTypeGoods,
}

var SchemeChineseNameFemale = CombinedDictionaryType{
	DictionaryTypeSingleSurnames,
	DictionaryTypeChineseFirstNameFemale,
}

var SchemeChineseNameMale = CombinedDictionaryType{
	DictionaryTypeSingleSurnames,
	DictionaryTypeChineseFirstNameMale,
}
