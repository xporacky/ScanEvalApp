package latex

// TemplateData represents the data used for replacing placeholders in a LaTeX template.
type TemplateData struct {
ID        string
Meno      string
Datum     string
Miestnost string
Cas       string
Bloky     int
QrCode    string
}
