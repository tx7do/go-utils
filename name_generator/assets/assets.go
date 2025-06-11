package assets

import _ "embed"

//go:embed adjective.txt
var Adjective []byte

//go:embed goods.txt
var Goods []byte

//go:embed name.txt
var Name []byte

//go:embed prefix.txt
var Prefix []byte

//go:embed role.txt
var Role []byte

//go:embed verb.txt
var Verb []byte

//go:embed sensitive.txt
var Sensitive []byte

//go:embed chinese_single_surnames.txt
var ChineseSingleSurnames []byte

//go:embed chinese_compound_surnames.txt
var ChineseCompoundSurnames []byte

//go:embed chinese_first_name_female.txt
var ChineseFirstNameFemale []byte

//go:embed chinese_first_name_male.txt
var ChineseFirstNameMale []byte

//go:embed english_first_name_female.txt
var EnglishFirstNameFemale []byte

//go:embed english_first_name_male.txt
var EnglishFirstNameMale []byte

//go:embed english_last_name.txt
var EnglishLastName []byte

//go:embed japanese_names_corpus.txt
var JapaneseNamesCorpus []byte

//go:embed japanese_surnames.txt
var JapaneseSurnames []byte

//go:embed japanese_last_name.txt
var JapaneseLastName []byte
