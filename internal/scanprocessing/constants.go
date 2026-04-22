package scanprocessing

const NUMBER_OF_CHOICES = 6
const NUMBER_OF_QUESTIONS_PER_PAGE = 15

const TEMP_IMAGE_PATH = "./assets/tmp/temp-image.png"
const TEMP_HEADER_IMAGE_PATH = "./assets/tmp/temp-header-image.png"
const PADDING = 10
const CHECKBOX_AREA_PADDING = -5
const CHECKBOX_PADDING = 5
const BORDER_RECTANGLE_AREA_SIZE = 1000000

// Group checkbox paddings
const GROUP_HEADER_HEIGHT = 350
const GROUP_SIDE_PADDING = 470
const GROUP_BOTTOM_PADDING = 250

const SUBGROUP_HEADER_HEIGHT = 210
const SUBGROUP_SIDE_PADDING = 635
const SUBGROUP_BOTTOM_PADDING = 120

// Checkbox area thresholds for 0.45cm boxes scanned at 300 DPI:
// 0.45cm * (300/2.54) ≈ 53px per side → area ≈ 2825 px²
const ANSWER_SQUARE_MIN_AREA_SIZE = 1800
const ANSWER_SQUARE_MAX_AREA_SIZE = 4500
const ANSWER_SQUARE_MIN_AREA_SIZE_8 = 1800
const ANSWER_SQUARE_MAX_AREA_SIZE_8 = 4500
const CONFIGS_DIR = "./configs/"

const (
	StateXFound      = iota
	StateCircleFound = iota
	StateEmpty       = iota
)
