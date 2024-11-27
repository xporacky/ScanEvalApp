package latex

// štruktúra na prácu s nahrádzaním hodnôt v šablóne
type TemplateData struct {
	ID        string
	Meno      string
	Datum     string
	Miestnost string
	Cas       string
	Bloky     int
	QrCode    string
}
