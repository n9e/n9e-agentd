package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestI18n(t *testing.T) {
	m := map[string]map[string]string{
		"zh": map[string]string{
			"hello world!": "你好 世界!",
		},
	}

	cases := [][2]string{
		[2]string{"hello world!", "你好 世界!"},
		[2]string{"hello world !", "hello world !"},
	}

	SetLangStrings(m)
	p := NewPrinter("zh")

	for _, c := range cases {
		got := p.Sprintf(c[0])
		assert.Equal(t, c[1], got)
	}

}
