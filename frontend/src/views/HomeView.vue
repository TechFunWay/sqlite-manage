<script setup>
import { ref, onMounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useDatabaseStore } from '../stores/database'
import { useToastStore } from '../stores/toast'
import { fileApi } from '../api'
import { 
  Database, Upload, FolderOpen, FileText, Clock, X, 
  Folder, FolderPlus, ChevronRight, ChevronLeft, Loader2
} from 'lucide-vue-next'
import Button from '../components/common/Button.vue'
import Modal from '../components/common/Modal.vue'
import Input from '../components/common/Input.vue'

const router = useRouter()
const store = useDatabaseStore()
const toast = useToastStore()

const fileInput = ref(null)
const showBrowseModal = ref(false)
const showCreateModal = ref(false)
const newDbName = ref('')
const newDbInputRef = ref(null)

// File browser state
const currentPath = ref('')
const parentPath = ref('')
const fileList = ref([])
const browsingLoading = ref(false)

const isDragging = ref(false)
const recentDatabases = ref(JSON.parse(localStorage.getItem('recentDatabases') || '[]'))

onMounted(async () => {
  await store.loadDatabases()
  
  // 处理从文件管理器右键打开的数据库文件
  const urlParams = new URLSearchParams(window.location.search)
  const dbPath = urlParams.get('path')
  if (dbPath) {
    // 检查是否已经打开过，避免重复
    const existingDb = store.databases.find(db => db.path === dbPath)
    if (existingDb) {
      await store.activateDatabase(existingDb.id)
    } else {
      const success = await store.openDatabase(dbPath)
      if (success) {
        addToRecent(dbPath)
      }
    }
    router.push('/database')
  }
})

function addToRecent(path) {
  const recent = recentDatabases.value.filter(d => d !== path)
  recent.unshift(path)
  recentDatabases.value = recent.slice(0, 5)
  localStorage.setItem('recentDatabases', JSON.stringify(recentDatabases.value))
}

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
  } catch (error) {
    toast.error(error.response?.data?.error || '无法访问目录')
  } finally {
    browsingLoading.value = false
  }
}

async function handleItemClick(file) {
  if (file.isDir) {
    await browseDirectory(file.path)
  } else {
    const success = await store.openDatabase(file.path)
    if (success) {
      addToRecent(file.path)
      showBrowseModal.value = false
      router.push('/database')
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
    router.push('/database')
  }
  event.target.value = ''
}

async function handleDrop(event) {
  isDragging.value = false
  const file = event.dataTransfer.files?.[0]
  if (file && (file.name.endsWith('.db') || file.name.endsWith('.sqlite') || file.name.endsWith('.sqlite3'))) {
    const success = await store.uploadDatabase(file)
    if (success) {
      router.push('/database')
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

async function openRecent(path) {
  const success = await store.openDatabase(path)
  if (success) {
    router.push('/database')
  }
}

async function selectDatabase(db) {
  await store.activateDatabase(db.id)
  router.push('/database')
}

async function closeDatabase(db, event) {
  event.stopPropagation()
  await store.closeDatabase(db.id)
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
    router.push('/database')
  }
}

function formatSize(bytes) {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}
</script>

<template>
  <div class="min-h-screen bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 flex items-center justify-center p-4">
    <div class="w-full max-w-xl">
      <!-- Logo -->
      <div class="text-center mb-10">
        <div class="inline-flex items-center justify-center w-20 h-20 bg-primary-500/10 rounded-2xl mb-6">
          <Database class="w-10 h-10 text-primary-400" />
        </div>
        <h1 class="text-4xl font-bold text-slate-100 mb-3">SQLite Manager</h1>
        <p class="text-slate-400 text-lg">简洁高效的 SQLite 数据库管理工具</p>
      </div>

      <!-- Main Card -->
      <div class="bg-slate-800/50 backdrop-blur-xl rounded-2xl border border-slate-700/50 p-6 shadow-2xl">
        <div class="space-y-4">
          <!-- Open from Server -->
          <button
            @click="openBrowseModal"
            class="w-full flex items-center gap-4 p-4 bg-slate-700/30 hover:bg-slate-700/50 rounded-xl transition-colors group"
          >
            <div class="w-12 h-12 bg-primary-500/10 rounded-xl flex items-center justify-center flex-shrink-0 group-hover:bg-primary-500/20 transition-colors">
              <FolderOpen class="w-6 h-6 text-primary-400" />
            </div>
            <div class="text-left">
              <p class="text-slate-100 font-medium">浏览服务器文件</p>
              <p class="text-sm text-slate-400">从服务器文件系统选择数据库</p>
            </div>
            <ChevronRight class="w-5 h-5 text-slate-500 ml-auto" />
          </button>

          <!-- Create New Database -->
          <button
            @click="openCreateModal"
            class="w-full flex items-center gap-4 p-4 bg-slate-700/30 hover:bg-slate-700/50 rounded-xl transition-colors group"
          >
            <div class="w-12 h-12 bg-emerald-500/10 rounded-xl flex items-center justify-center flex-shrink-0 group-hover:bg-emerald-500/20 transition-colors">
              <FolderPlus class="w-6 h-6 text-emerald-400" />
            </div>
            <div class="text-left">
              <p class="text-slate-100 font-medium">新建数据库</p>
              <p class="text-sm text-slate-400">在项目目录创建新的数据库文件</p>
            </div>
            <ChevronRight class="w-5 h-5 text-slate-500 ml-auto" />
          </button>

          <!-- Upload File (Drag & Drop) -->
          <div
            @drop.prevent="handleDrop"
            @dragover="handleDragOver"
            @dragleave="handleDragLeave"
            @click="handleBrowseClick"
            :class="[
              'relative border-2 border-dashed rounded-xl p-6 text-center cursor-pointer transition-all duration-200',
              isDragging
                ? 'border-primary-500 bg-primary-500/10'
                : 'border-slate-600 hover:border-slate-500 hover:bg-slate-700/30'
            ]"
          >
            <input
              ref="fileInput"
              type="file"
              accept=".db,.sqlite,.sqlite3"
              class="hidden"
              @change="handleFileUpload"
            />
            <Upload :class="['w-8 h-8 mx-auto mb-2', isDragging ? 'text-primary-400' : 'text-slate-500']" />
            <p class="text-slate-300 font-medium">
              {{ isDragging ? '放开以上传文件' : '拖拽或点击上传本地数据库' }}
            </p>
            <p class="text-xs text-slate-500 mt-1">支持 .db, .sqlite, .sqlite3 格式</p>
          </div>

          <!-- Existing Databases -->
          <div v-if="store.databases.length > 0" class="pt-2">
            <p class="text-sm text-slate-400 mb-2 px-1">已打开的数据库</p>
            <div class="space-y-2">
              <div
                v-for="db in store.databases"
                :key="db.id"
                @click="selectDatabase(db)"
                class="group flex items-center gap-3 px-4 py-3 rounded-lg bg-slate-700/30 hover:bg-slate-700/50 cursor-pointer transition-colors"
                :class="{ 'ring-1 ring-primary-500/50': db.active }"
              >
                <FileText class="w-5 h-5 text-slate-500 flex-shrink-0" />
                <div class="flex-1 min-w-0">
                  <p class="text-sm text-slate-200 truncate">{{ db.name }}</p>
                  <p class="text-xs text-slate-500 truncate" :title="db.path">{{ db.path }}</p>
                </div>
                <div class="flex items-center gap-2 flex-shrink-0">
                  <span v-if="db.active" class="px-1.5 py-0.5 bg-emerald-500/20 text-emerald-400 text-xs rounded">
                    活跃
                  </span>
                  <button
                    @click.stop="closeDatabase(db, $event)"
                    class="p-1.5 rounded text-slate-500 hover:text-red-400 hover:bg-red-500/10 opacity-0 group-hover:opacity-100 transition-all"
                    title="关闭"
                  >
                    <X class="w-4 h-4" />
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- Recent -->
          <div v-if="recentDatabases.length > 0" class="pt-2">
            <p class="text-sm text-slate-400 mb-2 px-1 flex items-center gap-1">
              <Clock class="w-3 h-3" />
              最近打开
            </p>
            <div class="space-y-1">
              <button
                v-for="path in recentDatabases.filter(p => !store.databases.find(d => d.path === p)).slice(0, 3)"
                :key="path"
                @click="openRecent(path)"
                class="w-full flex items-center gap-2 px-3 py-2 rounded-lg text-left hover:bg-slate-700/30 transition-colors"
              >
                <FileText class="w-4 h-4 text-slate-500" />
                <span class="text-sm text-slate-400 truncate">{{ path.split('/').pop() }}</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- File Browser Modal -->
    <Modal :show="showBrowseModal" title="选择数据库文件" size="large" @close="showBrowseModal = false">
      <div class="space-y-4">
        <!-- Current Path -->
        <div class="flex items-center gap-2 p-3 bg-slate-700/30 rounded-lg">
          <Folder class="w-4 h-4 text-slate-500 flex-shrink-0" />
          <span class="text-sm text-slate-300 truncate font-mono" :title="currentPath">
            {{ currentPath }}
          </span>
        </div>

        <!-- Navigation -->
        <div class="flex gap-2">
          <Button 
            size="small" 
            variant="secondary" 
            @click="goToParent" 
            :disabled="!parentPath || browsingLoading"
          >
            <ChevronLeft class="w-4 h-4" />
            上一级
          </Button>
        </div>

        <!-- File List -->
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
