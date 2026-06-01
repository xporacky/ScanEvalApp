package scanprocessing

const NUMBER_OF_CHOICES = 6
const NUMBER_OF_QUESTIONS_PER_PAGE = 15

const TEMP_IMAGE_PATH = "./assets/tmp/temp-image.png"
const TEMP_HEADER_IMAGE_PATH = "./assets/tmp/temp-header-image.png"
const PADDING = 10
const CHECKBOX_AREA_PADDING = -5
const CHECKBOX_PADDING = 5
const BORDER_RECTANGLE_AREA_SIZE = 1000000

// ID region starts at ~65% from left (tikz x=3.5 out of -9.2..9.2 range, 300 DPI A4).
// Using 55% to also capture the "ID:" label for OCR fallback.
const ID_REGION_LEFT_PERCENT = 55

// ID_SEPARATOR_Y_RATIO: estimated position of the header separator line (tikz y=-0.7)
// as a fraction of mat.Cols() — tunable if box crops land too high or too low.
// Decrease to shift crops UP, increase to shift DOWN.
// Now loaded from config file per scanner/printer configuration
var ID_SEPARATOR_Y_RATIO = 0.268

// ID_CENTER_X_OFFSET_RATIO: horizontal offset added to page center before tikz→pixel conversion.
// Fraction of mat.Cols() — e.g. 0.02 ≈ 4mm right on A4. Increase to shift crops RIGHT.
// Now loaded from config file per scanner/printer configuration
var ID_CENTER_X_OFFSET_RATIO = 0.01

// ID_SLASH_GAP_OFFSET_RATIO: extra rightward offset applied only to boxes 5-7 (after the slash separator).
// Compensates if the second group of digits is printed slightly further right than the tikz coords predict.
// Fraction of mat.Cols() — e.g. 0.01 ≈ 2mm on A4.
// Now loaded from config file per scanner/printer configuration
var ID_SLASH_GAP_OFFSET_RATIO = 0.006

// ID OCR retry configuration
// MAX_ID_OCR_RETRIES: maximum number of retry attempts with adjusted offsets
const MAX_ID_OCR_RETRIES = 20

// ID_OFFSET_ADJUSTMENT_STEP: amount to adjust horizontal offset on each retry (fraction of mat.Cols())
// Each retry will try different horizontal shifts (left/right)
const ID_OFFSET_ADJUSTMENT_STEP = 0.001 // ~0.4mm on A4 at 300 DPI

// Group checkbox paddings
const GROUP_HEADER_HEIGHT = 335
const GROUP_SIDE_PADDING = 370
const GROUP_BOTTOM_PADDING = 235
const GROUP_HALF_WIDTH = 416

const SUBGROUP_HEADER_HEIGHT = 210
const SUBGROUP_SIDE_PADDING = 540
const SUBGROUP_BOTTOM_PADDING = 110
const SUBGROUP_HALF_WIDTH = 261

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
