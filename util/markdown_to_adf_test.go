package util

import (
	"encoding/json"
	"testing"

	"github.com/ctreminiom/go-atlassian/pkg/infra/models"
)

func TestMarkdownToADF_Headings(t *testing.T) {
	md := "# Heading 1\n\n## Heading 2\n\n### Heading 3\n"
	doc := MarkdownToADF(md)

	assertNodeType(t, doc, "doc")
	assertContentLen(t, doc, 3)

	for i, level := range []int{1, 2, 3} {
		h := doc.Content[i]
		assertNodeType(t, h, "heading")
		gotLevel, ok := h.Attrs["level"].(int)
		if !ok {
			t.Fatalf("heading[%d] level is not int", i)
		}
		if gotLevel != level {
			t.Errorf("heading[%d] level = %d, want %d", i, gotLevel, level)
		}
	}
}

func TestMarkdownToADF_BoldItalicCode(t *testing.T) {
	md := "This is **bold** and *italic* and `code`.\n"
	doc := MarkdownToADF(md)

	assertNodeType(t, doc, "doc")
	assertContentLen(t, doc, 1)

	p := doc.Content[0]
	assertNodeType(t, p, "paragraph")

	var bold, italic, code bool
	for _, child := range p.Content {
		for _, mark := range child.Marks {
			switch mark.Type {
			case "strong":
				bold = true
				if child.Text != "bold" {
					t.Errorf("strong text = %q, want %q", child.Text, "bold")
				}
			case "em":
				italic = true
				if child.Text != "italic" {
					t.Errorf("em text = %q, want %q", child.Text, "italic")
				}
			case "code":
				code = true
				if child.Text != "code" {
					t.Errorf("code text = %q, want %q", child.Text, "code")
				}
			}
		}
	}
	if !bold {
		t.Error("expected bold mark, not found")
	}
	if !italic {
		t.Error("expected italic mark, not found")
	}
	if !code {
		t.Error("expected code mark, not found")
	}
}

func TestMarkdownToADF_BulletList(t *testing.T) {
	md := "- item one\n- item two\n- item three\n"
	doc := MarkdownToADF(md)

	assertNodeType(t, doc, "doc")
	assertContentLen(t, doc, 1)

	list := doc.Content[0]
	assertNodeType(t, list, "bulletList")
	assertContentLen(t, list, 3)

	for _, li := range list.Content {
		assertNodeType(t, li, "listItem")
	}
}

func TestMarkdownToADF_OrderedList(t *testing.T) {
	md := "1. first\n2. second\n3. third\n"
	doc := MarkdownToADF(md)

	assertNodeType(t, doc, "doc")
	assertContentLen(t, doc, 1)

	list := doc.Content[0]
	assertNodeType(t, list, "orderedList")
	assertContentLen(t, list, 3)
}

func TestMarkdownToADF_FencedCodeBlock(t *testing.T) {
	md := "```sql\nSELECT * FROM users;\n```\n"
	doc := MarkdownToADF(md)

	assertNodeType(t, doc, "doc")
	assertContentLen(t, doc, 1)

	cb := doc.Content[0]
	assertNodeType(t, cb, "codeBlock")

	lang, ok := cb.Attrs["language"].(string)
	if !ok || lang != "sql" {
		t.Errorf("codeBlock language = %q, want %q", lang, "sql")
	}

	if len(cb.Content) == 0 {
		t.Fatal("codeBlock has no content")
	}
	if cb.Content[0].Text != "SELECT * FROM users;\n" {
		t.Errorf("codeBlock text = %q, want %q", cb.Content[0].Text, "SELECT * FROM users;\n")
	}
}

func TestMarkdownToADF_FencedCodeBlockNoLang(t *testing.T) {
	md := "```\nhello world\n```\n"
	doc := MarkdownToADF(md)

	cb := doc.Content[0]
	assertNodeType(t, cb, "codeBlock")

	if cb.Attrs != nil {
		t.Errorf("codeBlock without language should have nil attrs, got %v", cb.Attrs)
	}
}

func TestMarkdownToADF_HorizontalRule(t *testing.T) {
	md := "Above\n\n---\n\nBelow\n"
	doc := MarkdownToADF(md)

	var foundRule bool
	for _, node := range doc.Content {
		if node.Type == "rule" {
			foundRule = true
		}
	}
	if !foundRule {
		t.Error("expected rule node, not found")
	}
}

func TestMarkdownToADF_Blockquote(t *testing.T) {
	md := "> This is a quote\n"
	doc := MarkdownToADF(md)

	assertContentLen(t, doc, 1)

	bq := doc.Content[0]
	assertNodeType(t, bq, "blockquote")

	if len(bq.Content) == 0 {
		t.Fatal("blockquote has no content")
	}
	assertNodeType(t, bq.Content[0], "paragraph")
}

func TestMarkdownToADF_Link(t *testing.T) {
	md := "Visit [Google](https://google.com) now.\n"
	doc := MarkdownToADF(md)

	p := doc.Content[0]
	assertNodeType(t, p, "paragraph")

	var foundLink bool
	for _, child := range p.Content {
		for _, mark := range child.Marks {
			if mark.Type == "link" {
				foundLink = true
				href, _ := mark.Attrs["href"].(string)
				if href != "https://google.com" {
					t.Errorf("link href = %q, want %q", href, "https://google.com")
				}
				if child.Text != "Google" {
					t.Errorf("link text = %q, want %q", child.Text, "Google")
				}
			}
		}
	}
	if !foundLink {
		t.Error("expected link mark, not found")
	}
}

func TestMarkdownToADF_MixedContent(t *testing.T) {
	md := `# Title

Some **bold** paragraph.

- bullet one
- bullet two

` + "```go\nfmt.Println(\"hello\")\n```\n"

	doc := MarkdownToADF(md)
	assertNodeType(t, doc, "doc")

	types := make([]string, len(doc.Content))
	for i, node := range doc.Content {
		types[i] = node.Type
	}

	expected := []string{"heading", "paragraph", "bulletList", "codeBlock"}
	if len(types) != len(expected) {
		t.Fatalf("content types = %v, want %v", types, expected)
	}
	for i, typ := range expected {
		if types[i] != typ {
			t.Errorf("content[%d] type = %q, want %q", i, types[i], typ)
		}
	}
}

func TestMarkdownToADF_PlainText(t *testing.T) {
	md := "Just a plain text without any markdown."
	doc := MarkdownToADF(md)

	assertNodeType(t, doc, "doc")
	assertContentLen(t, doc, 1)
	assertNodeType(t, doc.Content[0], "paragraph")

	p := doc.Content[0]
	if len(p.Content) == 0 {
		t.Fatal("paragraph has no text nodes")
	}
	if p.Content[0].Text != "Just a plain text without any markdown." {
		t.Errorf("text = %q, want %q", p.Content[0].Text, "Just a plain text without any markdown.")
	}
}

func TestMarkdownToADF_EmptyString(t *testing.T) {
	doc := MarkdownToADF("")
	assertNodeType(t, doc, "doc")
	if len(doc.Content) == 0 {
		t.Fatal("expected at least one content node for empty input")
	}
}

func TestMarkdownToADF_ValidJSON(t *testing.T) {
	md := "# Test\n\nParagraph with **bold**.\n\n- item\n\n```js\nconsole.log('hi');\n```\n"
	doc := MarkdownToADF(md)

	_, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("failed to marshal ADF to JSON: %v", err)
	}
}

// --- helpers ---

func assertNodeType(t *testing.T, node *models.CommentNodeScheme, expectedType string) {
	t.Helper()
	if node == nil {
		t.Fatalf("expected node of type %q, got nil", expectedType)
	}
	if node.Type != expectedType {
		t.Errorf("node type = %q, want %q", node.Type, expectedType)
	}
}

func assertContentLen(t *testing.T, node *models.CommentNodeScheme, expectedLen int) {
	t.Helper()
	if node == nil {
		t.Fatalf("expected node with %d children, got nil", expectedLen)
	}
	if len(node.Content) != expectedLen {
		t.Errorf("content length = %d, want %d", len(node.Content), expectedLen)
	}
}
