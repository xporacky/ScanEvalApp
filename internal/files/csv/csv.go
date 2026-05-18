package csv

import (
    "ScanEvalApp/internal/common"
    "ScanEvalApp/internal/config"
    "ScanEvalApp/internal/database/models"
    "ScanEvalApp/internal/database/repository"
    "ScanEvalApp/internal/logging"
    "encoding/csv"
    "encoding/json"
    "errors"
    "fmt"
    "log/slog"
    "os"
    "path/filepath"
    "sort"
    "strconv"
    "strings"
    "time"

    "gorm.io/gorm"
)

type ExamTemplate struct {
    Title             string
    SchoolYear        string
    DateTime          string
    QuestionCount     int
    OptionCount       int
    ShowName          bool
    Answers           []string
    SubgroupAnswers   map[string]string
    StudentCSVContent string
}

// ImportStudentsFromCSV parses student records from the given CSV content
// and stores them in the database.
func ImportStudentsFromCSV(db *gorm.DB, csvContent string, examID uint) error {
    exam, err := repository.GetExam(db, examID)
    if err != nil {
        return fmt.Errorf("chyba pri načítaní testu: %w", err)
    }

    logger := logging.GetLogger()
    errorLogger := logging.GetErrorLogger()

    reader := csv.NewReader(strings.NewReader(csvContent))
    rows, err := reader.ReadAll()
    if err != nil {
        errorLogger.Error("Chyba pri čítaní CSV súboru", slog.String("error", err.Error()))
        return err
    }

    for i, row := range rows {
        if i == 0 {
            continue // Preskoc hlavicku csv
        }
        birthDate, err := time.Parse("2006-01-02", row[2])
        if err != nil {
            errorLogger.Error("Chyba pri parsovaní dátumu narodenia", slog.String("error", err.Error()))
            return err
        }
        registrationNumber, err := strconv.Atoi(row[3])
        if err != nil {
            errorLogger.Error("Chyba pri parsovaní registračného čísla", slog.String("error", err.Error()))
            return err
        }

        student := models.Student{
            Name:               row[0],
            Surname:            row[1],
            BirthDate:          birthDate,
            RegistrationNumber: registrationNumber,
            Room:               row[4],
            ExamID:             examID,
            Answers:            strings.Repeat("0", exam.QuestionCount),
        }
        student.Answers = strings.Repeat("0", exam.QuestionCount)
        _, err = repository.GetStudentByRegistrationNumber(db, uint(registrationNumber), examID)
        if err == nil {
            logger.Info("Študent už je v teste, preskakujem import duplicitného záznamu",
                slog.Int("registrationNumber", registrationNumber),
                slog.Uint64("examID", uint64(examID)))
            continue
        }
        if !errors.Is(err, gorm.ErrRecordNotFound) {
            errorLogger.Error("Chyba pri overení existencie študenta",
                slog.Int("registrationNumber", registrationNumber),
                slog.String("error", err.Error()))
            return err
        }

        if err := repository.CreateStudent(db, &student); err != nil {
            errorLogger.Error("Chyba pri ukladaní študenta", slog.String("studentName", student.Name), slog.String("error", err.Error()))
            return err
        }
    }

    logger.Info("Import študentov z CSV dokončený", slog.Int("studentCount", len(rows)-1))

    return nil
}

// parseSlovakDate parses dates in DD.MM.YYYY format with or without leading zeros.
func parseSlovakDate(s string) (time.Time, error) {
    // normalize: remove spaces around dots ("30. 01. 2007" → "30.01.2007")
    s = strings.ReplaceAll(s, " ", "")
    formats := []string{"02.01.2006", "2.01.2006", "02.1.2006", "2.1.2006"}
    for _, f := range formats {
        if t, err := time.Parse(f, s); err == nil {
            return t, nil
        }
    }
    // manual fallback: split by "." and reconstruct as YYYY-MM-DD
    parts := strings.Split(s, ".")
    if len(parts) == 3 {
        normalized := fmt.Sprintf("%s-%02s-%02s", parts[2], parts[1], parts[0])
        if t, err := time.Parse("2006-01-02", normalized); err == nil {
            return t, nil
        }
    }
    return time.Time{}, fmt.Errorf("nepodporovaný formát dátumu: %q", s)
}

// ImportStudentsFromFullCSV parses the full 11-column export format (Por.;Priezvisko;Meno;Narodenie;Reg. č.;E-mail;Termín;Čas;Dátum;Miestnosť;Program)
// and stores students with per-student ExamDate, ExamTime and Room.
func ImportStudentsFromFullCSV(db *gorm.DB, csvContent string, examID uint) error {
    exam, err := repository.GetExam(db, examID)
    if err != nil {
        return fmt.Errorf("chyba pri načítaní testu: %w", err)
    }

    logger := logging.GetLogger()
    errorLogger := logging.GetErrorLogger()

    reader := csv.NewReader(strings.NewReader(csvContent))
    reader.Comma = ';'
    reader.LazyQuotes = true
    rows, err := reader.ReadAll()
    if err != nil {
        errorLogger.Error("Chyba pri čítaní full CSV súboru", slog.String("error", err.Error()))
        return err
    }

    imported := 0
    for i, row := range rows {
        if i == 0 {
            continue
        }
        if len(row) < 10 {
            errorLogger.Error("Krátky riadok v CSV, preskakujem", slog.Int("row", i))
            continue
        }

        // Stĺpce: 0=Por. 1=Priezvisko 2=Meno 3=Narodenie 4=Reg.č. 5=E-mail 6=Termín 7=Čas 8=Dátum 9=Miestnosť 10=Program
        name := strings.TrimSpace(row[2])
        surname := strings.TrimSpace(row[1])

        birthRaw := strings.TrimSpace(row[3])
        birthDate, err := parseSlovakDate(birthRaw)
        if err != nil {
            errorLogger.Error("Chyba pri parsovaní dátumu narodenia", slog.String("value", birthRaw), slog.String("error", err.Error()))
            return err
        }

        regRaw := strings.TrimSpace(row[4])
        regStr := strings.ReplaceAll(regRaw, "/", "")
        registrationNumber, err := strconv.Atoi(regStr)
        if err != nil {
            errorLogger.Error("Chyba pri parsovaní reg. čísla", slog.String("value", regRaw), slog.String("error", err.Error()))
            return err
        }

        examTimeStr := strings.TrimSpace(row[7])

        datumRaw := strings.TrimSpace(row[8])
        examDate, err := parseSlovakDate(datumRaw)
        if err != nil {
            errorLogger.Error("Chyba pri parsovaní dátumu skúšky", slog.String("value", datumRaw), slog.String("error", err.Error()))
            return err
        }

        miestnostRaw := strings.TrimSpace(row[9])
        room := strings.ToUpper(miestnostRaw)
        if len(room) > 5 {
            room = room[:5]
        }

        student := models.Student{
            Name:               name,
            Surname:            surname,
            BirthDate:          birthDate,
            RegistrationNumber: registrationNumber,
            Room:               room,
            ExamDate:           examDate,
            ExamTime:           examTimeStr,
            ExamID:             examID,
            Answers:            strings.Repeat("0", exam.QuestionCount),
        }

        _, err = repository.GetStudentByRegistrationNumber(db, uint(registrationNumber), examID)
        if err == nil {
            logger.Info("Študent už je v teste, preskakujem", slog.Int("registrationNumber", registrationNumber))
            continue
        }
        if !errors.Is(err, gorm.ErrRecordNotFound) {
            errorLogger.Error("Chyba pri overení existencie študenta", slog.String("error", err.Error()))
            return err
        }

        if err := repository.CreateStudent(db, &student); err != nil {
            errorLogger.Error("Chyba pri ukladaní študenta", slog.String("name", student.Name), slog.String("error", err.Error()))
            return err
        }
        imported++
    }

    logger.Info("Full CSV import dokončený", slog.Int("imported", imported))
    return nil
}

// ExportStudentsToCSV exports all students associated with the given exam
// into a CSV file.
func ExportStudentsToCSV(db *gorm.DB, exam models.Exam) (string, error) {
    logger := logging.GetLogger()
    errorLogger := logging.GetErrorLogger()

    var students []models.Student
    err := db.Where("exam_id = ?", exam.ID).Find(&students).Error
    if err != nil {
        errorLogger.Error("Chyba pri načítaní študentov", slog.String("error", err.Error()))
        return "", err
    }
    dirPath, err := config.LoadLastPath()
    if err != nil {
        errorLogger.Error("Chyba načítania configu", slog.String("error", err.Error()))
        return "", err
    }

    absDirPath, err := filepath.Abs(dirPath)
    if err != nil {
        errorLogger.Error("Chyba pri konverzii cesty", slog.String("error", err.Error()))
        return "", err
    }

    safeTitle := common.SanitizeFilename(exam.Title)

    fileName := filepath.Join(absDirPath, fmt.Sprintf("%s_ID%d.csv", safeTitle, exam.ID))

    file, err := os.Create(fileName)
    if err != nil {
        errorLogger.Error("Chyba pri vytváraní CSV súboru", slog.String("fileName", fileName), slog.String("error", err.Error()))
        return "", err
    }
    defer file.Close()

    file.Write([]byte{0xEF, 0xBB, 0xBF})

    writer := csv.NewWriter(file)
    defer writer.Flush()

    // Hlavicka CSV
    err = writer.Write([]string{"ID", "Meno", "Priezvisko", "Reg. číslo", "Skóre"})

    if err != nil {
        errorLogger.Error("Chyba pri zápise hlavičky CSV", slog.String("error", err.Error()))
        return "", err
    }

    for _, student := range students {
        record := []string{
            strconv.Itoa(int(student.ID)),
            student.Name,
            student.Surname,
            strconv.Itoa(student.RegistrationNumber),
            strconv.Itoa(student.Score),
        }
        err := writer.Write(record)

        if err != nil {
            errorLogger.Error("Chyba pri zápise záznamu študenta do CSV", slog.String("studentName", student.Name), slog.String("error", err.Error()))
            return "", err
        }
    }

    logger.Info("Export študentov do CSV úspešný", slog.String("fileName", fileName), slog.Int("studentCount", len(students)))

    return fileName, nil
}

// ExportMultiDayResultsCSV exports all students of a multi-day exam into one CSV.
// Columns: Meno, Priezvisko, Dátum, Čas, Miestnosť, Podskupina, Skóre, Odpovede študenta, Správne odpovede
func ExportMultiDayResultsCSV(db *gorm.DB, examID uint) (string, error) {
    errorLogger := logging.GetErrorLogger()

    var exam models.Exam
    if err := db.Preload("Students").First(&exam, examID).Error; err != nil {
        return "", err
    }

    // Deserializuj správne odpovede (JSON mapa alebo plain string)
    answersMap := map[string]string{}
    q := strings.TrimSpace(exam.Questions)
    if strings.HasPrefix(q, "{") {
        _ = json.Unmarshal([]byte(q), &answersMap)
    } else if q != "" {
        answersMap[""] = q
    }

    dirPath, err := config.LoadLastPath()
    if err != nil {
        return "", err
    }
    absDirPath, err := filepath.Abs(dirPath)
    if err != nil {
        return "", err
    }

    safeTitle := common.SanitizeFilename(exam.Title)
    fileName := filepath.Join(absDirPath, fmt.Sprintf("%s_ID%d_vysledky.csv", safeTitle, exam.ID))

    file, err := os.Create(fileName)
    if err != nil {
        errorLogger.Error("Chyba pri vytváraní CSV", slog.String("error", err.Error()))
        return "", err
    }
    defer file.Close()
    file.Write([]byte{0xEF, 0xBB, 0xBF})

    writer := csv.NewWriter(file)
    defer writer.Flush()

    header := []string{"Meno", "Priezvisko", "Reg. číslo", "Dátum skúšky", "Čas", "Miestnosť", "Podskupina", "Skóre", "Odpovede študenta", "Správne odpovede"}
    if err := writer.Write(header); err != nil {
        return "", err
    }

    // Zoraď: Dátum → Čas → Miestnosť → Priezvisko
    students := exam.Students
    sort.Slice(students, func(i, j int) bool {
        di, dj := students[i].ExamDate, students[j].ExamDate
        if !di.Equal(dj) {
            return di.Before(dj)
        }
        if students[i].ExamTime != students[j].ExamTime {
            return students[i].ExamTime < students[j].ExamTime
        }
        if students[i].Room != students[j].Room {
            return students[i].Room < students[j].Room
        }
        return students[i].Surname < students[j].Surname
    })

    for _, s := range students {
        correctAns := answersMap[s.Subgroup]
        if correctAns == "" && len(answersMap) == 1 {
            for _, v := range answersMap {
                correctAns = v
            }
        }

        dateStr := ""
        if !s.ExamDate.IsZero() {
            dateStr = s.ExamDate.Format("02.01.2006")
        }

        record := []string{
            s.Name,
            s.Surname,
            fmt.Sprintf("%04d/%03d", s.RegistrationNumber/1000, s.RegistrationNumber%1000),
            dateStr,
            s.ExamTime,
            s.Room,
            s.Subgroup,
            strconv.Itoa(s.Score),
            s.Answers,
            correctAns,
        }
        if err := writer.Write(record); err != nil {
            errorLogger.Error("Chyba pri zápise riadku", slog.String("error", err.Error()))
            return "", err
        }
    }

    return fileName, nil
}

func ExportExamTemplateCSV(template ExamTemplate) (string, error) {
    errorLogger := logging.GetErrorLogger()

    dirPath, err := config.LoadLastPath()
    if err != nil {
        errorLogger.Error("Chyba načítania configu", slog.String("error", err.Error()))
        return "", err
    }

    absDirPath, err := filepath.Abs(dirPath)
    if err != nil {
        errorLogger.Error("Chyba pri konverzii cesty", slog.String("error", err.Error()))
        return "", err
    }

    fileName := "test_sablona.csv"
    if strings.TrimSpace(template.Title) != "" {
        fileName = common.SanitizeFilename(template.Title) + "_sablona.csv"
    }
    filePath := filepath.Join(absDirPath, fileName)

    file, err := os.Create(filePath)
    if err != nil {
        errorLogger.Error("Chyba pri vytváraní CSV šablóny", slog.String("fileName", filePath), slog.String("error", err.Error()))
        return "", err
    }
    defer file.Close()

    file.Write([]byte{0xEF, 0xBB, 0xBF})

    writer := csv.NewWriter(file)
    defer writer.Flush()

    rows := [][]string{
        {"section", "key", "value"},
        {"meta", "title", template.Title},
        {"meta", "school_year", template.SchoolYear},
        {"meta", "date_time", template.DateTime},
        {"meta", "question_count", strconv.Itoa(template.QuestionCount)},
        {"meta", "option_count", strconv.Itoa(template.OptionCount)},
        {"meta", "show_name", strconv.FormatBool(template.ShowName)},
    }

    for index, answer := range template.Answers {
        rows = append(rows, []string{"answer", strconv.Itoa(index + 1), strings.TrimSpace(answer)})
    }

    for subgroup, answers := range template.SubgroupAnswers {
        rows = append(rows, []string{"subgroup", subgroup, strings.TrimSpace(answers)})
    }

    rows = append(rows, []string{"payload", "students_csv", template.StudentCSVContent})

    for _, row := range rows {
        if err := writer.Write(row); err != nil {
            errorLogger.Error("Chyba pri zápise CSV šablóny", slog.String("fileName", filePath), slog.String("error", err.Error()))
            return "", err
        }
    }

    return filePath, nil
}

func ParseExamTemplateCSV(csvContent string) (ExamTemplate, error) {
    reader := csv.NewReader(strings.NewReader(csvContent))
    rows, err := reader.ReadAll()
    if err != nil {
        return ExamTemplate{}, err
    }
    if len(rows) == 0 {
        return ExamTemplate{}, fmt.Errorf("csv sablona je prazdna")
    }

    template := ExamTemplate{
        OptionCount: 5,
        ShowName:    true,
    }

    for i, row := range rows {
        if len(row) < 3 {
            continue
        }
        section := strings.TrimSpace(strings.TrimPrefix(row[0], "\ufeff"))
        key := strings.TrimSpace(row[1])
        value := row[2]

        if i == 0 && strings.EqualFold(section, "section") {
            continue
        }

        switch section {
        case "meta":
            switch key {
            case "title":
                template.Title = value
            case "school_year":
                template.SchoolYear = value
            case "date_time":
                template.DateTime = value
            case "question_count":
                template.QuestionCount, err = strconv.Atoi(strings.TrimSpace(value))
                if err != nil {
                    return ExamTemplate{}, fmt.Errorf("neplatny question_count")
                }
            case "option_count":
                template.OptionCount, err = strconv.Atoi(strings.TrimSpace(value))
                if err != nil {
                    return ExamTemplate{}, fmt.Errorf("neplatny option_count")
                }
            case "show_name":
                template.ShowName, err = strconv.ParseBool(strings.TrimSpace(value))
                if err != nil {
                    return ExamTemplate{}, fmt.Errorf("neplatny show_name")
                }
            }
        case "answer":
            template.Answers = append(template.Answers, strings.TrimSpace(value))
        case "payload":
            if key == "students_csv" {
                template.StudentCSVContent = value
            }
        }
    }

    if template.QuestionCount <= 0 {
        template.QuestionCount = len(template.Answers)
    }
    if template.QuestionCount <= 0 {
        return ExamTemplate{}, fmt.Errorf("sablona neobsahuje pocet otazok")
    }

    for len(template.Answers) < template.QuestionCount {
        template.Answers = append(template.Answers, "")
    }
    if len(template.Answers) > template.QuestionCount {
        template.Answers = template.Answers[:template.QuestionCount]
    }

    if template.OptionCount < 2 {
        template.OptionCount = 2
    }
    if template.OptionCount > 8 {
        template.OptionCount = 8
    }

    return template, nil
}
