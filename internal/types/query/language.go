package query

import "fmt"

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

func (l Language) Parse() (Language, error) {
	if l == "" {
		return DefaultLanguage, nil
	}
	if _, ok := SupportedLanguages[l]; !ok {
		return "", fmt.Errorf("unsupported language: %s", l)
	}
	return l, nil
}
