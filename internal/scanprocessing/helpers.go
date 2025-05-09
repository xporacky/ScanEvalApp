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
	"strconv"

	"ScanEvalApp/internal/logging"
	"log/slog"

	"gocv.io/x/gocv"
	"gorm.io/gorm"
)

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
		panic(err)
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
	points := gocv.NewMat()
	defer points.Close()
	qrCode := gocv.NewMat()
	defer qrCode.Close()
	text := qrDetector.DetectAndDecode(*mat, &points, &qrCode)
	return text
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
		var id int
		_, err := fmt.Sscan(qrText, &id)
		if err != nil {
			errorLogger.Error("Chyba pri konverzii QR textu na ID", slog.String("qrText", qrText), slog.String("error", err.Error()))
			return nil, err
		}
		logger.Info("Id studenta bolo najdene z qr kodu", slog.Int("id", id))
		return repository.GetStudentById(db, uint(id), examID)

	}
	logger.Warn("QR kód nebol nájdený, pokúšame sa získať registrationNumber zo záhlavia")

	rect := image.Rectangle{Min: image.Point{PADDING, PADDING}, Max: image.Point{mat.Cols() - PADDING, (mat.Rows() / 4) - PADDING}}
	headerMat := mat.Region(rect)
	defer headerMat.Close()
	SaveMat(TEMP_HEADER_IMAGE_PATH, headerMat)
	registrationNumber, err := ocr.ExtractID(TEMP_HEADER_IMAGE_PATH)
	files.DeleteFile(TEMP_HEADER_IMAGE_PATH)
	if err != nil {
		errorLogger.Error("Chyba pri extrakcii registrationNumber zo záhlavia obrázku", slog.String("error", err.Error()))
		return nil, err
	}
	logger.Info("Registracne cislo bolo najdene z hlavicky", slog.Int("registrationNumber", registrationNumber))
	return repository.GetStudentByRegistrationNumber(db, uint(registrationNumber), examID)
}

func LoadConfig(configFile string) error {
	configPath := CONFIGS_DIR + configFile + ".json"
	file, err := os.Open(configPath)
	if err != nil {
		return fmt.Errorf("Chyba pri otváraní konfiguračného súboru: %w", err)
	}
	defer file.Close()

	var config struct {
		MeanIntensityXLowest  float64 `json:"mean_intensity_x_lowest"`
		MeanIntensityXHighest float64 `json:"mean_intensity_x_highest"`
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return fmt.Errorf("Chyba pri dekódovaní konfiguračného súboru: %w", err)
	}

	MEAN_INTENSITY_X_LOWEST = config.MeanIntensityXLowest
	MEAN_INTENSITY_X_HIGHEST = config.MeanIntensityXHighest
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
func AddFailedPage(failedPages *FailedPages, examID uint, pageNumber int) {
	failedPages.mu.Lock()
	defer failedPages.mu.Unlock()
	failedPages.data[examID] = append(failedPages.data[examID], pageNumber)
}
