<script setup>
import { RouterView } from 'vue-router'
import Toast from './components/common/Toast.vue'
</script>

<template>
  <div class="min-h-screen bg-slate-900">
    <RouterView />
    <Toast />
  </div>
</template>
