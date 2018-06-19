package localization

import (
	"context"
	"net/http"

	"golang.org/x/text/language"

	contextUtils "github.com/almerlucke/go-utils/server/context"
	"github.com/nicksnyder/go-i18n/i18n"
)

const (
	// LocalizationKey to get localization tag
	LocalizationKey = contextUtils.Key("localization")
)

// Localization localization data
type Localization struct {
	Translate i18n.TranslateFunc
	Tag       language.Tag
}

// Middleware middleware
type Middleware struct {
	Matcher language.Matcher
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

func (ware *Middleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	l, err := r.Cookie("lang")
	cookieLang := ""

	if err != http.ErrNoCookie && l != nil {
		cookieLang = l.String()
	}

	accept := r.Header.Get("Accept-Language")

	tag, _ := language.MatchStrings(ware.Matcher, cookieLang, accept)

	next(rw, r.WithContext(context.WithValue(r.Context(), LocalizationKey, &Localization{
		Tag:       tag,
		Translate: translateFunc(tag.String()),
	})))
}

// New language middleware
func New(m language.Matcher) *Middleware {
	return &Middleware{
		Matcher: m,
	}
}

// GetLocalization from context
func GetLocalization(ctx context.Context) *Localization {
	return ctx.Value(LocalizationKey).(*Localization)
}
