package scanprocessing

const NUMBER_OF_CHOICES = 6
const NUMBER_OF_QUESTIONS_PER_PAGE = 15

const TEMP_IMAGE_PATH = "./assets/tmp/temp-image.png"
const TEMP_HEADER_IMAGE_PATH = "./assets/tmp/temp-header-image.png"
const PADDING = 10
const CHECKBOX_AREA_PADDING = -5
const CHECKBOX_PADDING = 5
const BORDER_RECTANGLE_AREA_SIZE = 1000000
const ANSWER_SQUARE_MIN_AREA_SIZE = 700
const ANSWER_SQUARE_MAX_AREA_SIZE = 1800
const ANSWER_SQUARE_MIN_AREA_SIZE_8 = 700
const ANSWER_SQUARE_MAX_AREA_SIZE_8 = 1800
const CONFIGS_DIR = "./configs/"

const (
	StateXFound      = iota
	StateCircleFound = iota
	StateEmpty       = iota
)
