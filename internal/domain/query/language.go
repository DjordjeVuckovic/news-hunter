package query

type Language string

const (
	LanguageEnglish Language = "english"
	LanguageSpanish Language = "serbian"
)

var DefaultLanguage = LanguageEnglish

var SupportedLanguages = map[Language]bool{
	LanguageEnglish: true,
	LanguageSpanish: true,
}
