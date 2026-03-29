<script setup>
import { computed } from 'vue'
import { useDatabaseStore } from '../../stores/database'
import { HardDrive, FileText, Clock, Database, Table, Rows3, Hash } from 'lucide-vue-next'
import dayjs from 'dayjs'

const store = useDatabaseStore()

const info = computed(() => store.databaseInfo)

function formatSize(bytes) {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

function formatDate(date) {
  if (!date) return '-'
  return dayjs(date).format('YYYY-MM-DD HH:mm:ss')
}
</script>

<template>
  <div class="space-y-6">
    <h2 class="text-lg font-semibold text-slate-100">数据库信息</h2>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div class="bg-slate-800/50 rounded-lg p-4 border border-slate-700">
        <div class="flex items-center gap-3 mb-3">
          <div class="w-10 h-10 bg-primary-500/10 rounded-lg flex items-center justify-center">
            <FileText class="w-5 h-5 text-primary-400" />
          </div>
          <div>
            <p class="text-xs text-slate-500">文件名</p>
            <p class="text-slate-200 font-medium">{{ info?.name }}</p>
          </div>
        </div>
        <p class="text-xs text-slate-500 font-mono truncate" :title="info?.path">
          {{ info?.path }}
        </p>
      </div>

      <div class="bg-slate-800/50 rounded-lg p-4 border border-slate-700">
        <div class="flex items-center gap-3 mb-3">
          <div class="w-10 h-10 bg-emerald-500/10 rounded-lg flex items-center justify-center">
            <HardDrive class="w-5 h-5 text-emerald-400" />
          </div>
          <div>
            <p class="text-xs text-slate-500">文件大小</p>
            <p class="text-slate-200 font-medium">{{ formatSize(info?.size || 0) }}</p>
          </div>
        </div>
      </div>

      <div class="bg-slate-800/50 rounded-lg p-4 border border-slate-700">
        <div class="flex items-center gap-3 mb-3">
          <div class="w-10 h-10 bg-violet-500/10 rounded-lg flex items-center justify-center">
            <Hash class="w-5 h-5 text-violet-400" />
          </div>
          <div>
            <p class="text-xs text-slate-500">SQLite 版本</p>
            <p class="text-slate-200 font-medium">{{ info?.sqliteVersion }}</p>
          </div>
        </div>
      </div>

      <div class="bg-slate-800/50 rounded-lg p-4 border border-slate-700">
        <div class="flex items-center gap-3 mb-3">
          <div class="w-10 h-10 bg-cyan-500/10 rounded-lg flex items-center justify-center">
            <Table class="w-5 h-5 text-cyan-400" />
          </div>
          <div>
            <p class="text-xs text-slate-500">表数量</p>
            <p class="text-slate-200 font-medium">{{ info?.tableCount }} 个表</p>
          </div>
        </div>
      </div>

      <div class="bg-slate-800/50 rounded-lg p-4 border border-slate-700">
        <div class="flex items-center gap-3 mb-3">
          <div class="w-10 h-10 bg-amber-500/10 rounded-lg flex items-center justify-center">
            <Rows3 class="w-5 h-5 text-amber-400" />
          </div>
          <div>
            <p class="text-xs text-slate-500">总行数</p>
            <p class="text-slate-200 font-medium">{{ info?.totalRows?.toLocaleString() }} 行</p>
          </div>
        </div>
      </div>

      <div class="bg-slate-800/50 rounded-lg p-4 border border-slate-700">
        <div class="flex items-center gap-3 mb-3">
          <div class="w-10 h-10 bg-pink-500/10 rounded-lg flex items-center justify-center">
            <Clock class="w-5 h-5 text-pink-400" />
          </div>
          <div>
            <p class="text-xs text-slate-500">最后修改</p>
            <p class="text-slate-200 font-medium">{{ formatDate(info?.modifiedAt) }}</p>
          </div>
        </div>
      </div>
    </div>

    <div v-if="store.tables.length > 0" class="bg-slate-800/50 rounded-lg border border-slate-700 overflow-hidden">
      <div class="px-4 py-3 border-b border-slate-700 bg-slate-800">
        <h3 class="font-medium text-slate-200">表概览</h3>
      </div>
      <div class="divide-y divide-slate-700/50">
        <div
          v-for="table in store.tables"
          :key="table.name"
          @click="store.selectTable(table.name)"
          class="flex items-center justify-between px-4 py-3 hover:bg-slate-700/30 cursor-pointer transition-colors"
        >
          <div class="flex items-center gap-3">
            <Database class="w-4 h-4 text-slate-500" />
            <span class="text-slate-200">{{ table.name }}</span>
          </div>
          <span class="text-sm text-slate-500">{{ table.rows }} 行</span>
        </div>
      </div>
    </div>
  </div>
</template>
