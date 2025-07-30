package lexers

import (
	"github.com/Chi-Iroh/chroma-test-gitea/v2"
)

// HTML lexer.
var HTML = chroma.MustNewXMLLexer(embedded, "embedded/html.xml")
