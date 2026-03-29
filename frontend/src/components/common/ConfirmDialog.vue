<script setup>
import { AlertTriangle } from 'lucide-vue-next'
import Modal from './Modal.vue'
import Button from './Button.vue'

defineProps({
  show: Boolean,
  title: String,
  message: String,
  confirmText: {
    type: String,
    default: '确认'
  },
  cancelText: {
    type: String,
    default: '取消'
  },
  danger: {
    type: Boolean,
    default: true
  }
})

const emit = defineEmits(['confirm', 'cancel'])
</script>

<template>
  <Modal :show="show" :title="title" size="small" @close="emit('cancel')">
    <div class="flex gap-4">
      <div class="flex-shrink-0 w-10 h-10 bg-amber-500/10 rounded-full flex items-center justify-center">
        <AlertTriangle class="w-5 h-5 text-amber-400" />
      </div>
      <p class="text-slate-300">{{ message }}</p>
    </div>
    
    <template #footer>
      <div class="flex justify-end gap-3">
        <Button variant="secondary" @click="emit('cancel')">
          {{ cancelText }}
        </Button>
        <Button variant="danger" @click="emit('confirm')">
          {{ confirmText }}
        </Button>
      </div>
    </template>
  </Modal>
</template>
