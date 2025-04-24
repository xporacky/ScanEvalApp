package scanprocessing

const NUMBER_OF_CHOICES = 5
const NUMBER_OF_QUESTIONS_PER_PAGE = 20

// Paths for temporary image storage are also provided for intermediate OCR and visualization steps.
const TEMP_IMAGE_PATH = "./assets/tmp/temp-image.png"
const TEMP_HEADER_IMAGE_PATH = "./assets/tmp/temp-header-image.png"
const PADDING = 10
const CHECKBOX_AREA_PADDING = -5
const CHECKBOX_PADDING = 5
const BORDER_RECTANGLE_AREA_SIZE = 1000000
const ANSWER_SQUARE_MIN_AREA_SIZE = 1300
const ANSWER_SQUARE_MAX_AREA_SIZE = 2300
const MEAN_INTENSITY_X_LOWEST = 160
const MEAN_INTENSITY_X_HIGHEST = 220

// The states defined by iota represent internal processing flags used when analyzing answers:
//   - StateXFound: Indicates an 'X' has been detected in a checkbox.
//   - StateCircleFound: Indicates a filled circle or valid answer shape has been found.
//   - StateEmpty: Default state when no marking has been detected yet.
const (
	StateXFound      = iota
	StateCircleFound = iota
	StateEmpty       = iota
)
