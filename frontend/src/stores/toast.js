import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useToastStore = defineStore('toast', () => {
  const toasts = ref([])
  let idCounter = 0

  function addToast(message, type = 'info', duration = 3000) {
    const id = ++idCounter
    toasts.value.push({ id, message, type })
    
    if (duration > 0) {
      setTimeout(() => {
        removeToast(id)
      }, duration)
    }
    
    return id
  }

  function removeToast(id) {
    const index = toasts.value.findIndex(t => t.id === id)
    if (index !== -1) {
      toasts.value.splice(index, 1)
    }
  }

  function success(message, duration) {
    return addToast(message, 'success', duration)
  }

  function error(message, duration) {
    return addToast(message, 'error', duration)
  }

  function warning(message, duration) {
    return addToast(message, 'warning', duration)
  }

  function info(message, duration) {
    return addToast(message, 'info', duration)
  }

  return {
    toasts,
    addToast,
    removeToast,
    success,
    error,
    warning,
    info
  }
})
