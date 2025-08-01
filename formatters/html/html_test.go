package html

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
	"testing"

	assert "github.com/alecthomas/assert/v2"

	"github.com/Chi-Iroh/chroma-test-gitea/v2"
	"github.com/Chi-Iroh/chroma-test-gitea/v2/lexers"
	"github.com/Chi-Iroh/chroma-test-gitea/v2/styles"
)

func TestCompressStyle(t *testing.T) {
	style := "color: #888888; background-color: #faffff"
	actual := compressStyle(style)
	expected := "color:#888;background-color:#faffff"
	assert.Equal(t, expected, actual)
}

func BenchmarkHTMLFormatter(b *testing.B) {
	formatter := New()
	b.ResetTimer()
	for range b.N {
		it, err := lexers.Get("go").Tokenise(nil, "package main\nfunc main()\n{\nprintln(`hello world`)\n}\n")
		assert.NoError(b, err)
		err = formatter.Format(io.Discard, styles.Fallback, it)
		assert.NoError(b, err)
	}
}

func TestSplitTokensIntoLines(t *testing.T) {
	in := []chroma.Token{
		{Value: "hello", Type: chroma.NameKeyword},
		{Value: " world\nwhat?\n", Type: chroma.NameKeyword},
	}
	expected := [][]chroma.Token{
		{
			{Type: chroma.NameKeyword, Value: "hello"},
			{Type: chroma.NameKeyword, Value: " world\n"},
		},
		{
			{Type: chroma.NameKeyword, Value: "what?\n"},
		},
	}
	actual := chroma.SplitTokensIntoLines(in)
	assert.Equal(t, expected, actual)
}

func TestFormatterStyleToCSS(t *testing.T) {
	builder := styles.Get("github").Builder()
	builder.Add(chroma.LineHighlight, "bg:#ffffcc")
	builder.Add(chroma.LineNumbers, "bold")
	style, err := builder.Build()
	if err != nil {
		t.Error(err)
	}
	formatter := New(WithClasses(true))
	css := formatter.styleToCSS(style)
	for _, s := range css {
		if strings.HasPrefix(strings.TrimSpace(s), ";") {
			t.Errorf("rule starts with semicolon - expected valid css rule without semicolon: %v", s)
		}
	}
}

func TestClassPrefix(t *testing.T) {
	wantPrefix := "some-prefix-"
	withPrefix := New(WithClasses(true), ClassPrefix(wantPrefix))
	noPrefix := New(WithClasses(true))
	for st := range chroma.StandardTypes {
		if noPrefix.class(st) == "" {
			if got := withPrefix.class(st); got != "" {
				t.Errorf("Formatter.class(%v): prefix shouldn't be added to empty classes", st)
			}
		} else if got := withPrefix.class(st); !strings.HasPrefix(got, wantPrefix) {
			t.Errorf("Formatter.class(%v): %q should have a class prefix", st, got)
		}
	}

	var styleBuf bytes.Buffer
	err := withPrefix.WriteCSS(&styleBuf, styles.Fallback)
	assert.NoError(t, err)
	if !strings.Contains(styleBuf.String(), ".some-prefix-chroma ") {
		t.Error("Stylesheets should have a class prefix")
	}
}

func TestTableLineNumberNewlines(t *testing.T) {
	f := New(WithClasses(true), WithLineNumbers(true), LineNumbersInTable(true))
	it, err := lexers.Get("go").Tokenise(nil, "package main\nfunc main()\n{\nprintln(`hello world`)\n}\n")
	assert.NoError(t, err)

	var buf bytes.Buffer
	err = f.Format(&buf, styles.Fallback, it)
	assert.NoError(t, err)

	// Don't bother testing the whole output, just verify it's got line numbers
	// in a <pre>-friendly format.
	// Note: placing the newlines inside the <span> lets browser selections look
	// better, instead of "skipping" over the span margin.
	assert.Contains(t, buf.String(), `<span class="lnt">2
</span><span class="lnt">3
</span><span class="lnt">4
</span>`)
}

func TestTabWidthStyle(t *testing.T) {
	f := New(TabWidth(4), WithClasses(false))
	it, err := lexers.Get("bash").Tokenise(nil, "echo FOO")
	assert.NoError(t, err)

	var buf bytes.Buffer
	err = f.Format(&buf, styles.Fallback, it)
	assert.NoError(t, err)

	assert.True(t, regexp.MustCompile(`<pre.*style=".*background-color:[^;]+;-moz-tab-size:4;-o-tab-size:4;tab-size:4;[^"]*".+`).MatchString(buf.String()))
}

func TestWithCustomCSS(t *testing.T) {
	f := New(WithClasses(false), WithCustomCSS(map[chroma.TokenType]string{chroma.Line: `display: inline;`}))
	it, err := lexers.Get("bash").Tokenise(nil, "echo FOO")
	assert.NoError(t, err)

	var buf bytes.Buffer
	err = f.Format(&buf, styles.Fallback, it)
	assert.NoError(t, err)

	assert.True(t, regexp.MustCompile(`<span style="display:flex;display:inline;"><span><span style=".*">echo</span> FOO</span></span>`).MatchString(buf.String()))
}

func TestWithCustomCSSStyleInheritance(t *testing.T) {
	f := New(WithClasses(false), WithCustomCSS(map[chroma.TokenType]string{
		chroma.String:              `background: blue;`,
		chroma.LiteralStringDouble: `color: tomato;`,
	}))
	it, err := lexers.Get("bash").Tokenise(nil, `echo "FOO"`)
	assert.NoError(t, err)

	var buf bytes.Buffer
	err = f.Format(&buf, styles.Fallback, it)
	assert.NoError(t, err)

	assert.True(t, regexp.MustCompile(` <span style=".*;background:blue;color:tomato;">&#34;FOO&#34;</span>`).MatchString(buf.String()))
}

func TestWrapLongLines(t *testing.T) {
	f := New(WithClasses(false), WrapLongLines(true))
	it, err := lexers.Get("go").Tokenise(nil, "package main\nfunc main()\n{\nprintln(\"hello world\")\n}\n")
	assert.NoError(t, err)

	var buf bytes.Buffer
	err = f.Format(&buf, styles.Fallback, it)
	assert.NoError(t, err)

	assert.True(t, regexp.MustCompile(`<pre.*style=".*white-space:pre-wrap;word-break:break-word;`).MatchString(buf.String()))
}

func TestHighlightLines(t *testing.T) {
	f := New(WithClasses(true), HighlightLines([][2]int{{4, 5}}))
	it, err := lexers.Get("go").Tokenise(nil, "package main\nfunc main()\n{\nprintln(\"hello world\")\n}\n")
	assert.NoError(t, err)

	var buf bytes.Buffer
	err = f.Format(&buf, styles.Fallback, it)
	assert.NoError(t, err)

	assert.Contains(t, buf.String(), `<span class="line hl"><span class="cl">`)
}

func TestLineNumbers(t *testing.T) {
	f := New(WithClasses(true), WithLineNumbers(true))
	it, err := lexers.Get("bash").Tokenise(nil, "echo FOO")
	assert.NoError(t, err)

	var buf bytes.Buffer
	err = f.Format(&buf, styles.Fallback, it)
	assert.NoError(t, err)

	assert.Contains(t, buf.String(), `<span class="line"><span class="ln">1</span><span class="cl"><span class="nb">echo</span> FOO</span></span>`)
}

func TestPreWrapper(t *testing.T) {
	f := New(Standalone(true), WithClasses(true))
	it, err := lexers.Get("bash").Tokenise(nil, "echo FOO")
	assert.NoError(t, err)

	var buf bytes.Buffer
	err = f.Format(&buf, styles.Fallback, it)
	assert.NoError(t, err)

	assert.True(t, regexp.MustCompile("<body class=\"bg\">\n<pre.*class=\"chroma\"><code><span class=\"line\"><span class=\"cl\"><span class=\"nb\">echo</span> FOO</span></span></code></pre>\n</body>\n</html>").MatchString(buf.String()))
	assert.True(t, regexp.MustCompile(`\.bg { .+ }`).MatchString(buf.String()))
	assert.True(t, regexp.MustCompile(`\.chroma { .+ }`).MatchString(buf.String()))
}

func TestLinkeableLineNumbers(t *testing.T) {
	f := New(WithClasses(true), WithLineNumbers(true), WithLinkableLineNumbers(true, "line"), WithClasses(false))
	it, err := lexers.Get("go").Tokenise(nil, "package main\nfunc main()\n{\nprintln(\"hello world\")\n}\n")
	assert.NoError(t, err)

	var buf bytes.Buffer
	err = f.Format(&buf, styles.Fallback, it)
	assert.NoError(t, err)

	assert.Contains(t, buf.String(), `id="line1"><a style="outline:none;text-decoration:none;color:inherit" href="#line1">1</a>`)
	assert.Contains(t, buf.String(), `id="line5"><a style="outline:none;text-decoration:none;color:inherit" href="#line5">5</a>`)
}

func TestTableLinkeableLineNumbers(t *testing.T) {
	f := New(Standalone(true), WithClasses(true), WithLineNumbers(true), LineNumbersInTable(true), WithLinkableLineNumbers(true, "line"))
	it, err := lexers.Get("go").Tokenise(nil, "package main\nfunc main()\n{\nprintln(`hello world`)\n}\n")
	assert.NoError(t, err)

	var buf bytes.Buffer
	err = f.Format(&buf, styles.Fallback, it)
	assert.NoError(t, err)

	assert.Contains(t, buf.String(), `id="line1"><a class="lnlinks" href="#line1">1</a>`)
	assert.Contains(t, buf.String(), `id="line5"><a class="lnlinks" href="#line5">5</a>`)
	assert.Contains(t, buf.String(), `/* LineLink */ .chroma .lnlinks { outline: none; text-decoration: none; color: inherit }`, buf.String())
}

func TestTableLineNumberSpacing(t *testing.T) {
	testCases := []struct {
		baseLineNumber int
		expectedBuf    string
	}{{
		7,
		`<span class="lnt"> 7
</span><span class="lnt"> 8
</span><span class="lnt"> 9
</span><span class="lnt">10
</span><span class="lnt">11
</span>`,
	}, {
		6,
		`<span class="lnt"> 6
</span><span class="lnt"> 7
</span><span class="lnt"> 8
</span><span class="lnt"> 9
</span><span class="lnt">10
</span>`,
	}, {
		5,
		`<span class="lnt">5
</span><span class="lnt">6
</span><span class="lnt">7
</span><span class="lnt">8
</span><span class="lnt">9
</span>`,
	}}
	for i, testCase := range testCases {
		f := New(
			WithClasses(true),
			WithLineNumbers(true),
			LineNumbersInTable(true),
			BaseLineNumber(testCase.baseLineNumber),
		)
		it, err := lexers.Get("go").Tokenise(nil, "package main\nfunc main()\n{\nprintln(`hello world`)\n}\n")
		assert.NoError(t, err)
		var buf bytes.Buffer
		err = f.Format(&buf, styles.Fallback, it)
		assert.NoError(t, err, "Test Case %d", i)
		assert.Contains(t, buf.String(), testCase.expectedBuf, "Test Case %d", i)
	}
}

func TestWithPreWrapper(t *testing.T) {
	wrapper := preWrapper{
		start: func(code bool, styleAttr string) string {
			return fmt.Sprintf("<foo%s id=\"code-%t\">", styleAttr, code)
		},
		end: func(code bool) string {
			return "</foo>"
		},
	}

	format := func(f *Formatter) string {
		it, err := lexers.Get("bash").Tokenise(nil, "echo FOO")
		assert.NoError(t, err)

		var buf bytes.Buffer
		err = f.Format(&buf, styles.Fallback, it)
		assert.NoError(t, err)

		return buf.String()
	}

	t.Run("Regular", func(t *testing.T) {
		s := format(New(WithClasses(true)))
		assert.Equal(t, s, `<pre class="chroma"><code><span class="line"><span class="cl"><span class="nb">echo</span> FOO</span></span></code></pre>`)
	})

	t.Run("PreventSurroundingPre", func(t *testing.T) {
		s := format(New(PreventSurroundingPre(true), WithClasses(true)))
		assert.Equal(t, s, `<span class="nb">echo</span> FOO`)
	})

	t.Run("InlineCode", func(t *testing.T) {
		s := format(New(InlineCode(true), WithClasses(true)))
		assert.Equal(t, s, `<code class="chroma"><span class="nb">echo</span> FOO</code>`)
	})

	t.Run("InlineCode, inline styles", func(t *testing.T) {
		s := format(New(InlineCode(true)))
		assert.True(t, regexp.MustCompile(`<code style=".+?"><span style=".+?">echo</span> FOO</code>`).MatchString(s))
	})

	t.Run("Wrapper", func(t *testing.T) {
		s := format(New(WithPreWrapper(wrapper), WithClasses(true)))
		assert.Equal(t, s, `<foo class="chroma" id="code-true"><span class="line"><span class="cl"><span class="nb">echo</span> FOO</span></span></foo>`)
	})

	t.Run("Wrapper, LineNumbersInTable", func(t *testing.T) {
		s := format(New(WithPreWrapper(wrapper), WithClasses(true), WithLineNumbers(true), LineNumbersInTable(true)))

		assert.Equal(t, s, `<div class="chroma">
<table class="lntable"><tr><td class="lntd">
<foo class="chroma" id="code-false"><span class="lnt">1
</span></foo></td>
<td class="lntd">
<foo class="chroma" id="code-true"><span class="line"><span class="cl"><span class="nb">echo</span> FOO</span></span></foo></td></tr></table>
</div>
`)
	})
}

func TestReconfigureOptions(t *testing.T) {
	options := []Option{
		WithClasses(true),
		WithLineNumbers(true),
	}

	options = append(options, WithLineNumbers(false))

	f := New(options...)

	it, err := lexers.Get("bash").Tokenise(nil, "echo FOO")
	assert.NoError(t, err)

	var buf bytes.Buffer
	err = f.Format(&buf, styles.Fallback, it)

	assert.NoError(t, err)
	assert.Equal(t, `<pre class="chroma"><code><span class="line"><span class="cl"><span class="nb">echo</span> FOO</span></span></code></pre>`, buf.String())
}

func TestWriteCssWithAllClasses(t *testing.T) {
	formatter := New(WithAllClasses(true))

	var buf bytes.Buffer
	err := formatter.WriteCSS(&buf, styles.Fallback)

	assert.NoError(t, err)
	assert.NotContains(t, buf.String(), ".chroma . {", "Generated css doesn't contain invalid css")
}

func TestStyleCache(t *testing.T) {
	f := New()

	assert.True(t, len(styles.Registry) > styleCacheLimit)

	for _, style := range styles.Registry {
		var buf bytes.Buffer
		err := f.WriteCSS(&buf, style)
		assert.NoError(t, err)
	}

	assert.Equal(t, styleCacheLimit, len(f.styleCache.cache))
}
