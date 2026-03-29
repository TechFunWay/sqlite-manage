<script setup>
import { ref, watch } from 'vue'
import { Plus, Trash2 } from 'lucide-vue-next'
import Modal from '../common/Modal.vue'
import Input from '../common/Input.vue'
import Select from '../common/Select.vue'
import Button from '../common/Button.vue'
import { useDatabaseStore } from '../../stores/database'

const store = useDatabaseStore()

const props = defineProps({
  show: Boolean
})

const emit = defineEmits(['close'])

const tableName = ref('')
const columns = ref([
  { name: '', type: 'INTEGER', nullable: false, primaryKey: true, defaultValue: null }
])

const sqlTypes = [
  { value: 'TEXT', label: 'TEXT' },
  { value: 'INTEGER', label: 'INTEGER' },
  { value: 'REAL', label: 'REAL' },
  { value: 'BLOB', label: 'BLOB' },
  { value: 'VARCHAR(255)', label: 'VARCHAR(255)' },
  { value: 'BOOLEAN', label: 'BOOLEAN' },
  { value: 'DATE', label: 'DATE' },
  { value: 'DATETIME', label: 'DATETIME' },
  { value: 'INT', label: 'INT' },
  { value: 'BIGINT', label: 'BIGINT' },
  { value: 'FLOAT', label: 'FLOAT' },
  { value: 'DOUBLE', label: 'DOUBLE' },
  { value: 'DECIMAL(10,2)', label: 'DECIMAL(10,2)' }
]

watch(() => props.show, (newVal) => {
  if (newVal) {
    tableName.value = ''
    columns.value = [{ name: '', type: 'INTEGER', nullable: false, primaryKey: true, defaultValue: null }]
  }
})

function addColumn() {
  columns.value.push({ name: '', type: 'TEXT', nullable: true, primaryKey: false, defaultValue: null })
}

function removeColumn(index) {
  if (columns.value.length > 1) {
    columns.value.splice(index, 1)
  }
}

function setPrimaryKey(index) {
  columns.value.forEach((col, i) => {
    col.primaryKey = i === index
  })
}

async function createTable() {
  const validColumns = columns.value.filter(col => col.name.trim())
  
  if (!tableName.value.trim() || validColumns.length === 0) return
  
  const success = await store.createTable(tableName.value.trim(), validColumns)
  
  if (success) {
    emit('close')
  }
}
</script>

<template>
  <Modal :show="show" title="创建表" size="large" @close="emit('close')">
    <div class="space-y-4">
      <Input
        v-model="tableName"
        label="表名"
        placeholder="请输入表名"
        required
      />

      <div>
        <div class="flex items-center justify-between mb-2">
          <label class="block text-sm font-medium text-slate-300">字段</label>
          <Button size="small" variant="secondary" @click="addColumn">
            <Plus class="w-4 h-4" />
            添加字段
          </Button>
        </div>

        <div class="space-y-2 max-h-80 overflow-auto">
          <div
            v-for="(column, index) in columns"
            :key="index"
            class="flex items-center gap-2 p-3 bg-slate-700/30 rounded-lg"
          >
            <input
              v-model="column.name"
              placeholder="字段名"
              class="flex-1 px-3 py-2 bg-slate-700 border border-slate-600 rounded-lg text-slate-200 text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
            />
            <select
              v-model="column.type"
              class="px-3 py-2 bg-slate-700 border border-slate-600 rounded-lg text-slate-200 text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
            >
              <option v-for="t in sqlTypes" :key="t.value" :value="t.value">{{ t.label }}</option>
            </select>
            <label class="flex items-center gap-1 px-2 py-1 text-xs cursor-pointer" title="主键">
              <input
                type="checkbox"
                :checked="column.primaryKey"
                @change="setPrimaryKey(index)"
                class="w-3.5 h-3.5 rounded border-slate-600 bg-slate-700 text-amber-500 focus:ring-amber-500"
              />
              <span class="text-slate-400">PK</span>
            </label>
            <label class="flex items-center gap-1 px-2 py-1 text-xs cursor-pointer" title="可空">
              <input
                type="checkbox"
                v-model="column.nullable"
                class="w-3.5 h-3.5 rounded border-slate-600 bg-slate-700 text-emerald-500 focus:ring-emerald-500"
              />
              <span class="text-slate-400">NULL</span>
            </label>
            <button
              @click="removeColumn(index)"
              :disabled="columns.length <= 1"
              :class="[
                'p-1.5 rounded transition-colors',
                columns.length > 1
                  ? 'text-red-400 hover:bg-red-500/20'
                  : 'text-slate-600 cursor-not-allowed'
              ]"
            >
              <Trash2 class="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>

      <div class="mt-4 p-3 bg-slate-700/30 rounded-lg border border-slate-700">
        <p class="text-xs text-slate-500 mb-2">预览 SQL:</p>
        <code class="text-sm text-emerald-400 font-mono">
          CREATE TABLE "{{ tableName || 'table_name' }}" (<br>
          <span v-for="(col, i) in columns.filter(c => c.name)" :key="i">
            &nbsp;&nbsp;"{{ col.name }}" {{ col.type }}{{ col.primaryKey ? ' PRIMARY KEY' : '' }}{{ !col.nullable && !col.primaryKey ? ' NOT NULL' : '' }}{{ i < columns.filter(c => c.name).length - 1 ? ',' : '' }}<br>
          </span>
          );
        </code>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end gap-3">
        <Button variant="secondary" @click="emit('close')">取消</Button>
        <Button @click="createTable" :disabled="!tableName.trim() || !columns.some(c => c.name.trim())">
          创建
        </Button>
      </div>
    </template>
  </Modal>
</template>
