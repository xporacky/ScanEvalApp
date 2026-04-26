import './style.css';
import './app.css';

import {
  CreateExamWithCSV,
  CreateMultiDayExamWithCSV,
  PrintMultiDayExamPDFs,
  UpdateMultiDayAnswers,
  ExportMultiDayResultsCSV,
  DeleteExam,
  ListExams,
  ListStudents,
  PrintExamPDF,
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
  UpdateExamAnswers,
  PrintLegend,
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
  studentFilter: { name: '', surname: '', regNum: '' },
  mdTitle: '',
  mdSchoolYear: '',
  mdQuestionCount: 10,
  mdOptionCount: 5,
  mdShowName: true,
  mdCsvContent: '',
  mdCsvName: '',
  mdSubgroups: [{ name: '', answers: '' }],
  mdCheckboxMode: true,
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
        <button class="tab" data-tab="multiday">Multi-termín</button>
        <button class="tab" data-tab="students">Študenti</button>
        <button class="tab" data-tab="upload">Vyhodnotiť písomku</button>
        <button class="tab" data-tab="settings">Nastavenia</button>
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
    <div class="exams-page">
    <div class="exams-toolbar">
      <button class="btn secondary" id="btn-print-legend">Tlačiť legendu</button>
    </div>
    <div class="exams-table">
    <div class="panel-header">
      <span>Názov</span>
      <span>Školský rok</span>
      <span>Dátum</span>
      <span>Otázky</span>
      <span>Možnosti</span>
      <span>Študenti</span>
      <span>Tlačiť</span>
      <span>Odpovede</span>
      <span>Vyhodnotiť</span>
      <span>Štatistika PDF</span>
      <span>CSV</span>
      <span>Zmazať</span>
    </div>
    ${state.exams
      .map(
        (exam) => `
        <div class="row">
          <span class="cell title" data-label="Názov">${exam.title}${exam.isMultiDay ? ' <span class="badge-multi">multi</span>' : ''}</span>
          <span class="cell" data-label="Školský rok">${exam.schoolYear}</span>
          <span class="cell" data-label="Dátum">${formatDate(exam.date)}</span>
          <span class="cell" data-label="Otázky">${exam.questionCount}</span>
          <span class="cell" data-label="Možnosti">${exam.optionCount || '-'}</span>
          <span class="cell" data-label="Študenti">${exam.studentCount}</span>
          <span class="cell" data-label="Tlačiť">
            ${exam.isMultiDay
              ? `<button class="btn" data-action="print-multiday" data-exam-id="${exam.id}">Tlačiť po dňoch</button>`
              : `<button class="btn" data-action="print" data-exam-id="${exam.id}">Tlačiť</button>`
            }
          </span>
          <span class="cell" data-label="Odpovede">
            <button class="btn" data-action="answers" data-exam-id="${exam.id}">Odpovede</button>
          </span>
          <span class="cell" data-label="Vyhodnotiť">
            <button class="btn" data-action="evaluate" data-exam-id="${exam.id}">Vyhodnotiť</button>
          </span>
          <span class="cell" data-label="Štatistika PDF">
            <button class="btn" data-action="stats-pdf" data-exam-id="${exam.id}">Štatistika</button>
          </span>
          <span class="cell" data-label="CSV">
            ${exam.isMultiDay
              ? `<button class="btn" data-action="csv-multiday" data-exam-id="${exam.id}">Výsledky CSV</button>`
              : `<button class="btn" data-action="csv" data-exam-id="${exam.id}">CSV</button>`
            }
          </span>
          <span class="cell" data-label="Zmazať">
            <button class="btn danger" data-action="delete" data-exam-id="${exam.id}">Zmazať</button>
          </span>
        </div>
      `,
      )
      .join('')}
    </div>
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

function stripDiacritics(str) {
  return str.normalize('NFD').replace(/[\u0300-\u036f]/g, '');
}

function applyStudentFilter() {
  const nameFilter = stripDiacritics(state.studentFilter.name.toLowerCase().trim());
  const surnameFilter = stripDiacritics(state.studentFilter.surname.toLowerCase().trim());
  const regNumFilter = stripDiacritics(state.studentFilter.regNum.toLowerCase().trim());

  document.querySelectorAll('.students-table .row').forEach((row) => {
    const name = stripDiacritics((row.querySelector('[data-label="Meno"]')?.textContent || '').toLowerCase());
    const surname = stripDiacritics((row.querySelector('[data-label="Priezvisko"]')?.textContent || '').toLowerCase());
    const regNum = (row.querySelector('[data-label="Reg cislo"]')?.textContent || '').toLowerCase();

    const visible =
      name.includes(nameFilter) &&
      surname.includes(surnameFilter) &&
      regNum.includes(regNumFilter);

    row.style.display = visible ? '' : 'none';
  });
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
    <div class="students-filter">
      <input type="text" id="filter-name" placeholder="Hľadaj podľa mena…" value="${state.studentFilter.name}">
      <input type="text" id="filter-surname" placeholder="Hľadaj podľa priezviska…" value="${state.studentFilter.surname}">
      <input type="text" id="filter-regnum" placeholder="Hľadaj podľa reg. čísla…" value="${state.studentFilter.regNum}">
    </div>
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
          <span class="cell" data-label="Reg cislo">${String(student.registrationNumber).padStart(7, '0')}</span>
          <span class="cell" data-label="Test">${student.examId}</span>
          <span class="cell" data-label="Score">${student.score}</span>
          <span class="cell action" data-label="Akcie">
            <button class="btn" data-student-print="${student.id}">Nový hárok</button>
            ${student.pages ? `<button class="btn secondary" data-student-download="${student.id}">Vypracovaný hárok</button>` : ""}
          </span>
        </div>
      `,
      )
      .join('')}
    </div>
  `;

  applyStudentFilter();

  document.getElementById('filter-name').addEventListener('input', (e) => {
    state.studentFilter.name = e.target.value;
    applyStudentFilter();
  });
  document.getElementById('filter-surname').addEventListener('input', (e) => {
    state.studentFilter.surname = e.target.value;
    applyStudentFilter();
  });
  document.getElementById('filter-regnum').addEventListener('input', (e) => {
    state.studentFilter.regNum = e.target.value;
    applyStudentFilter();
  });

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
  const maxOptions = Math.min(Math.max(optionCount, 3), options.length);
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
          Datum (dd.MM.yyyy)
          <input id="date-time" type="text" placeholder="01.03.2025" value="${state.formDateTime}" required />
        </label>
        <label>
          Pocet otazok
          <input id="question-count" type="number" min="1" value="${state.formQuestionCount}" required />
        </label>
        <label>
          Pocet moznosti
          <input id="option-count" type="number" min="3" max="8" value="${state.formOptionCount}" required />
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
  const dateTimeEl = document.getElementById('date-time');
  const questionCountEl = document.getElementById('question-count');
  const optionCountEl = document.getElementById('option-count');
  const showNameEl = document.getElementById('show-name');

  titleEl.addEventListener('input', () => {
    state.formTitle = titleEl.value;
  });
  schoolYearEl.addEventListener('input', () => {
    state.formSchoolYear = schoolYearEl.value;
  });
  dateTimeEl.addEventListener('input', () => {
    state.formDateTime = dateTimeEl.value;
  });
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
      statusEl.textContent = 'Chyba pri ukladani.';
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
  if (state.activeTab === 'multiday') {
    renderMultiDay();
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
  }
}

function validAnswerChars(optionCount) {
  return Array.from({ length: optionCount }, (_, i) => String.fromCharCode(65 + i)).join('');
}

function validateSubgroupAnswers(answers, questionCount, optionCount) {
  if (answers.length !== questionCount) return false;
  const valid = validAnswerChars(optionCount).toLowerCase();
  return answers.toLowerCase().split('').every(c => valid.includes(c));
}

// Converts a string of answers (e.g. "AABCE") to per-question checkbox arrays
function answersToCheckboxes(answers, questionCount, optionCount) {
  const result = [];
  for (let i = 0; i < questionCount; i++) {
    result.push(i < answers.length ? answers[i].toUpperCase() : '');
  }
  return result;
}

// Converts per-question checkbox selections back to answer string
function checkboxesToAnswers(checkboxes) {
  return checkboxes.map(c => c || '').join('').toUpperCase();
}

function renderSubgroupCheckboxes(sg, idx, questionCount, optionCount) {
  const choices = Array.from({ length: optionCount }, (_, i) => String.fromCharCode(65 + i));
  const cbs = sg.checkboxAnswers || answersToCheckboxes(sg.answers, questionCount, optionCount);
  const rows = Array.from({ length: questionCount }, (_, q) => {
    const cells = choices.map(ch => {
      const checked = cbs[q] === ch ? 'checked' : '';
      return `<label class="sg-cb-label"><input type="radio" name="sg-${idx}-q${q}" value="${ch}" ${checked} />${ch}</label>`;
    }).join('');
    return `<div class="sg-cb-row"><span class="sg-cb-qnum">${q + 1}.</span>${cells}</div>`;
  }).join('');
  return `<div class="sg-cb-grid">${rows}</div>`;
}

function renderSubgroupRows(questionCount, optionCount) {
  const cbMode = state.mdCheckboxMode;
  return state.mdSubgroups.map((sg, idx) => {
    if (!sg.checkboxAnswers) {
      sg.checkboxAnswers = answersToCheckboxes(sg.answers, questionCount, optionCount);
    }
    const isValid = sg.answers.length > 0 && validateSubgroupAnswers(sg.answers, questionCount, optionCount);
    const isEmpty = sg.answers.length === 0;
    const indicator = isEmpty ? '' : (isValid ? '✓' : `✗ (${sg.answers.length}/${questionCount})`);
    const indicatorClass = isEmpty ? '' : (isValid ? 'sg-valid' : 'sg-invalid');
    const answerInput = cbMode
      ? renderSubgroupCheckboxes(sg, idx, questionCount, optionCount)
      : `<input class="sg-answers" type="text" placeholder="${validAnswerChars(optionCount).repeat(Math.ceil(questionCount / optionCount)).slice(0, questionCount)}" value="${sg.answers}" />`;
    return `
      <div class="sg-row-inline" data-sg-idx="${idx}">
        <input class="sg-name" type="text" placeholder="napr. A1" value="${sg.name}" maxlength="4" />
        <div class="sg-inline-answers">${answerInput}</div>
        <span class="sg-indicator ${indicatorClass}">${indicator}</span>
        <button type="button" class="btn danger sg-remove">Odstraniť</button>
      </div>
    `;
  }).join('');
}

function bindSubgroupEvents(questionCount, optionCount) {
  document.querySelectorAll('.sg-name').forEach((input, idx) => {
    input.addEventListener('input', () => {
      state.mdSubgroups[idx].name = input.value.toUpperCase();
      input.value = state.mdSubgroups[idx].name;
    });
  });

  if (state.mdCheckboxMode) {
    document.querySelectorAll('.sg-row-inline').forEach((row, idx) => {
      row.querySelectorAll('input[type="radio"]').forEach((radio) => {
        // Track "was already checked before click" for deselect support
        radio.addEventListener('mousedown', () => {
          radio.dataset.wasChecked = radio.checked ? '1' : '0';
        });
        radio.addEventListener('click', () => {
          const qMatch = radio.name.match(/q(\d+)$/);
          if (!qMatch) return;
          const q = parseInt(qMatch[1]);
          if (!state.mdSubgroups[idx].checkboxAnswers) {
            state.mdSubgroups[idx].checkboxAnswers = answersToCheckboxes(state.mdSubgroups[idx].answers, questionCount, optionCount);
          }
          if (radio.dataset.wasChecked === '1') {
            // Deselect: browser keeps it checked for radios, so manually uncheck
            radio.checked = false;
            state.mdSubgroups[idx].checkboxAnswers[q] = '';
          } else {
            // Browser already checked it
            state.mdSubgroups[idx].checkboxAnswers[q] = radio.value;
          }
          state.mdSubgroups[idx].answers = checkboxesToAnswers(state.mdSubgroups[idx].checkboxAnswers);
          const span = row.querySelector('.sg-indicator');
          if (span) {
            const val = state.mdSubgroups[idx].answers;
            const trimmed = val.replace(/ /g, '');
            if (!trimmed) { span.textContent = ''; span.className = 'sg-indicator'; return; }
            const ok = validateSubgroupAnswers(val, questionCount, optionCount);
            span.textContent = ok ? '✓' : `✗ (${val.length}/${questionCount})`;
            span.className = 'sg-indicator ' + (ok ? 'sg-valid' : 'sg-invalid');
          }
        });
      });
    });
  } else {
    document.querySelectorAll('.sg-answers').forEach((input, idx) => {
      input.addEventListener('input', () => {
        state.mdSubgroups[idx].answers = input.value.toUpperCase();
        input.value = state.mdSubgroups[idx].answers;
        state.mdSubgroups[idx].checkboxAnswers = answersToCheckboxes(state.mdSubgroups[idx].answers, questionCount, optionCount);
        const row = input.closest('.sg-row-inline');
        const span = row.querySelector('.sg-indicator');
        const val = state.mdSubgroups[idx].answers;
        if (!val) { span.textContent = ''; span.className = 'sg-indicator'; return; }
        const ok = validateSubgroupAnswers(val, questionCount, optionCount);
        span.textContent = ok ? '✓' : `✗ (${val.length}/${questionCount})`;
        span.className = 'sg-indicator ' + (ok ? 'sg-valid' : 'sg-invalid');
      });
    });
  }

  document.querySelectorAll('.sg-remove').forEach((btn, idx) => {
    btn.addEventListener('click', () => {
      state.mdSubgroups.splice(idx, 1);
      if (state.mdSubgroups.length === 0) state.mdSubgroups.push({ name: '', answers: '' });
      document.getElementById('sg-container').innerHTML = renderSubgroupRows(questionCount, optionCount);
      bindSubgroupEvents(questionCount, optionCount);
    });
  });
}


function renderMultiDay() {
  const content = document.getElementById('content');
  const qCount = state.mdQuestionCount;
  const oCount = state.mdOptionCount;

  content.innerHTML = `
    <form id="multiday-form" class="form">
      <div class="form-grid">
        <label>
          Nazov testu
          <input id="md-title" type="text" value="${state.mdTitle}" required />
        </label>
        <label>
          Skolsky rok (YYYY/YY)
          <input id="md-school-year" type="text" placeholder="2024/25" value="${state.mdSchoolYear}" required />
        </label>
        <label>
          Pocet otazok
          <input id="md-question-count" type="number" min="1" value="${qCount}" required />
        </label>
        <label>
          Pocet moznosti
          <input id="md-option-count" type="number" min="3" max="8" value="${oCount}" required />
        </label>
        <label class="toggle-field">
          Zobrazit meno
          <span class="toggle-row">
            <input id="md-show-name" type="checkbox" ${state.mdShowName ? 'checked' : ''} />
            <span class="md-toggle-text">${state.mdShowName ? 'Ano' : 'Nie'}</span>
          </span>
        </label>
        <label class="file">
          CSV zo systemu (s datumom, casom, miestnostou)
          <input id="md-csv-file" type="file" accept=".csv" />
          <span class="file-name">${state.mdCsvName || 'Ziadny subor'}</span>
        </label>
      </div>

      <div class="sg-section">
        <div class="sg-header">
          <span class="sg-title">Správne odpovede per podskupina</span>
          <span class="sg-hint">Povolené znaky: ${validAnswerChars(oCount)} &nbsp;|&nbsp; Dĺžka: ${qCount} znakov</span>
          <label class="toggle-field sg-mode-toggle">
            Checkboxy
            <span class="toggle-row">
              <input id="sg-checkbox-mode" type="checkbox" ${state.mdCheckboxMode ? 'checked' : ''} />
              <span class="md-toggle-text">${state.mdCheckboxMode ? 'Zap' : 'Vyp'}</span>
            </span>
          </label>
          <button type="button" id="sg-add" class="btn secondary">+ Pridať podskupinu</button>
        </div>
        ${state.mdCheckboxMode ? '' : `<div class="sg-labels"><span>Podskupina</span><span>Odpovede (${qCount} znakov)</span><span></span><span></span></div>`}
        <div id="sg-container">
          ${renderSubgroupRows(qCount, oCount)}
        </div>
      </div>

      <div class="form-actions">
        <button type="submit" class="primary">Vytvorit multi-terminovy test</button>
        <span id="md-status" class="status"></span>
      </div>
    </form>
  `;

  document.getElementById('md-title').addEventListener('input', (e) => { state.mdTitle = e.target.value; });
  document.getElementById('md-school-year').addEventListener('input', (e) => { state.mdSchoolYear = e.target.value; });
  document.getElementById('md-question-count').addEventListener('input', (e) => {
    state.mdQuestionCount = Number(e.target.value) || 0;
    renderMultiDay();
  });
  document.getElementById('md-option-count').addEventListener('input', (e) => {
    state.mdOptionCount = Number(e.target.value) || 0;
    renderMultiDay();
  });
  document.getElementById('md-show-name').addEventListener('change', (e) => {
    state.mdShowName = e.target.checked;
    const t = document.querySelector('.md-toggle-text');
    if (t) t.textContent = state.mdShowName ? 'Ano' : 'Nie';
  });

  document.getElementById('md-csv-file').addEventListener('change', async (e) => {
    const file = e.target.files?.[0];
    if (!file) { state.mdCsvContent = ''; state.mdCsvName = ''; return; }
    state.mdCsvName = file.name;
    state.mdCsvContent = await file.text();
    document.querySelector('#md-csv-file + .file-name').textContent = file.name;
  });

  document.getElementById('sg-checkbox-mode').addEventListener('change', (e) => {
    // Sync text↔checkboxes before switching mode
    if (!e.target.checked) {
      // switching TO text mode: checkboxAnswers → answers already up to date
    } else {
      // switching TO checkbox mode: ensure checkboxAnswers matches current answers
      state.mdSubgroups.forEach(sg => {
        sg.checkboxAnswers = answersToCheckboxes(sg.answers, qCount, oCount);
      });
    }
    state.mdCheckboxMode = e.target.checked;
    const t = e.target.nextElementSibling;
    if (t) t.textContent = state.mdCheckboxMode ? 'Zap' : 'Vyp';
    document.getElementById('sg-container').innerHTML = renderSubgroupRows(qCount, oCount);
    const labelsEl = document.querySelector('.sg-labels');
    if (state.mdCheckboxMode) {
      if (labelsEl) labelsEl.remove();
    } else {
      if (!labelsEl) {
        const container = document.getElementById('sg-container');
        const div = document.createElement('div');
        div.className = 'sg-labels';
        div.innerHTML = `<span>Podskupina</span><span>Odpovede (${qCount} znakov)</span><span></span><span></span>`;
        container.parentNode.insertBefore(div, container);
      }
    }
    bindSubgroupEvents(qCount, oCount);
  });

  document.getElementById('sg-add').addEventListener('click', () => {
    state.mdSubgroups.push({ name: '', answers: '' });
    document.getElementById('sg-container').innerHTML = renderSubgroupRows(qCount, oCount);
    bindSubgroupEvents(qCount, oCount);
  });

  bindSubgroupEvents(qCount, oCount);

  document.getElementById('multiday-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const statusEl = document.getElementById('md-status');

    const subgroupMap = {};
    for (const sg of state.mdSubgroups) {
      if (!sg.name.trim()) continue;
      // Allow empty answers (teacher may fill them in later)
      if (sg.answers.length > 0 && !validateSubgroupAnswers(sg.answers, qCount, oCount)) {
        statusEl.textContent = `Podskupina "${sg.name}": neplatné odpovede (${sg.answers.length}/${qCount} znakov, povolené: ${validAnswerChars(oCount)}).`;
        statusEl.className = 'status error';
        return;
      }
      subgroupMap[sg.name.trim()] = sg.answers.toUpperCase();
    }

    statusEl.textContent = 'Ukladam...';
    statusEl.className = 'status';

    try {
      await CreateMultiDayExamWithCSV(
        state.mdTitle.trim(),
        state.mdSchoolYear.trim(),
        Number(state.mdQuestionCount),
        Number(state.mdOptionCount),
        Boolean(state.mdShowName),
        state.mdCsvContent,
        subgroupMap,
      );
      statusEl.textContent = 'Ulozene.';
      statusEl.className = 'status success';
      state.mdTitle = '';
      state.mdSchoolYear = '';
      state.mdCsvContent = '';
      state.mdCsvName = '';
      state.mdQuestionCount = 10;
      state.mdOptionCount = 5;
      state.mdShowName = true;
      state.mdSubgroups = [{ name: '', answers: '' }];
      await refreshData();
      state.activeTab = 'exams';
      renderContent();
    } catch (err) {
      console.error(err);
      statusEl.textContent = 'Chyba: ' + (err?.message || err);
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
        try {
          const path = await PrintExamPDF(examId);
          if (path) {
            await OpenPath(path);
          }
        } catch (err) {
          console.error(err);
          window.alert('Tlač zlyhala. Skontroluj logs.');
        }
        return;
      }
      if (action === 'print-multiday') {
        try {
          const paths = await PrintMultiDayExamPDFs(examId);
          if (paths && paths.length > 0) {
            window.alert(`Vygenerovaných ${paths.length} PDF súborov v priečinku pdf_to_print.`);
            await OpenPath(paths[0]);
          }
        } catch (err) {
          console.error(err);
          window.alert('Tlač zlyhala. Skontroluj logs.');
        }
        return;
      }
      if (action === 'answers') {
        const exam = state.exams.find(e => e.id === examId);
        if (exam && exam.isMultiDay) {
          // Multi-day: podskupinový editor
          try {
            const answersRaw = await GetExamAnswers(examId);
            const qCount = exam.questionCount;
            const oCount = exam.optionCount || 5;
            let sgMap = {};
            try { sgMap = answersRaw ? JSON.parse(answersRaw) : {}; } catch (_) {}

            const renderSgModal = () => {
              // Local state for the modal
              const sgEntries = Object.entries(sgMap).map(([name, ans]) => ({
                name,
                answers: ans,
                cbAnswers: answersToCheckboxes(ans, qCount, oCount),
              }));
              if (sgEntries.length === 0) sgEntries.push({ name: '', answers: '', cbAnswers: Array(qCount).fill('') });

              let sgModalCbMode = true;

              const renderModalRows = () => sgEntries.map((sg, idx) => {
                const isValid = sg.answers.length > 0 && validateSubgroupAnswers(sg.answers, qCount, oCount);
                const isEmpty = sg.answers.length === 0;
                const ind = isEmpty ? '' : (isValid ? '✓' : `✗ ${sg.answers.replace(/ /g,'').length}/${qCount}`);
                const indClass = isEmpty ? '' : (isValid ? 'sg-valid' : 'sg-invalid');
                const choices = Array.from({ length: oCount }, (_, i) => String.fromCharCode(65 + i));
                const answerInput = sgModalCbMode
                  ? `<div class="sg-cb-grid">${Array.from({ length: qCount }, (_, q) => {
                      const cells = choices.map(ch => {
                        const checked = sg.cbAnswers[q] === ch ? 'checked' : '';
                        return `<label class="sg-cb-label"><input type="radio" name="modal-sg-${idx}-q${q}" value="${ch}" ${checked} />${ch}</label>`;
                      }).join('');
                      return `<div class="sg-cb-row"><span class="sg-cb-qnum">${q + 1}.</span>${cells}</div>`;
                    }).join('')}</div>`
                  : `<input class="sg-answers" type="text" placeholder="${validAnswerChars(oCount).repeat(Math.ceil(qCount / oCount)).slice(0, qCount)}" value="${sg.answers}" />`;
                return `
                  <div class="sg-row-inline" data-modal-idx="${idx}">
                    <input class="sg-name" type="text" placeholder="A1" value="${sg.name}" maxlength="4" />
                    <div class="sg-inline-answers">${answerInput}</div>
                    <span class="sg-indicator ${indClass}">${ind}</span>
                    <button type="button" class="btn danger sg-remove" data-modal-idx="${idx}">Odstraniť</button>
                  </div>`;
              }).join('');

              const updateIndicator = (row, val) => {
                const span = row.querySelector('.sg-indicator');
                if (!span) return;
                const trimmed = val.replace(/ /g, '');
                if (!trimmed) { span.textContent = ''; span.className = 'sg-indicator'; return; }
                const ok = validateSubgroupAnswers(val, qCount, oCount);
                span.textContent = ok ? '✓' : `✗ ${trimmed.length}/${qCount}`;
                span.className = 'sg-indicator ' + (ok ? 'sg-valid' : 'sg-invalid');
              };

              const bindModalEvents = () => {
                document.querySelectorAll('#modal-sg-container .sg-name').forEach((input, idx) => {
                  input.addEventListener('input', () => {
                    input.value = input.value.toUpperCase();
                    sgEntries[idx].name = input.value;
                  });
                });

                if (sgModalCbMode) {
                  document.querySelectorAll('#modal-sg-container .sg-row-inline').forEach((row, idx) => {
                    row.querySelectorAll('input[type="radio"]').forEach((radio) => {
                      radio.addEventListener('mousedown', () => {
                        radio.dataset.wasChecked = radio.checked ? '1' : '0';
                      });
                      radio.addEventListener('click', () => {
                        const qMatch = radio.name.match(/q(\d+)$/);
                        if (!qMatch) return;
                        const q = parseInt(qMatch[1]);
                        if (radio.dataset.wasChecked === '1') {
                          radio.checked = false;
                          sgEntries[idx].cbAnswers[q] = '';
                        } else {
                          sgEntries[idx].cbAnswers[q] = radio.value;
                        }
                        sgEntries[idx].answers = checkboxesToAnswers(sgEntries[idx].cbAnswers);
                        updateIndicator(row, sgEntries[idx].answers);
                      });
                    });
                  });
                } else {
                  document.querySelectorAll('#modal-sg-container .sg-answers').forEach((input, idx) => {
                    input.addEventListener('input', () => {
                      input.value = input.value.toUpperCase();
                      sgEntries[idx].answers = input.value;
                      sgEntries[idx].cbAnswers = answersToCheckboxes(input.value, qCount, oCount);
                      updateIndicator(input.closest('.sg-row-inline'), input.value);
                    });
                  });
                }

                document.querySelectorAll('#modal-sg-container .sg-remove').forEach((btn, idx) => {
                  btn.addEventListener('click', () => {
                    if (sgEntries.length <= 1) return;
                    sgEntries.splice(idx, 1);
                    document.getElementById('modal-sg-container').innerHTML = renderModalRows();
                    bindModalEvents();
                  });
                });
              };

              showModal('Odpovede podskupín', `
                <div class="answers-edit-form">
                  <div class="sg-modal-top-row">
                    <div class="sg-hint">Povolené: ${validAnswerChars(oCount)} &nbsp;|&nbsp; Dĺžka: ${qCount} znakov</div>
                    <label class="toggle-field sg-mode-toggle">
                      Checkboxy
                      <span class="toggle-row">
                        <input id="sg-modal-cb-mode" type="checkbox" ${sgModalCbMode ? 'checked' : ''} />
                        <span class="md-toggle-text">${sgModalCbMode ? 'Zap' : 'Vyp'}</span>
                      </span>
                    </label>
                  </div>
                  <div id="modal-sg-container">${renderModalRows()}</div>
                  <div class="form-actions" style="margin-top:12px">
                    <button id="modal-sg-add" class="btn secondary">+ Pridať podskupinu</button>
                    <button id="modal-sg-save" class="btn primary">Uložiť</button>
                    <span id="modal-sg-status" class="status"></span>
                  </div>
                </div>`);

              // Bind once-only listeners (on elements outside #modal-sg-container)
              document.getElementById('modal-sg-add').addEventListener('click', () => {
                sgEntries.push({ name: '', answers: '', cbAnswers: Array(qCount).fill('') });
                document.getElementById('modal-sg-container').innerHTML = renderModalRows();
                bindModalEvents();
              });

              document.getElementById('sg-modal-cb-mode').addEventListener('change', (e) => {
                if (e.target.checked) {
                  sgEntries.forEach(sg => { sg.cbAnswers = answersToCheckboxes(sg.answers, qCount, oCount); });
                }
                sgModalCbMode = e.target.checked;
                const t = e.target.nextElementSibling;
                if (t) t.textContent = sgModalCbMode ? 'Zap' : 'Vyp';
                document.getElementById('modal-sg-container').innerHTML = renderModalRows();
                bindModalEvents();
              });

              document.getElementById('modal-sg-save').addEventListener('click', async () => {
                const statusEl = document.getElementById('modal-sg-status');
                const newMap = {};
                let valid = true;
                sgEntries.forEach((sg) => {
                  const name = sg.name.trim();
                  const ans = sg.answers.trim().toUpperCase();
                  if (!name) return;
                  // Allow empty answers
                  if (ans.length > 0 && !validateSubgroupAnswers(ans, qCount, oCount)) {
                    statusEl.textContent = `Podskupina "${name}": neplatné odpovede.`;
                    statusEl.className = 'status error';
                    valid = false;
                  }
                  newMap[name] = ans;
                });
                if (!valid) return;
                try {
                  await UpdateMultiDayAnswers(examId, newMap);
                  hideModal();
                } catch (err) {
                  console.error(err);
                  if (statusEl) statusEl.textContent = 'Ukladanie zlyhalo.';
                }
              });

              bindModalEvents();
            };
            renderSgModal();
          } catch (err) {
            console.error(err);
            showModal('Odpovede testu', '<div class="error">Nacitavanie odpovedi zlyhalo.</div>');
          }
        } else {
          // Štandardný exam: radio button editor
          try {
            const questionCount = exam ? exam.questionCount : 0;
            const optionCount = exam ? (exam.optionCount || 5) : 5;
            const answers = await GetExamAnswers(examId);
            const currentAnswers = answers ? answers.split('') : [];
            const options = Array.from({length: optionCount}, (_, i) => String.fromCharCode(65 + i));
            let rows = '';
            for (let q = 0; q < questionCount; q++) {
              const cur = (currentAnswers[q] || '').toLowerCase();
              const opts = options.map(opt =>
                `<label class="radio-opt"><input type="radio" name="ans-q${q}" value="${opt.toLowerCase()}" ${cur === opt.toLowerCase() ? 'checked' : ''} />${opt}</label>`
              ).join('');
              rows += `<div class="answer-edit-row"><span class="answer-num">${q + 1}.</span><div class="radio-opts">${opts}</div></div>`;
            }
            showModal('Odpovede testu', `
              <div class="answers-edit-form">
                <div class="answers-edit-grid">${rows}</div>
                <div class="modal-actions">
                  <button id="save-answers-btn" class="btn primary">Uložiť odpovede</button>
                </div>
              </div>`);
            document.getElementById('save-answers-btn').addEventListener('click', async () => {
              const newAnswers = [];
              for (let q = 0; q < questionCount; q++) {
                const checked = document.querySelector(`input[name="ans-q${q}"]:checked`);
                newAnswers.push(checked ? checked.value : '');
              }
              try {
                await UpdateExamAnswers(examId, newAnswers);
                hideModal();
              } catch (err) {
                console.error(err);
                window.alert('Ukladanie odpovedi zlyhalo.');
              }
            });
          } catch (err) {
            console.error(err);
            showModal('Odpovede testu', '<div class="error">Nacitavanie odpovedi zlyhalo.</div>');
          }
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
          if (path) await OpenPath(path);
        } catch (err) {
          console.error(err);
          showModal('Export CSV', '<div class="error">Export studentov do CSV zlyhal.</div>');
        }
        return;
      }
      if (action === 'csv-multiday') {
        try {
          const path = await ExportMultiDayResultsCSV(examId);
          if (path) await OpenPath(path);
        } catch (err) {
          console.error(err);
          showModal('Export CSV', '<div class="error">Export vysledkov zlyhal.</div>');
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

  const legendBtn = document.getElementById('btn-print-legend');
  if (legendBtn) {
    legendBtn.addEventListener('click', async () => {
      try {
        const path = await PrintLegend();
        if (path) await OpenPath(path);
      } catch (err) {
        console.error(err);
        window.alert('Tlač legendy zlyhala. Skontroluj logs.');
      }
    });
  }
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

init();
