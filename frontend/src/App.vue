<template>
  <div class="app-container">
    <!-- Top Navigation Bar -->
    <nav class="navbar">
      <div class="nav-brand">
        <h1>📊 ScanEvalApp</h1>
        <div class="nav-subtitle">Profesionálny systém na hodnotenie testov</div>
      </div>
      
      <div class="nav-tabs">
        <button 
          v-for="tab in tabs" 
          :key="tab.id"
          :class="['tab-button', { 'active': activeTab === tab.id }]"
          @click="switchTab(tab.id)"
        >
          {{ tab.label }}
        </button>
      </div>
      
      <div class="nav-actions">
        <button class="action-btn" @click="showSettings">
          <span>⚙️</span> Nastavenia
        </button>
        <button class="action-btn" @click="exportData">
          <span>📤</span> Export
        </button>
      </div>
    </nav>

    <!-- Main Content -->
    <main class="main-content">
      <!-- Create Exam Tab (Active by default) -->
      <div v-if="activeTab === 'create'" class="tab-content">
        <CreateExam />
      </div>
      
      <!-- Exams List Tab -->
      <div v-if="activeTab === 'exams'" class="tab-content">
        <div class="placeholder-content">
          <h2>📚 Zoznam testov</h2>
          <p>Tu bude zoznam všetkých vytvorených testov.</p>
          <button class="btn" @click="switchTab('create')">
            ➕ Vytvoriť nový test
          </button>
        </div>
      </div>
      
      <!-- Evaluation Tab -->
      <div v-if="activeTab === 'evaluate'" class="tab-content">
        <div class="placeholder-content">
          <h2>📝 Vyhodnocovanie</h2>
          <p>Tu budete vyhodnocovať naskenované odpovede.</p>
        </div>
      </div>
      
      <!-- Statistics Tab -->
      <div v-if="activeTab === 'stats'" class="tab-content">
        <div class="placeholder-content">
          <h2>📈 Štatistiky</h2>
          <p>Tu budú štatistiky a grafy výsledkov.</p>
        </div>
      </div>
    </main>

    <!-- Status Bar -->
    <footer class="status-bar">
      <div class="status-left">
        <span v-if="isLoading" class="loading">
          🔄 Spracovávam...
        </span>
        <span v-else>✅ Pripravené</span>
      </div>
      <div class="status-right">
        <span>{{ currentTime }}</span>
      </div>
    </footer>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import CreateExam from './pages/MainPage.vue'

const tabs = [
  { id: 'create', label: '➕ Vytvoriť Test' },
  { id: 'exams', label: '📚 Testy' },
  { id: 'evaluate', label: '📝 Vyhodnotiť' },
  { id: 'stats', label: '📈 Štatistiky' }
]

const activeTab = ref('create')
const isLoading = ref(false)
const currentTime = ref('')

// Update time every minute
function updateTime() {
  const now = new Date()
  currentTime.value = now.toLocaleTimeString('sk-SK', { 
    hour: '2-digit', 
    minute: '2-digit' 
  })
}

function switchTab(tabId) {
  activeTab.value = tabId
  isLoading.value = true
  // Simulate loading
  setTimeout(() => {
    isLoading.value = false
  }, 300)
}

function showSettings() {
  alert('Nastavenia budú čoskoro dostupné')
}

function exportData() {
  alert('Export funkcia bude čoskoro dostupná')
}

// Lifecycle
onMounted(() => {
  updateTime()
  const timer = setInterval(updateTime, 60000)
  onUnmounted(() => clearInterval(timer))
})
</script>

<style scoped>
.app-container {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: #f8fafc;
}

.navbar {
  background: linear-gradient(135deg, #1e3a8a 0%, #1e40af 100%);
  color: white;
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 70px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.nav-brand h1 {
  font-size: 22px;
  font-weight: 700;
  margin: 0;
}

.nav-subtitle {
  font-size: 12px;
  opacity: 0.8;
  margin-top: 2px;
}

.nav-tabs {
  display: flex;
  gap: 4px;
  background: rgba(255, 255, 255, 0.1);
  padding: 4px;
  border-radius: 12px;
}

.tab-button {
  padding: 10px 20px;
  border: none;
  background: transparent;
  color: rgba(255, 255, 255, 0.8);
  border-radius: 8px;
  cursor: pointer;
  font-weight: 500;
  transition: all 0.2s;
}

.tab-button:hover {
  background: rgba(255, 255, 255, 0.1);
  color: white;
}

.tab-button.active {
  background: white;
  color: #1e40af;
  font-weight: 600;
}

.nav-actions {
  display: flex;
  gap: 12px;
}

.action-btn {
  padding: 8px 16px;
  border: 1px solid rgba(255, 255, 255, 0.3);
  background: rgba(255, 255, 255, 0.1);
  color: white;
  border-radius: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.2s;
}

.action-btn:hover {
  background: rgba(255, 255, 255, 0.2);
}

.main-content {
  flex: 1;
  padding: 24px;
  overflow-y: auto;
}

.tab-content {
  background: white;
  border-radius: 16px;
  padding: 24px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
  height: 100%;
}

.placeholder-content {
  text-align: center;
  padding: 60px 20px;
}

.placeholder-content h2 {
  color: #1e3a8a;
  margin-bottom: 16px;
}

.placeholder-content p {
  color: #64748b;
  margin-bottom: 24px;
}

.btn {
  padding: 12px 24px;
  background: #3b82f6;
  color: white;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s;
}

.btn:hover {
  background: #2563eb;
}

.status-bar {
  padding: 12px 24px;
  background: #1e293b;
  color: #cbd5e1;
  font-size: 14px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.loading {
  color: #60a5fa;
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-right {
  opacity: 0.7;
}
</style>
