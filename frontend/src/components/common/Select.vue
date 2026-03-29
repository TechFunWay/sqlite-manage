<script setup>
import { computed } from 'vue'

const props = defineProps({
  modelValue: [String, Number],
  label: String,
  options: {
    type: Array,
    required: true
  },
  placeholder: String,
  disabled: Boolean,
  required: Boolean
})

const emit = defineEmits(['update:modelValue'])

const selectClasses = computed(() => [
  'w-full px-4 py-2.5 bg-slate-700/50 border border-slate-600 rounded-lg text-slate-100',
  'focus:outline-none focus:ring-2 focus:ring-primary-500/50 focus:border-primary-500',
  'transition-all duration-200 cursor-pointer',
  props.disabled ? 'opacity-50 cursor-not-allowed' : ''
])
</script>

<template>
  <div class="space-y-1.5">
    <label v-if="label" class="block text-sm font-medium text-slate-300">
      {{ label }}
      <span v-if="required" class="text-red-400">*</span>
    </label>
    <select
      :value="modelValue"
      :disabled="disabled"
      :required="required"
      @change="emit('update:modelValue', $event.target.value)"
      :class="selectClasses"
    >
      <option v-if="placeholder" value="" disabled>{{ placeholder }}</option>
      <option
        v-for="option in options"
        :key="option.value"
        :value="option.value"
      >
        {{ option.label }}
      </option>
    </select>
  </div>
</template>
