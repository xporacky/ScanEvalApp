package latex

import (
	"os"
	"strings"
	"testing"
	"time"

	"ScanEvalApp/internal/database/models"
)

func TestBuildTemplateDataEscapesLatexSpecialChars(t *testing.T) {
	student := models.Student{
		RegistrationNumber: 1234567,
		Name:               "Anna &\nEva",
		Surname:            "Mrkvova_#1",
		Room:               "A&B_105%\t",
	}

	exam := models.Exam{
		Title:         "Matika & Fyzika_1",
		ShowName:      true,
		Date:          time.Date(2026, 3, 21, 8, 0, 0, 0, time.UTC),
		QuestionCount: 15,
	}

	data := buildTemplateData(student, exam)

	if strings.Contains(data.Meno, " & ") {
		t.Fatalf("name was not escaped: %q", data.Meno)
	}

	if strings.Contains(data.Meno, "\n") || strings.Contains(data.Miestnost, "\t") {
		t.Fatalf("control characters remained in escaped template data: name=%q room=%q", data.Meno, data.Miestnost)
	}

	for _, got := range []string{data.Meno, data.Miestnost, data.TestName} {
		if strings.Contains(got, "&") && !strings.Contains(got, `\&`) {
			t.Fatalf("raw ampersand remained in template data: %q", got)
		}
	}

	if !strings.Contains(data.Miestnost, `\&`) || !strings.Contains(data.Miestnost, `\_`) || !strings.Contains(data.Miestnost, `\%`) {
		t.Fatalf("room was not escaped correctly: %q", data.Miestnost)
	}

	if !strings.Contains(data.TestName, `\&`) || !strings.Contains(data.TestName, `\_`) {
		t.Fatalf("title was not escaped correctly: %q", data.TestName)
	}

	if data.Cas != "08:00" {
		t.Fatalf("time was not populated correctly: %q", data.Cas)
	}

	for _, expected := range []string{"ID:1234567", `MENO:Anna \& Eva Mrkvova\_\#1`, "DATUM:21.03.2026"} {
		if !strings.Contains(data.QrCode, expected) {
			t.Fatalf("qr payload is missing %q: %q", expected, data.QrCode)
		}
	}
}

func TestReplaceTemplatePlaceholdersKeepsRenderedLatexSafe(t *testing.T) {
	templateContent, err := os.ReadFile("../../assets/latex/templates/template_5.tex")
	if err != nil {
		t.Fatalf("failed to read template: %v", err)
	}

	data := TemplateData{
		ID:        "1234567",
		IDDigits:  []string{"1", "2", "3", "4", "5", "6", "7"},
		Meno:      latexEscape("Anna & Eva"),
		ShowName:  true,
		Datum:     latexEscape("21. 03. 2026"),
		Miestnost: latexEscape("A&B_105%"),
		Cas:       "",
		Bloky:     15,
		QrCode:    "1234567",
		TestName:  latexEscape("Matika & Fyzika_1"),
	}

	rendered, err := ReplaceTemplatePlaceholders(templateContent, data)
	if err != nil {
		t.Fatalf("failed to render template: %v", err)
	}

	renderedStr := string(rendered)

	for _, raw := range []string{
		"Anna & Eva",
		"A&B_105%",
		"Matika & Fyzika_1",
	} {
		if strings.Contains(renderedStr, raw) {
			t.Fatalf("rendered template contains unescaped value %q", raw)
		}
	}

	for _, escaped := range []string{
		`Anna \& Eva`,
		`A\&B\_105\%`,
		`Matika \& Fyzika\_1`,
		`\generateMultipleBlocks{ 15 }`,
	} {
		if !strings.Contains(renderedStr, escaped) {
			t.Fatalf("rendered template is missing expected value %q", escaped)
		}
	}
}
