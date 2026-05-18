(function(){const n=document.createElement("link").relList;if(n&&n.supports&&n.supports("modulepreload"))return;for(const i of document.querySelectorAll('link[rel="modulepreload"]'))a(i);new MutationObserver(i=>{for(const s of i)if(s.type==="childList")for(const r of s.addedNodes)r.tagName==="LINK"&&r.rel==="modulepreload"&&a(r)}).observe(document,{childList:!0,subtree:!0});function o(i){const s={};return i.integrity&&(s.integrity=i.integrity),i.referrerpolicy&&(s.referrerPolicy=i.referrerpolicy),i.crossorigin==="use-credentials"?s.credentials="include":i.crossorigin==="anonymous"?s.credentials="omit":s.credentials="same-origin",s}function a(i){if(i.ep)return;i.ep=!0;const s=o(i);fetch(i.href,s)}})();function se(t,n,o,a,i,s,r,l){return window.go.main.App.CreateExamWithCSV(t,n,o,a,i,s,r,l)}function oe(t,n,o,a,i,s,r){return window.go.main.App.CreateMultiDayExamWithCSV(t,n,o,a,i,s,r)}function ie(t){return window.go.main.App.DeleteExam(t)}function re(t){return window.go.main.App.DownloadStudentSheet(t)}function R(t,n,o){return window.go.main.App.EvaluateExam(t,n,o)}function le(t){return window.go.main.App.ExportExamStudentsCSV(t)}function de(t,n,o,a,i,s,r,l){return window.go.main.App.ExportExamTemplateCSV(t,n,o,a,i,s,r,l)}function ce(t,n,o,a,i,s){return window.go.main.App.ExportMultiDayExamTemplateCSV(t,n,o,a,i,s)}function ue(t){return window.go.main.App.ExportMultiDayResultsCSV(t)}function me(t,n){return window.go.main.App.GenerateExamStatisticsPDF(t,n)}function Y(t){return window.go.main.App.GetExamAnswers(t)}function pe(){return window.go.main.App.GetSavePath()}function ve(){return window.go.main.App.ListConfigs()}function ge(){return window.go.main.App.ListExams()}function be(){return window.go.main.App.ListStudents()}function fe(t,n){return window.go.main.App.MergeResultCSVs(t,n)}function E(t){return window.go.main.App.OpenPath(t)}function H(t){return window.go.main.App.ParseExamTemplateCSV(t)}function he(){return window.go.main.App.PickCSVFile()}function ye(){return window.go.main.App.PickFolder()}function Z(){return window.go.main.App.PickPDF()}function we(t){return window.go.main.App.PrintExamPDF(t)}function Ce(){return window.go.main.App.PrintLegend()}function Ee(t){return window.go.main.App.PrintMultiDayExamPDFs(t)}function xe(t){return window.go.main.App.PrintStudentSheet(t)}function ke(t){return window.go.main.App.SetSavePath(t)}function Se(t,n){return window.go.main.App.UpdateExamAnswers(t,n)}function $e(t,n){return window.go.main.App.UpdateMultiDayAnswers(t,n)}function Ne(t,n,o){return window.runtime.EventsOnMultiple(t,n,o)}function q(t,n){return Ne(t,n,-1)}const e={activeTab:"exams",exams:[],students:[],configs:[],answers:[],csvContent:"",csvName:"",templateCsvName:"",formTitle:"",formSchoolYear:"",formDateTime:"",formQuestionCount:10,formOptionCount:5,formShowName:!0,evalExamId:0,evalFilePath:"",evalConfig:"",savePath:"",uploadExamId:0,uploadFilePath:"",uploadConfig:"",statsExamId:0,statsSelected:{},studentFilter:{name:"",surname:"",regNum:""},mdTitle:"",mdSchoolYear:"",mdQuestionCount:10,mdOptionCount:5,mdShowName:!0,mdCsvContent:"",mdCsvName:"",mdSubgroups:[{name:"",answers:""}],mdCheckboxMode:!0,mdTemplateCsvName:"",mdTemplateCsvContent:"",mergeOutputPath:"",mergeCsvPaths:[],mergeCsvNames:[],mergeOutputName:""},Ie=document.querySelector("#app");function Pe(){Ie.innerHTML=`
    <div class="page">
      <header class="header">
        <div class="title">
          <div class="badge">ScanEval</div>
          <h1>ScanEvalApp</h1>
       
        </div>
      </header>
      <nav class="tabs">
        <button class="tab" data-tab="exams">P\xEDsomky</button>
        <button class="tab" data-tab="multiday">Vytvori\u0165 p\xEDsomku</button>
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
  `,document.querySelectorAll(".tab").forEach(t=>{t.addEventListener("click",()=>{e.activeTab=t.dataset.tab,z()})}),document.getElementById("modal-close").addEventListener("click",V),document.getElementById("modal").addEventListener("click",t=>{t.target.id==="modal"&&V()})}const _=["Maximum bodov","Minimum bodov","Priemer","Medi\xE1n","Graf rozdelenia bodov celkovo","Graf rozdelenia za jednotliv\xE9 pr\xEDklady","\xDAspe\u0161nos\u0165 absol\xFAtna aj relat\xEDvna","\xDAspe\u0161nos\u0165 absol\xFAtna aj relat\xEDvna pre jednotliv\xE9 pr\xEDklady","\u0160tatistika pod\u013Ea miestnosti","\u0160tatistika pod\u013Ea skup\xEDn"];function Q(t){if(!t)return"-";const n=new Date(t);return Number.isNaN(n.getTime())?"-":new Intl.DateTimeFormat("sk-SK",{year:"numeric",month:"2-digit",day:"2-digit"}).format(n)}function Le(){const t=document.getElementById("content");if(!e.exams.length){t.innerHTML=`
      <div class="empty">
        <div class="empty-title">Ziadne testy</div>
        <div class="empty-sub">Najprv si vytvor novy test v aplikacii.</div>
      </div>
    `;return}t.innerHTML=`
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
      <span>CSV</span>
      <span>Zmaza\u0165</span>
    </div>
    ${e.exams.map(n=>`
        <div class="row">
          <span class="cell title" data-label="N\xE1zov">${n.title}${n.isMultiDay?' <span class="badge-multi">multi</span>':""}</span>
          <span class="cell" data-label="\u0160kolsk\xFD rok">${n.schoolYear}</span>
          <span class="cell" data-label="D\xE1tum">${Q(n.date)}</span>
          <span class="cell" data-label="Ot\xE1zky">${n.questionCount}</span>
          <span class="cell" data-label="Mo\u017Enosti">${n.optionCount||"-"}</span>
          <span class="cell" data-label="\u0160tudenti">${n.studentCount}</span>
          <span class="cell" data-label="Tla\u010Di\u0165">
            ${n.isMultiDay?`<button class="btn" data-action="print-multiday" data-exam-id="${n.id}">Tla\u010Di\u0165 po d\u0148och</button>`:`<button class="btn" data-action="print" data-exam-id="${n.id}">Tla\u010Di\u0165</button>`}
          </span>
          <span class="cell" data-label="Odpovede">
            <button class="btn" data-action="answers" data-exam-id="${n.id}">Odpovede</button>
          </span>
          <span class="cell" data-label="Vyhodnoti\u0165">
            <button class="btn" data-action="evaluate" data-exam-id="${n.id}">Vyhodnoti\u0165</button>
          </span>
          <span class="cell" data-label="\u0160tatistika PDF">
            <button class="btn" data-action="stats-pdf" data-exam-id="${n.id}">\u0160tatistika</button>
          </span>
          <span class="cell" data-label="CSV">
            ${n.isMultiDay?`<button class="btn" data-action="csv-multiday" data-exam-id="${n.id}">V\xFDsledky CSV</button>`:`<button class="btn" data-action="csv" data-exam-id="${n.id}">CSV</button>`}
          </span>
          <span class="cell" data-label="Zmaza\u0165">
            <button class="btn danger" data-action="delete" data-exam-id="${n.id}">Zmaza\u0165</button>
          </span>
        </div>
      `).join("")}
    </div>
    </div>
  `}function x(t,n){const o=document.getElementById("modal"),a=document.getElementById("modal-body"),i=document.getElementById("modal-title");i&&(i.textContent=t),a.innerHTML=n,o.classList.remove("hidden")}function V(){document.getElementById("modal").classList.add("hidden")}function L(t){return t.normalize("NFD").replace(/[\u0300-\u036f]/g,"")}function M(){const t=L(e.studentFilter.name.toLowerCase().trim()),n=L(e.studentFilter.surname.toLowerCase().trim()),o=L(e.studentFilter.regNum.toLowerCase().trim());document.querySelectorAll(".students-table .row").forEach(a=>{var c,h,y;const i=L((((c=a.querySelector('[data-label="Meno"]'))==null?void 0:c.textContent)||"").toLowerCase()),s=L((((h=a.querySelector('[data-label="Priezvisko"]'))==null?void 0:h.textContent)||"").toLowerCase()),r=(((y=a.querySelector('[data-label="Reg cislo"]'))==null?void 0:y.textContent)||"").toLowerCase(),l=i.includes(t)&&s.includes(n)&&r.includes(o);a.style.display=l?"":"none"})}function Ae(){const t=document.getElementById("content");if(!e.students.length){t.innerHTML=`
      <div class="empty">
        <div class="empty-title">\u017Diadni \u0161tudenti</div>
        <div class="empty-sub">Importuj \u0161tudentov cez vytvorenie testu.</div>
      </div>
    `;return}t.innerHTML=`
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
    ${e.students.map(n=>`
        <div class="row">
          <span class="cell title" data-label="Meno">${n.name}</span>
          <span class="cell" data-label="Priezvisko">${n.surname}</span>
          <span class="cell" data-label="Datum narodenia">${Q(n.birthDate)}</span>
          <span class="cell" data-label="Miestnost">${n.room}</span>
          <span class="cell" data-label="Reg cislo">${String(n.registrationNumber).padStart(7,"0")}</span>
          <span class="cell" data-label="Test">${n.examId}</span>
          <span class="cell" data-label="Score">${n.score}</span>
          <span class="cell action" data-label="Akcie">
            <button class="btn" data-student-print="${n.id}">Nov\xFD h\xE1rok</button>
            ${n.pages?`<button class="btn secondary" data-student-download="${n.id}">Vypracovan\xFD h\xE1rok</button>`:""}
          </span>
        </div>
      `).join("")}
    </div>
  `,M(),document.getElementById("filter-name").addEventListener("input",n=>{e.studentFilter.name=n.target.value,M()}),document.getElementById("filter-surname").addEventListener("input",n=>{e.studentFilter.surname=n.target.value,M()}),document.getElementById("filter-regnum").addEventListener("input",n=>{e.studentFilter.regNum=n.target.value,M()}),document.querySelectorAll("[data-student-print]").forEach(n=>{n.addEventListener("click",async()=>{const o=Number(n.dataset.studentPrint);if(!!o)try{const a=await xe(o);a&&await E(a)}catch(a){console.error(a),window.alert("Tlac harku zlyhala. Skontroluj logs.")}})}),document.querySelectorAll("[data-student-download]").forEach(n=>{n.addEventListener("click",async()=>{const o=Number(n.dataset.studentDownload);if(!!o)try{const a=await re(o);a&&await E(a)}catch(a){console.error(a),window.alert("Stiahnutie harku zlyhalo. Skontroluj logs.")}})})}function j(t){const n=["A","B","C","D","E","F","G","H"],o=Number(document.getElementById("option-count").value)||5,a=Number(document.getElementById("question-count").value)||0,i=Math.min(Math.max(o,3),n.length),s=n.slice(0,i);e.answers=Array.from({length:a},(r,l)=>e.answers[l]||""),t.innerHTML=e.answers.map((r,l)=>`
        <div class="answer-row">
          <span>Otazka ${String(l+1).padStart(2,"0")}</span>
          ${s.map(c=>`
              <label class="radio">
                <input type="radio" name="q${l}" value="${c}" ${e.answers[l]===c?"checked":""}/>
                <span>${c}</span>
              </label>
            `).join("")}
        </div>
      `).join(""),t.querySelectorAll('input[type="radio"]').forEach(r=>{r.addEventListener("change",()=>{const l=Number(r.name.replace("q",""))||0;e.answers[l]=r.value})})}function U(){const t=Math.max(Number(e.formQuestionCount)||0,0);e.answers=Array.from({length:t},(n,o)=>e.answers[o]||"")}function Be(){return e.csvName?e.csvName:e.csvContent?"Nacitane zo sablony":"Ziadny subor"}function Te(){return e.templateCsvName||"Ziadny subor"}function A(){const t=document.getElementById("content");U(),t.innerHTML=`
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
          <span class="file-name">${Be()}</span>
        </label>
        <label class="file">
          CSV sablona testu
          <input id="template-csv-file" type="file" accept=".csv" />
          <span class="file-name">${Te()}</span>
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
  `;const n=document.getElementById("create-exam"),o=document.getElementById("form-status"),a=document.getElementById("answers"),i=document.getElementById("csv-file"),s=document.getElementById("template-csv-file"),r=document.getElementById("title"),l=document.getElementById("school-year"),c=document.getElementById("date-time"),h=document.getElementById("question-count"),y=document.getElementById("option-count"),p=document.getElementById("show-name");r.addEventListener("input",()=>{e.formTitle=r.value}),l.addEventListener("input",()=>{e.formSchoolYear=l.value}),c.addEventListener("input",()=>{e.formDateTime=c.value}),h.addEventListener("input",()=>{e.formQuestionCount=Number(h.value)||0,U()}),y.addEventListener("input",()=>{e.formOptionCount=Number(y.value)||0}),p.addEventListener("change",()=>{e.formShowName=p.checked;const u=document.querySelector(".toggle-text");u&&(u.textContent=e.formShowName?"Ano":"Nie")}),document.getElementById("generate").addEventListener("click",()=>{j(a)}),i.addEventListener("change",async()=>{var f;const u=(f=i.files)==null?void 0:f[0];if(!u){e.csvContent="",e.csvName="",A();return}e.csvName=u.name;const g=await u.text();e.csvContent=g,A()}),s.addEventListener("change",async()=>{var g;const u=(g=s.files)==null?void 0:g[0];if(!u){e.templateCsvName="",A();return}o.textContent="Nacitavam sablonu...",o.className="status";try{const f=await u.text(),b=await H(f);e.templateCsvName=u.name,e.formTitle=b.title||"",e.formSchoolYear=b.schoolYear||"",e.formDateTime=b.dateTime||"",e.formQuestionCount=Number(b.questionCount)||0,e.formOptionCount=Number(b.optionCount)||5,e.formShowName=Boolean(b.showName),e.answers=Array.isArray(b.answers)?b.answers:[],e.csvContent=b.studentCSVContent||"",e.csvName=b.studentCSVContent?"Nacitane zo sablony":"",A();const d=document.getElementById("form-status");d&&(d.textContent="Sablona nacitana.",d.className="status success")}catch(f){console.error(f),o.textContent="Chyba pri nacitani sablony.",o.className="status error"}}),document.getElementById("save-template").addEventListener("click",async()=>{o.textContent="Ukladam sablonu...",o.className="status",U();try{const u=await de(e.formTitle.trim(),e.formSchoolYear.trim(),e.formDateTime.trim(),Number(e.formQuestionCount)||0,Number(e.formOptionCount)||0,e.answers,e.csvContent,Boolean(e.formShowName));o.textContent=u?`Sablona ulozena: ${u}`:"Sablona ulozena.",o.className="status success"}catch(u){console.error(u),o.textContent="Chyba pri ukladani sablony.",o.className="status error"}}),n.addEventListener("submit",async u=>{u.preventDefault(),o.textContent="Ukladam...",o.className="status";const g=e.formTitle.trim(),f=e.formSchoolYear.trim(),b=e.formDateTime.trim(),d=Number(e.formQuestionCount),m=Number(e.formOptionCount),v=Boolean(e.formShowName);j(a);try{await se(g,f,b,d,m,e.answers,e.csvContent,v),o.textContent="Ulozene.",o.className="status success",e.csvContent="",e.csvName="",e.templateCsvName="",e.answers=[],e.formTitle="",e.formSchoolYear="",e.formDateTime="",e.formQuestionCount=10,e.formOptionCount=5,e.formShowName=!0,await T(),e.activeTab="exams",z()}catch(w){console.error(w),o.textContent="Chyba pri ukladani.",o.className="status error"}}),j(a)}function G(){const t=e.configs.map(i=>`<option value="${i}" ${i===e.evalConfig?"selected":""}>${i}</option>`).join("");x("Vyhodnotenie pisomiek",`
    <div class="eval-form">
      <label>
        Konfiguracia skenera
        <select id="eval-config">${t}</select>
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
    `);const n=document.getElementById("eval-config"),o=document.getElementById("eval-pick"),a=document.getElementById("eval-start");n&&n.addEventListener("change",()=>{e.evalConfig=n.value}),o&&o.addEventListener("click",async()=>{const i=await Z();i&&(e.evalFilePath=i,G())}),a&&a.addEventListener("click",async()=>{const i=document.getElementById("eval-status");if(!e.evalFilePath||!e.evalConfig){i&&(i.textContent="Vyber PDF a konfiguraciu.");return}i&&(i.textContent="Spustam...");try{await R(e.evalExamId,e.evalFilePath,e.evalConfig)}catch(s){console.error(s),i&&(i.textContent="Spustenie zlyhalo.")}})}function ze(){const t=_.map(o=>{const a=e.statsSelected[o]?"checked":"";return`
        <label class="stats-item">
          <input type="checkbox" data-stat="${o}" ${a} />
          <span>${o}</span>
        </label>
      `}).join("");x("Vyberte pozadovane statistiky",`
    <div class="stats-form">
      <div class="stats-list">
        ${t}
      </div>
      <div class="form-actions">
        <button id="stats-generate" class="btn primary">Generovat statistiky</button>
        <span id="stats-status" class="status"></span>
      </div>
    </div>
    `),document.querySelectorAll("[data-stat]").forEach(o=>{o.addEventListener("change",()=>{const a=o.dataset.stat;e.statsSelected[a]=o.checked})});const n=document.getElementById("stats-generate");n&&n.addEventListener("click",async()=>{const o=document.getElementById("stats-status"),a=_.filter(i=>e.statsSelected[i]);if(a.length===0){o&&(o.textContent="Vyber aspon jednu statistiku.");return}o&&(o.textContent="Generujem...");try{const i=await me(e.statsExamId,a);i?(await E(i),o&&(o.textContent="Hotovo.")):o&&(o.textContent="PDF sa nepodarilo vytvorit.")}catch(i){console.error(i),o&&(o.textContent="Generovanie zlyhalo.")}})}function z(){if(document.querySelectorAll(".tab").forEach(t=>{t.classList.toggle("active",t.dataset.tab===e.activeTab)}),e.activeTab==="exams"){Le(),Oe();return}if(e.activeTab==="create"){A();return}if(e.activeTab==="multiday"){$();return}if(e.activeTab==="students"){Ae();return}if(e.activeTab==="upload"){W();return}if(e.activeTab==="settings"){J();return}e.activeTab==="csvmerge"&&B()}function N(t){return Array.from({length:t},(n,o)=>String.fromCharCode(65+o)).join("")}function S(t,n,o){if(t.length!==n)return!1;const a=N(o).toLowerCase();return t.toLowerCase().split("").every(i=>a.includes(i))}function k(t,n,o){const a=[];for(let i=0;i<n;i++)a.push(i<t.length?t[i].toUpperCase():"");return a}function K(t){return t.map(n=>n||"").join("").toUpperCase()}function Me(t,n,o,a){const i=Array.from({length:a},(l,c)=>String.fromCharCode(65+c)),s=t.checkboxAnswers||k(t.answers,o);return`<div class="sg-cb-grid">${Array.from({length:o},(l,c)=>{const h=i.map(y=>{const p=s[c]===y?"checked":"";return`<label class="sg-cb-label"><input type="radio" name="sg-${n}-q${c}" value="${y}" ${p} />${y}</label>`}).join("");return`<div class="sg-cb-row"><span class="sg-cb-qnum">${c+1}.</span>${h}</div>`}).join("")}</div>`}function O(t,n){const o=e.mdCheckboxMode;return e.mdSubgroups.map((a,i)=>{a.checkboxAnswers||(a.checkboxAnswers=k(a.answers,t));const s=a.answers.length>0&&S(a.answers,t,n),r=a.answers.length===0,l=r?"":s?"\u2713":`\u2717 (${a.answers.length}/${t})`,c=r?"":s?"sg-valid":"sg-invalid",h=o?Me(a,i,t,n):`<input class="sg-answers" type="text" placeholder="${N(n).repeat(Math.ceil(t/n)).slice(0,t)}" value="${a.answers}" />`;return`
      <div class="sg-row-inline" data-sg-idx="${i}">
        <input class="sg-name" type="text" placeholder="napr. A1" value="${a.name}" maxlength="4" />
        <div class="sg-inline-answers">${h}</div>
        <span class="sg-indicator ${c}">${l}</span>
        <button type="button" class="btn danger sg-remove">Odstrani\u0165</button>
      </div>
    `}).join("")}function D(t,n){document.querySelectorAll(".sg-name").forEach((o,a)=>{o.addEventListener("input",()=>{e.mdSubgroups[a].name=o.value.toUpperCase(),o.value=e.mdSubgroups[a].name})}),e.mdCheckboxMode?document.querySelectorAll(".sg-row-inline").forEach((o,a)=>{o.querySelectorAll('input[type="radio"]').forEach(i=>{i.addEventListener("mousedown",()=>{i.dataset.wasChecked=i.checked?"1":"0"}),i.addEventListener("click",()=>{const s=i.name.match(/q(\d+)$/);if(!s)return;const r=parseInt(s[1]);e.mdSubgroups[a].checkboxAnswers||(e.mdSubgroups[a].checkboxAnswers=k(e.mdSubgroups[a].answers,t)),i.dataset.wasChecked==="1"?(i.checked=!1,e.mdSubgroups[a].checkboxAnswers[r]=""):e.mdSubgroups[a].checkboxAnswers[r]=i.value,e.mdSubgroups[a].answers=K(e.mdSubgroups[a].checkboxAnswers);const l=o.querySelector(".sg-indicator");if(l){const c=e.mdSubgroups[a].answers;if(!c.replace(/ /g,"")){l.textContent="",l.className="sg-indicator";return}const y=S(c,t,n);l.textContent=y?"\u2713":`\u2717 (${c.length}/${t})`,l.className="sg-indicator "+(y?"sg-valid":"sg-invalid")}})})}):document.querySelectorAll(".sg-answers").forEach((o,a)=>{o.addEventListener("input",()=>{e.mdSubgroups[a].answers=o.value.toUpperCase(),o.value=e.mdSubgroups[a].answers,e.mdSubgroups[a].checkboxAnswers=k(e.mdSubgroups[a].answers,t);const s=o.closest(".sg-row-inline").querySelector(".sg-indicator"),r=e.mdSubgroups[a].answers;if(!r){s.textContent="",s.className="sg-indicator";return}const l=S(r,t,n);s.textContent=l?"\u2713":`\u2717 (${r.length}/${t})`,s.className="sg-indicator "+(l?"sg-valid":"sg-invalid")})}),document.querySelectorAll(".sg-remove").forEach((o,a)=>{o.addEventListener("click",()=>{e.mdSubgroups.splice(a,1),e.mdSubgroups.length===0&&e.mdSubgroups.push({name:"",answers:""}),document.getElementById("sg-container").innerHTML=O(t,n),D(t,n)})})}function $(){const t=document.getElementById("content"),n=e.mdQuestionCount,o=e.mdOptionCount;t.innerHTML=`
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
          <input id="md-question-count" type="number" min="1" value="${n}" required />
        </label>
        <label>
          Pocet moznosti
          <input id="md-option-count" type="number" min="3" max="8" value="${o}" style="min-height:42px;width:100%;box-sizing:border-box;" required />
        </label>
        <label class="toggle-field">
          Zobrazit meno
          <span class="toggle-row">
            <input id="md-show-name" type="checkbox" ${e.mdShowName?"checked":""} />
            <span class="md-toggle-text">${e.mdShowName?"Ano":"Nie"}</span>
          </span>
        </label>
        <label class="file">
          CSV so studentami (s datumom, casom, miestnostou)
          <input id="md-csv-file" type="file" accept=".csv" />
          <span class="file-name">${e.mdCsvName||"Ziadny subor"}</span>
        </label>
        <label class="file" style="margin-top:-8px;">
          CSV sablona testu
          <input id="md-template-csv-file" type="file" accept=".csv" />
          <span class="file-name">${e.mdTemplateCsvName||"Ziadny subor"}</span>
        </label>
        <div style="display:flex;flex-direction:column;gap:8px;margin-top:-8px;">
          <span style="font-size:13px;color:rgba(238,243,251,0.8);">&nbsp;</span>
          <button type="button" id="md-load-template" ${e.mdTemplateCsvContent?"":"disabled"} style="padding:10px 16px;border-radius:10px;border:1px solid rgba(255,255,255,0.18);font-weight:600;font-size:13px;cursor:${e.mdTemplateCsvContent?"pointer":"not-allowed"};background:${e.mdTemplateCsvContent?"#3b82f6":"rgba(59,130,246,0.25)"};color:${e.mdTemplateCsvContent?"#fff":"rgba(255,255,255,0.35)"};">Generova\u0165 pod\u013Ea \u0161abl\xF3ny</button>
        </div>
      </div>

      <div class="sg-section">
        <div class="sg-header">
          <span class="sg-title">Spr\xE1vne odpovede per podskupina</span>
          <span class="sg-hint">Povolen\xE9 znaky: ${N(o)} &nbsp;|&nbsp; D\u013A\u017Eka: ${n} znakov</span>
          <label class="toggle-field sg-mode-toggle">
            Checkboxy
            <span class="toggle-row">
              <input id="sg-checkbox-mode" type="checkbox" ${e.mdCheckboxMode?"checked":""} />
              <span class="md-toggle-text">${e.mdCheckboxMode?"Zap":"Vyp"}</span>
            </span>
          </label>
          <button type="button" id="sg-add" class="btn secondary">+ Prida\u0165 podskupinu</button>
        </div>
        ${e.mdCheckboxMode?"":`<div class="sg-labels"><span>Podskupina</span><span>Odpovede (${n} znakov)</span><span></span><span></span></div>`}
        <div id="sg-container">
          ${O(n,o)}
        </div>
      </div>

      <div class="form-actions">
        <button type="button" id="md-save-template" class="secondary">Ulozit CSV sablonu</button>
        <button type="submit" class="primary">Vytvorit multi-terminovy test</button>
        <span id="md-status" class="status"></span>
      </div>
    </form>
  `,document.getElementById("md-title").addEventListener("input",a=>{e.mdTitle=a.target.value}),document.getElementById("md-school-year").addEventListener("input",a=>{e.mdSchoolYear=a.target.value}),document.getElementById("md-question-count").addEventListener("input",a=>{e.mdQuestionCount=Number(a.target.value)||0,$()}),document.getElementById("md-option-count").addEventListener("input",a=>{e.mdOptionCount=Number(a.target.value)||0,$()}),document.getElementById("md-show-name").addEventListener("change",a=>{e.mdShowName=a.target.checked;const i=document.querySelector(".md-toggle-text");i&&(i.textContent=e.mdShowName?"Ano":"Nie")}),document.getElementById("md-template-csv-file").addEventListener("change",async a=>{var s;const i=(s=a.target.files)==null?void 0:s[0];if(!i){e.mdTemplateCsvName="",e.mdTemplateCsvContent="",$();return}e.mdTemplateCsvName=i.name,e.mdTemplateCsvContent=await i.text(),$()}),document.getElementById("md-load-template").addEventListener("click",async()=>{if(!e.mdTemplateCsvContent)return;const a=document.getElementById("md-status");a.textContent="Na\u010D\xEDtavam \u0161abl\xF3nu...",a.className="status";try{const i=await H(e.mdTemplateCsvContent);e.mdTitle=i.title||"",e.mdSchoolYear=i.schoolYear||"",e.mdQuestionCount=Number(i.questionCount)||e.mdQuestionCount,e.mdOptionCount=Number(i.optionCount)||e.mdOptionCount,e.mdShowName=Boolean(i.showName),i.subgroupAnswers&&Object.keys(i.subgroupAnswers).length>0&&(e.mdSubgroups=Object.entries(i.subgroupAnswers).map(([r,l])=>({name:r,answers:l}))),$();const s=document.getElementById("md-status");s&&(s.textContent="\u0160abl\xF3na na\u010D\xEDtan\xE1.",s.className="status success")}catch(i){console.error(i);const s=document.getElementById("md-status");s&&(s.textContent="Chyba pri na\u010D\xEDtan\xED \u0161abl\xF3ny.",s.className="status error")}}),document.getElementById("md-save-template").addEventListener("click",async()=>{const a=document.getElementById("md-status");a.textContent="Ukladam sablonu...",a.className="status";const i={};for(const s of e.mdSubgroups)s.name.trim()&&(i[s.name.trim()]=s.answers.toUpperCase());try{const s=await ce(e.mdTitle.trim(),e.mdSchoolYear.trim(),Number(e.mdQuestionCount),Number(e.mdOptionCount),Boolean(e.mdShowName),i);a.textContent=s?`Sablona ulozena: ${s}`:"Sablona ulozena.",a.className="status success"}catch(s){console.error(s),a.textContent="Chyba pri ukladani sablony.",a.className="status error"}}),document.getElementById("md-csv-file").addEventListener("change",async a=>{var s;const i=(s=a.target.files)==null?void 0:s[0];if(!i){e.mdCsvContent="",e.mdCsvName="";return}e.mdCsvName=i.name,e.mdCsvContent=await i.text(),document.querySelector("#md-csv-file + .file-name").textContent=i.name}),document.getElementById("sg-checkbox-mode").addEventListener("change",a=>{a.target.checked&&e.mdSubgroups.forEach(r=>{r.checkboxAnswers=k(r.answers,n)}),e.mdCheckboxMode=a.target.checked;const i=a.target.nextElementSibling;i&&(i.textContent=e.mdCheckboxMode?"Zap":"Vyp"),document.getElementById("sg-container").innerHTML=O(n,o);const s=document.querySelector(".sg-labels");if(e.mdCheckboxMode)s&&s.remove();else if(!s){const r=document.getElementById("sg-container"),l=document.createElement("div");l.className="sg-labels",l.innerHTML=`<span>Podskupina</span><span>Odpovede (${n} znakov)</span><span></span><span></span>`,r.parentNode.insertBefore(l,r)}D(n,o)}),document.getElementById("sg-add").addEventListener("click",()=>{e.mdSubgroups.push({name:"",answers:""}),document.getElementById("sg-container").innerHTML=O(n,o),D(n,o)}),D(n,o),document.getElementById("multiday-form").addEventListener("submit",async a=>{a.preventDefault();const i=document.getElementById("md-status"),s={};for(const r of e.mdSubgroups)if(!!r.name.trim()){if(r.answers.length>0&&!S(r.answers,n,o)){i.textContent=`Podskupina "${r.name}": neplatn\xE9 odpovede (${r.answers.length}/${n} znakov, povolen\xE9: ${N(o)}).`,i.className="status error";return}s[r.name.trim()]=r.answers.toUpperCase()}i.textContent="Ukladam...",i.className="status";try{await oe(e.mdTitle.trim(),e.mdSchoolYear.trim(),Number(e.mdQuestionCount),Number(e.mdOptionCount),Boolean(e.mdShowName),e.mdCsvContent,s),i.textContent="Ulozene.",i.className="status success",e.mdTitle="",e.mdSchoolYear="",e.mdCsvContent="",e.mdCsvName="",e.mdQuestionCount=10,e.mdOptionCount=5,e.mdShowName=!0,e.mdSubgroups=[{name:"",answers:""}],e.mdTemplateCsvName="",e.mdTemplateCsvContent="",await T(),e.activeTab="exams",z()}catch(r){console.error(r),i.textContent="Chyba: "+((r==null?void 0:r.message)||r),i.className="status error"}})}function W(){const t=document.getElementById("content"),n=e.exams.map(l=>`<option value="${l.id}" ${Number(l.id)===Number(e.uploadExamId)?"selected":""}>${l.title}</option>`).join(""),o=e.configs.map(l=>`<option value="${l}" ${l===e.uploadConfig?"selected":""}>${l}</option>`).join("");t.innerHTML=`
    <div class="eval-form">
      <label>
        Test
        <select id="upload-exam">${n}</select>
      </label>
      <label>
        Konfiguracia skenera
        <select id="upload-config">${o}</select>
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
  `;const a=document.getElementById("upload-exam"),i=document.getElementById("upload-config"),s=document.getElementById("upload-pick"),r=document.getElementById("upload-start");a&&a.addEventListener("change",()=>{e.uploadExamId=Number(a.value)}),i&&i.addEventListener("change",()=>{e.uploadConfig=i.value}),s&&s.addEventListener("click",async()=>{const l=await Z();l&&(e.uploadFilePath=l,W())}),r&&r.addEventListener("click",async()=>{const l=document.getElementById("upload-status");if(!e.uploadExamId||!e.uploadConfig||!e.uploadFilePath){l&&(l.textContent="Vyber test, konfiguraciu a PDF.");return}l&&(l.textContent="Spustam...");try{await R(e.uploadExamId,e.uploadFilePath,e.uploadConfig)}catch(c){console.error(c),l&&(l.textContent="Spustenie zlyhalo.")}})}function J(){const t=document.getElementById("content"),n=e.savePath?`${e.savePath.replace(/[\\/]+$/,"")}/failed_pages`:"";t.innerHTML=`
    <div class="settings-form">
      <label>
        Miesto ukladania PDF
        <div class="file-row">
          <input id="settings-path" type="text" value="${e.savePath||""}" placeholder="/home/vbox/ScanEvalApp/output" />
          <button id="settings-pick" class="btn">Vybrat priecinok</button>
        </div>
      </label>
      <label>
        Failed pages
        <div class="file-row">
          <span class="file-path">${n||"Nenastavene"}</span>
          <button id="settings-open-failed" class="btn" ${n?"":"disabled"}>Otvorit priecinok</button>
        </div>
      </label>
      <div class="form-actions">
        <button id="settings-save" class="btn primary">Ulozit</button>
        <span id="settings-status" class="status"></span>
      </div>
    </div>
  `;const o=document.getElementById("settings-pick"),a=document.getElementById("settings-save"),i=document.getElementById("settings-open-failed"),s=document.getElementById("settings-path");s&&s.addEventListener("input",()=>{e.savePath=s.value.trim()}),o&&o.addEventListener("click",async()=>{const r=document.getElementById("settings-status");try{const l=await ye();l&&(e.savePath=l,J())}catch(l){console.error(l),r&&(r.textContent="Vyber priecinka zlyhal. Cestu mozes zadat rucne.")}}),a&&a.addEventListener("click",async()=>{const r=document.getElementById("settings-status");if(!e.savePath){r&&(r.textContent="Vyber priecinok.");return}r&&(r.textContent="Ukladam...");try{await ke(e.savePath),r&&(r.textContent="Ulozene.")}catch(l){console.error(l),r&&(r.textContent="Ulozenie zlyhalo.")}}),i&&n&&i.addEventListener("click",async()=>{const r=document.getElementById("settings-status");try{await E(n)}catch(l){console.error(l),r&&(r.textContent="Priecinok sa nepodarilo otvorit.")}})}function B(){const t=document.getElementById("content"),n=e.mergeCsvPaths.length?`<ul class="merge-file-list">
        ${e.mergeCsvNames.map((s,r)=>`
          <li class="merge-file-item">
            <span class="merge-file-idx">${r+1}.</span>
            <span class="merge-file-name" title="${e.mergeCsvPaths[r]}">${s}</span>
            <button class="btn danger small" data-remove-idx="${r}">\u2715</button>
          </li>`).join("")}
       </ul>`:'<div class="merge-empty-hint">Zatia\u013E \u017Eiadne s\xFAbory. Klikni \u201EPrida\u0165 CSV".</div>',o=e.mergeCsvPaths.length>=2&&e.mergeOutputName.trim()!=="";t.innerHTML=`
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
        <div class="merge-file-list-wrap">${n}</div>
      </div>

      <div class="form-actions">
        <button id="merge-run" class="btn primary" ${o?"":"disabled"}>
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
  `,document.getElementById("merge-output-name").addEventListener("input",s=>{e.mergeOutputName=s.target.value;const r=document.getElementById("merge-run");r&&(r.disabled=e.mergeCsvPaths.length<2||e.mergeOutputName.trim()==="")}),document.getElementById("merge-add").addEventListener("click",async()=>{const s=document.getElementById("merge-status");try{const r=await he();if(!r)return;if(e.mergeCsvPaths.includes(r)){s&&(s.textContent="Tento s\xFAbor je u\u017E v zozname.",s.className="status error");return}e.mergeCsvPaths.push(r),e.mergeCsvNames.push(r.split(/[/\\]/).pop()),s&&(s.textContent=""),B()}catch(r){console.error(r),s&&(s.textContent="Chyba pri v\xFDbere s\xFAboru.",s.className="status error")}}),document.querySelectorAll("[data-remove-idx]").forEach(s=>{s.addEventListener("click",()=>{const r=Number(s.dataset.removeIdx);e.mergeCsvPaths.splice(r,1),e.mergeCsvNames.splice(r,1),B()})});const a=document.getElementById("merge-clear");a&&a.addEventListener("click",()=>{e.mergeCsvPaths=[],e.mergeCsvNames=[],B()}),document.getElementById("merge-run").addEventListener("click",async()=>{const s=document.getElementById("merge-status");s&&(s.textContent="Zlu\u010Dujem\u2026",s.className="status");try{const r=await fe(e.mergeCsvPaths,e.mergeOutputName.trim());if(!r){s&&(s.textContent="Zru\u0161en\xE9.",s.className="status");return}e.mergeOutputPath=r,e.mergeCsvPaths=[],e.mergeCsvNames=[],s&&(s.textContent="Hotovo!",s.className="status success"),B()}catch(r){console.error(r),s&&(s.textContent="Chyba: "+((r==null?void 0:r.message)||r),s.className="status error")}});const i=document.getElementById("merge-open");i&&i.addEventListener("click",async()=>{e.mergeOutputPath&&await E(e.mergeOutputPath)})}function Oe(){document.querySelectorAll("[data-exam-id]").forEach(n=>{n.addEventListener("click",async()=>{const o=Number(n.dataset.examId);if(!o)return;const a=n.dataset.action;if(a==="print"){try{const s=await we(o);s&&await E(s)}catch(s){console.error(s),window.alert("Tla\u010D zlyhala. Skontroluj logs.")}return}if(a==="print-multiday"){try{const s=await Ee(o);s&&s.length>0&&(window.alert(`Vygenerovan\xFDch ${s.length} PDF s\xFAborov v prie\u010Dinku pdf_to_print.`),await E(s[0]))}catch(s){console.error(s),window.alert("Tla\u010D zlyhala. Skontroluj logs.")}return}if(a==="answers"){const s=e.exams.find(r=>r.id===o);if(s&&s.isMultiDay)try{const r=await Y(o),l=s.questionCount,c=s.optionCount||5;let h={};try{h=r?JSON.parse(r):{}}catch{}(()=>{const p=Object.entries(h).map(([d,m])=>({name:d,answers:m,cbAnswers:k(m,l,c)}));p.length===0&&p.push({name:"",answers:"",cbAnswers:Array(l).fill("")});let u=!0;const g=()=>p.map((d,m)=>{const v=d.answers.length>0&&S(d.answers,l,c),w=d.answers.length===0,C=w?"":v?"\u2713":`\u2717 ${d.answers.replace(/ /g,"").length}/${l}`,I=w?"":v?"sg-valid":"sg-invalid",X=Array.from({length:c},(te,P)=>String.fromCharCode(65+P)),ee=u?`<div class="sg-cb-grid">${Array.from({length:l},(te,P)=>{const ne=X.map(F=>{const ae=d.cbAnswers[P]===F?"checked":"";return`<label class="sg-cb-label"><input type="radio" name="modal-sg-${m}-q${P}" value="${F}" ${ae} />${F}</label>`}).join("");return`<div class="sg-cb-row"><span class="sg-cb-qnum">${P+1}.</span>${ne}</div>`}).join("")}</div>`:`<input class="sg-answers" type="text" placeholder="${N(c).repeat(Math.ceil(l/c)).slice(0,l)}" value="${d.answers}" />`;return`
                  <div class="sg-row-inline" data-modal-idx="${m}">
                    <input class="sg-name" type="text" placeholder="A1" value="${d.name}" maxlength="4" />
                    <div class="sg-inline-answers">${ee}</div>
                    <span class="sg-indicator ${I}">${C}</span>
                    <button type="button" class="btn danger sg-remove" data-modal-idx="${m}">Odstrani\u0165</button>
                  </div>`}).join(""),f=(d,m)=>{const v=d.querySelector(".sg-indicator");if(!v)return;const w=m.replace(/ /g,"");if(!w){v.textContent="",v.className="sg-indicator";return}const C=S(m,l,c);v.textContent=C?"\u2713":`\u2717 ${w.length}/${l}`,v.className="sg-indicator "+(C?"sg-valid":"sg-invalid")},b=()=>{document.querySelectorAll("#modal-sg-container .sg-name").forEach((d,m)=>{d.addEventListener("input",()=>{d.value=d.value.toUpperCase(),p[m].name=d.value})}),u?document.querySelectorAll("#modal-sg-container .sg-row-inline").forEach((d,m)=>{d.querySelectorAll('input[type="radio"]').forEach(v=>{v.addEventListener("mousedown",()=>{v.dataset.wasChecked=v.checked?"1":"0"}),v.addEventListener("click",()=>{const w=v.name.match(/q(\d+)$/);if(!w)return;const C=parseInt(w[1]);v.dataset.wasChecked==="1"?(v.checked=!1,p[m].cbAnswers[C]=""):p[m].cbAnswers[C]=v.value,p[m].answers=K(p[m].cbAnswers),f(d,p[m].answers)})})}):document.querySelectorAll("#modal-sg-container .sg-answers").forEach((d,m)=>{d.addEventListener("input",()=>{d.value=d.value.toUpperCase(),p[m].answers=d.value,p[m].cbAnswers=k(d.value,l,c),f(d.closest(".sg-row-inline"),d.value)})}),document.querySelectorAll("#modal-sg-container .sg-remove").forEach((d,m)=>{d.addEventListener("click",()=>{p.length<=1||(p.splice(m,1),document.getElementById("modal-sg-container").innerHTML=g(),b())})})};x("Odpovede podskup\xEDn",`
                <div class="answers-edit-form">
                  <div class="sg-modal-top-row">
                    <div class="sg-hint">Povolen\xE9: ${N(c)} &nbsp;|&nbsp; D\u013A\u017Eka: ${l} znakov</div>
                    <label class="toggle-field sg-mode-toggle">
                      Checkboxy
                      <span class="toggle-row">
                        <input id="sg-modal-cb-mode" type="checkbox" ${u?"checked":""} />
                        <span class="md-toggle-text">${u?"Zap":"Vyp"}</span>
                      </span>
                    </label>
                  </div>
                  <div id="modal-sg-container">${g()}</div>
                  <div class="form-actions" style="margin-top:12px">
                    <button id="modal-sg-add" class="btn secondary">+ Prida\u0165 podskupinu</button>
                    <button id="modal-sg-save" class="btn primary">Ulo\u017Ei\u0165</button>
                    <span id="modal-sg-status" class="status"></span>
                  </div>
                </div>`),document.getElementById("modal-sg-add").addEventListener("click",()=>{p.push({name:"",answers:"",cbAnswers:Array(l).fill("")}),document.getElementById("modal-sg-container").innerHTML=g(),b()}),document.getElementById("sg-modal-cb-mode").addEventListener("change",d=>{d.target.checked&&p.forEach(v=>{v.cbAnswers=k(v.answers,l,c)}),u=d.target.checked;const m=d.target.nextElementSibling;m&&(m.textContent=u?"Zap":"Vyp"),document.getElementById("modal-sg-container").innerHTML=g(),b()}),document.getElementById("modal-sg-save").addEventListener("click",async()=>{const d=document.getElementById("modal-sg-status"),m={};let v=!0;if(p.forEach(w=>{const C=w.name.trim(),I=w.answers.trim().toUpperCase();!C||(I.length>0&&!S(I,l,c)&&(d.textContent=`Podskupina "${C}": neplatn\xE9 odpovede.`,d.className="status error",v=!1),m[C]=I)}),!!v)try{await $e(o,m),V()}catch(w){console.error(w),d&&(d.textContent="Ukladanie zlyhalo.")}}),b()})()}catch(r){console.error(r),x("Odpovede testu",'<div class="error">Nacitavanie odpovedi zlyhalo.</div>')}else try{const r=s?s.questionCount:0,l=s&&s.optionCount||5,c=await Y(o),h=c?c.split(""):[],y=Array.from({length:l},(u,g)=>String.fromCharCode(65+g));let p="";for(let u=0;u<r;u++){const g=(h[u]||"").toLowerCase(),f=y.map(b=>`<label class="radio-opt"><input type="radio" name="ans-q${u}" value="${b.toLowerCase()}" ${g===b.toLowerCase()?"checked":""} />${b}</label>`).join("");p+=`<div class="answer-edit-row"><span class="answer-num">${u+1}.</span><div class="radio-opts">${f}</div></div>`}x("Odpovede testu",`
              <div class="answers-edit-form">
                <div class="answers-edit-grid">${p}</div>
                <div class="modal-actions">
                  <button id="save-answers-btn" class="btn primary">Ulo\u017Ei\u0165 odpovede</button>
                </div>
              </div>`),document.getElementById("save-answers-btn").addEventListener("click",async()=>{const u=[];for(let g=0;g<r;g++){const f=document.querySelector(`input[name="ans-q${g}"]:checked`);u.push(f?f.value:"")}try{await Se(o,u),V()}catch(g){console.error(g),window.alert("Ukladanie odpovedi zlyhalo.")}})}catch(r){console.error(r),x("Odpovede testu",'<div class="error">Nacitavanie odpovedi zlyhalo.</div>')}return}if(a==="evaluate"){e.evalExamId=o,e.evalFilePath="",!e.evalConfig&&e.configs.length>0&&(e.evalConfig=e.configs[0]),G();return}if(a==="stats-pdf"){e.statsExamId=o,Object.keys(e.statsSelected).length||_.forEach(s=>{e.statsSelected[s]=!0}),ze();return}if(a==="csv"){try{const s=await le(o);s&&await E(s)}catch(s){console.error(s),x("Export CSV",'<div class="error">Export studentov do CSV zlyhal.</div>')}return}if(a==="csv-multiday"){try{const s=await ue(o);s&&await E(s)}catch(s){console.error(s),x("Export CSV",'<div class="error">Export vysledkov zlyhal.</div>')}return}if(!!window.confirm("Naozaj chces zmazat tento test?"))try{await ie(o),await T(),z()}catch(s){console.error(s),window.alert("Zmazanie zlyhalo. Skontroluj logs.")}})});const t=document.getElementById("btn-print-legend");t&&t.addEventListener("click",async()=>{try{const n=await Ce();n&&await E(n)}catch(n){console.error(n),window.alert("Tla\u010D legendy zlyhala. Skontroluj logs.")}})}async function T(){const t={exams:[],students:[],configs:[],savePath:""};try{t.exams=await ge()||[]}catch(n){console.error("ListExams failed",n)}try{t.students=await be()||[]}catch(n){console.error("ListStudents failed",n)}try{t.configs=await ve()||[]}catch(n){console.error("ListConfigs failed",n)}try{t.savePath=await pe()||""}catch(n){console.error("GetSavePath failed",n)}e.exams=t.exams,e.students=t.students,e.configs=t.configs,!e.evalConfig&&e.configs.length>0&&(e.evalConfig=e.configs[0]),e.savePath=t.savePath,!e.uploadConfig&&e.configs.length>0&&(e.uploadConfig=e.configs[0]),!e.uploadExamId&&e.exams.length>0&&(e.uploadExamId=e.exams[0].id)}async function De(){Pe(),await T(),z(),q("evaluation_progress",t=>{const n=document.getElementById("eval-status");n&&(n.textContent=t);const o=document.getElementById("upload-status");o&&(o.textContent=t)}),q("evaluation_error",t=>{x("Vyhodnotenie pisomiek",`<div class="error">${t}</div>`)}),q("evaluation_done",t=>{let n="Vyhodnotenie dokoncene.",o="";if((t==null?void 0:t.hadFailures)&&t.failedPath&&(n=`Niektore strany sa nepodarilo spracovat (${t.failedCount||0} stran).`,o=`<div class="file-row"><span class="file-path">${t.failedPath}</span><button id="open-failed" class="btn">Otvorit PDF</button></div>`,(t==null?void 0:t.failedPages)&&t.failedPages.length>0&&(o+='<div style="margin-top: 20px;"><h3>Detaily zlyhan\xFDch str\xE1n:</h3>',o+='<div style="max-height: 400px; overflow-y: auto;">',o+='<table style="width: 100%; border-collapse: collapse; font-size: 12px;">',o+='<thead><tr style="background: #f0f0f0; position: sticky; top: 0; color: #333;">',o+='<th style="border: 1px solid #ddd; padding: 8px; text-align: left;">Strana PDF</th>',o+='<th style="border: 1px solid #ddd; padding: 8px; text-align: left;">Test / \u010Cas</th>',o+='<th style="border: 1px solid #ddd; padding: 8px; text-align: left;">D\xF4vod</th>',o+='<th style="border: 1px solid #ddd; padding: 8px; text-align: left;">Detail</th>',o+="</tr></thead><tbody>",t.failedPages.forEach(a=>{const i=(a.pageNumber||0)+1;let s=a.examTitle||"Nezn\xE1my test";a.examDate&&a.examTime?s+=`<br><small>${a.examDate} o ${a.examTime}</small>`:a.examDate&&(s+=`<br><small>${a.examDate}</small>`),a.room&&(s+=`<br><small>Miestnos\u0165: ${a.room}</small>`);const r=Ve(a.reason);let l=a.detailedReason||"";a.extractedAnswers&&a.reason==="PARTIAL_RECOGNITION"&&(l+=`<br><strong>Extrahovan\xE9 odpovede:</strong> ${a.extractedAnswers}`),a.unrecognizedQuestions&&a.unrecognizedQuestions.length>0&&(l+=`<br><strong>Nerozpoznan\xE9 ot\xE1zky:</strong> ${a.unrecognizedQuestions.join(", ")}`),o+="<tr>",o+=`<td style="border: 1px solid #ddd; padding: 8px;">${i}</td>`,o+=`<td style="border: 1px solid #ddd; padding: 8px;">${s}</td>`,o+=`<td style="border: 1px solid #ddd; padding: 8px;">${r}</td>`,o+=`<td style="border: 1px solid #ddd; padding: 8px;">${l}</td>`,o+="</tr>"}),o+="</tbody></table></div></div>")),x("Vyhodnotenie pisomiek",`<div class="status success">${n}</div>${o}`),t!=null&&t.failedPath){const a=document.getElementById("open-failed");a&&a.addEventListener("click",()=>E(t.failedPath))}T()})}function Ve(t){return{PANIC:"Kritick\xE1 chyba",IMAGE_EXTRACTION_ERROR:"Chyba pri extrakcii obr\xE1zka",ID_NOT_FOUND:"\u0160tudent nen\xE1jden\xFD",GROUP_NOT_RECOGNIZED:"Nerozpoznan\xE1 skupina",NO_ANSWERS_DETECTED:"\u017Diadne odpovede",NO_QUESTION_NUMBERS:"Nerozpoznan\xE9 \u010D\xEDsla ot\xE1zok",INVALID_QUESTION_COUNT:"Nespr\xE1vny po\u010Det ot\xE1zok",DB_UPDATE_ERROR:"Chyba datab\xE1zy",PARTIAL_RECOGNITION:"\u010Ciasto\u010Dn\xE9 rozpoznanie"}[t]||t}De();
