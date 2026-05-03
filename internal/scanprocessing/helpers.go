package scanprocessing

import (
	"ScanEvalApp/internal/database/models"
	"ScanEvalApp/internal/database/repository"
	"ScanEvalApp/internal/files"
	"ScanEvalApp/internal/ocr"
	"encoding/json"
	"fmt"
	"image"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"

	"ScanEvalApp/internal/logging"
	"log/slog"
	"strings"

	"gocv.io/x/gocv"
	"gorm.io/gorm"
)

var qrIDPattern = regexp.MustCompile(`(?:^|[|;\s])ID\s*:\s*(\d+)`)

// offsetAdjustment represents adjustments to ID box OCR offsets for retry attempts
type offsetAdjustment struct {
	separatorYDelta float64 // adjustment to ID_SEPARATOR_Y_RATIO
	centerXDelta    float64 // adjustment to ID_CENTER_X_OFFSET_RATIO
	slashGapDelta   float64 // adjustment to ID_SLASH_GAP_OFFSET_RATIO
}

// digitResult holds OCR result for a single digit box
type digitResult struct {
	digit      string  // the recognized character
	isValid    bool    // true if it's a valid digit (0-9)
	attempt    int     // which attempt produced this result
	confidence float64 // not used yet, but reserved for future OCR confidence scores
}

// extractIDByBoxOCR reads the 7 ID digits one box at a time using PSM 10 (single char).
// If the first attempt doesn't find all 7 digits, it retries with adjusted offsets.
// Box positions come from template_6.tex tikz coordinates.
// Coordinate math (A4, 0.8in margins, tikz range -9.2..9.2):
//
//	scale    = mat.Cols() * (1 - 2*20.32/210) / 18.4   [px per tikz unit]
//	pixelX   = mat.Cols()/2 + tikzX * scale
//	pixelY   = separatorY - (tikzY + 0.7) * scale       [separator line at tikz y=-0.7, ~28% down from top]
func extractIDByBoxOCR(mat *gocv.Mat) (int, error) {
	logger := logging.GetLogger()

	// Define retry strategy: horizontal shifts only (Y position stays constant)
	// Try progressively larger shifts left and right
	adjustments := []offsetAdjustment{
		{0.0, 0.0, 0.0},       // 0: Original values (baseline)
		{0.0, -0.001, 0.0},    // 1: Tiny shift left
		{0.0, 0.001, 0.0},     // 2: Tiny shift right
		{0.0, -0.002, 0.0},    // 3: Small shift left
		{0.0, 0.002, 0.0},     // 4: Small shift right
		{0.0, -0.003, 0.0},    // 5: Medium shift left
		{0.0, 0.003, 0.0},     // 6: Medium shift right
		{0.0, -0.004, 0.0},    // 7: Larger shift left
		{0.0, 0.004, 0.0},     // 8: Larger shift right
		{0.0, -0.005, 0.0},    // 9: Max shift left
		{0.0, 0.005, 0.0},     // 10: Max shift right
		{0.0, 0.0, -0.002},    // 11: Adjust slash gap left
		{0.0, 0.0, 0.002},     // 12: Adjust slash gap right
		{0.0, -0.003, -0.002}, // 13: Shift left + slash gap left
		{0.0, 0.003, 0.002},   // 14: Shift right + slash gap right
	}

	maxRetries := MAX_ID_OCR_RETRIES
	if maxRetries > len(adjustments) {
		maxRetries = len(adjustments)
	}

	// Store best results for each digit position (0-6)
	bestDigits := make([]digitResult, 7)
	for i := range bestDigits {
		bestDigits[i] = digitResult{digit: "?", isValid: false, attempt: -1}
	}

	var lastErr error
	foundAllDigits := false

	for attempt := 0; attempt < maxRetries && !foundAllDigits; attempt++ {
		adj := adjustments[attempt]

		logger.Info("ID extraction attempt",
			slog.Int("attempt", attempt),
			slog.Float64("separatorYDelta", adj.separatorYDelta),
			slog.Float64("centerXDelta", adj.centerXDelta),
			slog.Float64("slashGapDelta", adj.slashGapDelta))

		digits, err := tryExtractIDWithOffsets(mat, adj, attempt)

		// Merge results: keep valid digits from this attempt
		newValidCount := 0
		for i := 0; i < 7 && i < len(digits); i++ {
			if digits[i].isValid && !bestDigits[i].isValid {
				bestDigits[i] = digits[i]
				logger.Info("Found new valid digit",
					slog.Int("position", i),
					slog.String("digit", digits[i].digit),
					slog.Int("fromAttempt", attempt))
				newValidCount++
			}
		}

		// Check if we have all 7 digits now
		validCount := 0
		for _, d := range bestDigits {
			if d.isValid {
				validCount++
			}
		}

		logger.Info("Current digit status",
			slog.Int("attempt", attempt),
			slog.Int("totalValid", validCount),
			slog.Int("newThisAttempt", newValidCount))

		if validCount == 7 {
			foundAllDigits = true
			logger.Info("Successfully found all 7 digits",
				slog.Int("finalAttempt", attempt))
		}

		if err != nil {
			lastErr = err
		}
	}

	// Build final ID from best digits
	if foundAllDigits {
		idStr := ""
		for i, d := range bestDigits {
			if !d.isValid {
				return 0, fmt.Errorf("missing digit at position %d", i)
			}
			idStr += d.digit
		}

		logger.Info("Assembled ID from multiple attempts", slog.String("id", idStr))

		id, err := strconv.Atoi(idStr)
		if err != nil {
			return 0, fmt.Errorf("failed to parse assembled ID %q: %w", idStr, err)
		}
		return id, nil
	}

	// Count how many digits we found
	validCount := 0
	for _, d := range bestDigits {
		if d.isValid {
			validCount++
		}
	}

	return 0, fmt.Errorf("incomplete ID after %d attempts: found %d/7 digits, last error: %w", maxRetries, validCount, lastErr)
}

// tryExtractIDWithOffsets attempts to extract ID with specific offset adjustments
// Returns: array of digitResult for each of the 7 digit positions, and error (if any)
// Note: Only horizontal (X) adjustments are used; vertical (Y) position remains constant
func tryExtractIDWithOffsets(mat *gocv.Mat, adj offsetAdjustment, attemptNum int) ([]digitResult, error) {
	logger := logging.GetLogger()

	// Apply offset adjustments
	// Note: separatorYDelta is kept at 0 - we only adjust horizontal position
	separatorYRatio := ID_SEPARATOR_Y_RATIO + adj.separatorYDelta
	centerXOffsetRatio := ID_CENTER_X_OFFSET_RATIO + adj.centerXDelta
	slashGapOffsetRatio := ID_SLASH_GAP_OFFSET_RATIO + adj.slashGapDelta

	scale := float64(mat.Cols()) * (1.0 - 2.0*20.32/210.0) / 18.4
	centerX := float64(mat.Cols())/2.0 + float64(mat.Cols())*centerXOffsetRatio
	separatorY := float64(mat.Cols()) * separatorYRatio

	toPixelX := func(x float64) int { return int(centerX + x*scale) }
	toPixelY := func(y float64) int { return int(separatorY - (y+0.7)*scale) }

	// Box positions from template_6.tex: \draw[thick] (X,2.8) rectangle ++(0.6,0.6)
	boxXPositions := []float64{3.50, 4.22, 4.94, 5.66, 7.10, 7.82, 8.54}
	boxTop := toPixelY(3.4)
	boxBottom := toPixelY(2.8)
	// Crop 4px inside each border so the thick rectangle lines don't confuse OCR
	const borderPad = 4

	slashGapPx := int(float64(mat.Cols()) * slashGapOffsetRatio)

	// Initialize results for all 7 digit positions
	results := make([]digitResult, 7)
	for i := range results {
		results[i] = digitResult{digit: "?", isValid: false, attempt: attemptNum}
	}

	// Log crop positions for debugging
	if attemptNum == 0 {
		logger.Info("ID box crop positions",
			slog.Float64("separatorYRatio", separatorYRatio),
			slog.Float64("centerXOffsetRatio", centerXOffsetRatio),
			slog.Float64("slashGapOffsetRatio", slashGapOffsetRatio),
			slog.Int("boxTop", boxTop),
			slog.Int("boxBottom", boxBottom))
	}

	for i, bx := range boxXPositions {
		extraX := 0
		if i >= 4 { // boxes after the slash separator
			extraX = slashGapPx
		}
		region := image.Rectangle{
			Min: image.Point{toPixelX(bx) + borderPad + extraX, boxTop + borderPad},
			Max: image.Point{toPixelX(bx+0.6) - borderPad + extraX, boxBottom - borderPad},
		}
		if region.Min.X < 0 || region.Min.Y < 0 || region.Max.X > mat.Cols() || region.Max.Y > mat.Rows() {
			logger.Warn("Digit box out of bounds", slog.Int("box", i), slog.Any("region", region))
			continue
		}

		digitMat := mat.Region(region)
		// Save with attempt number so we can see how crops shift between retries
		tmpPath := fmt.Sprintf("./assets/tmp/id-digit-%d-attempt-%d.png", i, attemptNum)
		SaveMat(tmpPath, digitMat)
		digitMat.Close()

		// Log the exact crop coordinates for this box
		logger.Debug("ID digit box crop",
			slog.Int("attempt", attemptNum),
			slog.Int("box", i),
			slog.Int("x1", region.Min.X),
			slog.Int("y1", region.Min.Y),
			slog.Int("x2", region.Max.X),
			slog.Int("y2", region.Max.Y),
			slog.Int("width", region.Dx()),
			slog.Int("height", region.Dy()))

		raw, err := ocr.OcrSingleChar(tmpPath)
		//files.DeleteFile(tmpPath)
		if err != nil {
			logger.Warn("OCR failed for digit box",
				slog.Int("attempt", attemptNum),
				slog.Int("box", i),
				slog.String("error", err.Error()))
			continue
		}

		d := strings.TrimSpace(raw)
		logger.Debug("ID box OCR",
			slog.Int("attempt", attemptNum),
			slog.Int("box", i),
			slog.String("digit", d),
			slog.String("savedTo", tmpPath))

		// Check if it's a valid digit
		isValid := len(d) == 1 && d[0] >= '0' && d[0] <= '9'
		results[i] = digitResult{
			digit:   d,
			isValid: isValid,
			attempt: attemptNum,
		}
	}

	// Count valid digits for logging
	validCount := 0
	digitsStr := ""
	for _, r := range results {
		if r.isValid {
			validCount++
			digitsStr += r.digit
		} else {
			digitsStr += "?"
		}
	}

	logger.Debug("ID box OCR results",
		slog.Int("attempt", attemptNum),
		slog.String("digits", digitsStr),
		slog.Int("validCount", validCount))

	return results, nil
}

// FindContours detects external contours in the provided image using edge detection and morphological operations.
//
// The function applies Canny edge detection to highlight edges in the image, followed by morphological closing
// (dilation and erosion) to reduce noise and close gaps in detected edges. It then finds and returns external contours,
// which are typically used for shape detection and segmentation.
//
// Parameters:
//   - mat: A gocv.Mat representing the source image to process.
//
// Returns:
//   - gocv.PointsVector: A vector of detected contours, where each contour is represented as a slice of points.
func FindContours(mat gocv.Mat) gocv.PointsVector {
	// Use Canny edge detection
	canny := gocv.NewMat()
	defer canny.Close()
	gocv.Canny(mat, &canny, 100, 200)

	// Use morphological closing
	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
	defer kernel.Close()
	gocv.Dilate(canny, &canny, kernel)
	gocv.Erode(canny, &canny, kernel)

	// Find contours
	contours := gocv.FindContours(canny, gocv.RetrievalExternal, gocv.ChainApproxNone)
	return contours
}

// ImageToMat converts a Go image.RGBA to a gocv.Mat for OpenCV processing.
//
// The function reads pixel data from the input RGBA image and manually rearranges the color channels
// into BGR format, as required by OpenCV. It then creates a new gocv.Mat from the byte slice.
// If the conversion fails, the function logs the error and panics.
//
// Parameters:
//   - imgRGBA: A pointer to an image.RGBA containing the source image data.
//
// Returns:
//   - gocv.Mat: The converted image as a gocv.Mat in BGR format suitable for further OpenCV operations.
func ImageToMat(imgRGBA *image.RGBA) gocv.Mat {
	errorLogger := logging.GetErrorLogger()

	bounds := imgRGBA.Bounds()
	x := bounds.Dx()
	y := bounds.Dy()
	bytes := make([]byte, 0, x*y)
	for j := bounds.Min.Y; j < bounds.Max.Y; j++ {
		for i := bounds.Min.X; i < bounds.Max.X; i++ {
			r, g, b, _ := imgRGBA.At(i, j).RGBA()
			bytes = append(bytes, byte(b>>8))
			bytes = append(bytes, byte(g>>8))
			bytes = append(bytes, byte(r>>8))
		}
	}

	mat, err := gocv.NewMatFromBytes(y, x, gocv.MatTypeCV8UC3, bytes)
	if err != nil {
		errorLogger.Error("Chyba pri konverzii obrázka na Mat", slog.String("error", err.Error()))
		panic(err)
	}
	return mat
}

// MatToGrayscale converts a BGR image to a grayscale image.
//
// The function uses OpenCV's CvtColor to convert the input image from BGR color space to grayscale.
// This is commonly used in image processing tasks where color information is not needed and
// only intensity (brightness) is relevant.
//
// Parameters:
//   - mat: A gocv.Mat representing the input BGR image to be converted.
//
// Returns:
//   - gocv.Mat: A new gocv.Mat representing the grayscale version of the input image.
func MatToGrayscale(mat gocv.Mat) gocv.Mat {
	gray := gocv.NewMat()
	gocv.CvtColor(mat, &gray, gocv.ColorBGRToGray)
	return gray
}

// SaveMat saves a gocv.Mat image to a specified file path in PNG format.
//
// The function first attempts to delete any existing file at the given path. If the file is successfully
// removed, it proceeds to save the image using the OpenCV IMWrite function. If the path is empty, a default
// temporary image path is used. If the image is saved successfully, a log entry is created; otherwise,
// an error is logged and the function panics.
//
// Parameters:
//   - path: The file path where the image should be saved. If empty, the default temporary path is used.
//   - mat: The gocv.Mat image to be saved.
//
// Notes:
//   - If the file at the given path already exists, it will be deleted before saving the new image.
func SaveMat(path string, mat gocv.Mat) {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	if path == "" {
		path = TEMP_IMAGE_PATH
	}
	err := files.DeleteFile(path)
	if err != nil {
		errorLogger.Error("Chyba pri odstraňovaní existujúceho súboru", slog.String("path", path), slog.String("error", err.Error()))
		return
	}
	if gocv.IMWrite(path, mat) {
		logger.Info("Úspešne uložený obrázok", slog.String("path", path))
	} else {
		errorLogger.Error("Chyba pri ukladaní obrázka", slog.String("path", path))
	}
}

// ReadQR detects and decodes a QR code from a given image (gocv.Mat).
//
// The function uses OpenCV's QRCodeDetector to detect and decode the QR code in the input image.
// It returns the decoded text from the QR code. If no QR code is found or it cannot be decoded,
// an empty string is returned.
//
// Parameters:
//   - mat: A gocv.Mat representing the image that may contain a QR code.
//
// Returns:
//   - string: The decoded text from the QR code. If no QR code is detected, an empty string is returned.
func ReadQR(mat *gocv.Mat) string {
	qrDetector := gocv.NewQRCodeDetector()
	defer qrDetector.Close()

	if text := detectQR(qrDetector, *mat); text != "" {
		return text
	}

	candidates := buildQRFallbackCandidates(mat)
	defer func() {
		for _, candidate := range candidates {
			candidate.Close()
		}
	}()

	for _, candidate := range candidates {
		if text := detectQR(qrDetector, candidate); text != "" {
			return text
		}
	}

	return ""
}

func detectQR(qrDetector gocv.QRCodeDetector, mat gocv.Mat) string {
	points := gocv.NewMat()
	defer points.Close()
	qrCode := gocv.NewMat()
	defer qrCode.Close()
	return qrDetector.DetectAndDecode(mat, &points, &qrCode)
}

func buildQRFallbackCandidates(mat *gocv.Mat) []gocv.Mat {
	if mat == nil || mat.Empty() {
		return nil
	}

	cols := mat.Cols()
	rows := mat.Rows()
	rects := []image.Rectangle{
		image.Rect(cols*65/100, rows*8/100, cols-PADDING, rows*36/100),
		image.Rect(cols*55/100, rows*5/100, cols-PADDING, rows*42/100),
	}

	var candidates []gocv.Mat
	for _, rect := range rects {
		rect = rect.Intersect(image.Rect(0, 0, cols, rows))
		if rect.Empty() {
			continue
		}

		crop := mat.Region(rect)
		candidates = append(candidates, crop.Clone())

		gray := crop
		grayClone := gocv.NewMat()
		if crop.Channels() > 1 {
			gocv.CvtColor(crop, &grayClone, gocv.ColorBGRToGray)
			gray = grayClone
		}

		for _, scale := range []float64{2, 3, 4} {
			resized := gocv.NewMat()
			gocv.Resize(gray, &resized, image.Point{}, scale, scale, gocv.InterpolationCubic)
			candidates = append(candidates, resized)

			thresholded := gocv.NewMat()
			gocv.Threshold(resized, &thresholded, 0, 255, gocv.ThresholdBinary|gocv.ThresholdOtsu)
			candidates = append(candidates, thresholded)
		}

		if !grayClone.Empty() {
			grayClone.Close()
		}
		crop.Close()
	}

	return candidates
}

func extractRegistrationNumberFromQR(qrText string) (int, error) {
	if qrText == "" {
		return 0, fmt.Errorf("empty QR text")
	}

	if matches := qrIDPattern.FindStringSubmatch(qrText); len(matches) == 2 {
		return strconv.Atoi(matches[1])
	}

	var id int
	if _, err := fmt.Sscan(qrText, &id); err == nil {
		return id, nil
	}

	return 0, fmt.Errorf("QR text does not contain ID field: %s", qrText)
}

// GetStudent attempts to find and return a student from the provided image (gocv.Mat)
// using either a QR code or OCR to extract the student's ID or registration number.
//
// The function first tries to read the QR code from the image. If the QR code is successfully decoded,
// it extracts the student ID and retrieves the student information from the database using the ID.
// If no QR code is found, the function then attempts to extract the registration number from the image's header
// using OCR. Once the registration number is extracted, it retrieves the student information from the database.
//
// Parameters:
//   - mat: A gocv.Mat representing the image containing the QR code or header with registration number.
//   - db: A pointer to the gorm.DB object used for database access.
//   - examID: The exam ID to be associated with the student record lookup.
//
// Returns:
//   - *models.Student: The student object retrieved from the database, or nil if no student is found.
//   - error: An error if there was an issue reading the QR code, performing OCR, or querying the database.
func GetStudent(mat *gocv.Mat, db *gorm.DB, examID uint) (*models.Student, error) {

	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	qrText := ReadQR(mat)
	if qrText != "" {
		id, err := extractRegistrationNumberFromQR(qrText)
		if err != nil {
			errorLogger.Error("Chyba pri konverzii QR textu na ID", slog.String("qrText", qrText), slog.String("error", err.Error()))
			return nil, err
		}
		logger.Info("Id studenta bolo najdene z qr kodu", slog.Int("id", id))
		//return repository.GetStudentById(db, uint(id), examID)
		return repository.GetStudentByRegistrationNumber(db, uint(id), examID)

	}
	logger.Warn("QR kód nebol nájdený, pokúšame sa získať ID z boxov v hlavičke")

	// Primary fallback: crop each ID digit box individually (PSM 10, single char).
	registrationNumber, err := extractIDByBoxOCR(mat)
	if err != nil {
		logger.Warn("Box OCR fallback zlyhal, skúšam header OCR", slog.String("error", err.Error()))
		// Secondary fallback: crop right portion of header and run full OCR.
		idLeft := mat.Cols() * ID_REGION_LEFT_PERCENT / 100
		rect := image.Rectangle{Min: image.Point{idLeft, PADDING}, Max: image.Point{mat.Cols() - PADDING, (mat.Rows() / 4) - PADDING}}
		headerMat := mat.Region(rect)
		defer headerMat.Close()

		tmpFile, err := os.CreateTemp("./assets/tmp", "ocr-header-*.png")
		if err != nil {
			errorLogger.Error("Nepodarilo sa vytvoriť dočasný súbor pre OCR záhlavia", slog.String("error", err.Error()))
			return nil, err
		}
		tmpPath := tmpFile.Name()
		tmpFile.Close()
		defer files.DeleteFile(tmpPath)

		SaveMat(tmpPath, headerMat)
		registrationNumber, err = ocr.ExtractID(tmpPath)
		if err != nil {
			errorLogger.Error("Chyba pri extrakcii registrationNumber zo záhlavia obrázku", slog.String("error", err.Error()))
			return nil, err
		}
	}
	logger.Info("Registracne cislo bolo najdene", slog.Int("registrationNumber", registrationNumber))
	return repository.GetStudentByRegistrationNumber(db, uint(registrationNumber), examID)
}

func LoadConfig(configFile string) error {
	configPath := CONFIGS_DIR + configFile + ".json"
	file, err := os.Open(configPath)
	if err != nil {
		return fmt.Errorf("chyba pri otváraní konfiguračného súboru: %w", err)
	}
	defer file.Close()

	var config struct {
		MeanIntensityXLowest  float64 `json:"mean_intensity_x_lowest"`
		MeanIntensityXHighest float64 `json:"mean_intensity_x_highest"`
		IDSeparatorYRatio     float64 `json:"id_separator_y_ratio"`
		IDCenterXOffsetRatio  float64 `json:"id_center_x_offset_ratio"`
		IDSlashGapOffsetRatio float64 `json:"id_slash_gap_offset_ratio"`
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return fmt.Errorf("chyba pri dekódovaní konfiguračného súboru: %w", err)
	}

	MEAN_INTENSITY_X_LOWEST = config.MeanIntensityXLowest
	MEAN_INTENSITY_X_HIGHEST = config.MeanIntensityXHighest
	ID_SEPARATOR_Y_RATIO = config.IDSeparatorYRatio
	ID_CENTER_X_OFFSET_RATIO = config.IDCenterXOffsetRatio
	ID_SLASH_GAP_OFFSET_RATIO = config.IDSlashGapOffsetRatio
	return nil
}

func ExportFailedPagesToPDF(examTitle string, examID uint, pages []int, inputPDF string, outputPath string) error {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	if len(pages) == 0 {
		return nil
	}

	var pageArgs []string
	for _, p := range pages {
		pageArgs = append(pageArgs, strconv.Itoa(p+1))
	}

	cmdArgs := append([]string{inputPDF, "cat"}, pageArgs...)
	outputPDF := filepath.Join(outputPath, fmt.Sprintf("%s%d_failed_pages.pdf", examTitle, examID))
	cmdArgs = append(cmdArgs, "output", outputPDF)

	cmd := exec.Command("pdftk", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		errorLogger.Error("Chyba pri spájaní chybných stránok", "error", err.Error(), "output", string(output))
		return err
	}

	logger.Info("Chybné strany uložené do PDF", "output", outputPDF)
	return nil
}

// Adds a failed page into failedPagesMap with the use of locks
func AddFailedPage(failedPages *FailedPages, examID uint, pageNumber int, reason string) {
	failedPages.mu.Lock()
	defer failedPages.mu.Unlock()
	failedPages.data[examID] = append(failedPages.data[examID], FailedPageInfo{
		PageNumber: pageNumber,
		Reason:     reason,
	})
}

// AddFailedPageDetailed adds a failed page with detailed information
func AddFailedPageDetailed(failedPages *FailedPages, examID uint, info FailedPageInfo) {
	failedPages.mu.Lock()
	defer failedPages.mu.Unlock()
	failedPages.data[examID] = append(failedPages.data[examID], info)
}
