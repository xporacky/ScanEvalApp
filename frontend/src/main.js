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
  ExportExamTemplateCSV,
  GetSavePath,
  SetSavePath,
  GetExamStats,
  GenerateExamStatisticsPDF,
  ExportExamStudentsCSV,
  ParseExamTemplateCSV,
  UpdateExamAnswers,
  PrintLegend,
  CreateMultiVersionExam,
  GetExamVersions,
  UpdateVersionAnswers,
  DeleteExamVersion,
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
  multiTitle: '',
  multiSchoolYear: '',
  multiQuestionCount: 10,
  multiOptionCount: 5,
  multiShowName: true,
  multiCsvName: '',
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
        <button class="tab" data-tab="multicreate">Vytvoriť viacverziovú písomku</button>
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
          <span>Verzie</span>
          <span>Vyhodnotiť</span>
          <span>Štatistika PDF</span>
          <span>CSV</span>
          <span>Zmazať</span>
        </div>
        ${state.exams.map(exam => `
          <div class="row">
            <span class="cell title">${exam.title} ${exam.isMultiDay ? '<span class="badge">multiverzia</span>' : ''}</span>
            <span class="cell">${exam.schoolYear}</span>
            <span class="cell">${formatDate(exam.date)}</span>
            <span class="cell">${exam.questionCount}</span>
            <span class="cell">${exam.optionCount || '-'}</span>
            <span class="cell">${exam.studentCount}</span>
            <span class="cell"><button class="btn" data-action="print" data-exam-id="${exam.id}">Tlačiť</button></span>
            <span class="cell"><button class="btn" data-action="answers" data-exam-id="${exam.id}">Odpovede</button></span>
            <span class="cell">${exam.isMultiDay ? `<button class="btn" data-action="versions" data-exam-id="${exam.id}">Verzie</button>` : ''}</span>
            <span class="cell"><button class="btn" data-action="evaluate" data-exam-id="${exam.id}">Vyhodnotiť</button></span>
            <span class="cell"><button class="btn" data-action="stats-pdf" data-exam-id="${exam.id}">Štatistika</button></span>
            <span class="cell"><button class="btn" data-action="csv" data-exam-id="${exam.id}">CSV</button></span>
            <span class="cell"><button class="btn danger" data-action="delete" data-exam-id="${exam.id}">Zmazať</button></span>
          </div>
        `).join('')}
      </div>
    </div>
  `;
  bindExamActions();
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
        <span>Miestnosť</span>
        <span>Reg. číslo</span>
        <span>Písomka</span>
        <span>Skóre</span>
        <span>Akcie</span>
      </div>
      ${state.students.map(student => `
        <div class="row">
          <span class="cell title">${student.name}</span>
          <span class="cell">${student.surname}</span>
          <span class="cell">${formatDate(student.birthDate)}</span>
          <span class="cell">${student.room}</span>
          <span class="cell">${student.registrationNumber}</span>
          <span class="cell">${student.examId}</span>
          <span class="cell">${student.score}</span>
          <span class="cell action">
            <button class="btn" data-student-print="${student.id}">Nový hárok</button>
            ${student.pages ? `<button class="btn secondary" data-student-download="${student.id}">Vypracovaný hárok</button>` : ''}
          </span>
        </div>
      `).join('')}
    </div>
  `;

  document.querySelectorAll('[data-student-print]').forEach(btn => {
    btn.addEventListener('click', async () => {
      const id = Number(btn.dataset.studentPrint);
      if (!id) return;
      try {
        const path = await PrintStudentSheet(id);
        if (path) await OpenPath(path);
      } catch (err) {
        console.error(err);
        alert('Tlač hárku zlyhala.');
      }
    });
  });

  document.querySelectorAll('[data-student-download]').forEach(btn => {
    btn.addEventListener('click', async () => {
      const id = Number(btn.dataset.studentDownload);
      if (!id) return;
      try {
        const path = await DownloadStudentSheet(id);
        if (path) await OpenPath(path);
      } catch (err) {
        console.error(err);
        alert('Stiahnutie hárku zlyhalo.');
      }
    });
  });
}

function renderAnswerSelectors(container) {
  const options = ['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H'];
  const optionCount = Number(document.getElementById('option-count').value) || 5;
  const questionCount = Number(document.getElementById('question-count').value) || 0;
  const activeOptions = options.slice(0, Math.min(optionCount, options.length));
  state.answers = Array.from({ length: questionCount }, (_, i) => state.answers[i] || '');
  container.innerHTML = state.answers.map((_, idx) => `
    <div class="answer-row">
      <span>Otázka ${String(idx+1).padStart(2,'0')}</span>
      ${activeOptions.map(opt => `
        <label class="radio">
          <input type="radio" name="q${idx}" value="${opt}" ${state.answers[idx] === opt ? 'checked' : ''}>
          <span>${opt}</span>
        </label>
      `).join('')}
    </div>
  `).join('');
  container.querySelectorAll('input[type="radio"]').forEach(inp => {
    inp.addEventListener('change', () => {
      const q = Number(inp.name.replace('q',''));
      state.answers[q] = inp.value;
    });
  });
}

function syncAnswersState() {
  const qCount = Math.max(Number(state.formQuestionCount) || 0, 0);
  state.answers = Array.from({ length: qCount }, (_, i) => state.answers[i] || '');
}

function getStudentCsvLabel() {
  if (state.csvName) return state.csvName;
  if (state.csvContent) return 'Načítané zo šablóny';
  return 'Žiadny súbor';
}

function getTemplateCsvLabel() {
  return state.templateCsvName || 'Žiadny súbor';
}

function renderCreateExam() {
  const content = document.getElementById('content');
  syncAnswersState();
  content.innerHTML = `
    <form id="create-exam" class="form">
      <div class="form-grid">
        <label>Názov testu <input id="title" type="text" value="${state.formTitle}" required /></label>
        <label>Školský rok (YYYY/YY) <input id="school-year" type="text" placeholder="2024/25" value="${state.formSchoolYear}" required /></label>
        <label>Dátum (dd.MM.yyyy) <input id="date-time" type="text" placeholder="01.03.2025" value="${state.formDateTime}" required /></label>
        <label>Počet otázok <input id="question-count" type="number" min="1" value="${state.formQuestionCount}" required /></label>
        <label>Počet možností <input id="option-count" type="number" min="3" max="8" value="${state.formOptionCount}" required /></label>
        <label class="toggle-field">Zobraziť meno <input id="show-name" type="checkbox" ${state.formShowName ? 'checked' : ''} /></label>
        <label class="file">CSV so študentmi <input id="csv-file" type="file" accept=".csv" /><span class="file-name">${getStudentCsvLabel()}</span></label>
        <label class="file">CSV šablóna testu <input id="template-csv-file" type="file" accept=".csv" /><span class="file-name">${getTemplateCsvLabel()}</span></label>
      </div>
      <div class="form-actions">
        <button type="button" id="generate" class="secondary">Generovať otázky</button>
        <button type="button" id="save-template" class="secondary">Uložiť CSV šablónu</button>
        <button type="submit" class="primary">Vytvoriť test</button>
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
  const titleInp = document.getElementById('title');
  const yearInp = document.getElementById('school-year');
  const dateInp = document.getElementById('date-time');
  const qCountInp = document.getElementById('question-count');
  const optCountInp = document.getElementById('option-count');
  const showNameChk = document.getElementById('show-name');

  titleInp.addEventListener('input', () => state.formTitle = titleInp.value);
  yearInp.addEventListener('input', () => state.formSchoolYear = yearInp.value);
  dateInp.addEventListener('input', () => state.formDateTime = dateInp.value);
  qCountInp.addEventListener('input', () => { state.formQuestionCount = Number(qCountInp.value) || 0; syncAnswersState(); });
  optCountInp.addEventListener('input', () => state.formOptionCount = Number(optCountInp.value) || 0);
  showNameChk.addEventListener('change', () => state.formShowName = showNameChk.checked);

  document.getElementById('generate').addEventListener('click', () => renderAnswerSelectors(answersEl));
  fileInput.addEventListener('change', async (e) => {
    const file = e.target.files[0];
    if (!file) { state.csvContent = ''; state.csvName = ''; renderCreateExam(); return; }
    state.csvName = file.name;
    state.csvContent = await file.text();
    renderCreateExam();
  });
  templateFileInput.addEventListener('change', async (e) => {
    const file = e.target.files[0];
    if (!file) { state.templateCsvName = ''; renderCreateExam(); return; }
    statusEl.textContent = 'Načítavam šablónu...';
    try {
      const text = await file.text();
      const tmpl = await ParseExamTemplateCSV(text);
      state.templateCsvName = file.name;
      state.formTitle = tmpl.title || '';
      state.formSchoolYear = tmpl.schoolYear || '';
      state.formDateTime = tmpl.dateTime || '';
      state.formQuestionCount = tmpl.questionCount || 0;
      state.formOptionCount = tmpl.optionCount || 5;
      state.formShowName = !!tmpl.showName;
      state.answers = tmpl.answers || [];
      state.csvContent = tmpl.studentCSVContent || '';
      state.csvName = tmpl.studentCSVContent ? 'Načítané zo šablóny' : '';
      renderCreateExam();
      statusEl.textContent = 'Šablóna načítaná.';
      statusEl.className = 'status success';
    } catch (err) {
      console.error(err);
      statusEl.textContent = 'Chyba pri načítaní šablóny.';
      statusEl.className = 'status error';
    }
  });
  document.getElementById('save-template').addEventListener('click', async () => {
    statusEl.textContent = 'Ukladám šablónu...';
    syncAnswersState();
    try {
      const path = await ExportExamTemplateCSV(
        state.formTitle.trim(), state.formSchoolYear.trim(), state.formDateTime.trim(),
        Number(state.formQuestionCount), Number(state.formOptionCount),
        state.answers, state.csvContent, state.formShowName
      );
      statusEl.textContent = path ? `Šablóna uložená: ${path}` : 'Šablóna uložená.';
      statusEl.className = 'status success';
    } catch (err) {
      statusEl.textContent = 'Chyba pri ukladaní šablóny.';
      statusEl.className = 'status error';
    }
  });
  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    statusEl.textContent = 'Ukladám...';
    try {
      await CreateExamWithCSV(
        state.formTitle.trim(), state.formSchoolYear.trim(), state.formDateTime.trim(),
        Number(state.formQuestionCount), Number(state.formOptionCount),
        state.answers, state.csvContent, state.formShowName
      );
      statusEl.textContent = 'Uložené.';
      state.csvContent = ''; state.csvName = ''; state.templateCsvName = '';
      state.answers = []; state.formTitle = ''; state.formSchoolYear = ''; state.formDateTime = '';
      state.formQuestionCount = 10; state.formOptionCount = 5; state.formShowName = true;
      await refreshData();
      state.activeTab = 'exams';
      renderContent();
    } catch (err) {
      console.error(err);
      statusEl.textContent = 'Chyba pri ukladaní.';
      statusEl.className = 'status error';
    }
  });
  renderAnswerSelectors(answersEl);
}

function renderMultiCreateExam() {
  const content = document.getElementById('content');
  content.innerHTML = `
    <form id="create-multi-exam" class="form">
      <div class="form-grid">
        <label>Názov testu <input id="multi-title" type="text" value="${state.multiTitle}" required /></label>
        <label>Školský rok (YYYY/YY) <input id="multi-school-year" type="text" value="${state.multiSchoolYear}" required /></label>
        <label>Počet otázok <input id="multi-question-count" type="number" min="1" value="${state.multiQuestionCount}" required /></label>
        <label>Počet možností (3‑8) <input id="multi-option-count" type="number" min="3" max="8" value="${state.multiOptionCount}" required /></label>
        <label class="toggle-field">Zobraziť meno <input id="multi-show-name" type="checkbox" ${state.multiShowName ? 'checked' : ''} /></label>
        <label class="file">CSV so študentmi <input id="multi-csv-file" type="file" accept=".csv" /><span class="file-name">${state.multiCsvName || 'Žiadny súbor'}</span></label>
      </div>
      <div class="form-actions">
        <button type="submit" class="primary">Vytvoriť test</button>
        <span id="multi-status" class="status"></span>
      </div>
    </form>
  `;

  const form = document.getElementById('create-multi-exam');
  const titleInp = document.getElementById('multi-title');
  const yearInp = document.getElementById('multi-school-year');
  const qCountInp = document.getElementById('multi-question-count');
  const optCountInp = document.getElementById('multi-option-count');
  const showNameChk = document.getElementById('multi-show-name');
  const csvFile = document.getElementById('multi-csv-file');
  const fileNameSpan = document.querySelector('#create-multi-exam .file-name');

  titleInp.value = state.multiTitle;
  yearInp.value = state.multiSchoolYear;
  qCountInp.value = state.multiQuestionCount;
  optCountInp.value = state.multiOptionCount;
  showNameChk.checked = state.multiShowName;

  titleInp.addEventListener('input', () => state.multiTitle = titleInp.value);
  yearInp.addEventListener('input', () => state.multiSchoolYear = yearInp.value);
  qCountInp.addEventListener('input', () => state.multiQuestionCount = parseInt(qCountInp.value) || 0);
  optCountInp.addEventListener('input', () => state.multiOptionCount = parseInt(optCountInp.value) || 5);
  showNameChk.addEventListener('change', () => state.multiShowName = showNameChk.checked);

  let csvContent = '';
  let csvName = '';
  csvFile.addEventListener('change', async (e) => {
    const file = e.target.files[0];
    if (!file) {
      csvContent = '';
      csvName = '';
      fileNameSpan.textContent = 'Žiadny súbor';
      return;
    }
    csvName = file.name;
    fileNameSpan.textContent = csvName;
    csvContent = await file.text();
  });

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    const statusEl = document.getElementById('multi-status');
    statusEl.textContent = 'Ukladám...';
    try {
      await CreateMultiVersionExam(
        state.multiTitle.trim(),
        state.multiSchoolYear.trim(),
        state.multiQuestionCount,
        state.multiOptionCount,
        state.multiShowName,
        csvContent
      );
      statusEl.textContent = 'Test vytvorený.';
      state.multiTitle = '';
      state.multiSchoolYear = '';
      state.multiQuestionCount = 10;
      state.multiOptionCount = 5;
      state.multiShowName = true;
      csvContent = '';
      csvName = '';
      await refreshData();
      state.activeTab = 'exams';
      renderContent();
    } catch (err) {
      console.error(err);
      statusEl.textContent = 'Chyba pri vytváraní.';
    }
  });
}

async function showVersionsModal(examId, questionCount, optionCount) {
  const versions = await GetExamVersions(examId);
  const groups = ['A','B','C','D','E','F','G','H'];
  const numbers = [1,2,3,4,5];
  let gridHtml = '<div class="version-grid">';
  for (let g of groups) {
    gridHtml += `<div class="version-row"><strong>${g}</strong>`;
    for (let n of numbers) {
      const code = `${g}${n}`;
      const hasAnswers = versions[code] && versions[code].answers && versions[code].answers !== '';
      const cls = hasAnswers ? 'version-cell filled' : 'version-cell';
      gridHtml += `<button class="${cls}" data-version="${code}">${code}</button>`;
    }
    gridHtml += '</div>';
  }
  gridHtml += '</div><div class="modal-actions"><button id="close-versions" class="btn">Zavrieť</button></div>';
  showModal('Verzie testu', gridHtml);
 
  const modalClosed = new Promise((resolve) => {
    const closeBtn = document.getElementById('close-versions');
    const closeHandler = () => {
      hideModal();
      resolve();
    };
    closeBtn?.addEventListener('click', closeHandler, { once: true });
    const modalBg = document.getElementById('modal');
    const bgHandler = () => {
      if (modalBg.classList.contains('hidden')) {
        modalBg.removeEventListener('click', bgHandler);
        resolve();
      }
    };
    modalBg.addEventListener('click', bgHandler);
  });
 
  document.querySelectorAll('.version-cell').forEach(btn => {
    btn.addEventListener('click', async (e) => {
      e.stopPropagation();
      const version = btn.dataset.version;
      await showVersionEditorModal(examId, version, questionCount, optionCount);
      await showVersionsModal(examId, questionCount, optionCount);
    });
  });
 
  await modalClosed;
}

async function showVersionEditorModal(examId, versionCode, questionCount, optionCount) {
  const versions = await GetExamVersions(examId);
  const existing = versions[versionCode];
  const currentAnswers = existing && existing.answers ? existing.answers.split('') : new Array(questionCount).fill('');
 
  const currentDate = existing && existing.dateTime ? new Date(existing.dateTime) : new Date();
  const day = String(currentDate.getDate()).padStart(2, '0');
  const month = String(currentDate.getMonth() + 1).padStart(2, '0');
  const year = currentDate.getFullYear();
  const hours = String(currentDate.getHours()).padStart(2, '0');
  const minutes = String(currentDate.getMinutes()).padStart(2, '0');
  const dateTimeStr = `${day}.${month}.${year} ${hours}:${minutes}`;
 
  const letters = 'ABCDEFGH'.slice(0, optionCount);
  let answersHtml = '';
  for (let q = 0; q < questionCount; q++) {
    answersHtml += `<div class="answer-edit-row"><span>${q+1}.</span>`;
    for (let l of letters) {
      const checked = currentAnswers[q] === l ? 'checked' : '';
      answersHtml += `<label><input type="radio" name="q${q}" value="${l}" ${checked}> ${l}</label>`;
    }
    answersHtml += '</div>';
  }
  const modalContent = `
    <div class="version-editor">
      <h3>Verzia ${versionCode}</h3>
      <label>Dátum a čas (dd.MM.yyyy HH:mm) <input id="version-datetime" value="${dateTimeStr}" /></label>
      <div class="answers-edit-grid">${answersHtml}</div>
      <div class="modal-actions">
        <button id="save-version" class="btn primary">Uložiť</button>
        <button id="delete-version" class="btn danger">Zmazať verziu</button>
        <button id="cancel-version" class="btn">Zrušiť</button>
      </div>
    </div>
  `;
  showModal(`Verzia ${versionCode}`, modalContent);
 
  return new Promise((resolve) => {
    const saveBtn = document.getElementById('save-version');
    const deleteBtn = document.getElementById('delete-version');
    const cancelBtn = document.getElementById('cancel-version');
    const dtInput = document.getElementById('version-datetime');
    const modalCloseBtn = document.getElementById('modal-close');
   
    const cleanup = () => {
      saveBtn?.removeEventListener('click', saveHandler);
      deleteBtn?.removeEventListener('click', deleteHandler);
      cancelBtn?.removeEventListener('click', cancelHandler);
      modalCloseBtn?.removeEventListener('click', closeHandler);
    };
   
    const saveHandler = async () => {
      const newAnswers = [];
      for (let q = 0; q < questionCount; q++) {
        const sel = document.querySelector(`input[name="q${q}"]:checked`);
        newAnswers.push(sel ? sel.value : '');
      }
      const answersStr = newAnswers.join('');
      const dateTimeStr = dtInput.value;
     
      const dateTimeRegex = /^\d{2}\.\d{2}\.\d{4} \d{2}:\d{2}$/;
      if (!dateTimeRegex.test(dateTimeStr)) {
        alert('Dátum a čas musia byť vo formáte dd.MM.yyyy HH:mm (napr. 20.04.2026 14:30)');
        return;
      }
     
      try {
        await UpdateVersionAnswers(examId, versionCode, answersStr, dateTimeStr);
        hideModal();
        cleanup();
        resolve();
      } catch (err) {
        console.error('Save failed:', err);
        alert('Chyba pri ukladaní: ' + (err.message || err));
      }
    };
   
    const deleteHandler = async () => {
      if (confirm(`Naozaj zmazať verziu ${versionCode}?`)) {
        try {
          await DeleteExamVersion(examId, versionCode);
          hideModal();
          cleanup();
          resolve();
        } catch (err) {
          console.error('Delete failed:', err);
          alert('Chyba pri mazaní verzie.');
        }
      }
    };
   
    const cancelHandler = () => {
      hideModal();
      cleanup();
      resolve();
    };
   
    const closeHandler = () => {
      hideModal();
      cleanup();
      resolve();
    };
   
    saveBtn?.addEventListener('click', saveHandler);
    deleteBtn?.addEventListener('click', deleteHandler);
    cancelBtn?.addEventListener('click', cancelHandler);
    modalCloseBtn?.addEventListener('click', closeHandler);
  });
}

function renderUpload() {
  const content = document.getElementById('content');
  const examOptions = state.exams.map(exam => `<option value="${exam.id}" ${Number(exam.id) === state.uploadExamId ? 'selected' : ''}>${exam.title}</option>`).join('');
  const configOptions = state.configs.map(cfg => `<option value="${cfg}" ${cfg === state.uploadConfig ? 'selected' : ''}>${cfg}</option>`).join('');
  content.innerHTML = `
    <div class="eval-form">
      <label>Test <select id="upload-exam">${examOptions}</select></label>
      <label>Konfigurácia skenera <select id="upload-config">${configOptions}</select></label>
      <label>PDF sken <div class="file-row"><span class="file-path">${state.uploadFilePath || 'Žiadny súbor'}</span><button id="upload-pick" class="btn">Vybrať PDF</button></div></label>
      <div class="form-actions"><button id="upload-start" class="btn primary">Spustiť vyhodnotenie</button><span id="upload-status" class="status"></span></div>
    </div>
  `;
  const examSel = document.getElementById('upload-exam');
  const configSel = document.getElementById('upload-config');
  const pickBtn = document.getElementById('upload-pick');
  const startBtn = document.getElementById('upload-start');
  examSel?.addEventListener('change', () => state.uploadExamId = Number(examSel.value));
  configSel?.addEventListener('change', () => state.uploadConfig = configSel.value);
  pickBtn?.addEventListener('click', async () => {
    const path = await PickPDF();
    if (path) { state.uploadFilePath = path; renderUpload(); }
  });
  startBtn?.addEventListener('click', async () => {
    const statusEl = document.getElementById('upload-status');
    if (!state.uploadExamId || !state.uploadConfig || !state.uploadFilePath) {
      if (statusEl) statusEl.textContent = 'Vyber test, konfiguráciu a PDF.';
      return;
    }
    statusEl.textContent = 'Spúšťam...';
    try {
      await EvaluateExam(state.uploadExamId, state.uploadFilePath, state.uploadConfig);
    } catch (err) {
      console.error(err);
      statusEl.textContent = 'Spustenie zlyhalo.';
    }
  });
}

function renderSettings() {
  const content = document.getElementById('content');
  content.innerHTML = `
    <div class="settings-form">
      <label>Miesto ukladania PDF <div class="file-row"><span class="file-path">${state.savePath || 'Nenastavené'}</span><button id="settings-pick" class="btn">Vybrať priečinok</button></div></label>
      <div class="form-actions"><button id="settings-save" class="btn primary">Uložiť</button><span id="settings-status" class="status"></span></div>
    </div>
  `;
  const pickBtn = document.getElementById('settings-pick');
  const saveBtn = document.getElementById('settings-save');
  pickBtn?.addEventListener('click', async () => {
    const path = await PickFolder();
    if (path) { state.savePath = path; renderSettings(); }
  });
  saveBtn?.addEventListener('click', async () => {
    const statusEl = document.getElementById('settings-status');
    if (!state.savePath) { statusEl.textContent = 'Vyber priečinok.'; return; }
    statusEl.textContent = 'Ukladám...';
    try {
      await SetSavePath(state.savePath);
      statusEl.textContent = 'Uložené.';
    } catch (err) {
      statusEl.textContent = 'Uloženie zlyhalo.';
    }
  });
}

function renderContent() {
  document.querySelectorAll('.tab').forEach(btn => btn.classList.toggle('active', btn.dataset.tab === state.activeTab));
  if (state.activeTab === 'exams') { renderExams(); return; }
  if (state.activeTab === 'create') { renderCreateExam(); return; }
  if (state.activeTab === 'multicreate') { renderMultiCreateExam(); return; }
  if (state.activeTab === 'students') { renderStudents(); return; }
  if (state.activeTab === 'upload') { renderUpload(); return; }
  if (state.activeTab === 'settings') { renderSettings(); }
}

function bindExamActions() {
  document.querySelectorAll('[data-exam-id]').forEach(btn => {
    btn.addEventListener('click', async () => {
      const examId = Number(btn.dataset.examId);
      const action = btn.dataset.action;
      if (action === 'print') {
        try {
          const path = await PrintExamPDF(examId);
          if (path) await OpenPath(path);
        } catch (err) { alert('Tlač zlyhala.'); }
        return;
      }
      if (action === 'answers') {
        try {
          const exam = state.exams.find(e => e.id === examId);
          const qCount = exam.questionCount, optCount = exam.optionCount || 5;
          const answers = await GetExamAnswers(examId);
          const current = answers ? answers.split('') : [];
          const letters = Array.from({length: optCount}, (_,i) => String.fromCharCode(65+i));
          let rows = '';
          for (let q=0; q<qCount; q++) {
            const cur = (current[q] || '').toLowerCase();
            const opts = letters.map(l => `<label class="radio-opt"><input type="radio" name="ans-q${q}" value="${l.toLowerCase()}" ${cur===l.toLowerCase() ? 'checked' : ''}>${l}</label>`).join('');
            rows += `<div class="answer-edit-row"><span class="answer-num">${q+1}.</span><div class="radio-opts">${opts}</div></div>`;
          }
          showModal('Odpovede testu', `<div class="answers-edit-form"><div class="answers-edit-grid">${rows}</div><div class="modal-actions"><button id="save-answers-btn" class="btn primary">Uložiť odpovede</button></div></div>`);
          document.getElementById('save-answers-btn').addEventListener('click', async () => {
            const newAns = [];
            for (let q=0; q<qCount; q++) {
              const checked = document.querySelector(`input[name="ans-q${q}"]:checked`);
              newAns.push(checked ? checked.value : '');
            }
            try {
              await UpdateExamAnswers(examId, newAns);
              hideModal();
            } catch (err) { alert('Ukladanie zlyhalo.'); }
          });
        } catch (err) { showModal('Odpovede testu', '<div class="error">Načítavanie zlyhalo.</div>'); }
        return;
      }
      if (action === 'versions') {
        const exam = state.exams.find(e => e.id === examId);
        if (exam) await showVersionsModal(examId, exam.questionCount, exam.optionCount);
        return;
      }
      if (action === 'evaluate') {
        state.evalExamId = examId;
        state.evalFilePath = '';
        if (!state.evalConfig && state.configs.length) state.evalConfig = state.configs[0];
        renderEvalModal();
        return;
      }
      if (action === 'stats-pdf') {
        state.statsExamId = examId;
        if (!Object.keys(state.statsSelected).length) statisticsOptions.forEach(opt => state.statsSelected[opt] = true);
        renderStatsModal();
        return;
      }
      if (action === 'csv') {
        try {
          const path = await ExportExamStudentsCSV(examId);
          if (path) await OpenPath(path);
        } catch (err) { showModal('Export CSV', '<div class="error">Export zlyhal.</div>'); }
        return;
      }
      if (action === 'delete' && confirm('Naozaj chcete zmazať tento test?')) {
        try {
          await DeleteExam(examId);
          await refreshData();
          renderContent();
        } catch (err) { alert('Zmazanie zlyhalo.'); }
      }
    });
  });
  const legendBtn = document.getElementById('btn-print-legend');
  if (legendBtn) legendBtn.addEventListener('click', async () => {
    try { const path = await PrintLegend(); if (path) await OpenPath(path); } catch (err) { alert('Tlač legendy zlyhala.'); }
  });
}

function renderEvalModal() {
  const options = state.configs.map(cfg => `<option value="${cfg}" ${cfg === state.evalConfig ? 'selected' : ''}>${cfg}</option>`).join('');
  showModal('Vyhodnotenie písomiek', `
    <div class="eval-form">
      <label>Konfigurácia skenera <select id="eval-config">${options}</select></label>
      <label>PDF sken <div class="file-row"><span class="file-path">${state.evalFilePath || 'Žiadny súbor'}</span><button id="eval-pick" class="btn">Vybrať PDF</button></div></label>
      <div class="form-actions"><button id="eval-start" class="btn primary">Spustiť vyhodnotenie</button><span id="eval-status" class="status"></span></div>
    </div>
  `);
  const configSel = document.getElementById('eval-config');
  const pickBtn = document.getElementById('eval-pick');
  const startBtn = document.getElementById('eval-start');
  configSel?.addEventListener('change', () => state.evalConfig = configSel.value);
  pickBtn?.addEventListener('click', async () => { const path = await PickPDF(); if (path) { state.evalFilePath = path; renderEvalModal(); } });
  startBtn?.addEventListener('click', async () => {
    const statusEl = document.getElementById('eval-status');
    if (!state.evalFilePath || !state.evalConfig) { statusEl.textContent = 'Vyber PDF a konfiguráciu.'; return; }
    statusEl.textContent = 'Spúšťam...';
    try { await EvaluateExam(state.evalExamId, state.evalFilePath, state.evalConfig); } catch (err) { statusEl.textContent = 'Spustenie zlyhalo.'; }
  });
}

function renderStatsModal() {
  const items = statisticsOptions.map(opt => `<label class="stats-item"><input type="checkbox" data-stat="${opt}" ${state.statsSelected[opt] ? 'checked' : ''}><span>${opt}</span></label>`).join('');
  showModal('Vyberte požadované štatistiky', `
    <div class="stats-form"><div class="stats-list">${items}</div><div class="form-actions"><button id="stats-generate" class="btn primary">Generovať štatistiky</button><span id="stats-status" class="status"></span></div></div>
  `);
  document.querySelectorAll('[data-stat]').forEach(inp => inp.addEventListener('change', () => state.statsSelected[inp.dataset.stat] = inp.checked));
  const genBtn = document.getElementById('stats-generate');
  if (genBtn) genBtn.addEventListener('click', async () => {
    const statusEl = document.getElementById('stats-status');
    const selected = statisticsOptions.filter(opt => state.statsSelected[opt]);
    if (!selected.length) { statusEl.textContent = 'Vyber aspoň jednu štatistiku.'; return; }
    statusEl.textContent = 'Generujem...';
    try {
      const path = await GenerateExamStatisticsPDF(state.statsExamId, selected);
      if (path) { await OpenPath(path); statusEl.textContent = 'Hotovo.'; } else statusEl.textContent = 'PDF sa nepodarilo vytvoriť.';
    } catch (err) { statusEl.textContent = 'Generovanie zlyhalo.'; }
  });
}

async function refreshData() {
  try {
    const [exams, students, configs, savePath] = await Promise.all([ListExams(), ListStudents(), ListConfigs(), GetSavePath()]);
    state.exams = exams || [];
    state.students = students || [];
    state.configs = configs || [];
    if (!state.evalConfig && state.configs.length) state.evalConfig = state.configs[0];
    state.savePath = savePath || '';
    if (!state.uploadConfig && state.configs.length) state.uploadConfig = state.configs[0];
    if (!state.uploadExamId && state.exams.length) state.uploadExamId = state.exams[0].id;
  } catch (err) {
    console.error(err);
    state.exams = []; state.students = []; state.configs = []; state.savePath = '';
    state.uploadConfig = ''; state.uploadExamId = 0;
  }
}

async function init() {
  renderShell();
  await refreshData();
  renderContent();
  EventsOn('evaluation_progress', msg => {
    const s1 = document.getElementById('eval-status');
    const s2 = document.getElementById('upload-status');
    if (s1) s1.textContent = msg;
    if (s2) s2.textContent = msg;
  });
  EventsOn('evaluation_error', msg => showModal('Vyhodnotenie písomiek', `<div class="error">${msg}</div>`));
  EventsOn('evaluation_done', payload => {
    let msg = 'Vyhodnotenie dokončené.';
    let extra = '';
    if (payload?.hadFailures && payload.failedPath) {
      msg = 'Niektoré strany sa nepodarilo spracovať.';
      extra = `<div class="file-row"><span class="file-path">${payload.failedPath}</span><button id="open-failed" class="btn">Otvoriť PDF</button></div>`;
    }
    showModal('Vyhodnotenie písomiek', `<div class="status success">${msg}</div>${extra}`);
    if (payload?.failedPath) document.getElementById('open-failed')?.addEventListener('click', () => OpenPath(payload.failedPath));
    refreshData();
  });
}

init();
