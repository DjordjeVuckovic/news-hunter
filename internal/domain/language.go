package domain

type SearchLanguage string

const (
	LanguageEnglish SearchLanguage = "english"
	LanguageSpanish SearchLanguage = "serbian"
)

var DefaultSearchLanguage = LanguageEnglish

var SupportedLanguages = map[SearchLanguage]bool{
	LanguageEnglish: true,
	LanguageSpanish: true,
}
