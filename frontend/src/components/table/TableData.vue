<script setup>
import { ref, computed, watch, nextTick } from 'vue'
import { useDatabaseStore } from '../../stores/database'
import { useToastStore } from '../../stores/toast'
import { queryApi } from '../../api'
import { Plus, Trash2, Edit3, Check, X, ChevronLeft, ChevronRight, Download, Upload, Search, XCircle, Terminal, Loader, FileText, Database } from 'lucide-vue-next'
import Button from '../common/Button.vue'
import Modal from '../common/Modal.vue'
import Input from '../common/Input.vue'
import Select from '../common/Select.vue'
import ConfirmDialog from '../common/ConfirmDialog.vue'

const store = useDatabaseStore()
const toast = useToastStore()

const editingCell = ref(null)
const editValue = ref('')
const showAddRow = ref(false)
const newRow = ref({})
const firstInputRef = ref(null)
const showDeleteConfirm = ref(false)
const rowToDelete = ref(null)
const pageInput = ref('1')

const showImportModal = ref(false)
const showExportModal = ref(false)
const importFormat = ref('json')
const importJsonData = ref('')
const importCsvFile = ref(null)
const importResult = ref(null)
const importLoading = ref(false)

const exportFormat = ref('json')
const exportLoading = ref(false)

// SQL 执行
const showSqlPanel = ref(false)
const sqlInput = ref('')
const sqlResult = ref(null)
const sqlLoading = ref(false)

// 查询条件 - 支持多个条件（使用store中的状态）
const showQueryPanel = ref(false)

const operators = [
  { value: '=', label: '等于 (=)' },
  { value: '!=', label: '不等于 (!=)' },
  { value: '>', label: '大于 (>)' },
  { value: '>=', label: '大于等于 (>=)' },
  { value: '<', label: '小于 (<)' },
  { value: '<=', label: '小于等于 (<=)' },
  { value: 'LIKE', label: '包含 (LIKE)' },
  { value: 'NOT LIKE', label: '不包含 (NOT LIKE)' },
  { value: 'IS NULL', label: '为空 (IS NULL)' },
  { value: 'IS NOT NULL', label: '不为空 (IS NOT NULL)' }
]

const logicOptions = [
  { value: 'AND', label: '且 (AND)' },
  { value: 'OR', label: '或 (OR)' }
]

const totalPages = computed(() => Math.ceil(store.totalRows / store.pageSize) || 1)

const columnOptions = computed(() => {
  if (!store.currentSchema) return []
  return store.currentSchema.columns.map(col => ({
    value: col.name,
    label: col.name + (col.comment ? ` (${col.comment})` : '')
  }))
})

// 构建字段的 placeholder，优先使用备注
function getColumnPlaceholder(col) {
  if (col.comment) {
    return col.comment
  }
  if (col.nullable) return 'NULL'
  return ''
}

// 判断字段是否为数值类型
function isNumericType(colName) {
  const col = store.currentSchema?.columns.find(c => c.name === colName)
  if (!col) return false
  const type = col.type.toLowerCase()
  return type.includes('int') || type.includes('real') || type.includes('float') || 
         type.includes('double') || type.includes('numeric') || type.includes('decimal')
}

// 构建单个条件的 SQL
function buildConditionSQL(condition) {
  const { column, operator, value } = condition
  if (!column || !operator) return ''
  
  if (operator === 'IS NULL' || operator === 'IS NOT NULL') {
    return `${column} ${operator}`
  }
  
  if (operator === 'LIKE' || operator === 'NOT LIKE') {
    return `${column} ${operator} '%${value}%'`
  }
  
  // 数值类型不加引号，字符串加引号
  if (isNumericType(column)) {
    return `${column} ${operator} ${value}`
  }
  
  return `${column} ${operator} '${String(value).replace(/'/g, "''")}'`
}

// 构建完整的 WHERE 子句
const builtWhereClause = computed(() => {
  const validConditions = store.queryConditions
    .map(c => buildConditionSQL(c))
    .filter(sql => sql !== '')
  
  if (validConditions.length === 0) return ''
  return validConditions.join(` ${store.queryLogic} `)
})

// 添加条件
function addCondition() {
  store.queryConditions.push({
    column: '',
    operator: '=',
    value: ''
  })
}

// 删除条件
function removeCondition(index) {
  store.queryConditions.splice(index, 1)
}

function toggleQueryPanel() {
  showQueryPanel.value = !showQueryPanel.value
}

function applyQuery() {
  store.queryWhere = builtWhereClause.value
  store.executeQuery()
  // 不关闭面板，方便用户修改条件再次查询
}

function clearQuery() {
  store.clearQuery()
  showQueryPanel.value = false
}

function handleKeydown(e) {
  if (e.key === 'Enter' && editingCell.value) {
    saveEdit()
  } else if (e.key === 'Escape') {
    cancelEdit()
  }
}

// SQL 执行
function toggleSqlPanel() {
  showSqlPanel.value = !showSqlPanel.value
  if (showSqlPanel.value) {
    sqlInput.value = ''
    sqlResult.value = null
  }
}

async function executeSql() {
  if (!sqlInput.value.trim()) {
    toast.warning('请输入 SQL 语句')
    return
  }
  
  sqlLoading.value = true
  sqlResult.value = null
  
  try {
    const response = await queryApi.execute(sqlInput.value)
    sqlResult.value = response.data
    
    if (response.data.type === 'select') {
      toast.success(`查询完成，共 ${response.data.data.length} 条记录`)
    } else {
      toast.success(`执行成功，影响 ${response.data.rowsAffected} 行`)
    }
  } catch (error) {
    toast.error(error.response?.data?.error || 'SQL 执行失败')
    sqlResult.value = { error: error.response?.data?.error || 'SQL 执行失败' }
  } finally {
    sqlLoading.value = false
  }
}

function clearSqlResult() {
  sqlInput.value = ''
  sqlResult.value = null
}

// 监听 showAddRow，打开时聚焦第一个输入框
watch(showAddRow, (newVal) => {
  if (newVal) {
    newRow.value = {}
    nextTick(() => {
      firstInputRef.value?.focus()
    })
  }
})

function startEdit(row, column) {
  editingCell.value = { row, column }
  editValue.value = row[column] === null ? '' : String(row[column])
}

function cancelEdit() {
  editingCell.value = null
  editValue.value = ''
}

async function saveEdit() {
  if (!editingCell.value) return
  
  const { row, column } = editingCell.value
  const pkValue = row[store.primaryKey]
  const newData = { [column]: editValue.value === '' ? null : editValue.value }
  
  await store.updateRow(store.primaryKey, pkValue, newData)
  cancelEdit()
}

function confirmDelete(row) {
  rowToDelete.value = row
  showDeleteConfirm.value = true
}

async function deleteRow() {
  if (!rowToDelete.value || !store.primaryKey) return
  const pkValue = rowToDelete.value[store.primaryKey]
  await store.deleteRow(store.primaryKey, pkValue)
  rowToDelete.value = null
  showDeleteConfirm.value = false
}

async function addRow() {
  if (Object.keys(newRow.value).length === 0) {
    store.currentSchema.columns.forEach(col => {
      newRow.value[col.name] = ''
    })
  }
  await store.insertRow(newRow.value)
  newRow.value = {}
  showAddRow.value = false
}

function goToPage(page) {
  if (page < 1 || page > totalPages.value) return
  pageInput.value = String(page)
  store.loadPage(page)
}

function handlePageInput() {
  const page = parseInt(pageInput.value)
  if (!isNaN(page)) {
    goToPage(page)
  }
}

function exportCSV() {
  if (!store.currentData || !store.currentData.length) return
  
  const columns = Object.keys(store.currentData[0])
  const csv = [
    columns.join(','),
    ...store.currentData.map(row => 
      columns.map(col => {
        const val = row[col]
        if (val === null || val === undefined) return ''
        const str = String(val)
        if (str.includes(',') || str.includes('"') || str.includes('\n')) {
          return `"${str.replace(/"/g, '""')}"`
        }
        return str
      }).join(',')
    )
  ].join('\n')
  
  const blob = new Blob([csv], { type: 'text/csv' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${store.currentTable}.csv`
  a.click()
  URL.revokeObjectURL(url)
}

function openImportModal() {
  showImportModal.value = true
  importFormat.value = 'json'
  importJsonData.value = ''
  importCsvFile.value = null
  importResult.value = null
}

function handleCsvFileChange(event) {
  const file = event.target.files[0]
  if (file) {
    importCsvFile.value = file
  }
}

async function doImport() {
  if (importFormat.value === 'json') {
    if (!importJsonData.value.trim()) {
      toast.warning('请输入 JSON 数据')
      return
    }
    let parsed
    try {
      parsed = JSON.parse(importJsonData.value)
    } catch (e) {
      toast.error('JSON 格式错误')
      return
    }
    if (!Array.isArray(parsed) || parsed.length === 0) {
      toast.warning('JSON 数据必须是一个非空数组')
      return
    }
  } else if (importFormat.value === 'csv') {
    if (!importCsvFile.value) {
      toast.warning('请选择 CSV 文件')
      return
    }
  }

  importLoading.value = true
  importResult.value = null
  try {
    let result
    if (importFormat.value === 'json') {
      const parsed = JSON.parse(importJsonData.value)
      const response = await store.importTableData(store.currentTable, 'json', parsed)
      result = response
    } else {
      const response = await store.importTableData(store.currentTable, 'csv', importCsvFile.value)
      result = response
    }
    importResult.value = result
    if (result && (result.imported > 0 || result.message)) {
      toast.success(result.message || '导入完成')
    }
  } catch (error) {
    toast.error('导入失败: ' + (error.response?.data?.error || error.message))
  } finally {
    importLoading.value = false
  }
}

async function doExport(format) {
  exportLoading.value = true
  try {
    await store.exportTableData(store.currentTable, format)
  } catch (error) {
    toast.error('导出失败')
  } finally {
    exportLoading.value = false
  }
}

async function downloadDatabaseFile() {
  try {
    await store.downloadDatabase()
  } catch (error) {
    toast.error('下载失败')
  }
}
</script>

<template>
  <div class="h-full flex flex-col">
    <div class="flex items-center justify-between mb-4">
      <div>
        <h2 class="text-lg font-semibold text-slate-100">{{ store.currentTable }}</h2>
        <p class="text-sm text-slate-500">共 {{ store.totalRows }} 行</p>
      </div>
      <div class="flex items-center gap-2">
        <Button size="small" variant="secondary" @click="toggleQueryPanel">
          <Search class="w-4 h-4" />
          查询
          <span v-if="store.hasQueryCondition" class="ml-1 px-1.5 py-0.5 bg-primary-500/30 text-primary-400 text-xs rounded">
            有条件
          </span>
        </Button>
        <Button size="small" @click="showAddRow = true">
          <Plus class="w-4 h-4" />
          新增
        </Button>
        <Button size="small" variant="secondary" @click="openImportModal">
          <Upload class="w-4 h-4" />
          导入
        </Button>
        <div class="relative">
          <Button size="small" variant="secondary" @click="showExportModal = !showExportModal">
            <Download class="w-4 h-4" />
            导出
          </Button>
          <div v-if="showExportModal" class="absolute right-0 mt-1 w-48 bg-slate-800 border border-slate-700 rounded-lg shadow-lg z-50">
            <button @click="doExport('json'); showExportModal = false" class="w-full flex items-center gap-2 px-4 py-2 text-sm text-slate-300 hover:bg-slate-700/50 transition-colors rounded-t-lg">
              <FileText class="w-4 h-4" />
              导出为 JSON
            </button>
            <button @click="doExport('csv'); showExportModal = false" class="w-full flex items-center gap-2 px-4 py-2 text-sm text-slate-300 hover:bg-slate-700/50 transition-colors rounded-b-lg">
              <FileText class="w-4 h-4" />
              导出为 CSV
            </button>
          </div>
        </div>
        <Button size="small" variant="secondary" @click="downloadDatabaseFile">
          <Database class="w-4 h-4" />
          下载数据库
        </Button>
        <Button size="small" variant="secondary" @click="toggleSqlPanel">
          <Terminal class="w-4 h-4" />
          SQL
        </Button>
      </div>
    </div>

    <!-- 查询条件面板 -->
    <div v-if="showQueryPanel" class="mb-4 p-4 bg-slate-800/50 rounded-lg border border-slate-700">
      <div class="flex items-center gap-2 mb-3">
        <span class="text-sm text-slate-300">查询条件</span>
        <Select
          v-model="queryLogic"
          :options="logicOptions"
          class="w-24"
        />
        <Button size="small" variant="secondary" @click="addCondition">
          <Plus class="w-4 h-4" />
          添加条件
        </Button>
      </div>
      
      <!-- 条件列表 -->
      <div class="space-y-2">
        <div 
          v-for="(condition, index) in store.queryConditions" 
          :key="index"
          class="flex items-center gap-2 flex-wrap"
        >
          <!-- 逻辑连接符 -->
          <span v-if="index > 0" class="text-sm text-primary-400 font-medium min-w-[40px]">
            {{ store.queryLogic }}
          </span>
          
          <Select
            v-model="condition.column"
            :options="columnOptions"
            placeholder="字段"
            class="w-36"
          />
          <Select
            v-model="condition.operator"
            :options="operators"
            class="w-32"
          />
          <Input
            v-if="condition.operator !== 'IS NULL' && condition.operator !== 'IS NOT NULL'"
            v-model="condition.value"
            placeholder="值"
            class="flex-1 min-w-[120px]"
            @keyup.enter="applyQuery"
          />
          <button
            @click="removeCondition(index)"
            class="p-1.5 text-slate-500 hover:text-red-400 hover:bg-red-500/10 rounded transition-colors"
            title="删除条件"
          >
            <X class="w-4 h-4" />
          </button>
        </div>
        
        <!-- 空状态 -->
        <div v-if="store.queryConditions.length === 0" class="text-sm text-slate-500 py-2">
          点击"添加条件"开始查询
        </div>
      </div>
      
      <!-- SQL 预览和操作按钮 -->
      <div class="flex items-center justify-between mt-4 pt-3 border-t border-slate-700">
        <div v-if="builtWhereClause" class="text-xs text-slate-400 flex-1 mr-4">
          <span class="text-slate-500">WHERE:</span>
          <code class="bg-slate-700 px-1.5 py-0.5 rounded ml-1">{{ builtWhereClause }}</code>
        </div>
        <div v-else class="flex-1"></div>
        <div class="flex items-center gap-2">
          <Button size="small" variant="secondary" @click="clearQuery">
            <XCircle class="w-4 h-4" />
            清除
          </Button>
          <Button size="small" @click="applyQuery" :disabled="!builtWhereClause">
            <Search class="w-4 h-4" />
            查询
          </Button>
        </div>
      </div>
    </div>

    <div class="flex-1 overflow-auto rounded-lg border border-slate-700">
      <table class="w-full text-sm">
        <thead class="bg-slate-800 sticky top-0">
          <tr>
            <th
              v-for="column in store.currentSchema?.columns"
              :key="column.name"
              class="px-4 py-3 text-left font-medium text-slate-300 border-b border-slate-700 whitespace-nowrap"
            >
              <div class="flex items-center gap-2">
                <span>{{ column.name }}</span>
                <span class="text-xs text-slate-500">{{ column.type }}</span>
                <span v-if="column.primaryKey" class="px-1.5 py-0.5 bg-amber-500/20 text-amber-400 text-xs rounded">PK</span>
              </div>
            </th>
            <th class="px-4 py-3 w-20 border-b border-slate-700"></th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(row, index) in store.currentData"
            :key="index"
            class="group hover:bg-slate-800/50 transition-colors"
          >
            <td
              v-for="column in store.currentSchema?.columns"
              :key="column.name"
              class="px-4 py-2 border-b border-slate-700/50"
            >
              <div v-if="editingCell?.row === row && editingCell?.column === column.name" class="flex items-center gap-2">
                <input
                  v-model="editValue"
                  @keydown="handleKeydown"
                  class="flex-1 px-2 py-1 bg-slate-700 border border-slate-600 rounded text-slate-100 text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
                  autofocus
                />
                <button @click="saveEdit" class="p-1 text-emerald-400 hover:bg-emerald-500/20 rounded">
                  <Check class="w-4 h-4" />
                </button>
                <button @click="cancelEdit" class="p-1 text-slate-400 hover:bg-slate-700 rounded">
                  <X class="w-4 h-4" />
                </button>
              </div>
              <div
                v-else
                @dblclick="startEdit(row, column.name)"
                :class="[
                  'cursor-pointer px-2 py-1 -mx-2 rounded hover:bg-slate-700/50',
                  row[column.name] === null ? 'text-slate-500 italic' : 'text-slate-300'
                ]"
              >
                {{ row[column.name] === null ? 'NULL' : row[column.name] }}
              </div>
            </td>
            <td class="px-4 py-2 border-b border-slate-700/50">
              <button
                @click="confirmDelete(row)"
                class="p-1 text-red-400 opacity-0 group-hover:opacity-100 hover:bg-red-500/20 rounded transition-all"
              >
                <Trash2 class="w-4 h-4" />
              </button>
            </td>
          </tr>
          <tr v-if="store.currentData?.length === 0">
            <td :colspan="(store.currentSchema?.columns?.length || 0) + 1" class="px-4 py-12 text-center text-slate-500">
              暂无数据
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="flex items-center justify-between mt-4 text-sm">
      <p class="text-slate-500">
        显示 {{ (store.currentPage - 1) * store.pageSize + 1 }} - {{ Math.min(store.currentPage * store.pageSize, store.totalRows) }} / {{ store.totalRows }}
      </p>
      <div class="flex items-center gap-2">
        <Button
          size="small"
          variant="ghost"
          :disabled="store.currentPage <= 1"
          @click="goToPage(store.currentPage - 1)"
        >
          <ChevronLeft class="w-4 h-4" />
        </Button>
        <div class="flex items-center gap-1">
          <input
            v-model="pageInput"
            @change="handlePageInput"
            type="text"
            class="w-12 px-2 py-1 bg-slate-700 border border-slate-600 rounded text-center text-slate-200 text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
          />
          <span class="text-slate-500">/ {{ totalPages }}</span>
        </div>
        <Button
          size="small"
          variant="ghost"
          :disabled="store.currentPage >= totalPages"
          @click="goToPage(store.currentPage + 1)"
        >
          <ChevronRight class="w-4 h-4" />
        </Button>
      </div>
    </div>

    <Modal :show="showAddRow" title="新增数据" @close="showAddRow = false">
      <div class="space-y-3">
        <Input
          v-for="(column, index) in store.currentSchema?.columns"
          :key="column.name"
          :ref="index === 0 ? 'firstInputRef' : undefined"
          :label="column.name + (column.comment ? ` (${column.comment})` : '')"
          v-model="newRow[column.name]"
          :placeholder="getColumnPlaceholder(column)"
          :autofocus="index === 0"
          @keyup.enter="addRow"
        />
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <Button variant="secondary" @click="showAddRow = false">取消</Button>
          <Button @click="addRow">新增</Button>
        </div>
      </template>
    </Modal>

    <Modal :show="showImportModal" title="导入数据" size="large" @close="showImportModal = false">
      <div class="space-y-4">
        <div class="flex gap-4">
          <button 
            @click="importFormat = 'json'"
            :class="['flex-1 py-2 px-4 rounded-lg text-sm font-medium transition-colors', importFormat === 'json' ? 'bg-primary-500 text-white' : 'bg-slate-700 text-slate-300 hover:bg-slate-600']"
          >
            JSON
          </button>
          <button 
            @click="importFormat = 'csv'"
            :class="['flex-1 py-2 px-4 rounded-lg text-sm font-medium transition-colors', importFormat === 'csv' ? 'bg-primary-500 text-white' : 'bg-slate-700 text-slate-300 hover:bg-slate-600']"
          >
            CSV
          </button>
        </div>

        <div v-if="importFormat === 'json'">
          <label class="block text-sm font-medium text-slate-300 mb-2">JSON 数据</label>
          <textarea
            v-model="importJsonData"
            rows="8"
            placeholder='[{"name": "John", "age": 30}, {"name": "Jane", "age": 25}]'
            class="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-lg text-slate-200 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-primary-500 resize-none"
          ></textarea>
          <p class="text-xs text-slate-500 mt-1">JSON 格式：对象数组，字段名需与表字段匹配</p>
        </div>

        <div v-if="importFormat === 'csv'">
          <label class="block text-sm font-medium text-slate-300 mb-2">CSV 文件</label>
          <input
            type="file"
            accept=".csv"
            @change="handleCsvFileChange"
            class="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-lg text-slate-200 text-sm file:mr-4 file:py-1 file:px-4 file:rounded file:border-0 file:text-sm file:bg-slate-600 file:text-slate-200 file:hover:bg-slate-500"
          />
          <p v-if="importCsvFile" class="text-xs text-slate-400 mt-1">已选择: {{ importCsvFile.name }}</p>
          <p class="text-xs text-slate-500 mt-1">CSV 格式：第一行为字段名（表头），后续行为数据</p>
        </div>

        <div v-if="importResult" class="p-3 rounded-lg text-sm">
          <p :class="importResult.failed > 0 ? 'text-amber-400' : 'text-emerald-400'">
            {{ importResult.message }}
          </p>
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <Button variant="secondary" @click="showImportModal = false">取消</Button>
          <Button @click="doImport" :disabled="importLoading">
            {{ importLoading ? '导入中...' : '导入' }}
          </Button>
        </div>
      </template>
    </Modal>

    <!-- SQL 执行弹窗 - 全屏 -->
    <Teleport to="body">
      <div v-if="showSqlPanel" class="fixed inset-0 z-50 bg-slate-900/95 flex flex-col">
        <!-- 头部 -->
        <div class="flex items-center justify-between px-6 py-4 border-b border-slate-700">
          <h2 class="text-lg font-semibold text-slate-100">SQL 执行器</h2>
          <button
            @click="showSqlPanel = false"
            class="p-2 text-slate-400 hover:text-slate-100 hover:bg-slate-700 rounded-lg transition-colors"
          >
            <X class="w-5 h-5" />
          </button>
        </div>
        
        <!-- SQL 输入区 -->
        <div class="px-6 py-4 border-b border-slate-700 bg-slate-800/50">
          <div class="flex gap-2 mb-2">
            <Button @click="executeSql" :disabled="sqlLoading || !sqlInput.trim()">
              <Terminal class="w-4 h-4" />
              {{ sqlLoading ? '执行中...' : '执行 (Ctrl+Enter)' }}
            </Button>
            <Button variant="secondary" @click="clearSqlResult">
              <XCircle class="w-4 h-4" />
              清空
            </Button>
          </div>
          <textarea
            v-model="sqlInput"
            rows="3"
            placeholder="输入 SQL 语句，如: SELECT * FROM users WHERE id > 10"
            class="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-lg text-slate-100 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent resize-none"
            @keydown.ctrl.enter="executeSql"
          ></textarea>
          <p class="text-xs text-slate-500 mt-1">支持 SELECT、INSERT、UPDATE、DELETE、PRAGMA 等语句 (Ctrl+Enter 执行)</p>
        </div>
        
        <!-- 结果区 -->
        <div class="flex-1 overflow-hidden p-6">
          <div v-if="sqlLoading" class="flex items-center justify-center h-full">
            <Loader class="w-8 h-8 text-primary-400 animate-spin" />
          </div>
          
          <div v-else-if="sqlResult">
            <div v-if="sqlResult.error" class="h-full p-6 bg-red-500/10 border border-red-500/30 rounded-lg">
              <p class="text-red-400 text-lg font-medium">执行错误</p>
              <p class="text-red-300 text-sm mt-2 font-mono">{{ sqlResult.error }}</p>
            </div>
            
            <div v-else-if="sqlResult.type === 'select'" class="h-full flex flex-col">
              <div class="flex items-center justify-between mb-3">
                <span class="text-sm text-slate-400">查询结果 ({{ sqlResult.data.length }} 行)</span>
              </div>
              <div class="flex-1 overflow-auto border border-slate-700 rounded-lg">
                <table class="w-full text-sm">
                  <thead class="bg-slate-800 sticky top-0">
                    <tr>
                      <th
                        v-for="col in (sqlResult.data[0] ? Object.keys(sqlResult.data[0]) : [])"
                        :key="col"
                        class="px-4 py-3 text-left font-medium text-slate-300 border-b border-slate-700 whitespace-nowrap"
                      >
                        {{ col }}
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr
                      v-for="(row, index) in sqlResult.data"
                      :key="index"
                      class="hover:bg-slate-800/50"
                    >
                      <td
                        v-for="col in Object.keys(row)"
                        :key="col"
                        class="px-4 py-2 border-b border-slate-700/50 max-w-xs truncate"
                      >
                        <span :class="row[col] === null ? 'text-slate-500 italic' : 'text-slate-300'">
                          {{ row[col] === null ? 'NULL' : row[col] }}
                        </span>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
            
            <div v-else class="p-6 bg-emerald-500/10 border border-emerald-500/30 rounded-lg">
              <p class="text-emerald-400 text-lg">
                执行成功，影响 <span class="font-medium">{{ sqlResult.rowsAffected }}</span> 行
              </p>
            </div>
          </div>
          
          <div v-else class="flex items-center justify-center h-full text-slate-500">
            输入 SQL 语句并点击执行按钮查看结果
          </div>
        </div>
      </div>
    </Teleport>

    <ConfirmDialog
      :show="showDeleteConfirm"
      title="删除数据"
      message="确定要删除这条数据吗？此操作不可撤销。"
      @confirm="deleteRow"
      @cancel="showDeleteConfirm = false"
    />
  </div>
</template>
