package latex

const (
	// Specifies the permission mode used for generated files (read/write for owner, read-only for group and others).
	FILE_PERMISSION = 0644

	// Path to the main LaTeX template used for PDF generation
	TEMPLATE_PATH = "./assets/latex/main.tex"

	// Directory where generated PDF files are stored
	OUTPUT_PDF_PATH = "./assets/tmp"

	// Temporary directory used during PDF generation
	TEMPORARY_PDF_PATH = "./assets/tmp"
)
