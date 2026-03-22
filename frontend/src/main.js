import './style.css';
import './app.css';

import {
  CreateExamWithCSV,
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
  GetSavePath,
  SetSavePath,
  GetExamStats,
  GenerateExamStatisticsPDF,
  ExportExamStudentsCSV,
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
            <button class="btn" data-action="answers" data-exam-id="${exam.id}">Odpovede</button>
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

function renderCreateExam() {
  const content = document.getElementById('content');
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
          Datum a cas (dd.MM.yyyy HH:mm)
          <input id="date-time" type="text" placeholder="01.03.2025 10:00" value="${state.formDateTime}" required />
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
          <span class="file-name">${state.csvName || 'Ziadny subor'}</span>
        </label>
      </div>
      <div class="form-actions">
        <button type="button" id="generate" class="secondary">Generovat otazky</button>
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
      if (action === 'answers') {
        try {
          const answers = await GetExamAnswers(examId);
          if (answers) {
            const items = answers
              .split('')
              .map(
                (ans, idx) =>
                  `<div class="answer-item"><span class="answer-num">${idx + 1}.</span><span class="answer-val">${ans.toUpperCase()}</span></div>`,
              )
              .join('');
            showModal('Odpovede testu', `<div class="answers-grid">${items}</div>`);
          } else {
            showModal('Odpovede testu', '<div class="empty">Odpovede neboli najdene.</div>');
          }
        } catch (err) {
          console.error(err);
          showModal('Odpovede testu', '<div class="error">Nacitavanie odpovedi zlyhalo.</div>');
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

init();
