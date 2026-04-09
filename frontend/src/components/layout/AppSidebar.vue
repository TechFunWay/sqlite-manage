<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useDatabaseStore } from '../../stores/database'
import { useToastStore } from '../../stores/toast'
import { recentApi } from '../../api'
import { 
  Table as TableIcon, Search, Plus, Trash2, ChevronRight, ChevronDown, 
  HardDrive, Database, Loader2, Clock, X, FolderOpen, FolderPlus 
} from 'lucide-vue-next'
import Button from '../common/Button.vue'
import ConfirmDialog from '../common/ConfirmDialog.vue'

const emit = defineEmits(['create-table', 'select-database', 'create-database'])

const store = useDatabaseStore()
const toast = useToastStore()

const searchQuery = ref('')
const showDeleteConfirm = ref(false)
const tableToDelete = ref(null)
const expandedDatabases = ref(new Set())
const loadingDb = ref(null)
const showMenu = ref(false)

// Recent databases from backend
const recentDatabases = ref([])

// Load recent databases from backend
async function loadRecentDatabases() {
  try {
    const response = await recentApi.get()
    recentDatabases.value = response.data || []
  } catch (error) {
    console.error('Failed to load recent databases', error)
    recentDatabases.value = []
  }
}

// Add to recent databases
async function addToRecent(path, name) {
  try {
    await recentApi.add({ path, name })
    await loadRecentDatabases()
  } catch (error) {
    console.error('Failed to add recent database', error)
  }
}

// Auto-expand active database
watch(() => store.databaseInfo?.id, (newId) => {
  if (newId) {
    expandedDatabases.value.add(newId)
  }
}, { immediate: true })

// Filter databases
const filteredDatabases = computed(() => {
  if (!searchQuery.value) return store.databases
  const query = searchQuery.value.toLowerCase()
  return store.databases.filter(db => 
    db.name.toLowerCase().includes(query)
  )
})

// Filter recent databases (exclude already opened)
const filteredRecentDatabases = computed(() => {
  const openPaths = store.databases.map(db => db.path)
  let recent = recentDatabases.value.filter(r => !openPaths.includes(r.path))
  
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    recent = recent.filter(r => r.name.toLowerCase().includes(query) || r.path.toLowerCase().includes(query))
  }
  
  return recent.slice(0, 10)
})

onMounted(() => {
  loadRecentDatabases()
})

async function toggleDatabase(dbId) {
  if (expandedDatabases.value.has(dbId)) {
    expandedDatabases.value.delete(dbId)
  } else {
    expandedDatabases.value.add(dbId)
    await loadDatabaseTables(dbId)
  }
}

async function loadDatabaseTables(dbId) {
  const db = store.databases.find(d => d.id === dbId)
  if (db && !db.tables) {
    loadingDb.value = dbId
    try {
      const wasActive = db.active
      if (!wasActive) {
        await store.activateDatabase(dbId)
      }
      db.tables = [...store.tables]
      if (!wasActive) {
        const activeDb = store.databases.find(d => d.active && d.id !== dbId)
        if (activeDb) {
          await store.activateDatabase(activeDb.id)
        }
      }
    } catch (error) {
      console.error('Failed to load tables', error)
    } finally {
      loadingDb.value = null
    }
  }
}

function selectTable(table, db) {
  if (!db.active) {
    store.activateDatabase(db.id).then(() => {
      store.selectTable(table.name)
    })
  } else {
    store.selectTable(table.name)
  }
}

async function openRecentDatabase(record) {
  const success = await store.openDatabase(record.path)
  if (success) {
    await addToRecent(record.path, record.name)
  }
}

function handleDropTable(table) {
  tableToDelete.value = table.name
  showDeleteConfirm.value = true
}

async function confirmDelete() {
  if (tableToDelete.value) {
    await store.dropTable(tableToDelete.value)
    tableToDelete.value = null
  }
  showDeleteConfirm.value = false
}

function getTables(db) {
  return db.tables || []
}

async function closeDatabase(db, event) {
  event.stopPropagation()
  await store.closeDatabase(db.id)
}

function formatPath(path) {
  if (!path) return ''
  const parts = path.split('/')
  return parts.length > 3 ? '...' + parts.slice(-2).join('/') : path
}

function handleOpenDatabase() {
  showMenu.value = false
  emit('select-database')
}

function handleCreateDatabase() {
  showMenu.value = false
  emit('create-database')
}

// Expose addToRecent for parent component
defineExpose({ addToRecent, loadRecentDatabases })
</script>

<template>
  <aside class="w-72 bg-slate-800/50 border-r border-slate-700 flex flex-col overflow-hidden">
    <div class="p-3 border-b border-slate-700">
      <div class="relative">
        <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500" />
        <input
          v-model="searchQuery"
          type="text"
          placeholder="搜索数据库..."
          class="w-full pl-9 pr-3 py-2 bg-slate-700/50 border border-slate-600 rounded-lg text-sm text-slate-200 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-primary-500/50 focus:border-primary-500"
        />
      </div>
    </div>

    <div class="flex-1 overflow-auto">
      <!-- Active Databases -->
      <div v-if="filteredDatabases.length > 0" class="p-3">
        <p class="text-xs text-slate-500 mb-2 px-1 flex items-center gap-1">
          <Database class="w-3 h-3" />
          已打开的数据库
        </p>
        <div class="space-y-1">
          <div
            v-for="db in filteredDatabases"
            :key="db.id"
          >
            <div
              @click="toggleDatabase(db.id)"
              :class="[
                'group flex items-center gap-2 px-3 py-2 rounded-lg cursor-pointer transition-all duration-200',
                db.active
                  ? 'bg-primary-500/10 text-primary-400'
                  : 'text-slate-300 hover:bg-slate-700/50'
              ]"
            >
              <ChevronRight 
                class="w-4 h-4 flex-shrink-0 transition-transform" 
                :class="{ 'rotate-90': expandedDatabases.has(db.id) }"
              />
              <Database class="w-4 h-4 flex-shrink-0" />
              <div class="flex-1 min-w-0">
                <p class="text-sm font-medium truncate">{{ db.name }}</p>
              </div>
              <Loader2 v-if="loadingDb === db.id" class="w-4 h-4 animate-spin" />
              <span v-else-if="db.active" class="px-1.5 py-0.5 bg-emerald-500/20 text-emerald-400 text-xs rounded">
                活跃
              </span>
              <button
                v-if="!loadingDb || loadingDb !== db.id"
                @click.stop="closeDatabase(db, $event)"
                class="p-1 rounded text-slate-500 hover:text-red-400 hover:bg-red-500/10 opacity-0 group-hover:opacity-100 transition-all"
                title="关闭"
              >
                <X class="w-3 h-3" />
              </button>
            </div>

            <!-- Tables (expanded) -->
            <div v-if="expandedDatabases.has(db.id)" class="ml-4 mt-1 space-y-0.5">
              <div
                v-for="table in getTables(db)"
                :key="table.name"
                @click="selectTable(table, db)"
                :class="[
                  'group flex items-center gap-2 px-3 py-1.5 rounded-lg cursor-pointer transition-all duration-200 border-l-2 border-slate-700 ml-2',
                  store.currentTable === table.name && db.active
                    ? 'bg-primary-500/10 text-primary-400 border-primary-500'
                    : 'text-slate-400 hover:text-slate-300 hover:bg-slate-700/30'
                ]"
              >
                <TableIcon class="w-3.5 h-3.5 flex-shrink-0" />
                <div class="flex-1 min-w-0">
                  <p class="text-sm truncate">{{ table.name }}</p>
                </div>
                <span v-if="table.rows !== undefined" class="text-xs text-slate-500">{{ table.rows }}</span>
                <button
                  v-if="db.active"
                  @click.stop="handleDropTable(table)"
                  class="p-1 rounded opacity-0 group-hover:opacity-100 hover:bg-red-500/20 hover:text-red-400 transition-all"
                  title="删除表"
                >
                  <Trash2 class="w-3 h-3" />
                </button>
              </div>
              
              <div v-if="getTables(db).length === 0" class="ml-4 py-2 text-xs text-slate-500">
                暂无表
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Recent Databases -->
      <div v-if="filteredRecentDatabases.length > 0" class="p-3 border-t border-slate-700">
        <p class="text-xs text-slate-500 mb-2 px-1 flex items-center gap-1">
          <Clock class="w-3 h-3" />
          最近打开
        </p>
        <div class="space-y-1">
          <div
            v-for="record in filteredRecentDatabases"
            :key="record.id"
            @click="openRecentDatabase(record)"
            class="group flex items-center gap-2 px-3 py-2 rounded-lg cursor-pointer text-slate-400 hover:text-slate-300 hover:bg-slate-700/30 transition-colors"
            :title="record.path"
          >
            <Database class="w-4 h-4 flex-shrink-0 text-slate-500" />
            <div class="flex-1 min-w-0">
              <p class="text-sm truncate">{{ record.name }}</p>
              <p class="text-xs text-slate-600 truncate">{{ formatPath(record.path) }}</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Empty State -->
      <div v-if="filteredDatabases.length === 0 && filteredRecentDatabases.length === 0" class="p-8 text-center text-slate-500">
        <Database class="w-10 h-10 mx-auto mb-3 opacity-50" />
        <p class="text-sm">暂无数据库</p>
        <p class="text-xs mt-1">点击下方按钮打开或创建</p>
      </div>
    </div>

    <div class="p-3 border-t border-slate-700 space-y-2">
      <!-- 打开/新建数据库按钮 -->
      <div class="relative">
        <Button @click="showMenu = !showMenu" size="small" fullWidth variant="secondary">
          <HardDrive class="w-4 h-4" />
          打开/新建数据库
          <ChevronDown class="w-3 h-3 ml-auto transition-transform" :class="{ 'rotate-180': showMenu }" />
        </Button>
        
        <!-- 下拉菜单 -->
        <Transition name="dropdown">
          <div 
            v-if="showMenu" 
            class="absolute bottom-full left-0 right-0 mb-1 bg-slate-700 rounded-lg border border-slate-600 shadow-xl overflow-hidden z-50"
          >
            <button
              @click="handleOpenDatabase"
              class="w-full flex items-center gap-2 px-4 py-2.5 text-sm text-slate-200 hover:bg-slate-600 transition-colors"
            >
              <FolderOpen class="w-4 h-4 text-primary-400" />
              浏览/打开数据库
            </button>
            <button
              @click="handleCreateDatabase"
              class="w-full flex items-center gap-2 px-4 py-2.5 text-sm text-slate-200 hover:bg-slate-600 transition-colors"
            >
              <FolderPlus class="w-4 h-4 text-emerald-400" />
              新建数据库
            </button>
          </div>
        </Transition>
      </div>
      
      <Button @click="emit('create-table')" size="small" fullWidth :disabled="!store.databaseInfo">
        <Plus class="w-4 h-4" />
        新建表
      </Button>
    </div>

    <ConfirmDialog
      :show="showDeleteConfirm"
      title="删除表"
      :message="`确定要删除表 '${tableToDelete}' 吗？此操作不可撤销。`"
      confirm-text="删除"
      @confirm="confirmDelete"
      @cancel="showDeleteConfirm = false"
    />
  </aside>
</template>

<style scoped>
.dropdown-enter-active,
.dropdown-leave-active {
  transition: all 0.15s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: translateY(10px);
}
</style>
