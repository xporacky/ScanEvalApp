<template>
  <div class="create-exam-page">
    <div class="page-header">
      <h2>📝 Vytvorenie nového testu</h2>
      <p class="subtitle">Vyplňte základné informácie o teste</p>
    </div>

    <!-- Basic Information Form -->
    <div class="form-section">
      <h3><span>📋</span> Základné informácie</h3>
      
      <div class="form-grid">
        <!-- Exam Name -->
        <div class="form-group">
          <label for="examName">
            <span class="label-icon">🏷️</span>
            Názov testu *
          </label>
          <input
            id="examName"
            v-model="examData.name"
            type="text"
            placeholder="Napríklad: Matematika - Stredná skúška"
            class="form-input"
            :class="{ 'error': !examData.name }"
          >
          <div v-if="!examData.name" class="error-message">
            Zadajte názov testu
          </div>
        </div>

        <!-- School Year -->
        <div class="form-group">
          <label for="schoolYear">
            <span class="label-icon">📅</span>
            Školský rok *
          </label>
          <input
            id="schoolYear"
            v-model="examData.schoolYear"
            type="text"
            placeholder="2024/25"
            class="form-input"
            :class="{ 'error': !isValidSchoolYear }"
          >
          <div v-if="!isValidSchoolYear && examData.schoolYear" class="error-message">
            Formát: YYYY/YY
          </div>
          <div class="input-hint" v-else>Formát: 2024/25</div>
        </div>

        <!-- Date and Time -->
        <div class="form-group">
          <label for="examDate">
            <span class="label-icon">⏰</span>
            Dátum a čas *
          </label>
          <input
            id="examDate"
            v-model="examData.dateTime"
            type="text"
            placeholder="15.12.2024 09:00"
            class="form-input"
            :class="{ 'error': !isValidDateTime }"
          >
          <div class="input-hint">Formát: DD.MM.RRRR HH:MM</div>
        </div>

        <!-- Question Count -->
        <div class="form-group">
          <label for="questionCount">
            <span class="label-icon">❓</span>
            Počet otázok *
          </label>
          <div class="number-input-wrapper">
            <button 
              class="number-btn"
              @click="decreaseQuestionCount"
              :disabled="examData.questionCount <= 1"
            >
              -
            </button>
            <input
              id="questionCount"
              v-model.number="examData.questionCount"
              type="number"
              min="1"
              max="100"
              class="form-input number-input"
            >
            <button 
              class="number-btn"
              @click="increaseQuestionCount"
              :disabled="examData.questionCount >= 100"
            >
              +
            </button>
          </div>
        </div>

        <!-- Option Count -->
        <div class="form-group">
          <label>
            <span class="label-icon">🔢</span>
            Počet možností *
          </label>
          <div class="option-selector">
            <button
              v-for="option in [3, 4, 5, 6, 7, 8]"
              :key="option"
              :class="['option-btn', { 'active': examData.optionCount === option }]"
              @click="examData.optionCount = option"
            >
              {{ option }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- File Upload Section -->
    <div class="form-section">
      <h3><span>📁</span> Nahratie študentov</h3>
      
      <div 
        class="upload-area"
        @click="triggerFileUpload"
        @dragover.prevent="dragOver = true"
        @dragleave="dragOver = false"
        @drop.prevent="handleFileDrop"
        :class="{ 'drag-over': dragOver }"
      >
        <div class="upload-icon">
          <span>📤</span>
        </div>
        <div class="upload-text">
          <h4>Nahrať študentov (.csv)</h4>
          <p>Kliknite alebo potiahnite súbor sem</p>
          <p class="upload-hint">Podporovaný formát: CSV so študentmi</p>
        </div>
        <input
          type="file"
          ref="fileInput"
          accept=".csv"
          @change="handleFileSelect"
          style="display: none;"
        >
      </div>

      <!-- Selected File Info -->
      <div v-if="selectedFile" class="file-info">
        <div class="file-info-content">
          <div class="file-icon">
            📄
          </div>
          <div class="file-details">
            <div class="file-name">{{ selectedFile.name }}</div>
            <div class="file-size">{{ formatFileSize(selectedFile.size) }}</div>
          </div>
          <button class="clear-btn" @click="clearFile" title="Odstrániť">
            ✕
          </button>
        </div>
      </div>
    </div>

    <!-- Action Buttons -->
    <div class="action-buttons">
      <button 
        class="btn btn-primary btn-large"
        @click="generateQuestions"
        :disabled="!canGenerateQuestions"
        :class="{ 'disabled': !canGenerateQuestions }"
      >
        <span class="btn-icon">➕</span>
        Generovať otázky
      </button>

      <button 
        v-if="showQuestions"
        class="btn btn-success btn-large"
        @click="submitExam"
        :disabled="!canSubmitExam"
        :class="{ 'disabled': !canSubmitExam }"
      >
        <span class="btn-icon">💾</span>
        Vytvoriť test
      </button>
    </div>

    <!-- Questions Section (Dynamic) -->
    <div v-if="showQuestions" class="questions-section">
      <div class="questions-header">
        <h3><span>📝</span> Správne odpovede ({{ examData.questionCount }} otázok)</h3>
        <p>Vyberte správnu odpoveď pre každú otázku</p>
      </div>

      <div class="questions-grid">
        <div 
          v-for="question in questionForms" 
          :key="question.id"
          class="question-item"
        >
          <div class="question-number">
            {{ question.number }}
          </div>
          <div class="question-options">
            <div
              v-for="option in availableOptions"
              :key="option"
              class="option-wrapper"
            >
              <input
                type="radio"
                :name="'question-' + question.id"
                :id="'q' + question.id + '-' + option"
                :value="option"
                v-model="question.selectedOption"
                class="option-radio"
              >
              <label 
                :for="'q' + question.id + '-' + option"
                class="option-label"
              >
                {{ option }}
              </label>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Status Messages -->
    <div v-if="statusMessage" class="status-message" :class="statusType">
      {{ statusMessage }}
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'

// Reactive state
const examData = reactive({
  name: '',
  schoolYear: '',
  dateTime: '',
  questionCount: 20,
  optionCount: 5
})

const selectedFile = ref(null)
const showQuestions = ref(false)
const dragOver = ref(false)
const statusMessage = ref('')
const statusType = ref('info')
const fileInput = ref(null)

// Generate questions array
const questionForms = ref([])

// Computed properties
const isValidSchoolYear = computed(() => {
  return /^\d{4}\/\d{2}$/.test(examData.schoolYear)
})

const isValidDateTime = computed(() => {
  return /^\d{2}\.\d{2}\.\d{4} \d{2}:\d{2}$/.test(examData.dateTime)
})

const availableOptions = computed(() => {
  const letters = 'ABCDEFGH'
  return letters.slice(0, examData.optionCount).split('')
})

const canGenerateQuestions = computed(() => {
  return examData.name && 
         isValidSchoolYear.value && 
         isValidDateTime.value && 
         examData.questionCount > 0
})

const canSubmitExam = computed(() => {
  return showQuestions.value && 
         selectedFile.value && 
         questionForms.value.every(q => q.selectedOption)
})

// Methods
function decreaseQuestionCount() {
  if (examData.questionCount > 1) {
    examData.questionCount--
  }
}

function increaseQuestionCount() {
  if (examData.questionCount < 100) {
    examData.questionCount++
  }
}

function generateQuestions() {
  if (!canGenerateQuestions.value) {
    showStatus('Vyplňte všetky povinné polia', 'error')
    return
  }

  questionForms.value = Array.from(
    { length: examData.questionCount },
    (_, i) => ({
      id: i + 1,
      number: String(i + 1).padStart(2, '0'),
      selectedOption: null
    })
  )

  showQuestions.value = true
  showStatus(`Vygenerovaných ${examData.questionCount} otázok`, 'success')
}

function triggerFileUpload() {
  fileInput.value.click()
}

async function handleFileSelect(event) {
  const file = event.target.files[0]
  if (file && file.name.endsWith('.csv')) {
    selectedFile.value = file
    showStatus(`Súbor "${file.name}" vybraný`, 'success')
  } else {
    showStatus('Vyberte platný CSV súbor', 'error')
  }
}

function handleFileDrop(event) {
  dragOver.value = false
  const file = event.dataTransfer.files[0]
  if (file && file.name.endsWith('.csv')) {
    selectedFile.value = file
    showStatus(`Súbor "${file.name}" nahratý`, 'success')
  } else {
    showStatus('Potiahnite platný CSV súbor', 'error')
  }
}

function clearFile() {
  selectedFile.value = null
  if (fileInput.value) {
    fileInput.value.value = ''
  }
}

function formatFileSize(bytes) {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

function showStatus(message, type = 'info') {
  statusMessage.value = message
  statusType.value = type
  setTimeout(() => {
    statusMessage.value = ''
  }, 3000)
}

async function submitExam() {
  if (!canSubmitExam.value) {
    showStatus('Vyplňte všetky otázky a nahrajte súbor', 'error')
    return
  }

  // In real app, this would call Wails backend
  showStatus('Vytváram test...', 'info')
  
  // Simulate API call
  setTimeout(() => {
    const answers = questionForms.value
      .map(q => q.selectedOption)
      .join('')
    
    console.log('Exam Data:', { ...examData, answers })
    console.log('File:', selectedFile.value?.name)
    
    showStatus('Test úspešne vytvorený!', 'success')
    
    // Reset form after success
    setTimeout(() => {
      resetForm()
    }, 2000)
    
  }, 1500)
}

function resetForm() {
  examData.name = ''
  examData.schoolYear = ''
  examData.dateTime = ''
  examData.questionCount = 20
  examData.optionCount = 5
  selectedFile.value = null
  showQuestions.value = false
  questionForms.value = []
}

// Initialize with current date/time
onMounted(() => {
  const now = new Date()
  examData.dateTime = now.toLocaleDateString('sk-SK', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric'
  }) + ' ' + now.getHours().toString().padStart(2, '0') + ':' + 
        now.getMinutes().toString().padStart(2, '0')
})
</script>

<style scoped>
.create-exam-page {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.page-header {
  margin-bottom: 16px;
}

.page-header h2 {
  color: #1e3a8a;
  font-size: 24px;
  margin: 0;
  display: flex;
  align-items: center;
  gap: 12px;
}

.subtitle {
  color: #64748b;
  margin: 8px 0 0 0;
}

.form-section {
  background: white;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.form-section h3 {
  color: #334155;
  font-size: 18px;
  margin: 0 0 20px 0;
  display: flex;
  align-items: center;
  gap: 10px;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 20px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-group label {
  font-weight: 600;
  color: #475569;
  display: flex;
  align-items: center;
  gap: 8px;
}

.label-icon {
  font-size: 18px;
}

.form-input {
  padding: 12px 16px;
  border: 2px solid #e2e8f0;
  border-radius: 8px;
  font-size: 16px;
  transition: all 0.2s;
}

.form-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-input.error {
  border-color: #ef4444;
}

.error-message {
  color: #ef4444;
  font-size: 14px;
  margin-top: 4px;
}

.input-hint {
  color: #94a3b8;
  font-size: 14px;
  margin-top: 4px;
}

.number-input-wrapper {
  display: flex;
  align-items: center;
  gap: 8px;
}

.number-btn {
  width: 40px;
  height: 40px;
  border: 2px solid #e2e8f0;
  background: white;
  border-radius: 8px;
  font-size: 18px;
  font-weight: bold;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.number-btn:hover:not(:disabled) {
  background: #f8fafc;
  border-color: #3b82f6;
}

.number-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.number-input {
  flex: 1;
  text-align: center;
}

.option-selector {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 4px;
}

.option-btn {
  padding: 10px 16px;
  border: 2px solid #e2e8f0;
  background: white;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  min-width: 50px;
}

.option-btn:hover {
  border-color: #3b82f6;
  background: #f0f9ff;
}

.option-btn.active {
  background: #3b82f6;
  color: white;
  border-color: #3b82f6;
}

.upload-area {
  border: 3px dashed #cbd5e1;
  border-radius: 12px;
  padding: 48px 24px;
  text-align: center;
  cursor: pointer;
  transition: all 0.3s;
  background: #f8fafc;
}

.upload-area:hover, .upload-area.drag-over {
  border-color: #3b82f6;
  background: #f0f9ff;
}

.upload-icon {
  font-size: 48px;
  margin-bottom: 16px;
  opacity: 0.7;
}

.upload-text h4 {
  margin: 0 0 8px 0;
  color: #1e293b;
}

.upload-text p {
  margin: 0;
  color: #64748b;
}

.upload-hint {
  font-size: 14px;
  margin-top: 8px !important;
}

.file-info {
  margin-top: 16px;
  background: #f1f5f9;
  border-radius: 8px;
  padding: 16px;
}

.file-info-content {
  display: flex;
  align-items: center;
  gap: 16px;
}

.file-icon {
  font-size: 32px;
}

.file-details {
  flex: 1;
}

.file-name {
  font-weight: 600;
  color: #1e293b;
}

.file-size {
  color: #64748b;
  font-size: 14px;
  margin-top: 4px;
}

.clear-btn {
  background: none;
  border: none;
  color: #94a3b8;
  font-size: 20px;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
}

.clear-btn:hover {
  background: #e2e8f0;
  color: #64748b;
}

.action-buttons {
  display: flex;
  gap: 16px;
  margin-top: 16px;
}

.btn {
  padding: 16px 32px;
  border: none;
  border-radius: 10px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  transition: all 0.2s;
  flex: 1;
}

.btn-large {
  padding: 18px 36px;
}

.btn-primary {
  background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%);
  color: white;
}

.btn-success {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
}

.btn:hover:not(.disabled) {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.15);
}

.btn.disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none !important;
}

.btn-icon {
  font-size: 18px;
}

.questions-section {
  background: white;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  margin-top: 16px;
}

.questions-header {
  margin-bottom: 24px;
}

.questions-header h3 {
  color: #1e3a8a;
  margin: 0 0 8px 0;
  display: flex;
  align-items: center;
  gap: 10px;
}

.questions-header p {
  color: #64748b;
  margin: 0;
}

.questions-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 16px;
  max-height: 400px;
  overflow-y: auto;
  padding: 8px;
}

.question-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
  background: #f8fafc;
  border-radius: 8px;
  transition: background 0.2s;
}

.question-item:hover {
  background: #f1f5f9;
}

.question-number {
  font-size: 18px;
  font-weight: 700;
  color: #3b82f6;
  min-width: 40px;
}

.question-options {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.option-wrapper {
  display: flex;
  align-items: center;
}

.option-radio {
  display: none;
}

.option-label {
  width: 40px;
  height: 40px;
  border: 2px solid #cbd5e1;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  user-select: none;
}

.option-radio:checked + .option-label {
  background: #3b82f6;
  color: white;
  border-color: #3b82f6;
}

.option-label:hover {
  border-color: #3b82f6;
  background: #f0f9ff;
}

.status-message {
  padding: 16px;
  border-radius: 8px;
  margin-top: 16px;
  font-weight: 500;
  animation: fadeIn 0.3s ease;
}

.status-message.info {
  background: #dbeafe;
  color: #1e40af;
  border-left: 4px solid #3b82f6;
}

.status-message.success {
  background: #d1fae5;
  color: #065f46;
  border-left: 4px solid #10b981;
}

.status-message.error {
  background: #fee2e2;
  color: #991b1b;
  border-left: 4px solid #ef4444;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(-10px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
