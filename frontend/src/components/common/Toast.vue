<script setup>
import { useToastStore } from '../../stores/toast'
import { CheckCircle, XCircle, AlertTriangle, Info, X } from 'lucide-vue-next'

const toastStore = useToastStore()

const icons = {
  success: CheckCircle,
  error: XCircle,
  warning: AlertTriangle,
  info: Info
}

const colors = {
  success: 'bg-emerald-500/10 border-emerald-500/30 text-emerald-400',
  error: 'bg-red-500/10 border-red-500/30 text-red-400',
  warning: 'bg-amber-500/10 border-amber-500/30 text-amber-400',
  info: 'bg-cyan-500/10 border-cyan-500/30 text-cyan-400'
}
</script>

<template>
  <div class="fixed top-4 right-4 z-50 flex flex-col gap-2 max-w-sm">
    <TransitionGroup name="fade">
      <div
        v-for="toast in toastStore.toasts"
        :key="toast.id"
        :class="['flex items-center gap-3 px-4 py-3 rounded-lg border backdrop-blur-sm shadow-lg', colors[toast.type]]"
      >
        <component :is="icons[toast.type]" class="w-5 h-5 flex-shrink-0" />
        <span class="flex-1 text-sm">{{ toast.message }}</span>
        <button
          @click="toastStore.removeToast(toast.id)"
          class="p-1 hover:opacity-70 transition-opacity"
        >
          <X class="w-4 h-4" />
        </button>
      </div>
    </TransitionGroup>
  </div>
</template>
