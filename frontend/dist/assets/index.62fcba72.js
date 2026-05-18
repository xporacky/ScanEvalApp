(function(){const a=document.createElement("link").relList;if(a&&a.supports&&a.supports("modulepreload"))return;for(const r of document.querySelectorAll('link[rel="modulepreload"]'))s(r);new MutationObserver(r=>{for(const o of r)if(o.type==="childList")for(const i of o.addedNodes)i.tagName==="LINK"&&i.rel==="modulepreload"&&s(i)}).observe(document,{childList:!0,subtree:!0});function t(r){const o={};return r.integrity&&(o.integrity=r.integrity),r.referrerpolicy&&(o.referrerPolicy=r.referrerpolicy),r.crossorigin==="use-credentials"?o.credentials="include":r.crossorigin==="anonymous"?o.credentials="omit":o.credentials="same-origin",o}function s(r){if(r.ep)return;r.ep=!0;const o=t(r);fetch(r.href,o)}})();function se(n,a,t,s,r,o,i,l){return window.go.main.App.CreateExamWithCSV(n,a,t,s,r,o,i,l)}function oe(n,a,t,s,r,o,i){return window.go.main.App.CreateMultiDayExamWithCSV(n,a,t,s,r,o,i)}function re(n){return window.go.main.App.DeleteExam(n)}function ie(n){return window.go.main.App.DownloadStudentSheet(n)}function R(n,a,t){return window.go.main.App.EvaluateExam(n,a,t)}function le(n){return window.go.main.App.ExportExamAnswersCSV(n)}function de(n){return window.go.main.App.ExportExamStudentsCSV(n)}function ce(n,a,t,s,r,o,i,l){return window.go.main.App.ExportExamTemplateCSV(n,a,t,s,r,o,i,l)}function ue(n){return window.go.main.App.ExportMultiDayResultsCSV(n)}function me(n,a){return window.go.main.App.GenerateExamStatisticsPDF(n,a)}function Y(n){return window.go.main.App.GetExamAnswers(n)}function pe(){return window.go.main.App.GetSavePath()}function ve(){return window.go.main.App.ListConfigs()}function ge(){return window.go.main.App.ListExams()}function be(){return window.go.main.App.ListStudents()}function fe(n,a){return window.go.main.App.MergeResultCSVs(n,a)}function k(n){return window.go.main.App.OpenPath(n)}function H(n){return window.go.main.App.ParseExamTemplateCSV(n)}function he(){return window.go.main.App.PickCSVFile()}function ye(){return window.go.main.App.PickFolder()}function Z(){return window.go.main.App.PickPDF()}function we(n){return window.go.main.App.PrintExamPDF(n)}function Ce(){return window.go.main.App.PrintLegend()}function Ee(n){return window.go.main.App.PrintMultiDayExamPDFs(n)}function ke(n){return window.go.main.App.PrintStudentSheet(n)}function xe(n){return window.go.main.App.SetSavePath(n)}function Se(n,a){return window.go.main.App.UpdateExamAnswers(n,a)}function Ne(n,a){return window.go.main.App.UpdateMultiDayAnswers(n,a)}function $e(n,a,t){return window.runtime.EventsOnMultiple(n,a,t)}function F(n,a){return $e(n,a,-1)}const e={activeTab:"exams",exams:[],students:[],configs:[],answers:[],csvContent:"",csvName:"",templateCsvName:"",formTitle:"",formSchoolYear:"",formDateTime:"",formQuestionCount:10,formOptionCount:5,formShowName:!0,evalExamId:0,evalFilePath:"",evalConfig:"",savePath:"",uploadExamId:0,uploadFilePath:"",uploadConfig:"",statsExamId:0,statsSelected:{},studentFilter:{name:"",surname:"",regNum:""},mdTitle:"",mdSchoolYear:"",mdQuestionCount:10,mdOptionCount:5,mdShowName:!0,mdCsvContent:"",mdCsvName:"",mdTemplateCsvName:"",mdSubgroups:[{name:"",answers:""}],mdCheckboxMode:!0,mergeOutputPath:"",mergeCsvPaths:[],mergeCsvNames:[],mergeOutputName:""},Ie=document.querySelector("#app");function Ae(){Ie.innerHTML=`
    <div class="page">
      <header class="header">
        <div class="title">
          <div class="badge">ScanEval</div>
          <h1>ScanEvalApp</h1>
       
        </div>
      </header>
      <nav class="tabs">
        <button class="tab" data-tab="exams">P\xEDsomky</button>
        <button class="tab" data-tab="create">Vytvori\u0165 p\xEDsomku</button>
        <button class="tab" data-tab="multiday">Multi-term\xEDn</button>
        <button class="tab" data-tab="students">\u0160tudenti</button>
        <button class="tab" data-tab="upload">Vyhodnoti\u0165 p\xEDsomku</button>
        <button class="tab" data-tab="csvmerge">CSV Merge</button>
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
  `,document.querySelectorAll(".tab").forEach(n=>{n.addEventListener("click",()=>{e.activeTab=n.dataset.tab,B()})}),document.getElementById("modal-close").addEventListener("click",V),document.getElementById("modal").addEventListener("click",n=>{n.target.id==="modal"&&V()})}const _=["Maximum bodov","Minimum bodov","Priemer","Medi\xE1n","Graf rozdelenia bodov celkovo","Graf rozdelenia za jednotliv\xE9 pr\xEDklady","\xDAspe\u0161nos\u0165 absol\xFAtna aj relat\xEDvna","\xDAspe\u0161nos\u0165 absol\xFAtna aj relat\xEDvna pre jednotliv\xE9 pr\xEDklady"];function G(n){if(!n)return"-";const a=new Date(n);return Number.isNaN(a.getTime())?"-":new Intl.DateTimeFormat("sk-SK",{year:"numeric",month:"2-digit",day:"2-digit"}).format(a)}function Le(){const n=document.getElementById("content");if(!e.exams.length){n.innerHTML=`
      <div class="empty">
        <div class="empty-title">Ziadne testy</div>
        <div class="empty-sub">Najprv si vytvor novy test v aplikacii.</div>
      </div>
    `;return}n.innerHTML=`
    <div class="exams-page">
    <div class="exams-toolbar">
      <button class="btn secondary" id="btn-print-legend">Tla\u010Di\u0165 legendu</button>
    </div>
    <div class="exams-table">
    <div class="panel-header">
      <span>N\xE1zov</span>
      <span>\u0160kolsk\xFD rok</span>
      <span>D\xE1tum</span>
      <span>Ot\xE1zky</span>
      <span>Mo\u017Enosti</span>
      <span>\u0160tudenti</span>
      <span>Tla\u010Di\u0165</span>
      <span>Odpovede</span>
      <span>Vyhodnoti\u0165</span>
      <span>\u0160tatistika PDF</span>
      <span>Export odpoved\xED</span>
      <span>CSV</span>
      <span>Zmaza\u0165</span>
    </div>
    ${e.exams.map(a=>`
        <div class="row">
          <span class="cell title" data-label="N\xE1zov">${a.title}${a.isMultiDay?' <span class="badge-multi">multi</span>':""}</span>
          <span class="cell" data-label="\u0160kolsk\xFD rok">${a.schoolYear}</span>
          <span class="cell" data-label="D\xE1tum">${G(a.date)}</span>
          <span class="cell" data-label="Ot\xE1zky">${a.questionCount}</span>
          <span class="cell" data-label="Mo\u017Enosti">${a.optionCount||"-"}</span>
          <span class="cell" data-label="\u0160tudenti">${a.studentCount}</span>
          <span class="cell" data-label="Tla\u010Di\u0165">
            ${a.isMultiDay?`<button class="btn" data-action="print-multiday" data-exam-id="${a.id}">Tla\u010Di\u0165 po d\u0148och</button>`:`<button class="btn" data-action="print" data-exam-id="${a.id}">Tla\u010Di\u0165</button>`}
          </span>
          <span class="cell" data-label="Odpovede">
            <button class="btn" data-action="answers" data-exam-id="${a.id}">Odpovede</button>
          </span>
          <span class="cell" data-label="Vyhodnoti\u0165">
            <button class="btn" data-action="evaluate" data-exam-id="${a.id}">Vyhodnoti\u0165</button>
          </span>
              <span class="cell" data-label="\u0160tatistika PDF">
            <button class="btn" data-action="stats-pdf" data-exam-id="${a.id}">\u0160tatistika</button>
          </span>
          <span class="cell" data-label="Export odpoved\xED">
            <button class="btn" data-action="export-answers" data-exam-id="${a.id}">Uloz odpovede do CSV</button>
          </span>
          <span class="cell" data-label="CSV">
            ${a.isMultiDay?`<button class="btn" data-action="csv-multiday" data-exam-id="${a.id}">V\xFDsledky CSV</button>`:`<button class="btn" data-action="csv" data-exam-id="${a.id}">CSV</button>`}
          </span>
          <span class="cell" data-label="Zmaza\u0165">
            <button class="btn danger" data-action="delete" data-exam-id="${a.id}">Zmaza\u0165</button>
          </span>
        </div>
      `).join("")}
    </div>
    </div>
  `}function E(n,a){const t=document.getElementById("modal"),s=document.getElementById("modal-body"),r=document.getElementById("modal-title");r&&(r.textContent=n),s.innerHTML=a,t.classList.remove("hidden")}function V(){document.getElementById("modal").classList.add("hidden")}function A(n){return n.normalize("NFD").replace(/[\u0300-\u036f]/g,"")}function z(){const n=A(e.studentFilter.name.toLowerCase().trim()),a=A(e.studentFilter.surname.toLowerCase().trim()),t=A(e.studentFilter.regNum.toLowerCase().trim());document.querySelectorAll(".students-table .row").forEach(s=>{var d,b,v;const r=A((((d=s.querySelector('[data-label="Meno"]'))==null?void 0:d.textContent)||"").toLowerCase()),o=A((((b=s.querySelector('[data-label="Priezvisko"]'))==null?void 0:b.textContent)||"").toLowerCase()),i=(((v=s.querySelector('[data-label="Reg cislo"]'))==null?void 0:v.textContent)||"").toLowerCase(),l=r.includes(n)&&o.includes(a)&&i.includes(t);s.style.display=l?"":"none"})}function Pe(){const n=document.getElementById("content");if(!e.students.length){n.innerHTML=`
      <div class="empty">
        <div class="empty-title">\u017Diadni \u0161tudenti</div>
        <div class="empty-sub">Importuj \u0161tudentov cez vytvorenie testu.</div>
      </div>
    `;return}n.innerHTML=`
    <div class="students-filter">
      <input type="text" id="filter-name" placeholder="H\u013Eadaj pod\u013Ea mena\u2026" value="${e.studentFilter.name}">
      <input type="text" id="filter-surname" placeholder="H\u013Eadaj pod\u013Ea priezviska\u2026" value="${e.studentFilter.surname}">
      <input type="text" id="filter-regnum" placeholder="H\u013Eadaj pod\u013Ea reg. \u010D\xEDsla\u2026" value="${e.studentFilter.regNum}">
    </div>
    <div class="students-table">
    <div class="panel-header">
      <span>Meno</span>
      <span>Priezvisko</span>
      <span>D\xE1tum narodenia</span>
      <span>Mestnos\u0165</span>
      <span>Reg. \u010D\xEDslo</span>
      <span>P\xEDsomka</span>
      <span>Sk\xF3re</span>
      <span>Akcie</span>
    </div>
    ${e.students.map(a=>`
        <div class="row">
          <span class="cell title" data-label="Meno">${a.name}</span>
          <span class="cell" data-label="Priezvisko">${a.surname}</span>
          <span class="cell" data-label="Datum narodenia">${G(a.birthDate)}</span>
          <span class="cell" data-label="Miestnost">${a.room}</span>
          <span class="cell" data-label="Reg cislo">${String(a.registrationNumber).padStart(7,"0")}</span>
          <span class="cell" data-label="Test">${a.examId}</span>
          <span class="cell" data-label="Score">${a.score}</span>
          <span class="cell action" data-label="Akcie">
            <button class="btn" data-student-print="${a.id}">Nov\xFD h\xE1rok</button>
            ${a.pages?`<button class="btn secondary" data-student-download="${a.id}">Vypracovan\xFD h\xE1rok</button>`:""}
          </span>
        </div>
      `).join("")}
    </div>
  `,z(),document.getElementById("filter-name").addEventListener("input",a=>{e.studentFilter.name=a.target.value,z()}),document.getElementById("filter-surname").addEventListener("input",a=>{e.studentFilter.surname=a.target.value,z()}),document.getElementById("filter-regnum").addEventListener("input",a=>{e.studentFilter.regNum=a.target.value,z()}),document.querySelectorAll("[data-student-print]").forEach(a=>{a.addEventListener("click",async()=>{const t=Number(a.dataset.studentPrint);if(!!t)try{const s=await ke(t);s&&await k(s)}catch(s){console.error(s),window.alert("Tlac harku zlyhala. Skontroluj logs.")}})}),document.querySelectorAll("[data-student-download]").forEach(a=>{a.addEventListener("click",async()=>{const t=Number(a.dataset.studentDownload);if(!!t)try{const s=await ie(t);s&&await k(s)}catch(s){console.error(s),window.alert("Stiahnutie harku zlyhalo. Skontroluj logs.")}})})}function j(n){const a=["A","B","C","D","E","F","G","H"],t=Number(document.getElementById("option-count").value)||5,s=Number(document.getElementById("question-count").value)||0,r=Math.min(Math.max(t,3),a.length),o=a.slice(0,r);e.answers=Array.from({length:s},(i,l)=>e.answers[l]||""),n.innerHTML=e.answers.map((i,l)=>`
        <div class="answer-row">
          <span>Otazka ${String(l+1).padStart(2,"0")}</span>
          ${o.map(d=>`
              <label class="radio">
                <input type="radio" name="q${l}" value="${d}" ${e.answers[l]===d?"checked":""}/>
                <span>${d}</span>
              </label>
            `).join("")}
        </div>
      `).join(""),n.querySelectorAll('input[type="radio"]').forEach(i=>{i.addEventListener("change",()=>{const l=Number(i.name.replace("q",""))||0;e.answers[l]=i.value})})}function U(){const n=Math.max(Number(e.formQuestionCount)||0,0);e.answers=Array.from({length:n},(a,t)=>e.answers[t]||"")}function Te(){return e.csvName?e.csvName:e.csvContent?"Nacitane zo sablony":"Ziadny subor"}function Be(){return e.templateCsvName||"Ziadny subor"}function ze(n){var r,o,i,l;const a=n.replace(/\r\n/g,`
`).replace(/\r/g,`
`).split(`
`).map(d=>d.trim()).filter(d=>d.length>0).map(d=>{var f,y;const b=d.split(","),v=(f=b[0])==null?void 0:f.trim(),m=(y=b[1])==null?void 0:y.trim(),u=b.slice(2).join(",").trim();return{section:v,key:m,value:u}}),t={title:"",schoolYear:"",dateTime:"",questionCount:0,optionCount:5,showName:!0,answers:[],subgroupAnswers:{},studentCSVContent:""};if(a.length===0)return t;const s=a[0];for(((r=s.section)==null?void 0:r.toLowerCase())==="section"&&((o=s.key)==null?void 0:o.toLowerCase())==="key"&&((i=s.value)==null?void 0:i.toLowerCase())==="value"&&a.shift(),a.forEach(d=>{var u;const b=(u=d.section)==null?void 0:u.toLowerCase(),v=d.key,m=d.value;if(!(!b||!v))switch(b){case"meta":switch(v.toLowerCase()){case"title":t.title=m;break;case"school_year":t.schoolYear=m;break;case"date_time":t.dateTime=m;break;case"question_count":t.questionCount=Number(m)||t.questionCount;break;case"option_count":t.optionCount=Number(m)||t.optionCount;break;case"show_name":t.showName=m.toLowerCase()==="true";break}break;case"answer":t.answers[Number(v)-1]=m.toUpperCase();break;case"subgroup":t.subgroupAnswers[v]=m.toUpperCase();break;case"payload":v==="students_csv"&&(t.studentCSVContent=m);break}}),t.questionCount<=0&&(t.questionCount=t.answers.length||((l=Object.values(t.subgroupAnswers)[0])==null?void 0:l.length)||0),t.answers.length>t.questionCount&&(t.answers=t.answers.slice(0,t.questionCount));t.answers.length<t.questionCount;)t.answers.push("");return t}function L(){const n=document.getElementById("content");U(),n.innerHTML=`
    <form id="create-exam" class="form">
      <div class="form-grid">
        <label>
          Nazov testu
          <input id="title" type="text" value="${e.formTitle}" required />
        </label>
        <label>
          Skolsky rok (YYYY/YY)
          <input id="school-year" type="text" placeholder="2024/25" value="${e.formSchoolYear}" required />
        </label>
        <label>
          Datum (dd.MM.yyyy)
          <input id="date-time" type="text" placeholder="01.03.2025" value="${e.formDateTime}" required />
        </label>
        <label>
          Pocet otazok
          <input id="question-count" type="number" min="1" value="${e.formQuestionCount}" required />
        </label>
        <label>
          Pocet moznosti
          <input id="option-count" type="number" min="3" max="8" value="${e.formOptionCount}" required />
        </label>
        <label class="toggle-field">
          Zobrazit meno
          <span class="toggle-row">
            <input id="show-name" type="checkbox" ${e.formShowName?"checked":""} />
            <span class="toggle-text">${e.formShowName?"Ano":"Nie"}</span>
          </span>
        </label>
        <label class="file">
          CSV so studentmi
          <input id="csv-file" type="file" accept=".csv" />
          <span class="file-name">${Te()}</span>
        </label>
        <label class="file">
          Import odpovedi z CSV
          <input id="template-csv-file" type="file" accept=".csv" />
          <span class="file-name">${Be()}</span>
        </label>
      </div>
      <div class="form-actions">
        <button type="button" id="generate" class="secondary">Generovat otazky</button>
        <button type="button" id="save-template" class="secondary">Ulozit odpovede do CSV</button>
        <button type="submit" class="primary">Vytvorit test</button>
        <span id="form-status" class="status"></span>
      </div>
      <div id="answers" class="answers"></div>
    </form>
  `;const a=document.getElementById("create-exam"),t=document.getElementById("form-status"),s=document.getElementById("answers"),r=document.getElementById("csv-file"),o=document.getElementById("template-csv-file"),i=document.getElementById("title"),l=document.getElementById("school-year"),d=document.getElementById("date-time"),b=document.getElementById("question-count"),v=document.getElementById("option-count"),m=document.getElementById("show-name");i.addEventListener("input",()=>{e.formTitle=i.value}),l.addEventListener("input",()=>{e.formSchoolYear=l.value}),d.addEventListener("input",()=>{e.formDateTime=d.value}),b.addEventListener("input",()=>{e.formQuestionCount=Number(b.value)||0,U()}),v.addEventListener("input",()=>{e.formOptionCount=Number(v.value)||0}),m.addEventListener("change",()=>{e.formShowName=m.checked;const u=document.querySelector(".toggle-text");u&&(u.textContent=e.formShowName?"Ano":"Nie")}),document.getElementById("generate").addEventListener("click",()=>{j(s)}),r.addEventListener("change",async()=>{var y;const u=(y=r.files)==null?void 0:y[0];if(!u){e.csvContent="",e.csvName="",L();return}e.csvName=u.name;const f=await u.text();e.csvContent=f,L()}),o.addEventListener("change",async()=>{var f;const u=(f=o.files)==null?void 0:f[0];if(!u){e.templateCsvName="",L();return}t.textContent="Nacitavam sablonu...",t.className="status";try{const y=await u.text(),h=await H(y);e.templateCsvName=u.name,e.formTitle=h.title||"",e.formSchoolYear=h.schoolYear||"",e.formDateTime=h.dateTime||"",e.formQuestionCount=Number(h.questionCount)||0,e.formOptionCount=Number(h.optionCount)||5,e.formShowName=Boolean(h.showName),e.answers=Array.isArray(h.answers)?h.answers:[],e.csvContent=h.studentCSVContent||"",e.csvName=h.studentCSVContent?"Nacitane zo sablony":"",L();const c=document.getElementById("form-status");c&&(c.textContent="Sablona nacitana.",c.className="status success")}catch(y){console.error(y),t.textContent="Chyba pri nacitani sablony.",t.className="status error"}}),document.getElementById("save-template").addEventListener("click",async()=>{t.textContent="Ukladam sablonu...",t.className="status",U();try{const u=await ce(e.formTitle.trim(),e.formSchoolYear.trim(),e.formDateTime.trim(),Number(e.formQuestionCount)||0,Number(e.formOptionCount)||0,e.answers,e.csvContent,Boolean(e.formShowName));t.textContent=u?`Sablona ulozena: ${u}`:"Sablona ulozena.",t.className="status success"}catch(u){console.error(u),t.textContent="Chyba pri ukladani sablony.",t.className="status error"}}),a.addEventListener("submit",async u=>{u.preventDefault(),t.textContent="Ukladam...",t.className="status";const f=e.formTitle.trim(),y=e.formSchoolYear.trim(),h=e.formDateTime.trim(),c=Number(e.formQuestionCount),p=Number(e.formOptionCount),g=Boolean(e.formShowName);j(s);try{await se(f,y,h,c,p,e.answers,e.csvContent,g),t.textContent="Ulozene.",t.className="status success",e.csvContent="",e.csvName="",e.templateCsvName="",e.answers=[],e.formTitle="",e.formSchoolYear="",e.formDateTime="",e.formQuestionCount=10,e.formOptionCount=5,e.formShowName=!0,await T(),e.activeTab="exams",B()}catch(w){console.error(w),t.textContent="Chyba pri ukladani.",t.className="status error"}}),j(s)}function Q(){const n=e.configs.map(r=>`<option value="${r}" ${r===e.evalConfig?"selected":""}>${r}</option>`).join("");E("Vyhodnotenie pisomiek",`
    <div class="eval-form">
      <label>
        Konfiguracia skenera
        <select id="eval-config">${n}</select>
      </label>
      <label>
        PDF sken
        <div class="file-row">
          <span class="file-path">${e.evalFilePath||"Ziadny subor"}</span>
          <button id="eval-pick" class="btn">Vybrat PDF</button>
        </div>
      </label>
      <div class="form-actions">
        <button id="eval-start" class="btn primary">Spustit vyhodnotenie</button>
        <span id="eval-status" class="status"></span>
      </div>
    </div>
    `);const a=document.getElementById("eval-config"),t=document.getElementById("eval-pick"),s=document.getElementById("eval-start");a&&a.addEventListener("change",()=>{e.evalConfig=a.value}),t&&t.addEventListener("click",async()=>{const r=await Z();r&&(e.evalFilePath=r,Q())}),s&&s.addEventListener("click",async()=>{const r=document.getElementById("eval-status");if(!e.evalFilePath||!e.evalConfig){r&&(r.textContent="Vyber PDF a konfiguraciu.");return}r&&(r.textContent="Spustam...");try{await R(e.evalExamId,e.evalFilePath,e.evalConfig)}catch(o){console.error(o),r&&(r.textContent="Spustenie zlyhalo.")}})}function Me(){const n=_.map(t=>{const s=e.statsSelected[t]?"checked":"";return`
        <label class="stats-item">
          <input type="checkbox" data-stat="${t}" ${s} />
          <span>${t}</span>
        </label>
      `}).join("");E("Vyberte pozadovane statistiky",`
    <div class="stats-form">
      <div class="stats-list">
        ${n}
      </div>
      <div class="form-actions">
        <button id="stats-generate" class="btn primary">Generovat statistiky</button>
        <span id="stats-status" class="status"></span>
      </div>
    </div>
    `),document.querySelectorAll("[data-stat]").forEach(t=>{t.addEventListener("change",()=>{const s=t.dataset.stat;e.statsSelected[s]=t.checked})});const a=document.getElementById("stats-generate");a&&a.addEventListener("click",async()=>{const t=document.getElementById("stats-status"),s=_.filter(r=>e.statsSelected[r]);if(s.length===0){t&&(t.textContent="Vyber aspon jednu statistiku.");return}t&&(t.textContent="Generujem...");try{const r=await me(e.statsExamId,s);r?(await k(r),t&&(t.textContent="Hotovo.")):t&&(t.textContent="PDF sa nepodarilo vytvorit.")}catch(r){console.error(r),t&&(t.textContent="Generovanie zlyhalo.")}})}function B(){if(document.querySelectorAll(".tab").forEach(n=>{n.classList.toggle("active",n.dataset.tab===e.activeTab)}),e.activeTab==="exams"){Le(),De();return}if(e.activeTab==="create"){L();return}if(e.activeTab==="multiday"){D();return}if(e.activeTab==="students"){Pe();return}if(e.activeTab==="upload"){W();return}if(e.activeTab==="settings"){J();return}e.activeTab==="csvmerge"&&P()}function N(n){return Array.from({length:n},(a,t)=>String.fromCharCode(65+t)).join("")}function S(n,a,t){if(n.length!==a)return!1;const s=N(t).toLowerCase();return n.toLowerCase().split("").every(r=>s.includes(r))}function x(n,a,t){const s=[];for(let r=0;r<a;r++)s.push(r<n.length?n[r].toUpperCase():"");return s}function K(n){return n.map(a=>a||"").join("").toUpperCase()}function Oe(n,a,t,s){const r=Array.from({length:s},(l,d)=>String.fromCharCode(65+d)),o=n.checkboxAnswers||x(n.answers,t);return`<div class="sg-cb-grid">${Array.from({length:t},(l,d)=>{const b=r.map(v=>{const m=o[d]===v?"checked":"";return`<label class="sg-cb-label"><input type="radio" name="sg-${a}-q${d}" value="${v}" ${m} />${v}</label>`}).join("");return`<div class="sg-cb-row"><span class="sg-cb-qnum">${d+1}.</span>${b}</div>`}).join("")}</div>`}function M(n,a){const t=e.mdCheckboxMode;return e.mdSubgroups.map((s,r)=>{s.checkboxAnswers||(s.checkboxAnswers=x(s.answers,n));const o=s.answers.length>0&&S(s.answers,n,a),i=s.answers.length===0,l=i?"":o?"\u2713":`\u2717 (${s.answers.length}/${n})`,d=i?"":o?"sg-valid":"sg-invalid",b=t?Oe(s,r,n,a):`<input class="sg-answers" type="text" placeholder="${N(a).repeat(Math.ceil(n/a)).slice(0,n)}" value="${s.answers}" />`;return`
      <div class="sg-row-inline" data-sg-idx="${r}">
        <input class="sg-name" type="text" placeholder="napr. A1" value="${s.name}" maxlength="4" />
        <div class="sg-inline-answers">${b}</div>
        <span class="sg-indicator ${d}">${l}</span>
        <button type="button" class="btn danger sg-remove">Odstrani\u0165</button>
      </div>
    `}).join("")}function O(n,a){document.querySelectorAll(".sg-name").forEach((t,s)=>{t.addEventListener("input",()=>{e.mdSubgroups[s].name=t.value.toUpperCase(),t.value=e.mdSubgroups[s].name})}),e.mdCheckboxMode?document.querySelectorAll(".sg-row-inline").forEach((t,s)=>{t.querySelectorAll('input[type="radio"]').forEach(r=>{r.addEventListener("mousedown",()=>{r.dataset.wasChecked=r.checked?"1":"0"}),r.addEventListener("click",()=>{const o=r.name.match(/q(\d+)$/);if(!o)return;const i=parseInt(o[1]);e.mdSubgroups[s].checkboxAnswers||(e.mdSubgroups[s].checkboxAnswers=x(e.mdSubgroups[s].answers,n)),r.dataset.wasChecked==="1"?(r.checked=!1,e.mdSubgroups[s].checkboxAnswers[i]=""):e.mdSubgroups[s].checkboxAnswers[i]=r.value,e.mdSubgroups[s].answers=K(e.mdSubgroups[s].checkboxAnswers);const l=t.querySelector(".sg-indicator");if(l){const d=e.mdSubgroups[s].answers;if(!d.replace(/ /g,"")){l.textContent="",l.className="sg-indicator";return}const v=S(d,n,a);l.textContent=v?"\u2713":`\u2717 (${d.length}/${n})`,l.className="sg-indicator "+(v?"sg-valid":"sg-invalid")}})})}):document.querySelectorAll(".sg-answers").forEach((t,s)=>{t.addEventListener("input",()=>{e.mdSubgroups[s].answers=t.value.toUpperCase(),t.value=e.mdSubgroups[s].answers,e.mdSubgroups[s].checkboxAnswers=x(e.mdSubgroups[s].answers,n);const o=t.closest(".sg-row-inline").querySelector(".sg-indicator"),i=e.mdSubgroups[s].answers;if(!i){o.textContent="",o.className="sg-indicator";return}const l=S(i,n,a);o.textContent=l?"\u2713":`\u2717 (${i.length}/${n})`,o.className="sg-indicator "+(l?"sg-valid":"sg-invalid")})}),document.querySelectorAll(".sg-remove").forEach((t,s)=>{t.addEventListener("click",()=>{e.mdSubgroups.splice(s,1),e.mdSubgroups.length===0&&e.mdSubgroups.push({name:"",answers:""}),document.getElementById("sg-container").innerHTML=M(n,a),O(n,a)})})}function D(){const n=document.getElementById("content"),a=e.mdQuestionCount,t=e.mdOptionCount;n.innerHTML=`
    <form id="multiday-form" class="form">
      <div class="form-grid">
        <label>
          Nazov testu
          <input id="md-title" type="text" value="${e.mdTitle}" required />
        </label>
        <label>
          Skolsky rok (YYYY/YY)
          <input id="md-school-year" type="text" placeholder="2024/25" value="${e.mdSchoolYear}" required />
        </label>
        <label>
          Pocet otazok
          <input id="md-question-count" type="number" min="1" value="${a}" required />
        </label>
        <label>
          Pocet moznosti
          <input id="md-option-count" type="number" min="3" max="8" value="${t}" required />
        </label>
        <label class="toggle-field">
          Zobrazit meno
          <span class="toggle-row">
            <input id="md-show-name" type="checkbox" ${e.mdShowName?"checked":""} />
            <span class="md-toggle-text">${e.mdShowName?"Ano":"Nie"}</span>
          </span>
        </label>
        <label class="file">
          CSV zo systemu (s datumom, casom, miestnostou)
          <input id="md-csv-file" type="file" accept=".csv" />
          <span class="file-name">${e.mdCsvName||"Ziadny subor"}</span>
        </label>
        <label class="file">
          Import odpovedi z CSV
          <input id="md-template-csv-file" type="file" accept=".csv" />
          <span class="file-name">${e.mdTemplateCsvName||"Ziadny subor"}</span>
        </label>
      </div>

      <div class="sg-section">
        <div class="sg-header">
          <span class="sg-title">Spr\xE1vne odpovede per podskupina</span>
          <span class="sg-hint">Povolen\xE9 znaky: ${N(t)} &nbsp;|&nbsp; D\u013A\u017Eka: ${a} znakov</span>
          <label class="toggle-field sg-mode-toggle">
            Checkboxy
            <span class="toggle-row">
              <input id="sg-checkbox-mode" type="checkbox" ${e.mdCheckboxMode?"checked":""} />
              <span class="md-toggle-text">${e.mdCheckboxMode?"Zap":"Vyp"}</span>
            </span>
          </label>
          <button type="button" id="sg-add" class="btn secondary">+ Prida\u0165 podskupinu</button>
        </div>
        ${e.mdCheckboxMode?"":`<div class="sg-labels"><span>Podskupina</span><span>Odpovede (${a} znakov)</span><span></span><span></span></div>`}
        <div id="sg-container">
          ${M(a,t)}
        </div>
      </div>

      <div class="form-actions">
        <button type="submit" class="primary">Vytvorit multi-terminovy test</button>
        <span id="md-status" class="status"></span>
      </div>
    </form>
  `,document.getElementById("md-title").addEventListener("input",s=>{e.mdTitle=s.target.value}),document.getElementById("md-school-year").addEventListener("input",s=>{e.mdSchoolYear=s.target.value}),document.getElementById("md-question-count").addEventListener("input",s=>{e.mdQuestionCount=Number(s.target.value)||0,D()}),document.getElementById("md-option-count").addEventListener("input",s=>{e.mdOptionCount=Number(s.target.value)||0,D()}),document.getElementById("md-show-name").addEventListener("change",s=>{e.mdShowName=s.target.checked;const r=document.querySelector(".md-toggle-text");r&&(r.textContent=e.mdShowName?"Ano":"Nie")}),document.getElementById("md-csv-file").addEventListener("change",async s=>{var o;const r=(o=s.target.files)==null?void 0:o[0];if(!r){e.mdCsvContent="",e.mdCsvName="";return}e.mdCsvName=r.name,e.mdCsvContent=await r.text(),document.querySelector("#md-csv-file + .file-name").textContent=r.name}),document.getElementById("md-template-csv-file").addEventListener("change",async s=>{var i;const r=(i=s.target.files)==null?void 0:i[0],o=document.querySelector("#md-template-csv-file + .file-name");if(!r){e.mdTemplateCsvName="",o&&(o.textContent="Ziadny subor");return}e.mdTemplateCsvName=r.name,o&&(o.textContent=r.name);try{const l=await r.text();let d=await H(l);(!d.subgroupAnswers||Object.keys(d.subgroupAnswers).length===0)&&(d=ze(l)),e.mdTitle=d.title||"",e.mdSchoolYear=d.schoolYear||"",e.mdQuestionCount=Number(d.questionCount)||0,e.mdOptionCount=Number(d.optionCount)||5,e.mdShowName=Boolean(d.showName),d.subgroupAnswers&&Object.keys(d.subgroupAnswers).length>0?e.mdSubgroups=Object.entries(d.subgroupAnswers).map(([b,v])=>({name:b,answers:v.toUpperCase()})):Array.isArray(d.answers)&&d.answers.length>0?e.mdSubgroups=[{name:"A",answers:d.answers.join("").toUpperCase()}]:e.mdSubgroups=[{name:"",answers:""}],D()}catch(l){console.error(l),window.alert("Chyba pri nacitani CSV sablony.")}}),document.getElementById("sg-checkbox-mode").addEventListener("change",s=>{s.target.checked&&e.mdSubgroups.forEach(i=>{i.checkboxAnswers=x(i.answers,a)}),e.mdCheckboxMode=s.target.checked;const r=s.target.nextElementSibling;r&&(r.textContent=e.mdCheckboxMode?"Zap":"Vyp"),document.getElementById("sg-container").innerHTML=M(a,t);const o=document.querySelector(".sg-labels");if(e.mdCheckboxMode)o&&o.remove();else if(!o){const i=document.getElementById("sg-container"),l=document.createElement("div");l.className="sg-labels",l.innerHTML=`<span>Podskupina</span><span>Odpovede (${a} znakov)</span><span></span><span></span>`,i.parentNode.insertBefore(l,i)}O(a,t)}),document.getElementById("sg-add").addEventListener("click",()=>{e.mdSubgroups.push({name:"",answers:""}),document.getElementById("sg-container").innerHTML=M(a,t),O(a,t)}),O(a,t),document.getElementById("multiday-form").addEventListener("submit",async s=>{s.preventDefault();const r=document.getElementById("md-status"),o={};for(const i of e.mdSubgroups)if(!!i.name.trim()){if(i.answers.length>0&&!S(i.answers,a,t)){r.textContent=`Podskupina "${i.name}": neplatn\xE9 odpovede (${i.answers.length}/${a} znakov, povolen\xE9: ${N(t)}).`,r.className="status error";return}o[i.name.trim()]=i.answers.toUpperCase()}r.textContent="Ukladam...",r.className="status";try{await oe(e.mdTitle.trim(),e.mdSchoolYear.trim(),Number(e.mdQuestionCount),Number(e.mdOptionCount),Boolean(e.mdShowName),e.mdCsvContent,o),r.textContent="Ulozene.",r.className="status success",e.mdTitle="",e.mdSchoolYear="",e.mdCsvContent="",e.mdCsvName="",e.mdQuestionCount=10,e.mdOptionCount=5,e.mdShowName=!0,e.mdSubgroups=[{name:"",answers:""}],await T(),e.activeTab="exams",B()}catch(i){console.error(i),r.textContent="Chyba: "+((i==null?void 0:i.message)||i),r.className="status error"}})}function W(){const n=document.getElementById("content"),a=e.exams.map(l=>`<option value="${l.id}" ${Number(l.id)===Number(e.uploadExamId)?"selected":""}>${l.title}</option>`).join(""),t=e.configs.map(l=>`<option value="${l}" ${l===e.uploadConfig?"selected":""}>${l}</option>`).join("");n.innerHTML=`
    <div class="eval-form">
      <label>
        Test
        <select id="upload-exam">${a}</select>
      </label>
      <label>
        Konfiguracia skenera
        <select id="upload-config">${t}</select>
      </label>
      <label>
        PDF sken
        <div class="file-row">
          <span class="file-path">${e.uploadFilePath||"Ziadny subor"}</span>
          <button id="upload-pick" class="btn">Vybrat PDF</button>
        </div>
      </label>
      <div class="form-actions">
        <button id="upload-start" class="btn primary">Spustit vyhodnotenie</button>
        <span id="upload-status" class="status"></span>
      </div>
    </div>
  `;const s=document.getElementById("upload-exam"),r=document.getElementById("upload-config"),o=document.getElementById("upload-pick"),i=document.getElementById("upload-start");s&&s.addEventListener("change",()=>{e.uploadExamId=Number(s.value)}),r&&r.addEventListener("change",()=>{e.uploadConfig=r.value}),o&&o.addEventListener("click",async()=>{const l=await Z();l&&(e.uploadFilePath=l,W())}),i&&i.addEventListener("click",async()=>{const l=document.getElementById("upload-status");if(!e.uploadExamId||!e.uploadConfig||!e.uploadFilePath){l&&(l.textContent="Vyber test, konfiguraciu a PDF.");return}l&&(l.textContent="Spustam...");try{await R(e.uploadExamId,e.uploadFilePath,e.uploadConfig)}catch(d){console.error(d),l&&(l.textContent="Spustenie zlyhalo.")}})}function J(){const n=document.getElementById("content");n.innerHTML=`
    <div class="settings-form">
      <label>
        Miesto ukladania PDF
        <div class="file-row">
          <span class="file-path">${e.savePath||"Nenastavene"}</span>
          <button id="settings-pick" class="btn">Vybrat priecinok</button>
        </div>
      </label>
      <div class="form-actions">
        <button id="settings-save" class="btn primary">Ulozit</button>
        <span id="settings-status" class="status"></span>
      </div>
    </div>
  `;const a=document.getElementById("settings-pick"),t=document.getElementById("settings-save");a&&a.addEventListener("click",async()=>{const s=await ye();s&&(e.savePath=s,J())}),t&&t.addEventListener("click",async()=>{const s=document.getElementById("settings-status");if(!e.savePath){s&&(s.textContent="Vyber priecinok.");return}s&&(s.textContent="Ukladam...");try{await xe(e.savePath),s&&(s.textContent="Ulozene.")}catch(r){console.error(r),s&&(s.textContent="Ulozenie zlyhalo.")}})}function P(){const n=document.getElementById("content"),a=e.mergeCsvPaths.length?`<ul class="merge-file-list">
        ${e.mergeCsvNames.map((o,i)=>`
          <li class="merge-file-item">
            <span class="merge-file-idx">${i+1}.</span>
            <span class="merge-file-name" title="${e.mergeCsvPaths[i]}">${o}</span>
            <button class="btn danger small" data-remove-idx="${i}">\u2715</button>
          </li>`).join("")}
       </ul>`:'<div class="merge-empty-hint">Zatia\u013E \u017Eiadne s\xFAbory. Klikni \u201EPrida\u0165 CSV".</div>',t=e.mergeCsvPaths.length>=2&&e.mergeOutputName.trim()!=="";n.innerHTML=`
    <div class="settings-form csv-merge-form">
      <h2 class="merge-title">Zl\xFA\u010Denie v\xFDsledkov\xFDch CSV s\xFAborov</h2>
      <p class="merge-desc">
        Pridajte \u013Eubovo\u013En\xFD po\u010Det v\xFDsledkov\xFDch CSV s\xFAborov. Z\xE1klad tvoria riadky z
        <strong>prv\xE9ho</strong> pridan\xE9ho s\xFAboru \u2013 pre ka\u017Ed\xE9ho \u0161tudenta (pod\u013Ea Reg. \u010D\xEDsla)
        sa doplnia st\u013Apce <em>Podskupina, Sk\xF3re, Odpovede \u0161tudenta, Spr\xE1vne odpovede</em>
        zo s\xFAborov, kde m\xE1 dan\xFD \u0161tudent vyplnen\xE9 odpovede.
        V\xFDsledok sa ulo\u017E\xED do prie\u010Dinka nastaven\xE9ho v z\xE1lo\u017Eke Nastavenia.
      </p>

      <div class="merge-section">
        <label class="merge-name-label">
          N\xE1zov v\xFDsledn\xE9ho s\xFAboru
          <div class="merge-name-row">
            <input id="merge-output-name" class="merge-name-input" type="text"
              placeholder="napr. vysledky_merged"
              value="${e.mergeOutputName}" />
            <span class="merge-ext">.csv</span>
          </div>
        </label>
      </div>

      <div class="merge-section">
        <div class="merge-section-header">
          <span class="merge-section-label">Vstupn\xE9 CSV s\xFAbory (${e.mergeCsvPaths.length})</span>
          <div class="merge-btns">
            <button id="merge-add" class="btn">+ Prida\u0165 CSV</button>
            ${e.mergeCsvPaths.length>0?'<button id="merge-clear" class="btn secondary">Vymaza\u0165 zoznam</button>':""}
          </div>
        </div>
        <div class="merge-file-list-wrap">${a}</div>
      </div>

      <div class="form-actions">
        <button id="merge-run" class="btn primary" ${t?"":"disabled"}>
          Zl\xFA\u010Di\u0165 a ulo\u017Ei\u0165
        </button>
        <span id="merge-status" class="status"></span>
      </div>

      ${e.mergeOutputPath?`
        <div class="merge-result">
          <span class="merge-result-label">Naposledy ulo\u017Een\xE9:</span>
          <span class="merge-result-path">${e.mergeOutputPath}</span>
          <button id="merge-open" class="btn secondary">Otvori\u0165</button>
        </div>`:""}
    </div>
  `,document.getElementById("merge-output-name").addEventListener("input",o=>{e.mergeOutputName=o.target.value;const i=document.getElementById("merge-run");i&&(i.disabled=e.mergeCsvPaths.length<2||e.mergeOutputName.trim()==="")}),document.getElementById("merge-add").addEventListener("click",async()=>{const o=document.getElementById("merge-status");try{const i=await he();if(!i)return;if(e.mergeCsvPaths.includes(i)){o&&(o.textContent="Tento s\xFAbor je u\u017E v zozname.",o.className="status error");return}e.mergeCsvPaths.push(i),e.mergeCsvNames.push(i.split(/[/\\]/).pop()),o&&(o.textContent=""),P()}catch(i){console.error(i),o&&(o.textContent="Chyba pri v\xFDbere s\xFAboru.",o.className="status error")}}),document.querySelectorAll("[data-remove-idx]").forEach(o=>{o.addEventListener("click",()=>{const i=Number(o.dataset.removeIdx);e.mergeCsvPaths.splice(i,1),e.mergeCsvNames.splice(i,1),P()})});const s=document.getElementById("merge-clear");s&&s.addEventListener("click",()=>{e.mergeCsvPaths=[],e.mergeCsvNames=[],P()}),document.getElementById("merge-run").addEventListener("click",async()=>{const o=document.getElementById("merge-status");o&&(o.textContent="Zlu\u010Dujem\u2026",o.className="status");try{const i=await fe(e.mergeCsvPaths,e.mergeOutputName.trim());if(!i){o&&(o.textContent="Zru\u0161en\xE9.",o.className="status");return}e.mergeOutputPath=i,e.mergeCsvPaths=[],e.mergeCsvNames=[],o&&(o.textContent="Hotovo!",o.className="status success"),P()}catch(i){console.error(i),o&&(o.textContent="Chyba: "+((i==null?void 0:i.message)||i),o.className="status error")}});const r=document.getElementById("merge-open");r&&r.addEventListener("click",async()=>{e.mergeOutputPath&&await k(e.mergeOutputPath)})}function De(){document.querySelectorAll("[data-exam-id]").forEach(a=>{a.addEventListener("click",async()=>{const t=Number(a.dataset.examId);if(!t)return;const s=a.dataset.action;if(s==="print"){try{const o=await we(t);o&&await k(o)}catch(o){console.error(o),window.alert("Tla\u010D zlyhala. Skontroluj logs.")}return}if(s==="print-multiday"){try{const o=await Ee(t);o&&o.length>0&&(window.alert(`Vygenerovan\xFDch ${o.length} PDF s\xFAborov v prie\u010Dinku pdf_to_print.`),await k(o[0]))}catch(o){console.error(o),window.alert("Tla\u010D zlyhala. Skontroluj logs.")}return}if(s==="answers"){const o=e.exams.find(i=>i.id===t);if(o&&o.isMultiDay)try{const i=await Y(t),l=o.questionCount,d=o.optionCount||5;let b={};try{b=i?JSON.parse(i):{}}catch{}(()=>{const m=Object.entries(b).map(([c,p])=>({name:c,answers:p,cbAnswers:x(p,l,d)}));m.length===0&&m.push({name:"",answers:"",cbAnswers:Array(l).fill("")});let u=!0;const f=()=>m.map((c,p)=>{const g=c.answers.length>0&&S(c.answers,l,d),w=c.answers.length===0,C=w?"":g?"\u2713":`\u2717 ${c.answers.replace(/ /g,"").length}/${l}`,$=w?"":g?"sg-valid":"sg-invalid",X=Array.from({length:d},(te,I)=>String.fromCharCode(65+I)),ee=u?`<div class="sg-cb-grid">${Array.from({length:l},(te,I)=>{const ne=X.map(q=>{const ae=c.cbAnswers[I]===q?"checked":"";return`<label class="sg-cb-label"><input type="radio" name="modal-sg-${p}-q${I}" value="${q}" ${ae} />${q}</label>`}).join("");return`<div class="sg-cb-row"><span class="sg-cb-qnum">${I+1}.</span>${ne}</div>`}).join("")}</div>`:`<input class="sg-answers" type="text" placeholder="${N(d).repeat(Math.ceil(l/d)).slice(0,l)}" value="${c.answers}" />`;return`
                  <div class="sg-row-inline" data-modal-idx="${p}">
                    <input class="sg-name" type="text" placeholder="A1" value="${c.name}" maxlength="4" />
                    <div class="sg-inline-answers">${ee}</div>
                    <span class="sg-indicator ${$}">${C}</span>
                    <button type="button" class="btn danger sg-remove" data-modal-idx="${p}">Odstrani\u0165</button>
                  </div>`}).join(""),y=(c,p)=>{const g=c.querySelector(".sg-indicator");if(!g)return;const w=p.replace(/ /g,"");if(!w){g.textContent="",g.className="sg-indicator";return}const C=S(p,l,d);g.textContent=C?"\u2713":`\u2717 ${w.length}/${l}`,g.className="sg-indicator "+(C?"sg-valid":"sg-invalid")},h=()=>{document.querySelectorAll("#modal-sg-container .sg-name").forEach((c,p)=>{c.addEventListener("input",()=>{c.value=c.value.toUpperCase(),m[p].name=c.value})}),u?document.querySelectorAll("#modal-sg-container .sg-row-inline").forEach((c,p)=>{c.querySelectorAll('input[type="radio"]').forEach(g=>{g.addEventListener("mousedown",()=>{g.dataset.wasChecked=g.checked?"1":"0"}),g.addEventListener("click",()=>{const w=g.name.match(/q(\d+)$/);if(!w)return;const C=parseInt(w[1]);g.dataset.wasChecked==="1"?(g.checked=!1,m[p].cbAnswers[C]=""):m[p].cbAnswers[C]=g.value,m[p].answers=K(m[p].cbAnswers),y(c,m[p].answers)})})}):document.querySelectorAll("#modal-sg-container .sg-answers").forEach((c,p)=>{c.addEventListener("input",()=>{c.value=c.value.toUpperCase(),m[p].answers=c.value,m[p].cbAnswers=x(c.value,l,d),y(c.closest(".sg-row-inline"),c.value)})}),document.querySelectorAll("#modal-sg-container .sg-remove").forEach((c,p)=>{c.addEventListener("click",()=>{m.length<=1||(m.splice(p,1),document.getElementById("modal-sg-container").innerHTML=f(),h())})})};E("Odpovede podskup\xEDn",`
                <div class="answers-edit-form">
                  <div class="sg-modal-top-row">
                    <div class="sg-hint">Povolen\xE9: ${N(d)} &nbsp;|&nbsp; D\u013A\u017Eka: ${l} znakov</div>
                    <label class="toggle-field sg-mode-toggle">
                      Checkboxy
                      <span class="toggle-row">
                        <input id="sg-modal-cb-mode" type="checkbox" ${u?"checked":""} />
                        <span class="md-toggle-text">${u?"Zap":"Vyp"}</span>
                      </span>
                    </label>
                  </div>
                  <div id="modal-sg-container">${f()}</div>
                  <div class="form-actions" style="margin-top:12px">
                    <button id="modal-sg-add" class="btn secondary">+ Prida\u0165 podskupinu</button>
                    <button id="modal-sg-save" class="btn primary">Ulo\u017Ei\u0165</button>
                    <span id="modal-sg-status" class="status"></span>
                  </div>
                </div>`),document.getElementById("modal-sg-add").addEventListener("click",()=>{m.push({name:"",answers:"",cbAnswers:Array(l).fill("")}),document.getElementById("modal-sg-container").innerHTML=f(),h()}),document.getElementById("sg-modal-cb-mode").addEventListener("change",c=>{c.target.checked&&m.forEach(g=>{g.cbAnswers=x(g.answers,l,d)}),u=c.target.checked;const p=c.target.nextElementSibling;p&&(p.textContent=u?"Zap":"Vyp"),document.getElementById("modal-sg-container").innerHTML=f(),h()}),document.getElementById("modal-sg-save").addEventListener("click",async()=>{const c=document.getElementById("modal-sg-status"),p={};let g=!0;if(m.forEach(w=>{const C=w.name.trim(),$=w.answers.trim().toUpperCase();!C||($.length>0&&!S($,l,d)&&(c.textContent=`Podskupina "${C}": neplatn\xE9 odpovede.`,c.className="status error",g=!1),p[C]=$)}),!!g)try{await Ne(t,p),V()}catch(w){console.error(w),c&&(c.textContent="Ukladanie zlyhalo.")}}),h()})()}catch(i){console.error(i),E("Odpovede testu",'<div class="error">Nacitavanie odpovedi zlyhalo.</div>')}else try{const i=o?o.questionCount:0,l=o&&o.optionCount||5,d=await Y(t),b=d?d.split(""):[],v=Array.from({length:l},(u,f)=>String.fromCharCode(65+f));let m="";for(let u=0;u<i;u++){const f=(b[u]||"").toLowerCase(),y=v.map(h=>`<label class="radio-opt"><input type="radio" name="ans-q${u}" value="${h.toLowerCase()}" ${f===h.toLowerCase()?"checked":""} />${h}</label>`).join("");m+=`<div class="answer-edit-row"><span class="answer-num">${u+1}.</span><div class="radio-opts">${y}</div></div>`}E("Odpovede testu",`
              <div class="answers-edit-form">
                <div class="answers-edit-grid">${m}</div>
                <div class="modal-actions">
                  <button id="save-answers-btn" class="btn primary">Ulo\u017Ei\u0165 odpovede</button>
                </div>
              </div>`),document.getElementById("save-answers-btn").addEventListener("click",async()=>{const u=[];for(let f=0;f<i;f++){const y=document.querySelector(`input[name="ans-q${f}"]:checked`);u.push(y?y.value:"")}try{await Se(t,u),V()}catch(f){console.error(f),window.alert("Ukladanie odpovedi zlyhalo.")}})}catch(i){console.error(i),E("Odpovede testu",'<div class="error">Nacitavanie odpovedi zlyhalo.</div>')}return}if(s==="evaluate"){e.evalExamId=t,e.evalFilePath="",!e.evalConfig&&e.configs.length>0&&(e.evalConfig=e.configs[0]),Q();return}if(s==="stats-pdf"){e.statsExamId=t,Object.keys(e.statsSelected).length||_.forEach(o=>{e.statsSelected[o]=!0}),Me();return}if(s==="export-answers"){try{const o=await le(t);o&&await k(o)}catch(o){console.error(o),E("Export odpoved\xED",'<div class="error">Export odpoved\xED do CSV zlyhal.</div>')}return}if(s==="csv"){try{const o=await de(t);o&&await k(o)}catch(o){console.error(o),E("Export CSV",'<div class="error">Export studentov do CSV zlyhal.</div>')}return}if(s==="csv-multiday"){try{const o=await ue(t);o&&await k(o)}catch(o){console.error(o),E("Export CSV",'<div class="error">Export vysledkov zlyhal.</div>')}return}if(!!window.confirm("Naozaj chces zmazat tento test?"))try{await re(t),await T(),B()}catch(o){console.error(o),window.alert("Zmazanie zlyhalo. Skontroluj logs.")}})});const n=document.getElementById("btn-print-legend");n&&n.addEventListener("click",async()=>{try{const a=await Ce();a&&await k(a)}catch(a){console.error(a),window.alert("Tla\u010D legendy zlyhala. Skontroluj logs.")}})}async function T(){const n={exams:[],students:[],configs:[],savePath:""};try{n.exams=await ge()||[]}catch(a){console.error("ListExams failed",a)}try{n.students=await be()||[]}catch(a){console.error("ListStudents failed",a)}try{n.configs=await ve()||[]}catch(a){console.error("ListConfigs failed",a)}try{n.savePath=await pe()||""}catch(a){console.error("GetSavePath failed",a)}e.exams=n.exams,e.students=n.students,e.configs=n.configs,!e.evalConfig&&e.configs.length>0&&(e.evalConfig=e.configs[0]),e.savePath=n.savePath,!e.uploadConfig&&e.configs.length>0&&(e.uploadConfig=e.configs[0]),!e.uploadExamId&&e.exams.length>0&&(e.uploadExamId=e.exams[0].id)}async function Ve(){Ae(),await T(),B(),F("evaluation_progress",n=>{const a=document.getElementById("eval-status");a&&(a.textContent=n);const t=document.getElementById("upload-status");t&&(t.textContent=n)}),F("evaluation_error",n=>{E("Vyhodnotenie pisomiek",`<div class="error">${n}</div>`)}),F("evaluation_done",n=>{let a="Vyhodnotenie dokoncene.",t="";if((n==null?void 0:n.hadFailures)&&n.failedPath&&(a=`Niektore strany sa nepodarilo spracovat (${n.failedCount||0} stran).`,t=`<div class="file-row"><span class="file-path">${n.failedPath}</span><button id="open-failed" class="btn">Otvorit PDF</button></div>`,(n==null?void 0:n.failedPages)&&n.failedPages.length>0&&(t+='<div style="margin-top: 20px;"><h3>Detaily zlyhan\xFDch str\xE1n:</h3>',t+='<div style="max-height: 400px; overflow-y: auto;">',t+='<table style="width: 100%; border-collapse: collapse; font-size: 12px;">',t+='<thead><tr style="background: #f0f0f0; position: sticky; top: 0; color: #333;">',t+='<th style="border: 1px solid #ddd; padding: 8px; text-align: left;">Strana PDF</th>',t+='<th style="border: 1px solid #ddd; padding: 8px; text-align: left;">Test / \u010Cas</th>',t+='<th style="border: 1px solid #ddd; padding: 8px; text-align: left;">D\xF4vod</th>',t+='<th style="border: 1px solid #ddd; padding: 8px; text-align: left;">Detail</th>',t+="</tr></thead><tbody>",n.failedPages.forEach(s=>{const r=(s.pageNumber||0)+1;let o=s.examTitle||"Nezn\xE1my test";s.examDate&&s.examTime?o+=`<br><small>${s.examDate} o ${s.examTime}</small>`:s.examDate&&(o+=`<br><small>${s.examDate}</small>`),s.room&&(o+=`<br><small>Miestnos\u0165: ${s.room}</small>`);const i=qe(s.reason);let l=s.detailedReason||"";s.extractedAnswers&&s.reason==="PARTIAL_RECOGNITION"&&(l+=`<br><strong>Extrahovan\xE9 odpovede:</strong> ${s.extractedAnswers}`),s.unrecognizedQuestions&&s.unrecognizedQuestions.length>0&&(l+=`<br><strong>Nerozpoznan\xE9 ot\xE1zky:</strong> ${s.unrecognizedQuestions.join(", ")}`),t+="<tr>",t+=`<td style="border: 1px solid #ddd; padding: 8px;">${r}</td>`,t+=`<td style="border: 1px solid #ddd; padding: 8px;">${o}</td>`,t+=`<td style="border: 1px solid #ddd; padding: 8px;">${i}</td>`,t+=`<td style="border: 1px solid #ddd; padding: 8px;">${l}</td>`,t+="</tr>"}),t+="</tbody></table></div></div>")),E("Vyhodnotenie pisomiek",`<div class="status success">${a}</div>${t}`),n!=null&&n.failedPath){const s=document.getElementById("open-failed");s&&s.addEventListener("click",()=>k(n.failedPath))}T()})}function qe(n){return{PANIC:"Kritick\xE1 chyba",IMAGE_EXTRACTION_ERROR:"Chyba pri extrakcii obr\xE1zka",ID_NOT_FOUND:"\u0160tudent nen\xE1jden\xFD",GROUP_NOT_RECOGNIZED:"Nerozpoznan\xE1 skupina",NO_ANSWERS_DETECTED:"\u017Diadne odpovede",NO_QUESTION_NUMBERS:"Nerozpoznan\xE9 \u010D\xEDsla ot\xE1zok",INVALID_QUESTION_COUNT:"Nespr\xE1vny po\u010Det ot\xE1zok",DB_UPDATE_ERROR:"Chyba datab\xE1zy",PARTIAL_RECOGNITION:"\u010Ciasto\u010Dn\xE9 rozpoznanie"}[n]||n}Ve();
