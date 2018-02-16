package localization

import (
	"net/http"

	"github.com/nicksnyder/go-i18n/i18n"
	"golang.org/x/text/language"
)

// var matcher = language.NewMatcher([]language.Tag{
// 	language.English, // The first language is used as fallback.
// 	language.MustParse("en-AU"),
// 	language.Danish,
// 	language.Chinese,
// })

// Localization info
type Localization struct {
	Translate i18n.TranslateFunc
	Tag       language.Tag
}

func translateFunc(acceptLang string) i18n.TranslateFunc {
	T, err := i18n.Tfunc(acceptLang)
	if err != nil {
		T = func(translationID string, args ...interface{}) string {
			return translationID
		}
	}

	return T
}

// GetLocalizationForRequest create a localization object containing a translate
// function and a language tag from the http.Request accept language
func GetLocalizationForRequest(request *http.Request, languageMatcher language.Matcher) *Localization {
	localization := &Localization{}

	requestedLanguage := request.Header.Get("Accept-Language")
	tag, _ := language.MatchStrings(languageMatcher, requestedLanguage)

	localization.Tag = tag
	localization.Translate = translateFunc(tag.String())

	return localization
}
