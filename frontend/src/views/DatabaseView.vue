<script setup>
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useDatabaseStore } from '../stores/database'
import { useToastStore } from '../stores/toast'
import { fileApi, recentApi } from '../api'
import AppHeader from '../components/layout/AppHeader.vue'
import AppSidebar from '../components/layout/AppSidebar.vue'
import TableData from '../components/table/TableData.vue'
import TableSchema from '../components/table/TableSchema.vue'
import CreateTableModal from '../components/table/CreateTableModal.vue'
import DatabaseInfo from '../components/database/DatabaseInfo.vue'
import Modal from '../components/common/Modal.vue'
import Button from '../components/common/Button.vue'
import Input from '../components/common/Input.vue'
import { 
  Table, Columns, Info, Loader2, Database, Upload, FolderOpen, FileText, 
  Folder, FolderPlus, ChevronRight, ChevronLeft, X
} from 'lucide-vue-next'

const router = useRouter()
const store = useDatabaseStore()
const toast = useToastStore()

const sidebarRef = ref(null)
const activeTab = ref('data')
const showCreateTable = ref(false)
const fileInput = ref(null)
const showBrowseModal = ref(false)
const showCreateModal = ref(false)
const newDbName = ref('')
const newDbInputRef = ref(null)
const isDragging = ref(false)

// 监听当前表变化，自动切换到数据Tab
watch(() => store.currentTable, (newTable) => {
  if (newTable) {
    activeTab.value = 'data'
  }
})

// 监听 store.activeTab 变化（用于"打开新数据库"按钮）
watch(() => store.activeTab, (newTab) => {
  activeTab.value = newTab
})

// File browser state
const currentPath = ref('')
const parentPath = ref('')
const fileList = ref([])
const shareDirs = ref([])  // 共享目录列表
const browsingLoading = ref(false)

const tabs = [
  { id: 'data', label: '数据', icon: Table },
  { id: 'schema', label: '结构', icon: Columns },
  { id: 'info', label: '信息', icon: Info }
]

const showContent = computed(() => store.currentTable !== null)

onMounted(async () => {
  await store.loadDatabases()
  // If has databases, set active tab to data
  if (store.databases.length > 0) {
    activeTab.value = 'data'
  }
})

// File browser functions
async function openBrowseModal() {
  showBrowseModal.value = true
  await browseDirectory('')
}

async function browseDirectory(path) {
  browsingLoading.value = true
  try {
    const response = await fileApi.browse(path)
    currentPath.value = response.data.currentPath
    parentPath.value = response.data.parent
    fileList.value = response.data.files || []
    shareDirs.value = response.data.shareDirs || []
  } catch (error) {
    toast.error(error.response?.data?.error || '无法访问目录')
  } finally {
    browsingLoading.value = false
  }
}

// 跳转到共享目录
async function goToShare(sharePath) {
  await browseDirectory(sharePath)
}

async function handleItemClick(file) {
  if (file.isDir) {
    await browseDirectory(file.path)
  } else {
    const success = await store.openDatabase(file.path)
    if (success) {
      showBrowseModal.value = false
      activeTab.value = 'data'
      // Refresh recent databases in sidebar
      if (sidebarRef.value) {
        sidebarRef.value.addToRecent(file.path, file.name)
      }
    }
  }
}

async function goToParent() {
  if (parentPath.value) {
    await browseDirectory(parentPath.value)
  }
}

// Upload functions
function handleBrowseClick() {
  fileInput.value?.click()
}

async function handleFileUpload(event) {
  const file = event.target.files?.[0]
  if (!file) return
  const success = await store.uploadDatabase(file)
  if (success) {
    activeTab.value = 'data'
  }
  event.target.value = ''
}

async function handleDrop(event) {
  isDragging.value = false
  const file = event.dataTransfer.files?.[0]
  if (file && (file.name.endsWith('.db') || file.name.endsWith('.sqlite') || file.name.endsWith('.sqlite3'))) {
    const success = await store.uploadDatabase(file)
    if (success) {
      activeTab.value = 'data'
    }
  }
}

function handleDragOver(event) {
  event.preventDefault()
  isDragging.value = true
}

function handleDragLeave() {
  isDragging.value = false
}

function openCreateModal() {
  newDbName.value = ''
  showCreateModal.value = true
  nextTick(() => {
    newDbInputRef.value?.focus()
  })
}

async function handleCreateDatabase() {
  if (!newDbName.value.trim()) {
    toast.warning('请输入数据库名称')
    return
  }
  const success = await store.createDatabase(newDbName.value.trim())
  if (success) {
    showCreateModal.value = false
    activeTab.value = 'data'
  }
}

function formatSize(bytes) {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

function formatSharePath(path) {
  if (!path) return ''
  // 路径较长时，截断开头部分
  if (path.length > 25) {
    return '...' + path.slice(-22)
  }
  return path
}
</script>

<template>
  <div class="h-screen flex flex-col bg-slate-900">
    <AppHeader />
    
    <div class="flex-1 flex overflow-hidden">
      <!-- Sidebar - always show -->
      <AppSidebar 
        ref="sidebarRef"
        @create-table="showCreateTable = true" 
        @select-database="openBrowseModal"
        @create-database="openCreateModal"
      />
      
      <main class="flex-1 flex flex-col overflow-hidden">
        <!-- Tabs -->
        <div class="border-b border-slate-700 bg-slate-800/50">
          <nav class="flex px-4">
            <button
              v-for="tab in tabs"
              :key="tab.id"
              @click="activeTab = tab.id"
              :disabled="!store.currentTable"
              :class="[
                'flex items-center gap-2 px-4 py-3 text-sm font-medium border-b-2 -mb-px transition-colors',
                activeTab === tab.id
                  ? 'border-primary-500 text-primary-400'
                  : 'border-transparent text-slate-400 hover:text-slate-200',
                !store.currentTable ? 'opacity-50 cursor-not-allowed' : ''
              ]"
            >
              <component :is="tab.icon" class="w-4 h-4" />
              {{ tab.label }}
            </button>
          </nav>
        </div>
        
        <!-- Content Area -->
        <div class="flex-1 overflow-auto">
          <!-- Loading -->
          <div v-if="store.loading" class="h-full flex items-center justify-center">
            <Loader2 class="w-8 h-8 text-primary-400 animate-spin" />
          </div>
          
          <!-- Table Content -->
          <template v-else-if="showContent">
            <div class="p-4 h-full">
              <TableData v-if="activeTab === 'data'" />
              <TableSchema v-else-if="activeTab === 'schema'" />
              <DatabaseInfo v-else-if="activeTab === 'info'" />
            </div>
          </template>
          
          <!-- No database or table selected - show welcome -->
          <div v-else class="h-full flex items-center justify-center p-8">
            <div class="text-center max-w-md">
              <div class="inline-flex items-center justify-center w-20 h-20 bg-primary-500/10 rounded-2xl mb-6">
                <Database class="w-10 h-10 text-primary-400" />
              </div>
              <h2 class="text-2xl font-bold text-slate-100 mb-2">欢迎使用 SQLite Manager</h2>
              <p class="text-slate-400 mb-8">从左侧选择一张表开始，或打开一个数据库</p>
              
              <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
                <button
                  @click="openBrowseModal"
                  class="flex flex-col items-center gap-2 p-4 bg-slate-800/50 hover:bg-slate-700/50 rounded-xl transition-colors"
                >
                  <FolderOpen class="w-6 h-6 text-primary-400" />
                  <span class="text-sm text-slate-300">浏览文件</span>
                </button>
                
                <button
                  @click="openCreateModal"
                  class="flex flex-col items-center gap-2 p-4 bg-slate-800/50 hover:bg-slate-700/50 rounded-xl transition-colors"
                >
                  <FolderPlus class="w-6 h-6 text-emerald-400" />
                  <span class="text-sm text-slate-300">新建数据库</span>
                </button>
                
                <button
                  @click="handleBrowseClick"
                  class="flex flex-col items-center gap-2 p-4 bg-slate-800/50 hover:bg-slate-700/50 rounded-xl transition-colors"
                >
                  <Upload class="w-6 h-6 text-cyan-400" />
                  <span class="text-sm text-slate-300">上传文件</span>
                </button>
              </div>
              
              <input
                ref="fileInput"
                type="file"
                accept=".db,.sqlite,.sqlite3"
                class="hidden"
                @change="handleFileUpload"
              />
            </div>
          </div>
        </div>
      </main>
    </div>

    <CreateTableModal
      :show="showCreateTable"
      @close="showCreateTable = false"
    />

    <!-- File Browser Modal -->
    <Modal :show="showBrowseModal" title="选择数据库文件" size="large" @close="showBrowseModal = false">
      <div class="space-y-4">
        <div class="flex items-center gap-2 p-3 bg-slate-700/30 rounded-lg">
          <Folder class="w-4 h-4 text-slate-500 flex-shrink-0" />
          <span class="text-sm text-slate-300 truncate font-mono" :title="currentPath">
            {{ currentPath }}
          </span>
        </div>

        <div class="flex gap-2">
          <Button 
            size="small" 
            variant="secondary" 
            @click="goToParent" 
            :disabled="!parentPath || browsingLoading"
            class="whitespace-nowrap flex-shrink-0"
          >
            <ChevronLeft class="w-4 h-4" />
            上一级
          </Button>
          
          <!-- 共享目录下拉选择 -->
          <select 
            v-if="shareDirs.length > 0"
            @change="goToShare($event.target.value); $event.target.selectedIndex = 0"
            class="min-w-0 flex-1 px-3 py-1.5 bg-slate-700 border border-slate-600 rounded-lg text-sm text-slate-200 focus:outline-none focus:ring-2 focus:ring-primary-500"
          >
            <option value="" disabled selected>切换共享目录...</option>
            <option 
              v-for="share in shareDirs" 
              :key="share.path" 
              :value="share.path"
              :title="share.path"
              class="truncate"
            >
              {{ share.name }} ({{ formatSharePath(share.path) }})
            </option>
          </select>
        </div>

        <div class="border border-slate-700 rounded-lg overflow-hidden">
          <div v-if="browsingLoading" class="p-8 text-center">
            <Loader2 class="w-6 h-6 text-primary-400 animate-spin mx-auto" />
          </div>
          <div v-else-if="fileList.length === 0" class="p-8 text-center text-slate-500">
            此目录为空或没有找到数据库文件
          </div>
          <div v-else class="max-h-80 overflow-auto">
            <div
              v-for="file in fileList"
              :key="file.path"
              @click="handleItemClick(file)"
              class="flex items-center gap-3 px-4 py-3 hover:bg-slate-700/50 cursor-pointer transition-colors border-b border-slate-700/50 last:border-0"
            >
              <Folder v-if="file.isDir" class="w-5 h-5 text-amber-400" />
              <FileText v-else class="w-5 h-5 text-primary-400" />
              <div class="flex-1 min-w-0">
                <p class="text-sm text-slate-200 truncate">{{ file.name }}</p>
              </div>
              <span v-if="!file.isDir" class="text-xs text-slate-500">
                {{ formatSize(file.size) }}
              </span>
              <ChevronRight v-if="file.isDir" class="w-4 h-4 text-slate-500" />
            </div>
          </div>
        </div>

        <p class="text-xs text-slate-500 text-center">
          点击文件夹进入，点击数据库文件打开
        </p>
      </div>
    </Modal>

    <!-- Create Database Modal -->
    <Modal :show="showCreateModal" title="新建数据库" @close="showCreateModal = false">
      <div class="space-y-4">
        <Input
          ref="newDbInputRef"
          v-model="newDbName"
          label="数据库名称"
          placeholder="例如: my_database"
          required
          autofocus
          @keyup.enter="handleCreateDatabase"
        />
        <p class="text-xs text-slate-500">
          数据库将保存到 <code class="bg-slate-700 px-1 py-0.5 rounded">databases/</code> 目录
        </p>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <Button variant="secondary" @click="showCreateModal = false">取消</Button>
          <Button @click="handleCreateDatabase" :disabled="!newDbName.trim()">创建</Button>
        </div>
      </template>
    </Modal>
  </div>
</template>
