package latex

// TemplateData represents the data used for replacing placeholders in a LaTeX template.
// Each field corresponds to a piece of data that will be inserted into the template.
// This structure is used to dynamically populate the LaTeX template during PDF generation.
type TemplateData struct {
	// ID represents the unique identifier for the document or the entity.
	ID string

	// Meno represents the name associated with the document or the entity.
	Meno string

	// Datum represents the date when the document was generated or when the event occurred.
	Datum string

	// Miestnost represents the room or location associated with the event or meeting.
	Miestnost string

	// Cas represents the time of the event or meeting.
	Cas string

	// Bloky represents the number of blocks or sessions in the event.
	Bloky int

	// QrCode represents the data or image path for a QR code, which can be embedded into the document.
	QrCode string
}
