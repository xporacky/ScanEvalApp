import './style.css';
import './app.css';

import {
  CreateExamWithCSV,
  DeleteExam,
  ListExams,
  ListStudents,
  PrintExamPDF,
  PrintExamPDFWithLegend,
  OpenPath,
  GetExamAnswers,
  ListConfigs,
  EvaluateExam,
  PrintStudentSheet,
  DownloadStudentSheet,
  PickPDF,
  PickFolder,
  ExportExamTemplateCSV,
  GetSavePath,
  SetSavePath,
  GetExamStats,
  GenerateExamStatisticsPDF,
  ExportExamStudentsCSV,
  ParseExamTemplateCSV,
  PrintLegendPDF,
  UpdateExamAnswers,
} from '../wailsjs/go/main/App';
import { EventsOn } from '../wailsjs/runtime/runtime';

const state = {
  activeTab: 'exams',
  exams: [],
  students: [],
  configs: [],
  answers: [],
  csvContent: '',
  csvName: '',
  templateCsvName: '',
  formTitle: '',
  formSchoolYear: '',
  formDateTime: '',
  formQuestionCount: 10,
  formOptionCount: 5,
  formShowName: true,
  evalExamId: 0,
  evalFilePath: '',
  evalConfig: '',
  savePath: '',
  uploadExamId: 0,
  uploadFilePath: '',
  uploadConfig: '',
  statsExamId: 0,
  statsSelected: {},
};

const appEl = document.querySelector('#app');

function renderShell() {
  appEl.innerHTML = `
    <div class="page">
      <header class="header">
        <div class="title">
          <div class="badge">ScanEval</div>
          <h1>ScanEvalApp</h1>
       
        </div>
      </header>
      <nav class="tabs">
        <button class="tab" data-tab="exams">Písomky</button>
        <button class="tab" data-tab="create">Vytvoriť písomku</button>
        <button class="tab" data-tab="students">Študenti</button>
        <button class="tab" data-tab="upload">Vyhodnotiť písomku</button>
        <button class="tab" data-tab="settings">Nastavenia</button>
        <button class="tab" data-tab="legenda">Legenda</button>
      </nav>
      <section class="panel">
        <div id="content" class="panel-body"></div>
      </section>
    </div>
    <div id="modal" class="modal hidden">
      <div class="modal-card">
        <div class="modal-header">
          <h2 id="modal-title">Odpovede testu</h2>
          <button id="modal-close" class="btn ghost">Zatvorit</button>
        </div>
        <div id="modal-body" class="modal-body"></div>
      </div>
    </div>
  `;

  document.querySelectorAll('.tab').forEach((btn) => {
    btn.addEventListener('click', () => {
      state.activeTab = btn.dataset.tab;
      renderContent();
    });
  });

  document.getElementById('modal-close').addEventListener('click', hideModal);
  document.getElementById('modal').addEventListener('click', (event) => {
    if (event.target.id === 'modal') hideModal();
  });
}

const statisticsOptions = [
  'Maximum bodov',
  'Minimum bodov',
  'Priemer',
  'Medián',
  'Graf rozdelenia bodov celkovo',
  'Graf rozdelenia za jednotlivé príklady',
  'Úspešnosť absolútna aj relatívna',
  'Úspešnosť absolútna aj relatívna pre jednotlivé príklady',
];

// "DD.MM.YYYY HH:mm" -> "YYYY-MM-DD" for date input
function toDateInput(dtStr) {
  if (!dtStr) return '';
  const m = dtStr.match(/^(\d{2})\.(\d{2})\.(\d{4}) (\d{2}):(\d{2})$/);
  if (!m) return '';
  return `${m[3]}-${m[2]}-${m[1]}`;
}

// "DD.MM.YYYY HH:mm" -> "HH:mm" for time input
function toHourInput(dtStr) {
  if (!dtStr) return '08';
  const m = dtStr.match(/^(\d{2})\.(\d{2})\.(\d{4}) (\d{2}):(\d{2})$/);
  return m ? m[4] : '08';
}

function toMinuteInput(dtStr) {
  if (!dtStr) return '00';
  const m = dtStr.match(/^(\d{2})\.(\d{2})\.(\d{4}) (\d{2}):(\d{2})$/);
  if (!m) return '00';
  const min = parseInt(m[5], 10);
  return String(Math.round(min / 5) * 5).padStart(2, '0');
}

function makeHourOptions(selected) {
  return Array.from({ length: 24 }, (_, i) => {
    const v = String(i).padStart(2, '0');
    return `<option value="${v}" ${v === selected ? 'selected' : ''}>${v}</option>`;
  }).join('');
}

function makeMinuteOptions(selected) {
  return Array.from({ length: 12 }, (_, i) => {
    const v = String(i * 5).padStart(2, '0');
    return `<option value="${v}" ${v === selected ? 'selected' : ''}>${v}</option>`;
  }).join('');
}

// "YYYY-MM-DD" + "HH:mm" -> "DD.MM.YYYY HH:mm" for backend
function combineDatetime(dateStr, timeStr) {
  if (!dateStr || !timeStr) return '';
  const d = dateStr.match(/^(\d{4})-(\d{2})-(\d{2})$/);
  if (!d) return '';
  return `${d[3]}.${d[2]}.${d[1]} ${timeStr}`;
}

function formatDate(value) {
  if (!value) return '-';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return '-';
  return new Intl.DateTimeFormat('sk-SK', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  }).format(date);
}

// ==================== PRINT DIALOG ====================

async function showPrintDialog(examId) {
  const modalContent = `
    <div class="print-dialog">
      <div class="info-banner">
        <div class="info-icon">🖨️</div>
        <div class="info-text">
          <strong>Voľby tlače</strong><br>
          Vyberte, či chcete vytlačiť test s legendou alebo bez legendy.
        </div>
      </div>
     
      <div class="print-options">
        <label class="print-option">
          <input type="radio" name="printOption" value="with" checked>
          <div class="option-content">
            <span class="option-icon">📚</span>
            <div>
              <strong>S legendou</strong>
              <small>Prvá strana s návodom na vyplnenie</small>
            </div>
          </div>
        </label>
       
        <label class="print-option">
          <input type="radio" name="printOption" value="without">
          <div class="option-content">
            <span class="option-icon">📄</span>
            <div>
              <strong>Bez legendy</strong>
              <small>Iba testovacie hárky</small>
            </div>
          </div>
        </label>
      </div>
     
      <div class="form-actions">
        <button class="btn primary" id="confirm-print">Tlačiť</button>
        <button class="btn ghost" onclick="hideModal()">Zrušiť</button>
      </div>
      <div id="print-status" class="status"></div>
    </div>
  `;
 
  showModal('Tlač testu', modalContent);
 
  const confirmBtn = document.getElementById('confirm-print');
  if (confirmBtn) {
    confirmBtn.addEventListener('click', async () => {
      const selectedOption = document.querySelector('input[name="printOption"]:checked').value;
      const statusEl = document.getElementById('print-status');
     
      statusEl.textContent = 'Generujem PDF...';
      statusEl.className = 'status';
     
      try {
        let path;
        if (selectedOption === 'with') {
          path = await PrintExamPDFWithLegend(examId);
        } else {
          path = await PrintExamPDF(examId);
        }
       
        if (path) {
          statusEl.textContent = 'Hotovo!';
          statusEl.className = 'status success';
          await OpenPath(path);
          setTimeout(() => hideModal(), 1000);
        } else {
          statusEl.textContent = 'Chyba pri generovaní PDF';
          statusEl.className = 'status error';
        }
      } catch (err) {
        console.error(err);
        statusEl.textContent = 'Chyba: ' + (err?.message || err || 'neznáma chyba');
        statusEl.className = 'status error';
      }
    });
  }
}

function renderExams() {
  const content = document.getElementById('content');
  if (!state.exams.length) {
    content.innerHTML = `
      <div class="empty">
        <div class="empty-title">Ziadne testy</div>
        <div class="empty-sub">Najprv si vytvor novy test v aplikacii.</div>
      </div>
    `;
    return;
  }

  content.innerHTML = `
    <div class="exams-table">
    <div class="panel-header">
      <span>Názov</span>
      <span>Školský rok</span>
      <span>Dátum</span>
      <span>Otázky</span>
      <span>Možnosti</span>
      <span>Študenti</span>
      <span>Tlačiť</span>
      <span>Upraviť odpovede</span>
      <span>Vyhodnotiť</span>
      <span>Štatistika PDF</span>
      <span>CSV</span>
      <span>Zmazať</span>
    </div>
    ${state.exams
      .map(
        (exam) => `
        <div class="row">
          <span class="cell title" data-label="Názov">${exam.title}</span>
          <span class="cell" data-label="Školský rok">${exam.schoolYear}</span>
          <span class="cell" data-label="Dátum">${formatDate(exam.date)}</span>
          <span class="cell" data-label="Otázky">${exam.questionCount}</span>
          <span class="cell" data-label="Možnosti">${exam.optionCount || '-'}</span>
          <span class="cell" data-label="Študenti">${exam.studentCount}</span>
          <span class="cell" data-label="Tlačiť">
            <button class="btn" data-action="print" data-exam-id="${exam.id}">Tlačiť</button>
          </span>
          <span class="cell" data-label="Odpovede">
            <button class="btn" data-action="answers" data-exam-id="${exam.id}">Upraviť odpovede</button>
          </span>
          <span class="cell" data-label="Vyhodnotiť">
            <button class="btn" data-action="evaluate" data-exam-id="${exam.id}">Vyhodnotiť</button>
          </span>
          <span class="cell" data-label="Štatistika PDF">
            <button class="btn" data-action="stats-pdf" data-exam-id="${exam.id}">Štatistika</button>
          </span>
          <span class="cell" data-label="CSV">
            <button class="btn" data-action="csv" data-exam-id="${exam.id}">CSV</button>
          </span>
          <span class="cell" data-label="Zmazať">
            <button class="btn danger" data-action="delete" data-exam-id="${exam.id}">Zmazať</button>
          </span>
        </div>
      `,
      )
      .join('')}
    </div>
  `;
}

function showModal(title, content) {
  const modal = document.getElementById('modal');
  const body = document.getElementById('modal-body');
  const titleEl = document.getElementById('modal-title');
  if (titleEl) titleEl.textContent = title;
  body.innerHTML = content;
  modal.classList.remove('hidden');
}

function hideModal() {
  const modal = document.getElementById('modal');
  modal.classList.add('hidden');
}

function renderStudents() {
  const content = document.getElementById('content');
  if (!state.students.length) {
    content.innerHTML = `
      <div class="empty">
        <div class="empty-title">Žiadni študenti</div>
        <div class="empty-sub">Importuj študentov cez vytvorenie testu.</div>
      </div>
    `;
    return;
  }

  content.innerHTML = `
    <div class="students-table">
    <div class="panel-header">
      <span>Meno</span>
      <span>Priezvisko</span>
      <span>Dátum narodenia</span>
      <span>Mestnosť</span>
      <span>Reg. číslo</span>
      <span>Písomka</span>
      <span>Skóre</span>
      <span>Akcie</span>
    </div>
    ${state.students
      .map(
        (student) => `
        <div class="row">
          <span class="cell title" data-label="Meno">${student.name}</span>
          <span class="cell" data-label="Priezvisko">${student.surname}</span>
          <span class="cell" data-label="Datum narodenia">${formatDate(student.birthDate)}</span>
          <span class="cell" data-label="Miestnost">${student.room}</span>
          <span class="cell" data-label="Reg cislo">${student.registrationNumber}</span>
          <span class="cell" data-label="Test">${student.examId}</span>
          <span class="cell" data-label="Score">${student.score}</span>
          <span class="cell action" data-label="Akcie">
            <button class="btn" data-student-print="${student.id}">Tlačiť hárok</button>
         
          </span>
        </div>
      `,
      )
      .join('')}
    </div>
  `;

  document.querySelectorAll('[data-student-print]').forEach((btn) => {
    btn.addEventListener('click', async () => {
      const studentId = Number(btn.dataset.studentPrint);
      if (!studentId) return;
      try {
        const path = await PrintStudentSheet(studentId);
        if (path) {
          await OpenPath(path);
        }
      } catch (err) {
        console.error(err);
        window.alert('Tlac harku zlyhala. Skontroluj logs.');
      }
    });
  });

  document.querySelectorAll('[data-student-download]').forEach((btn) => {
    btn.addEventListener('click', async () => {
      const studentId = Number(btn.dataset.studentDownload);
      if (!studentId) return;
      try {
        const path = await DownloadStudentSheet(studentId);
        if (path) {
          await OpenPath(path);
        }
      } catch (err) {
        console.error(err);
        window.alert('Stiahnutie harku zlyhalo. Skontroluj logs.');
      }
    });
  });
}

function renderAnswerSelectors(container) {
  const options = ['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H'];
  const optionCount = Number(document.getElementById('option-count').value) || 5;
  const questionCount = Number(document.getElementById('question-count').value) || 0;
  const maxOptions = Math.min(Math.max(optionCount, 2), options.length);
  const activeOptions = options.slice(0, maxOptions);

  state.answers = Array.from({ length: questionCount }, (_, i) => state.answers[i] || '');
  container.innerHTML = state.answers
    .map(
      (_, idx) => `
        <div class="answer-row">
          <span>Otazka ${String(idx + 1).padStart(2, '0')}</span>
          ${activeOptions
            .map(
              (opt) => `
              <label class="radio">
                <input type="radio" name="q${idx}" value="${opt}" ${state.answers[idx] === opt ? 'checked' : ''}/>
                <span>${opt}</span>
              </label>
            `,
            )
            .join('')}
        </div>
      `,
    )
    .join('');

  container.querySelectorAll('input[type="radio"]').forEach((input) => {
    input.addEventListener('change', () => {
      const qIndex = Number(input.name.replace('q', '')) || 0;
      state.answers[qIndex] = input.value;
    });
  });
}

function syncAnswersState() {
  const questionCount = Math.max(Number(state.formQuestionCount) || 0, 0);
  state.answers = Array.from({ length: questionCount }, (_, i) => state.answers[i] || '');
}

function getStudentCsvLabel() {
  if (state.csvName) return state.csvName;
  if (state.csvContent) return 'Nacitane zo sablony';
  return 'Ziadny subor';
}

function getTemplateCsvLabel() {
  return state.templateCsvName || 'Ziadny subor';
}

function renderCreateExam() {
  const content = document.getElementById('content');
  syncAnswersState();
  content.innerHTML = `
    <form id="create-exam" class="form">
      <div class="form-grid">
        <label>
          Nazov testu
          <input id="title" type="text" value="${state.formTitle}" required />
        </label>
        <label>
          Skolsky rok (YYYY/YY)
          <input id="school-year" type="text" placeholder="2024/25" value="${state.formSchoolYear}" required />
        </label>
        <label>
          Dátum a čas
          <div class="datetime-row">
            <input id="date-part" type="date" value="${toDateInput(state.formDateTime)}" required />
            <div class="time-selects">
              <select id="hour-part">${makeHourOptions(toHourInput(state.formDateTime))}</select>
              <span class="time-sep">:</span>
              <select id="minute-part">${makeMinuteOptions(toMinuteInput(state.formDateTime))}</select>
            </div>
          </div>
        </label>
        <label>
          Pocet otazok
          <input id="question-count" type="number" min="1" value="${state.formQuestionCount}" required />
        </label>
        <label>
          Pocet moznosti
          <input id="option-count" type="number" min="2" max="8" value="${state.formOptionCount}" required />
        </label>
        <label class="toggle-field">
          Zobrazit meno
          <span class="toggle-row">
            <input id="show-name" type="checkbox" ${state.formShowName ? 'checked' : ''} />
            <span class="toggle-text">${state.formShowName ? 'Ano' : 'Nie'}</span>
          </span>
        </label>
        <label class="file">
          CSV so studentmi
          <input id="csv-file" type="file" accept=".csv" />
          <span class="file-name">${getStudentCsvLabel()}</span>
        </label>
        <label class="file">
          CSV sablona testu
          <input id="template-csv-file" type="file" accept=".csv" />
          <span class="file-name">${getTemplateCsvLabel()}</span>
        </label>
      </div>
      <div class="form-actions">
        <button type="button" id="generate" class="secondary">Generovat otazky</button>
        <button type="button" id="save-template" class="secondary">Ulozit CSV sablonu</button>
        <button type="submit" class="primary">Vytvorit test</button>
        <span id="form-status" class="status"></span>
      </div>
      <div id="answers" class="answers"></div>
    </form>
  `;

  const form = document.getElementById('create-exam');
  const statusEl = document.getElementById('form-status');
  const answersEl = document.getElementById('answers');
  const fileInput = document.getElementById('csv-file');
  const templateFileInput = document.getElementById('template-csv-file');

  const titleEl = document.getElementById('title');
  const schoolYearEl = document.getElementById('school-year');
  const datePartEl = document.getElementById('date-part');
  const hourPartEl = document.getElementById('hour-part');
  const minutePartEl = document.getElementById('minute-part');
  const questionCountEl = document.getElementById('question-count');
  const optionCountEl = document.getElementById('option-count');
  const showNameEl = document.getElementById('show-name');

  titleEl.addEventListener('input', () => {
    state.formTitle = titleEl.value;
  });
  schoolYearEl.addEventListener('input', () => {
    state.formSchoolYear = schoolYearEl.value;
  });
  const syncDateTime = () => {
    state.formDateTime = combineDatetime(datePartEl.value, `${hourPartEl.value}:${minutePartEl.value}`);
  };
  datePartEl.addEventListener('change', syncDateTime);
  hourPartEl.addEventListener('change', syncDateTime);
  minutePartEl.addEventListener('change', syncDateTime);
  questionCountEl.addEventListener('input', () => {
    state.formQuestionCount = Number(questionCountEl.value) || 0;
    syncAnswersState();
  });
  optionCountEl.addEventListener('input', () => {
    state.formOptionCount = Number(optionCountEl.value) || 0;
  });
  showNameEl.addEventListener('change', () => {
    state.formShowName = showNameEl.checked;
    const toggleText = document.querySelector('.toggle-text');
    if (toggleText) toggleText.textContent = state.formShowName ? 'Ano' : 'Nie';
  });

  document.getElementById('generate').addEventListener('click', () => {
    renderAnswerSelectors(answersEl);
  });

  fileInput.addEventListener('change', async () => {
    const file = fileInput.files?.[0];
    if (!file) {
      state.csvContent = '';
      state.csvName = '';
      renderCreateExam();
      return;
    }
    state.csvName = file.name;
    const text = await file.text();
    state.csvContent = text;
    renderCreateExam();
  });

  templateFileInput.addEventListener('change', async () => {
    const file = templateFileInput.files?.[0];
    if (!file) {
      state.templateCsvName = '';
      renderCreateExam();
      return;
    }

    statusEl.textContent = 'Nacitavam sablonu...';
    statusEl.className = 'status';

    try {
      const text = await file.text();
      const template = await ParseExamTemplateCSV(text);
      state.templateCsvName = file.name;
      state.formTitle = template.title || '';
      state.formSchoolYear = template.schoolYear || '';
      state.formDateTime = template.dateTime || '';
      state.formQuestionCount = Number(template.questionCount) || 0;
      state.formOptionCount = Number(template.optionCount) || 5;
      state.formShowName = Boolean(template.showName);
      state.answers = Array.isArray(template.answers) ? template.answers : [];
      state.csvContent = template.studentCSVContent || '';
      state.csvName = template.studentCSVContent ? 'Nacitane zo sablony' : '';
      renderCreateExam();
      const nextStatusEl = document.getElementById('form-status');
      if (nextStatusEl) {
        nextStatusEl.textContent = 'Sablona nacitana.';
        nextStatusEl.className = 'status success';
      }
    } catch (err) {
      console.error(err);
      statusEl.textContent = 'Chyba pri nacitani sablony.';
      statusEl.className = 'status error';
    }
  });

  document.getElementById('save-template').addEventListener('click', async () => {
    statusEl.textContent = 'Ukladam sablonu...';
    statusEl.className = 'status';
    syncAnswersState();

    try {
      const path = await ExportExamTemplateCSV(
        state.formTitle.trim(),
        state.formSchoolYear.trim(),
        state.formDateTime.trim(),
        Number(state.formQuestionCount) || 0,
        Number(state.formOptionCount) || 0,
        state.answers,
        state.csvContent,
        Boolean(state.formShowName),
      );
      statusEl.textContent = path ? `Sablona ulozena: ${path}` : 'Sablona ulozena.';
      statusEl.className = 'status success';
    } catch (err) {
      console.error(err);
      statusEl.textContent = 'Chyba pri ukladani sablony.';
      statusEl.className = 'status error';
    }
  });

  form.addEventListener('submit', async (event) => {
    event.preventDefault();
    statusEl.textContent = 'Ukladam...';
    statusEl.className = 'status';

    const title = state.formTitle.trim();
    const schoolYear = state.formSchoolYear.trim();
    const dateTime = state.formDateTime.trim();
    const questionCount = Number(state.formQuestionCount);
    const optionCount = Number(state.formOptionCount);
    const showName = Boolean(state.formShowName);

    renderAnswerSelectors(answersEl);

    try {
      await CreateExamWithCSV(title, schoolYear, dateTime, questionCount, optionCount, state.answers, state.csvContent, showName);
      statusEl.textContent = 'Ulozene.';
      statusEl.className = 'status success';
      state.csvContent = '';
      state.csvName = '';
      state.templateCsvName = '';
      state.answers = [];
      state.formTitle = '';
      state.formSchoolYear = '';
      state.formDateTime = '';
      state.formQuestionCount = 10;
      state.formOptionCount = 5;
      state.formShowName = true;
      await refreshData();
      state.activeTab = 'exams';
      renderContent();
    } catch (err) {
      console.error(err);
      statusEl.textContent = 'Chyba pri ukladani: ' + (err?.message || err || 'neznama chyba');
      statusEl.className = 'status error';
    }
  });

  renderAnswerSelectors(answersEl);
}

function renderPlaceholder(title, text) {
  const content = document.getElementById('content');
  content.innerHTML = `
    <div class="empty">
      <div class="empty-title">${title}</div>
      <div class="empty-sub">${text}</div>
    </div>
  `;
}

function renderEvalModal() {
  const options = state.configs
    .map((cfg) => `<option value="${cfg}" ${cfg === state.evalConfig ? 'selected' : ''}>${cfg}</option>`)
    .join('');

  showModal(
    'Vyhodnotenie pisomiek',
    `
    <div class="eval-form">
      <label>
        Konfiguracia skenera
        <select id="eval-config">${options}</select>
      </label>
      <label>
        PDF sken
        <div class="file-row">
          <span class="file-path">${state.evalFilePath || 'Ziadny subor'}</span>
          <button id="eval-pick" class="btn">Vybrat PDF</button>
        </div>
      </label>
      <div class="form-actions">
        <button id="eval-start" class="btn primary">Spustit vyhodnotenie</button>
        <span id="eval-status" class="status"></span>
      </div>
    </div>
    `,
  );

  const configEl = document.getElementById('eval-config');
  const pickBtn = document.getElementById('eval-pick');
  const startBtn = document.getElementById('eval-start');

  if (configEl) {
    configEl.addEventListener('change', () => {
      state.evalConfig = configEl.value;
    });
  }

  if (pickBtn) {
    pickBtn.addEventListener('click', async () => {
      const path = await PickPDF();
      if (path) {
        state.evalFilePath = path;
        renderEvalModal();
      }
    });
  }

  if (startBtn) {
    startBtn.addEventListener('click', async () => {
      const statusEl = document.getElementById('eval-status');
      if (!state.evalFilePath || !state.evalConfig) {
        if (statusEl) statusEl.textContent = 'Vyber PDF a konfiguraciu.';
        return;
      }
      if (statusEl) statusEl.textContent = 'Spustam...';
      try {
        await EvaluateExam(state.evalExamId, state.evalFilePath, state.evalConfig);
      } catch (err) {
        console.error(err);
        if (statusEl) statusEl.textContent = 'Spustenie zlyhalo.';
      }
    });
  }
}

function renderStatsModal() {
  const items = statisticsOptions
    .map((opt) => {
      const checked = state.statsSelected[opt] ? 'checked' : '';
      return `
        <label class="stats-item">
          <input type="checkbox" data-stat="${opt}" ${checked} />
          <span>${opt}</span>
        </label>
      `;
    })
    .join('');

  showModal(
    'Vyberte pozadovane statistiky',
    `
    <div class="stats-form">
      <div class="stats-list">
        ${items}
      </div>
      <div class="form-actions">
        <button id="stats-generate" class="btn primary">Generovat statistiky</button>
        <span id="stats-status" class="status"></span>
      </div>
    </div>
    `,
  );

  document.querySelectorAll('[data-stat]').forEach((input) => {
    input.addEventListener('change', () => {
      const key = input.dataset.stat;
      state.statsSelected[key] = input.checked;
    });
  });

  const generateBtn = document.getElementById('stats-generate');
  if (generateBtn) {
    generateBtn.addEventListener('click', async () => {
      const statusEl = document.getElementById('stats-status');
      const selected = statisticsOptions.filter((opt) => state.statsSelected[opt]);
      if (selected.length === 0) {
        if (statusEl) statusEl.textContent = 'Vyber aspon jednu statistiku.';
        return;
      }
      if (statusEl) statusEl.textContent = 'Generujem...';
      try {
        const path = await GenerateExamStatisticsPDF(state.statsExamId, selected);
        if (path) {
          await OpenPath(path);
          if (statusEl) statusEl.textContent = 'Hotovo.';
        } else if (statusEl) {
          statusEl.textContent = 'PDF sa nepodarilo vytvorit.';
        }
      } catch (err) {
        console.error(err);
        if (statusEl) statusEl.textContent = 'Generovanie zlyhalo.';
      }
    });
  }
}

function renderContent() {
  document.querySelectorAll('.tab').forEach((btn) => {
    btn.classList.toggle('active', btn.dataset.tab === state.activeTab);
  });

  if (state.activeTab === 'exams') {
    renderExams();
    bindExamActions();
    return;
  }
  if (state.activeTab === 'create') {
    renderCreateExam();
    return;
  }
  if (state.activeTab === 'students') {
    renderStudents();
    return;
  }
  if (state.activeTab === 'upload') {
    renderUpload();
    return;
  }
  if (state.activeTab === 'settings') {
    renderSettings();
    return;
  }
  if (state.activeTab === 'legenda') {
    renderLegenda();
  }
}

function renderLegenda() {
  const content = document.getElementById('content');
  content.innerHTML = `
    <div class="legenda-page">
      <h2>Legenda</h2>
      <p>Vygeneruje PDF s návodom na vyplnenie testu a uloží ho do nastaveného priečinka.</p>
      <button class="btn primary" id="print-legenda">Tlačiť legendu</button>
      <span id="legenda-status" class="status"></span>
    </div>
  `;

  document.getElementById('print-legenda').addEventListener('click', async () => {
    const statusEl = document.getElementById('legenda-status');
    statusEl.textContent = 'Generujem PDF...';
    statusEl.className = 'status';
    try {
      const path = await PrintLegendPDF();
      statusEl.textContent = 'Ulozene: ' + path;
      statusEl.className = 'status success';
      if (path) await OpenPath(path);
    } catch (err) {
      console.error(err);
      statusEl.textContent = 'Chyba: ' + (err?.message || err || 'neznama chyba');
      statusEl.className = 'status error';
    }
  });
}

function renderUpload() {
  const content = document.getElementById('content');
  const examOptions = state.exams
    .map(
      (exam) =>
        `<option value="${exam.id}" ${Number(exam.id) === Number(state.uploadExamId) ? 'selected' : ''}>${exam.title}</option>`,
    )
    .join('');
  const configOptions = state.configs
    .map((cfg) => `<option value="${cfg}" ${cfg === state.uploadConfig ? 'selected' : ''}>${cfg}</option>`)
    .join('');

  content.innerHTML = `
    <div class="eval-form">
      <label>
        Test
        <select id="upload-exam">${examOptions}</select>
      </label>
      <label>
        Konfiguracia skenera
        <select id="upload-config">${configOptions}</select>
      </label>
      <label>
        PDF sken
        <div class="file-row">
          <span class="file-path">${state.uploadFilePath || 'Ziadny subor'}</span>
          <button id="upload-pick" class="btn">Vybrat PDF</button>
        </div>
      </label>
      <div class="form-actions">
        <button id="upload-start" class="btn primary">Spustit vyhodnotenie</button>
        <span id="upload-status" class="status"></span>
      </div>
    </div>
  `;

  const examEl = document.getElementById('upload-exam');
  const configEl = document.getElementById('upload-config');
  const pickBtn = document.getElementById('upload-pick');
  const startBtn = document.getElementById('upload-start');

  if (examEl) {
    examEl.addEventListener('change', () => {
      state.uploadExamId = Number(examEl.value);
    });
  }

  if (configEl) {
    configEl.addEventListener('change', () => {
      state.uploadConfig = configEl.value;
    });
  }

  if (pickBtn) {
    pickBtn.addEventListener('click', async () => {
      const path = await PickPDF();
      if (path) {
        state.uploadFilePath = path;
        renderUpload();
      }
    });
  }

  if (startBtn) {
    startBtn.addEventListener('click', async () => {
      const statusEl = document.getElementById('upload-status');
      if (!state.uploadExamId || !state.uploadConfig || !state.uploadFilePath) {
        if (statusEl) statusEl.textContent = 'Vyber test, konfiguraciu a PDF.';
        return;
      }
      if (statusEl) statusEl.textContent = 'Spustam...';
      try {
        await EvaluateExam(state.uploadExamId, state.uploadFilePath, state.uploadConfig);
      } catch (err) {
        console.error(err);
        if (statusEl) statusEl.textContent = 'Spustenie zlyhalo.';
      }
    });
  }
}

function renderSettings() {
  const content = document.getElementById('content');
  content.innerHTML = `
    <div class="settings-form">
      <label>
        Miesto ukladania PDF
        <div class="file-row">
          <span class="file-path">${state.savePath || 'Nenastavene'}</span>
          <button id="settings-pick" class="btn">Vybrat priecinok</button>
        </div>
      </label>
      <div class="form-actions">
        <button id="settings-save" class="btn primary">Ulozit</button>
        <span id="settings-status" class="status"></span>
      </div>
    </div>
  `;

  const pickBtn = document.getElementById('settings-pick');
  const saveBtn = document.getElementById('settings-save');

  if (pickBtn) {
    pickBtn.addEventListener('click', async () => {
      const path = await PickFolder();
      if (path) {
        state.savePath = path;
        renderSettings();
      }
    });
  }

  if (saveBtn) {
    saveBtn.addEventListener('click', async () => {
      const statusEl = document.getElementById('settings-status');
      if (!state.savePath) {
        if (statusEl) statusEl.textContent = 'Vyber priecinok.';
        return;
      }
      if (statusEl) statusEl.textContent = 'Ukladam...';
      try {
        await SetSavePath(state.savePath);
        if (statusEl) statusEl.textContent = 'Ulozene.';
      } catch (err) {
        console.error(err);
        if (statusEl) statusEl.textContent = 'Ulozenie zlyhalo.';
      }
    });
  }
}

function bindExamActions() {
  document.querySelectorAll('[data-exam-id]').forEach((btn) => {
    btn.addEventListener('click', async () => {
      const examId = Number(btn.dataset.examId);
      if (!examId) return;

      const action = btn.dataset.action;
      if (action === 'print') {
        // Show the print dialog instead of directly printing
        showPrintDialog(examId);
        return;
      }
      if (action === 'answers') {
        try {
          const exam = state.exams.find((e) => e.id === examId);
          const currentAnswers = await GetExamAnswers(examId);
          const options = ['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H'];
          const questionCount = exam ? exam.questionCount : (currentAnswers ? currentAnswers.length : 0);
          const optionCount = exam ? (exam.optionCount || 5) : 5;
          const activeOptions = options.slice(0, Math.min(Math.max(optionCount, 2), options.length));
          const answersArr = Array.from({ length: questionCount }, (_, i) =>
            currentAnswers && currentAnswers[i] ? currentAnswers[i].toUpperCase() : '',
          );

          const rows = answersArr
            .map(
              (ans, idx) => `
              <div class="answer-row">
                <span>Otazka ${String(idx + 1).padStart(2, '0')}</span>
                ${activeOptions
                  .map(
                    (opt) => `
                  <label class="radio">
                    <input type="radio" name="eq${idx}" value="${opt}" ${ans === opt ? 'checked' : ''} />
                    <span>${opt}</span>
                  </label>`,
                  )
                  .join('')}
              </div>`,
            )
            .join('');

          showModal(
            'Upraviť odpovede',
            `<div class="answers" id="edit-answers-container">${rows}</div>
             <div style="margin-top:12px;display:flex;align-items:center;gap:12px;">
               <button class="btn primary" id="save-answers-btn">Uložiť</button>
               <span id="save-answers-status" class="status"></span>
             </div>`,
          );

          const editAnswers = [...answersArr];
          document.querySelectorAll('#edit-answers-container input[type="radio"]').forEach((input) => {
            input.addEventListener('change', () => {
              const qIndex = Number(input.name.replace('eq', '')) || 0;
              editAnswers[qIndex] = input.value;
            });
          });

          document.getElementById('save-answers-btn').addEventListener('click', async () => {
            const statusEl = document.getElementById('save-answers-status');
            statusEl.textContent = 'Ukladam...';
            statusEl.className = 'status';
            try {
              await UpdateExamAnswers(examId, editAnswers);
              statusEl.textContent = 'Ulozene.';
              statusEl.className = 'status success';
            } catch (err) {
              console.error(err);
              statusEl.textContent = 'Chyba: ' + (err?.message || err || 'neznama chyba');
              statusEl.className = 'status error';
            }
          });
        } catch (err) {
          console.error(err);
          showModal('Upraviť odpovede', '<div class="error">Nacitavanie odpovedi zlyhalo.</div>');
        }
        return;
      }
      if (action === 'evaluate') {
        state.evalExamId = examId;
        state.evalFilePath = '';
        if (!state.evalConfig && state.configs.length > 0) {
          state.evalConfig = state.configs[0];
        }
        renderEvalModal();
        return;
      }
      if (action === 'stats-pdf') {
        state.statsExamId = examId;
        if (!Object.keys(state.statsSelected).length) {
          statisticsOptions.forEach((opt) => {
            state.statsSelected[opt] = true;
          });
        }
        renderStatsModal();
        return;
      }
      if (action === 'csv') {
        try {
          const path = await ExportExamStudentsCSV(examId);
          if (path) {
            await OpenPath(path);
          }
        } catch (err) {
          console.error(err);
          showModal('Export CSV', '<div class="error">Export studentov do CSV zlyhal.</div>');
        }
        return;
      }

      const confirmed = window.confirm('Naozaj chces zmazat tento test?');
      if (!confirmed) return;

      try {
        await DeleteExam(examId);
        await refreshData();
        renderContent();
      } catch (err) {
        console.error(err);
        window.alert('Zmazanie zlyhalo. Skontroluj logs.');
      }
    });
  });
}

async function refreshData() {
  try {
    const [exams, students, configs, savePath] = await Promise.all([
      ListExams(),
      ListStudents(),
      ListConfigs(),
      GetSavePath(),
    ]);
    state.exams = exams || [];
    state.students = students || [];
    state.configs = configs || [];
    if (!state.evalConfig && state.configs.length > 0) {
      state.evalConfig = state.configs[0];
    }
    state.savePath = savePath || '';
    if (!state.uploadConfig && state.configs.length > 0) {
      state.uploadConfig = state.configs[0];
    }
    if (!state.uploadExamId && state.exams.length > 0) {
      state.uploadExamId = state.exams[0].id;
    }
  } catch (err) {
    console.error(err);
    state.exams = [];
    state.students = [];
    state.configs = [];
    state.savePath = '';
    state.uploadConfig = '';
    state.uploadExamId = 0;
  }
}

async function init() {
  renderShell();
  await refreshData();
  renderContent();

  EventsOn('evaluation_progress', (msg) => {
    const statusEl = document.getElementById('eval-status');
    if (statusEl) statusEl.textContent = msg;
    const uploadStatus = document.getElementById('upload-status');
    if (uploadStatus) uploadStatus.textContent = msg;
  });
  EventsOn('evaluation_error', (msg) => {
    showModal('Vyhodnotenie pisomiek', `<div class="error">${msg}</div>`);
  });
  EventsOn('evaluation_done', (payload) => {
    let message = 'Vyhodnotenie dokoncene.';
    let extra = '';
    if (payload?.hadFailures && payload.failedPath) {
      message = 'Niektore strany sa nepodarilo spracovat.';
      extra = `<div class="file-row"><span class="file-path">${payload.failedPath}</span><button id="open-failed" class="btn">Otvorit PDF</button></div>`;
    }
    showModal('Vyhodnotenie pisomiek', `<div class="status success">${message}</div>${extra}`);
    if (payload?.failedPath) {
      const btn = document.getElementById('open-failed');
      if (btn) {
        btn.addEventListener('click', () => OpenPath(payload.failedPath));
      }
    }
    refreshData();
  });
}


window.hideModal = hideModal;
window.showPrintDialog = showPrintDialog;

init();
