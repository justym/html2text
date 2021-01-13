package html2text

import (
	"bytes"
	"github.com/ssor/bom"
	"golang.org/x/net/html"
	"golang.org/x/xerrors"
	"io"
)

type Handler func(textBuffer bytes.Buffer, node *html.Node) error

type Context struct {
	Node       *html.Node
	HTMLBuffer bytes.Buffer
	TextBuffer bytes.Buffer
	Handlers   map[html.NodeType]Handler
}

//FIXME: Functional Options Pattern ぽくする (Handlersを対象に)
// Refs: https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
func NewContext(r io.Reader) (*Context, error) {
	readerWithoutBOM, err := bom.NewReaderWithoutBom(r)
	if err != nil {
		return nil, xerrors.Errorf(": %w", err)
	}

	node, err := html.Parse(readerWithoutBOM)
	if err != nil {
		return nil, xerrors.Errorf(": %w", err)
	}

	ctx := &Context{
		Node: node,
	}

	_, err = ctx.HTMLBuffer.ReadFrom(readerWithoutBOM)
	if err != nil {
		return nil, xerrors.Errorf(": %w", err)
	}

	return ctx, nil
}

//FIXME: Functional Options Pattern ぽくする
func (ctx *Context) SetHandlers(nodeType html.NodeType, handler Handler) {
	ctx.Handlers[nodeType] = handler
}

//TODO: traverse中にContext.TextBufferに書き込むようにする。(現状はtraverseしているだけのようにみえる)
func (ctx *Context) Convert() (string, error) {
	// ctx.Handlers should be set at least one handler, if you want to Convert.
	// If not, return html
	if ctx.Handlers == nil {
		return ctx.HTMLBuffer.String(), nil
	}

	ctx.traverse(ctx.Node)

	return "", nil
}

// ctx.Handlersに設定してあるHandlerがあればそれを使ってtraverseする。なければ、そのまま子要素へtraverseする
func (ctx *Context) traverse(node *html.Node) error {

	if ctx.Handlers[node.Type] == nil {
		return ctx.traverseChildren(node)
	}

	return ctx.Handlers[node.Type](ctx.TextBuffer, node)
}

// ctx.Nodeを最後までtraverseしていく
func (ctx *Context) traverseChildren(node *html.Node) error {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if err := ctx.traverse(child); err != nil {
			return xerrors.Errorf(": %w", err)
		}
	}

	return nil
}

// 書き込み用の関数として用意したけど使うかわからない
func render(buf bytes.Buffer, node *html.Node) error {
	writer := io.Writer(&buf)

	return html.Render(writer, node)
}
