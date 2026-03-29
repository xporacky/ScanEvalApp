package latex

// TemplateData represents the data used for replacing placeholders in a LaTeX template.
type TemplateData struct {
	ID        string
	IDDigits  []string
	Meno      string
	ShowName  bool
	Datum     string
	Miestnost string
	Cas       string
	Bloky     int
	QrCode    string
	TestName  string
}
