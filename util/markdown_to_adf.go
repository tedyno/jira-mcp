package util

import (
	"bytes"
	"strings"

	"github.com/ctreminiom/go-atlassian/pkg/infra/models"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
)

// MarkdownToADF converts a markdown string to an Atlassian Document Format (ADF) tree.
// It parses the markdown using goldmark and walks the AST to build the ADF structure.
func MarkdownToADF(markdown string) *models.CommentNodeScheme {
	source := []byte(markdown)

	md := goldmark.New(
		goldmark.WithExtensions(extension.Strikethrough),
	)
	reader := text.NewReader(source)
	doc := md.Parser().Parse(reader)

	adfDoc := &models.CommentNodeScheme{
		Version: 1,
		Type:    "doc",
	}

	for child := doc.FirstChild(); child != nil; child = child.NextSibling() {
		if node := convertNode(child, source); node != nil {
			adfDoc.Content = append(adfDoc.Content, node)
		}
	}

	// Ensure at least one paragraph so ADF is never empty
	if len(adfDoc.Content) == 0 {
		adfDoc.Content = []*models.CommentNodeScheme{
			{Type: "paragraph"},
		}
	}

	return adfDoc
}

// convertNode dispatches a goldmark AST node to the appropriate ADF builder.
func convertNode(n ast.Node, source []byte) *models.CommentNodeScheme {
	switch n.Kind() {
	case ast.KindParagraph:
		return convertParagraph(n, source)
	case ast.KindHeading:
		return convertHeading(n.(*ast.Heading), source)
	case ast.KindFencedCodeBlock:
		return convertFencedCodeBlock(n.(*ast.FencedCodeBlock), source)
	case ast.KindCodeBlock:
		return convertCodeBlock(n, source)
	case ast.KindBlockquote:
		return convertBlockquote(n, source)
	case ast.KindList:
		return convertList(n.(*ast.List), source)
	case ast.KindThematicBreak:
		return &models.CommentNodeScheme{Type: "rule"}
	default:
		// Fallback: try to render children as a paragraph
		return convertParagraph(n, source)
	}
}

func convertParagraph(n ast.Node, source []byte) *models.CommentNodeScheme {
	p := &models.CommentNodeScheme{Type: "paragraph"}
	p.Content = convertInlineChildren(n, source, nil)
	if len(p.Content) == 0 {
		return nil
	}
	return p
}

func convertHeading(n *ast.Heading, source []byte) *models.CommentNodeScheme {
	h := &models.CommentNodeScheme{
		Type:  "heading",
		Attrs: map[string]interface{}{"level": n.Level},
	}
	h.Content = convertInlineChildren(n, source, nil)
	return h
}

func convertFencedCodeBlock(n *ast.FencedCodeBlock, source []byte) *models.CommentNodeScheme {
	cb := &models.CommentNodeScheme{Type: "codeBlock"}

	lang := string(n.Language(source))
	if lang != "" {
		cb.Attrs = map[string]interface{}{"language": lang}
	}

	var buf bytes.Buffer
	lines := n.Lines()
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		buf.Write(seg.Value(source))
	}

	text := buf.String()
	if text != "" {
		cb.Content = []*models.CommentNodeScheme{
			{Type: "text", Text: text},
		}
	}

	return cb
}

func convertCodeBlock(n ast.Node, source []byte) *models.CommentNodeScheme {
	cb := &models.CommentNodeScheme{Type: "codeBlock"}

	var buf bytes.Buffer
	lines := n.Lines()
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		buf.Write(seg.Value(source))
	}

	text := buf.String()
	if text != "" {
		cb.Content = []*models.CommentNodeScheme{
			{Type: "text", Text: text},
		}
	}

	return cb
}

func convertBlockquote(n ast.Node, source []byte) *models.CommentNodeScheme {
	bq := &models.CommentNodeScheme{Type: "blockquote"}
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		if node := convertNode(child, source); node != nil {
			bq.Content = append(bq.Content, node)
		}
	}
	return bq
}

func convertList(n *ast.List, source []byte) *models.CommentNodeScheme {
	listType := "bulletList"
	if n.IsOrdered() {
		listType = "orderedList"
	}

	list := &models.CommentNodeScheme{Type: listType}

	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		if child.Kind() == ast.KindListItem {
			li := convertListItem(child, source)
			if li != nil {
				list.Content = append(list.Content, li)
			}
		}
	}

	return list
}

func convertListItem(n ast.Node, source []byte) *models.CommentNodeScheme {
	li := &models.CommentNodeScheme{Type: "listItem"}
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		if node := convertNode(child, source); node != nil {
			li.Content = append(li.Content, node)
		}
	}
	return li
}

// convertInlineChildren walks a block node's inline children and produces ADF text/hardBreak nodes.
// parentMarks are accumulated when descending into emphasis/strong nodes.
func convertInlineChildren(n ast.Node, source []byte, parentMarks []*models.MarkScheme) []*models.CommentNodeScheme {
	var result []*models.CommentNodeScheme

	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		nodes := convertInline(child, source, parentMarks)
		result = append(result, nodes...)
	}

	return result
}

func convertInline(n ast.Node, source []byte, parentMarks []*models.MarkScheme) []*models.CommentNodeScheme {
	switch n.Kind() {
	case ast.KindText:
		tn := n.(*ast.Text)
		seg := tn.Segment
		txt := string(seg.Value(source))
		if txt == "" {
			return nil
		}

		node := &models.CommentNodeScheme{
			Type: "text",
			Text: txt,
		}
		if len(parentMarks) > 0 {
			node.Marks = copyMarks(parentMarks)
		}

		result := []*models.CommentNodeScheme{node}

		// Handle soft/hard line breaks within text
		if tn.HardLineBreak() {
			result = append(result, &models.CommentNodeScheme{Type: "hardBreak"})
		} else if tn.SoftLineBreak() {
			result = append(result, &models.CommentNodeScheme{Type: "hardBreak"})
		}

		return result

	case ast.KindString:
		txt := string(n.Text(source))
		if txt == "" {
			return nil
		}
		node := &models.CommentNodeScheme{
			Type: "text",
			Text: txt,
		}
		if len(parentMarks) > 0 {
			node.Marks = copyMarks(parentMarks)
		}
		return []*models.CommentNodeScheme{node}

	case ast.KindEmphasis:
		em := n.(*ast.Emphasis)
		markType := "em"
		if em.Level == 2 {
			markType = "strong"
		}
		newMarks := append(copyMarks(parentMarks), &models.MarkScheme{Type: markType})
		return convertInlineChildren(n, source, newMarks)

	case ast.KindCodeSpan:
		var buf bytes.Buffer
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			if tn, ok := child.(*ast.Text); ok {
				buf.Write(tn.Segment.Value(source))
			}
		}
		txt := buf.String()
		if txt == "" {
			return nil
		}
		marks := append(copyMarks(parentMarks), &models.MarkScheme{Type: "code"})
		return []*models.CommentNodeScheme{
			{Type: "text", Text: txt, Marks: marks},
		}

	case ast.KindLink:
		link := n.(*ast.Link)
		href := string(link.Destination)
		mark := &models.MarkScheme{
			Type:  "link",
			Attrs: map[string]interface{}{"href": href},
		}
		newMarks := append(copyMarks(parentMarks), mark)
		return convertInlineChildren(n, source, newMarks)

	case ast.KindAutoLink:
		al := n.(*ast.AutoLink)
		url := string(al.URL(source))
		mark := &models.MarkScheme{
			Type:  "link",
			Attrs: map[string]interface{}{"href": url},
		}
		marks := append(copyMarks(parentMarks), mark)
		return []*models.CommentNodeScheme{
			{Type: "text", Text: url, Marks: marks},
		}

	case ast.KindImage:
		img := n.(*ast.Image)
		href := string(img.Destination)
		altText := strings.TrimSpace(string(n.Text(source)))
		if altText == "" {
			altText = href
		}
		mark := &models.MarkScheme{
			Type:  "link",
			Attrs: map[string]interface{}{"href": href},
		}
		marks := append(copyMarks(parentMarks), mark)
		return []*models.CommentNodeScheme{
			{Type: "text", Text: altText, Marks: marks},
		}

	default:
		// Check GFM extension nodes
		if n.Kind() == extast.KindStrikethrough {
			newMarks := append(copyMarks(parentMarks), &models.MarkScheme{Type: "strike"})
			return convertInlineChildren(n, source, newMarks)
		}
		// Unknown inline — try to render children
		return convertInlineChildren(n, source, parentMarks)
	}
}

func copyMarks(marks []*models.MarkScheme) []*models.MarkScheme {
	if len(marks) == 0 {
		return nil
	}
	out := make([]*models.MarkScheme, len(marks))
	copy(out, marks)
	return out
}
