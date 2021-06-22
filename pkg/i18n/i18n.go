package i18n

import (
	"io"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var (
	defPrinter = message.NewPrinter(language.Chinese)
	matcher    = language.NewMatcher([]language.Tag{
		language.Chinese, // zh
		language.English, // en
	})
)

// lang: BCP 47 - Tags for Identifying Languages
// http://tools.ietf.org/html/bcp47
func NewPrinter(lang ...string) *message.Printer {
	tag, _ := language.MatchStrings(matcher, lang...)
	return message.NewPrinter(tag)
}

func SetDefaultPrinter(lang ...string) {
	defPrinter = NewPrinter(lang...)
}

func SetMatcher(langs ...language.Tag) {
	matcher = language.NewMatcher(langs)
}

func SetLangStrings(dict map[string]map[string]string) {
	for lang, kv := range dict {
		SetStrings(lang, kv)
	}
}

func SetStrings(lang string, kv map[string]string) {
	tag, _ := language.MatchStrings(matcher, lang)
	for k, v := range kv {
		message.SetString(tag, k, v)
	}
}

// Fprintf is like fmt.Fprintf, but using language-specific formatting.
func Fprintf(w io.Writer, key message.Reference, a ...interface{}) (n int, err error) {
	return defPrinter.Fprintf(w, key, a...)
}

// Printf is like fmt.Printf, but using language-specific formatting.
func Printf(format string, a ...interface{}) {
	_, _ = defPrinter.Printf(format, a...)
}

// Sprintf formats according to a format specifier and returns the resulting string.
func Sprintf(format string, a ...interface{}) string {
	return defPrinter.Sprintf(format, a...)
}

// Sprint is like fmt.Sprint, but using language-specific formatting.
func Sprint(a ...interface{}) string {
	return defPrinter.Sprint(a...)
}
