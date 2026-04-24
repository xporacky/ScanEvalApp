package models

import (
"encoding/json"
"time"

"gorm.io/gorm"
)

// VersionData holds answers and date for a specific test version.
type VersionData struct {
Answers  string    `json:"answers"`
DateTime time.Time `json:"dateTime"`
}

// Exam represents a test (single or multi-version).
type Exam struct {
gorm.Model
Title         string    `gorm:"not null"`
SchoolYear    string    `gorm:"not null"`
Date          time.Time `gorm:"not null"` // used for single‑version tests
QuestionCount int       `gorm:"not null"`
Questions     string    // correct answers for single‑version
Students      []Student `gorm:"foreignKey:ExamID"`
OptionCount   int       `gorm:"not null"`
ShowName      bool      `gorm:"not null"`
IsMultiDay    bool      `gorm:"default:false"` // true = multi‑version test
VersionsInfo  string    `gorm:"type:text"`     // JSON map[string]VersionData
}

// GetVersionsMap returns the parsed versions map.
func (e *Exam) GetVersionsMap() (map[string]VersionData, error) {
if e.VersionsInfo == "" {
return make(map[string]VersionData), nil
}
var m map[string]VersionData
err := json.Unmarshal([]byte(e.VersionsInfo), &m)
return m, err
}

// SetVersionsMap stores the versions map as JSON.
func (e *Exam) SetVersionsMap(m map[string]VersionData) error {
data, err := json.Marshal(m)
if err != nil {
return err
}
e.VersionsInfo = string(data)
return nil
}

// GetVersionAnswers returns the answer string for a given version code (e.g. "A1").
func (e *Exam) GetVersionAnswers(versionCode string) (string, error) {
m, err := e.GetVersionsMap()
if err != nil {
return "", err
}
if vd, ok := m[versionCode]; ok {
return vd.Answers, nil
}
return "", nil
}

// UpdateVersionAnswers sets answers and optionally dateTime for a version.
func (e *Exam) UpdateVersionAnswers(versionCode, answers string, dateTime time.Time) error {
m, err := e.GetVersionsMap()
if err != nil {
return err
}
m[versionCode] = VersionData{
Answers:  answers,
DateTime: dateTime,
}
return e.SetVersionsMap(m)
}

// DeleteVersion removes a version.
func (e *Exam) DeleteVersion(versionCode string) error {
m, err := e.GetVersionsMap()
if err != nil {
return err
}
delete(m, versionCode)
return e.SetVersionsMap(m)
}




