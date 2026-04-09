<script setup>
import { computed, ref, onMounted } from 'vue'

const props = defineProps({
  modelValue: [String, Number],
  label: String,
  placeholder: String,
  type: {
    type: String,
    default: 'text'
  },
  error: String,
  disabled: Boolean,
  required: Boolean,
  autofocus: Boolean
})

const emit = defineEmits(['update:modelValue', 'keyup'])

const inputRef = ref(null)

const inputClasses = computed(() => [
  'w-full px-4 py-2.5 bg-slate-700/50 border rounded-lg text-slate-100 placeholder-slate-500',
  'focus:outline-none focus:ring-2 focus:ring-primary-500/50 focus:border-primary-500',
  'transition-all duration-200',
  props.error ? 'border-red-500' : 'border-slate-600',
  props.disabled ? 'opacity-50 cursor-not-allowed' : ''
])

defineExpose({
  focus: () => inputRef.value?.focus()
})

onMounted(() => {
  if (props.autofocus && inputRef.value) {
    inputRef.value.focus()
  }
})
</script>

<template>
  <div class="space-y-1.5">
    <label v-if="label" class="block text-sm font-medium text-slate-300">
      {{ label }}
      <span v-if="required" class="text-red-400">*</span>
    </label>
    <input
      ref="inputRef"
      :type="type"
      :value="modelValue"
      :placeholder="placeholder"
      :disabled="disabled"
      :required="required"
      @input="emit('update:modelValue', $event.target.value)"
      @keyup="$emit('keyup', $event)"
      :class="inputClasses"
    />
    <p v-if="error" class="text-xs text-red-400">{{ error }}</p>
  </div>
</template>
