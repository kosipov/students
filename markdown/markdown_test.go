package markdown

import (
	"strings"
	"testing"
)

func TestRenderTable(t *testing.T) {
	html, err := Render([]byte("| № | Тема |\n|---|---|\n| 1 | Интернет-магазин |\n"))
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	for _, want := range []string{"<table>", "<th>Тема</th>", "<td>Интернет-магазин</td>"} {
		if !strings.Contains(string(html), want) {
			t.Errorf("Render() = %s, want it to contain %s", html, want)
		}
	}
}

func TestRenderOmitsUnsafeContent(t *testing.T) {
	source := "<script>alert(1)</script>\n\n" +
		"<img src=x onerror=alert(1)>\n\n" +
		"[ссылка](javascript:alert(1))\n"

	html, err := Render([]byte(source))
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	for _, unsafe := range []string{"<script", "onerror", "javascript:"} {
		if strings.Contains(string(html), unsafe) {
			t.Errorf("Render() = %s, must not contain %s", html, unsafe)
		}
	}
}
