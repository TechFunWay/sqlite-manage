<script setup>
import { ref, nextTick } from 'vue'
import { useDatabaseStore } from '../../stores/database'
import { Plus, Trash2, Key, Hash, Type, ToggleLeft, List, Calendar, Check, Edit3 } from 'lucide-vue-next'
import Button from '../common/Button.vue'
import Modal from '../common/Modal.vue'
import Input from '../common/Input.vue'
import Select from '../common/Select.vue'
import ConfirmDialog from '../common/ConfirmDialog.vue'

const store = useDatabaseStore()

const showAddColumn = ref(false)
const showAddIndex = ref(false)
const showDeleteConfirm = ref(false)
const itemToDelete = ref(null)
const deleteType = ref('column')

const newColumn = ref({
  name: '',
  type: 'TEXT',
  nullable: true,
  primaryKey: false,
  defaultValue: null
})
const columnNameRef = ref(null)

// 编辑字段
const showEditColumn = ref(false)
const editingOldName = ref('')
const editingPrimaryKey = ref(false)
const editColumn = ref({
  name: '',
  type: 'TEXT',
  nullable: true,
  defaultValue: null,
  comment: ''
})
const editColumnNameRef = ref(null)

const newIndex = ref({
  name: '',
  columns: [],
  unique: false
})
const indexNameRef = ref(null)

const sqlTypes = [
  { value: 'TEXT', label: 'TEXT' },
  { value: 'INTEGER', label: 'INTEGER' },
  { value: 'REAL', label: 'REAL' },
  { value: 'BLOB', label: 'BLOB' },
  { value: 'NUMERIC', label: 'NUMERIC' },
  { value: 'VARCHAR(255)', label: 'VARCHAR(255)' },
  { value: 'BOOLEAN', label: 'BOOLEAN' },
  { value: 'DATE', label: 'DATE' },
  { value: 'DATETIME', label: 'DATETIME' },
  { value: 'TIMESTAMP', label: 'TIMESTAMP' },
  { value: 'INT', label: 'INT' },
  { value: 'BIGINT', label: 'BIGINT' },
  { value: 'FLOAT', label: 'FLOAT' },
  { value: 'DOUBLE', label: 'DOUBLE' },
  { value: 'DECIMAL(10,2)', label: 'DECIMAL(10,2)' }
]

function getTypeIcon(type) {
  const upperType = type.toUpperCase()
  if (upperType.includes('INT')) return Hash
  if (upperType.includes('TEXT') || upperType.includes('VARCHAR') || upperType.includes('CHAR')) return Type
  if (upperType.includes('REAL') || upperType.includes('FLOAT') || upperType.includes('DOUBLE') || upperType.includes('DECIMAL')) return List
  if (upperType.includes('DATE') || upperType.includes('TIME')) return Calendar
  if (upperType.includes('BOOL')) return ToggleLeft
  return List
}

function openAddColumn() {
  newColumn.value = {
    name: '',
    type: 'TEXT',
    nullable: true,
    primaryKey: false,
    defaultValue: null
  }
  showAddColumn.value = true
  nextTick(() => {
    columnNameRef.value?.focus()
  })
}

async function addColumn() {
  const success = await store.addColumn(newColumn.value)
  if (success) {
    showAddColumn.value = false
  }
}

function openEditColumn(column) {
  editingOldName.value = column.name
  editingPrimaryKey.value = column.primaryKey
  editColumn.value = {
    name: column.name,
    type: column.type,
    nullable: column.nullable,
    defaultValue: column.defaultValue ?? null,
    comment: column.comment || ''
  }
  showEditColumn.value = true
  nextTick(() => {
    editColumnNameRef.value?.focus()
  })
}

async function saveEditColumn() {
  if (!editColumn.value.name || !editColumn.value.type) return
  const payload = {
    name: editColumn.value.name,
    type: editColumn.value.type,
    nullable: editColumn.value.nullable,
    primaryKey: editingPrimaryKey.value,
    defaultValue: editColumn.value.defaultValue === '' ? null : editColumn.value.defaultValue,
    comment: editColumn.value.comment || null
  }
  const success = await store.updateColumn(editingOldName.value, payload)
  if (success) {
    showEditColumn.value = false
  }
}

function openAddIndex() {
  newIndex.value = {
    name: `${store.currentTable}_idx_${Date.now()}`,
    columns: [],
    unique: false
  }
  showAddIndex.value = true
  nextTick(() => {
    indexNameRef.value?.focus()
  })
}

async function addIndex() {
  const success = await store.createIndex(
    newIndex.value.name,
    newIndex.value.columns,
    newIndex.value.unique
  )
  if (success) {
    showAddIndex.value = false
  }
}

function confirmDelete(type, item) {
  deleteType.value = type
  itemToDelete.value = item
  showDeleteConfirm.value = true
}

async function deleteItem() {
  if (deleteType.value === 'column') {
    await store.dropColumn(itemToDelete.value)
  } else if (deleteType.value === 'index') {
    await store.dropIndex(itemToDelete.value)
  }
  itemToDelete.value = null
  showDeleteConfirm.value = false
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h2 class="text-lg font-semibold text-slate-100">表结构</h2>
      <Button size="small" @click="openAddColumn">
        <Plus class="w-4 h-4" />
        添加字段
      </Button>
    </div>

    <div class="bg-slate-800/50 rounded-lg border border-slate-700 overflow-x-auto">
      <table class="w-full text-sm min-w-[640px]">
        <thead class="bg-slate-800">
          <tr>
            <th class="px-4 py-3 text-left font-medium text-slate-300">字段名</th>
            <th class="px-4 py-3 text-left font-medium text-slate-300">类型</th>
            <th class="px-4 py-3 text-left font-medium text-slate-300">可空</th>
            <th class="px-4 py-3 text-left font-medium text-slate-300">默认值</th>
            <th class="px-4 py-3 text-left font-medium text-slate-300">键</th>
            <th class="px-4 py-3 text-left font-medium text-slate-300">备注</th>
            <th class="px-4 py-3 w-24"></th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="column in store.currentSchema?.columns"
            :key="column.name"
            class="group hover:bg-slate-800/30 transition-colors"
          >
            <td class="px-4 py-3">
              <div class="flex items-center gap-2">
                <component :is="getTypeIcon(column.type)" class="w-4 h-4 text-slate-500" />
                <span class="font-medium text-slate-200">{{ column.name }}</span>
              </div>
            </td>
            <td class="px-4 py-3 text-slate-400 font-mono">{{ column.type }}</td>
            <td class="px-4 py-3">
              <span :class="column.nullable ? 'text-emerald-400' : 'text-red-400'">
                {{ column.nullable ? 'YES' : 'NO' }}
              </span>
            </td>
            <td class="px-4 py-3 text-slate-400 font-mono">
              {{ column.defaultValue || '-' }}
            </td>
            <td class="px-4 py-3">
              <span v-if="column.primaryKey" class="px-2 py-0.5 bg-amber-500/20 text-amber-400 text-xs rounded font-medium">
                PRIMARY KEY
              </span>
              <span v-else class="text-slate-600">-</span>
            </td>
            <td class="px-4 py-3 text-slate-400 max-w-[220px]">
              <span v-if="column.comment" class="block truncate" :title="column.comment">{{ column.comment }}</span>
              <span v-else class="text-slate-600">-</span>
            </td>
            <td class="px-4 py-3">
              <div class="flex items-center gap-1">
                <button
                  @click="openEditColumn(column)"
                  class="p-1 text-slate-400 opacity-100 lg:opacity-0 lg:group-hover:opacity-100 hover:text-primary-400 hover:bg-primary-500/10 rounded transition-all"
                  title="修改字段"
                >
                  <Edit3 class="w-4 h-4" />
                </button>
                <button
                  @click="confirmDelete('column', column.name)"
                  class="p-1 text-red-400 opacity-100 lg:opacity-0 lg:group-hover:opacity-100 hover:bg-red-500/20 rounded transition-all"
                  title="删除字段"
                >
                  <Trash2 class="w-4 h-4" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div>
      <div class="flex items-center justify-between mb-3">
        <h3 class="text-md font-semibold text-slate-100">索引</h3>
        <Button size="small" variant="secondary" @click="openAddIndex">
          <Plus class="w-4 h-4" />
          添加索引
        </Button>
      </div>

      <div v-if="store.currentSchema?.indexes?.length > 0" class="space-y-2">
        <div
          v-for="index in store.currentSchema.indexes"
          :key="index.name"
          class="group flex items-center justify-between px-4 py-3 bg-slate-800/50 rounded-lg border border-slate-700"
        >
          <div class="flex items-center gap-3">
            <Key class="w-4 h-4 text-slate-500" />
            <div>
              <p class="text-slate-200 font-medium">{{ index.name }}</p>
              <p class="text-xs text-slate-500">
                {{ index.columns.join(', ') }}
                <span v-if="index.unique" class="text-amber-400 ml-1">UNIQUE</span>
              </p>
            </div>
          </div>
          <button
            @click="confirmDelete('index', index.name)"
            class="p-1 text-red-400 opacity-100 lg:opacity-0 lg:group-hover:opacity-100 hover:bg-red-500/20 rounded transition-all"
          >
            <Trash2 class="w-4 h-4" />
          </button>
        </div>
      </div>
      <div v-else class="text-center py-8 text-slate-500 text-sm bg-slate-800/30 rounded-lg border border-slate-700 border-dashed">
        暂无索引
      </div>
    </div>

    <Modal :show="showAddColumn" title="添加字段" @close="showAddColumn = false">
      <div class="space-y-4">
        <Input
          ref="columnNameRef"
          v-model="newColumn.name"
          label="字段名"
          placeholder="请输入字段名"
          required
          autofocus
          @keyup.enter="addColumn"
        />
        <Select
          v-model="newColumn.type"
          label="数据类型"
          :options="sqlTypes"
          required
        />
        <div class="flex items-center gap-4">
          <label class="flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              v-model="newColumn.nullable"
              class="w-4 h-4 rounded border-slate-600 bg-slate-700 text-primary-500 focus:ring-primary-500"
            />
            <span class="text-sm text-slate-300">允许为空</span>
          </label>
          <label class="flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              v-model="newColumn.primaryKey"
              class="w-4 h-4 rounded border-slate-600 bg-slate-700 text-primary-500 focus:ring-primary-500"
            />
            <span class="text-sm text-slate-300">主键</span>
          </label>
        </div>
        <Input
          v-model="newColumn.defaultValue"
          label="默认值 (可选)"
          placeholder="留空表示无默认值"
        />
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <Button variant="secondary" @click="showAddColumn = false">取消</Button>
          <Button @click="addColumn" :disabled="!newColumn.name || !newColumn.type">添加</Button>
        </div>
      </template>
    </Modal>

    <Modal :show="showEditColumn" title="修改字段" @close="showEditColumn = false">
      <div class="space-y-4">
        <Input
          ref="editColumnNameRef"
          v-model="editColumn.name"
          label="字段名"
          placeholder="请输入字段名"
          required
          @keyup.enter="saveEditColumn"
        />
        <Input
          v-model="editColumn.type"
          label="数据类型"
          placeholder="如 TEXT、INTEGER、VARCHAR(255)"
          required
        />
        <Input
          v-model="editColumn.comment"
          label="备注 (可选)"
          placeholder="字段说明"
        />
        <div class="flex items-center gap-4">
          <label class="flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              v-model="editColumn.nullable"
              :disabled="editingPrimaryKey"
              class="w-4 h-4 rounded border-slate-600 bg-slate-700 text-primary-500 focus:ring-primary-500 disabled:opacity-50"
            />
            <span class="text-sm text-slate-300">允许为空</span>
          </label>
          <span v-if="editingPrimaryKey" class="px-2 py-0.5 bg-amber-500/20 text-amber-400 text-xs rounded font-medium">
            主键（不可修改）
          </span>
        </div>
        <Input
          v-model="editColumn.defaultValue"
          label="默认值 (可选)"
          placeholder="留空表示无默认值"
        />
        <p class="text-xs text-slate-500">
          提示：修改类型/可空/默认值会重建表（数据与索引会保留）；仅改名或备注则不会重建。
        </p>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <Button variant="secondary" @click="showEditColumn = false">取消</Button>
          <Button @click="saveEditColumn" :disabled="!editColumn.name || !editColumn.type">保存</Button>
        </div>
      </template>
    </Modal>

    <Modal :show="showAddIndex" title="添加索引" @close="showAddIndex = false">
      <div class="space-y-4">
        <Input
          ref="indexNameRef"
          v-model="newIndex.name"
          label="索引名"
          placeholder="请输入索引名"
          required
          autofocus
          @keyup.enter="addIndex"
        />
        <div class="space-y-2">
          <label class="block text-sm font-medium text-slate-300">选择字段</label>
          <div class="flex flex-wrap gap-2">
            <label
              v-for="column in store.currentSchema?.columns"
              :key="column.name"
              class="flex items-center gap-2 px-3 py-2 bg-slate-700/50 rounded-lg cursor-pointer hover:bg-slate-700"
            >
              <input
                type="checkbox"
                :value="column.name"
                v-model="newIndex.columns"
                class="w-4 h-4 rounded border-slate-600 bg-slate-700 text-primary-500 focus:ring-primary-500"
              />
              <span class="text-sm text-slate-300">{{ column.name }}</span>
            </label>
          </div>
        </div>
        <label class="flex items-center gap-2 cursor-pointer">
          <input
            type="checkbox"
            v-model="newIndex.unique"
            class="w-4 h-4 rounded border-slate-600 bg-slate-700 text-primary-500 focus:ring-primary-500"
          />
          <span class="text-sm text-slate-300">唯一索引</span>
        </label>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <Button variant="secondary" @click="showAddIndex = false">取消</Button>
          <Button @click="addIndex" :disabled="!newIndex.name || newIndex.columns.length === 0">添加</Button>
        </div>
      </template>
    </Modal>

    <ConfirmDialog
      :show="showDeleteConfirm"
      :title="deleteType === 'column' ? '删除字段' : '删除索引'"
      :message="deleteType === 'column' 
        ? `确定要删除字段 '${itemToDelete}' 吗？这可能会导致数据丢失。` 
        : `确定要删除索引 '${itemToDelete}' 吗？`"
      @confirm="deleteItem"
      @cancel="showDeleteConfirm = false"
    />
  </div>
</template>
