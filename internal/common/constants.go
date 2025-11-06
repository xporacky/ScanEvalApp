package common

import (
"fmt"
"path/filepath"
"ScanEvalApp/internal/logging" // ADD THIS IMPORT
	//"log/slog" 
)

const (
FILE_PERMISSION           = 0644
TEMPLATE_BASE_PATH        = "./assets/latex/templates/"
TEMPLATE_PATH             = TEMPLATE_BASE_PATH + "template_5.tex"
TEMPORARY_PDF_PATH        = "./assets/tmp"
QUESTION_NUMBER_NOT_FOUND = -1
DEFAULT_OPTION_COUNT      = 5
MIN_OPTION_COUNT          = 3
MAX_OPTION_COUNT          = 8
)

const (
SUCCESS         = 0
FILE_NOT_FOUND  = 1
INVALID_FORMAT  = 2
PROCESSING_FAIL = 3
)

// Template and option functions ONLY
func GetTemplatePath(optionCount int) string {
	logger := logging.GetLogger()
	
	logger.Debug("GetTemplatePath called", "input_option_count", optionCount)
	
	if optionCount < MIN_OPTION_COUNT || optionCount > MAX_OPTION_COUNT {
		logger.Warn("Invalid option count, using default", 
			"input_option_count", optionCount,
			"min", MIN_OPTION_COUNT,
			"max", MAX_OPTION_COUNT,
			"default", DEFAULT_OPTION_COUNT)
		optionCount = DEFAULT_OPTION_COUNT
	}
	
	templateFile := fmt.Sprintf("template_%d.tex", optionCount)
	fullPath := filepath.Join(TEMPLATE_BASE_PATH, templateFile)
	
	logger.Debug("Template path construction", 
		"template_file", templateFile,
		"base_path", TEMPLATE_BASE_PATH,
		"full_path", fullPath)
	
	// Convert to absolute path for better debugging
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		logger.Error("Error converting to absolute path", 
			"relative_path", fullPath,
			"error", err.Error())
		return fullPath
	}
	
	logger.Debug("GetTemplatePath result", "absolute_path", absPath)
	return absPath
}

func ValidateOptionCount(optionCount int) bool {
return optionCount >= MIN_OPTION_COUNT && optionCount <= MAX_OPTION_COUNT
}

func GetOptionCountFromIndex(index int) int {
if index < 0 || index > (MAX_OPTION_COUNT-MIN_OPTION_COUNT) {
return DEFAULT_OPTION_COUNT
}
return index + MIN_OPTION_COUNT
}

func GetOptionCountIndex(optionCount int) int {
if !ValidateOptionCount(optionCount) {
return DEFAULT_OPTION_COUNT - MIN_OPTION_COUNT
}
return optionCount - MIN_OPTION_COUNT
}

func GetOptionValues() []string {
options := make([]string, MAX_OPTION_COUNT-MIN_OPTION_COUNT+1)
for i := MIN_OPTION_COUNT; i <= MAX_OPTION_COUNT; i++ {
options[i-MIN_OPTION_COUNT] = fmt.Sprintf("%d", i)
}
return options
}

func GetOptionLetters(optionCount int) []string {
if !ValidateOptionCount(optionCount) {
optionCount = DEFAULT_OPTION_COUNT
}

letters := make([]string, optionCount)
for i := 0; i < optionCount; i++ {
letters[i] = string(rune('A' + i))
}
return letters
}
