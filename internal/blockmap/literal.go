package blockmap

import (
	"bytes"
	"regexp"

	"github.com/yuin/goldmark/ast"
)

// templateOpenerRE matches an MDX expression that opens with a template
// literal: `{` then a backtick, as in code={`…`} or <CodeBlock>{`…`}.
var templateOpenerRE = regexp.MustCompile("\\{\\s*`")

// literalChain freezes the paragraphs of an MDX template literal that holds
// blank lines.
//
// A code sample passed to a component as a template literal, as an attribute
// (code={`…`}) or as children (<CodeBlock>{`…`}</CodeBlock>), often holds
// blank lines. goldmark ends a paragraph at each one, so the literal's text
// becomes paragraphs that read as prose, and reflowing them rewrites the code
// the page shows. A paragraph that opens such an expression and leaves the
// backtick count odd starts a chain; the chain then counts unescaped backticks
// in the raw source, and while the count is odd, each paragraph that starts
// there is inside the literal. The paragraph that opens the chain and every
// paragraph inside it are frozen (SkipTemplateLiteral). The first paragraph
// that starts with the count even ends the chain and is handled as usual.
//
// Only backticks are counted, not JavaScript or JSX syntax. A line that opens
// with an unescaped run of three backticks cannot be inside a real template
// literal, where it would have to be escaped, so it ends the chain: a count
// thrown off by a backtick in a quoted string then stops at the next code
// fence instead of running on out of phase. A blank line elsewhere in an
// expression (between attributes, inside an object literal) keeps the old
// handling. The chain is idempotent: every paragraph it reads is frozen, and
// the blocks between them never reflow.
type literalChain struct {
	active bool
	pos    int  // source offset the count has reached
	odd    bool // the count of unescaped backticks up to pos is odd
}

// advance counts unescaped backticks in source[lc.pos:to]. It reports false,
// leaving the count meaningless, if it meets a line that opens with an
// unescaped code fence.
func (lc *literalChain) advance(source []byte, to int) bool {
	for ; lc.pos < to; lc.pos++ {
		if lc.pos > 0 && source[lc.pos-1] == '\n' && opensWithFence(source[lc.pos:]) {
			return false
		}
		if source[lc.pos] == '`' && !backslashEscaped(source, lc.pos) {
			lc.odd = !lc.odd
		}
	}
	return true
}

// inside reports whether paragraph p, the next one in document order, opens
// or lies inside a template literal that holds a blank line, and so must be
// frozen. collect also passes HTML blocks, which never reflow but can open a
// literal, and ignores the answer.
func (lc *literalChain) inside(p ast.Node, source []byte) bool {
	lines := p.Lines()
	if lines.Len() == 0 {
		return false
	}
	start, stop := lines.At(0).Start, lines.At(lines.Len()-1).Stop
	if lc.active && lc.advance(source, start) && lc.odd {
		return true
	}
	*lc = literalChain{}
	if !templateOpenerRE.Match(source[start:stop]) {
		return false
	}
	*lc = literalChain{active: true, pos: start}
	if lc.advance(source, stop) && lc.odd {
		return true
	}
	*lc = literalChain{}
	return false
}

// opensWithFence reports whether line, after leading spaces and tabs, opens
// with a run of three or more backticks.
func opensWithFence(line []byte) bool {
	return bytes.HasPrefix(bytes.TrimLeft(line, " \t"), []byte("```"))
}

// backslashEscaped reports whether the byte at source[i] is preceded by an
// odd run of backslashes.
func backslashEscaped(source []byte, i int) bool {
	n := 0
	for j := i - 1; j >= 0 && source[j] == '\\'; j-- {
		n++
	}
	return n%2 == 1
}
