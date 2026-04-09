<script setup>
import { ref, computed, watch, nextTick } from 'vue'
import { useDatabaseStore } from '../../stores/database'
import { Plus, Trash2, Edit3, Check, X, ChevronLeft, ChevronRight, Download, Upload } from 'lucide-vue-next'
import Button from '../common/Button.vue'
import Modal from '../common/Modal.vue'
import Input from '../common/Input.vue'
import Select from '../common/Select.vue'
import ConfirmDialog from '../common/ConfirmDialog.vue'

const store = useDatabaseStore()

const editingCell = ref(null)
const editValue = ref('')
const showAddRow = ref(false)
const newRow = ref({})
const firstInputRef = ref(null)
const showDeleteConfirm = ref(false)
const rowToDelete = ref(null)
const pageInput = ref('1')

const totalPages = computed(() => Math.ceil(store.totalRows / store.pageSize) || 1)

// 监听 showAddRow，打开时聚焦第一个输入框
watch(showAddRow, (newVal) => {
  if (newVal) {
    newRow.value = {}
    nextTick(() => {
      firstInputRef.value?.focus()
    })
  }
})

const columnOptions = computed(() => {
  if (!store.currentSchema) return []
  return store.currentSchema.columns.map(col => ({
    value: col.name,
    label: col.name
  }))
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

function handleKeydown(e) {
  if (e.key === 'Enter') {
    saveEdit()
  } else if (e.key === 'Escape') {
    cancelEdit()
  }
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
</script>

<template>
  <div class="h-full flex flex-col">
    <div class="flex items-center justify-between mb-4">
      <div>
        <h2 class="text-lg font-semibold text-slate-100">{{ store.currentTable }}</h2>
        <p class="text-sm text-slate-500">共 {{ store.totalRows }} 行</p>
      </div>
      <div class="flex items-center gap-2">
        <Button size="small" @click="showAddRow = true">
          <Plus class="w-4 h-4" />
          新增
        </Button>
        <Button size="small" variant="secondary" @click="exportCSV">
          <Download class="w-4 h-4" />
          导出
        </Button>
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
          :label="column.name + ' (' + column.type + ')'"
          v-model="newRow[column.name]"
          :placeholder="column.nullable ? 'NULL' : ''"
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

    <ConfirmDialog
      :show="showDeleteConfirm"
      title="删除数据"
      message="确定要删除这条数据吗？此操作不可撤销。"
      @confirm="deleteRow"
      @cancel="showDeleteConfirm = false"
    />
  </div>
</template>
