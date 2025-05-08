package common

const (
	// Specifies the permission mode used for generated files (read/write for owner, read-only for group and others).
	FILE_PERMISSION = 0644

	// Path to the main LaTeX template used for PDF generation
	TEMPLATE_PATH = "./assets/latex/main.tex"

	// Temporary directory used during PDF generation for LaTeX compiling purposes
	TEMPORARY_PDF_PATH = "./assets/tmp"

	// Represents a value which means no question number was found
	QUESTION_NUMBER_NOT_FOUND = -1
)

// TODO - prerobit nejako, aby sa vracali normalne spravy, ked nastane error, toto je zatial ako SABLONA
const (
	SUCCESS         = 0
	FILE_NOT_FOUND  = 1
	INVALID_FORMAT  = 2
	PROCESSING_FAIL = 3
)
