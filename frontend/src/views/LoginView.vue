<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { systemApi } from '../api'
import { Database, Lock, User, LogIn, UserPlus } from 'lucide-vue-next'
import Button from '../components/common/Button.vue'
import Input from '../components/common/Input.vue'

const router = useRouter()
const authStore = useAuthStore()

const username = ref('')
const password = ref('')
const confirmPassword = ref('')
const isRegisterMode = ref(false)
const appVersion = ref('')

onMounted(async () => {
  await authStore.checkAuth()
  if (authStore.hasAdmin) {
    isRegisterMode.value = false
  } else {
    isRegisterMode.value = true
  }
  
  // Get version
  try {
    const response = await systemApi.getVersion()
    appVersion.value = response.data.appVersion
  } catch (error) {
    console.error('Failed to get version', error)
  }
})

async function handleSubmit() {
  if (!username.value.trim() || !password.value.trim()) {
    return
  }

  if (isRegisterMode.value) {
    if (password.value !== confirmPassword.value) {
      return
    }
    const success = await authStore.register(username.value.trim(), password.value)
    if (success) {
      router.push('/')
    }
  } else {
    const success = await authStore.login(username.value.trim(), password.value)
    if (success) {
      router.push('/')
    }
  }
}

function toggleMode() {
  if (!authStore.hasAdmin) {
    return // Cannot toggle if no admin exists
  }
  isRegisterMode.value = !isRegisterMode.value
  username.value = ''
  password.value = ''
  confirmPassword.value = ''
}
</script>

<template>
  <div class="min-h-screen bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 flex items-center justify-center p-4">
    <div class="w-full max-w-md">
      <!-- Logo -->
      <div class="text-center mb-8">
        <div class="inline-flex items-center justify-center w-16 h-16 bg-primary-500/10 rounded-2xl mb-4">
          <Database class="w-8 h-8 text-primary-400" />
        </div>
        <h1 class="text-2xl font-bold text-slate-100">SQLite Manager</h1>
        <p v-if="appVersion" class="text-xs text-slate-500 mt-1">v{{ appVersion }}</p>
        <p class="text-slate-400 mt-1">
          {{ isRegisterMode ? '创建管理员账号' : '登录您的账号' }}
        </p>
      </div>

      <!-- Form Card -->
      <div class="bg-slate-800/50 backdrop-blur-xl rounded-2xl border border-slate-700/50 p-6 shadow-2xl">
        <form @submit.prevent="handleSubmit" class="space-y-5">
          <!-- Username -->
          <div class="space-y-1.5">
            <label class="block text-sm font-medium text-slate-300">用户名</label>
            <div class="relative">
              <User class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500" />
              <input
                v-model="username"
                type="text"
                placeholder="请输入用户名"
                required
                class="w-full pl-10 pr-4 py-2.5 bg-slate-700/50 border border-slate-600 rounded-lg text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-primary-500/50 focus:border-primary-500"
              />
            </div>
          </div>

          <!-- Password -->
          <div class="space-y-1.5">
            <label class="block text-sm font-medium text-slate-300">密码</label>
            <div class="relative">
              <Lock class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500" />
              <input
                v-model="password"
                type="password"
                placeholder="请输入密码"
                required
                class="w-full pl-10 pr-4 py-2.5 bg-slate-700/50 border border-slate-600 rounded-lg text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-primary-500/50 focus:border-primary-500"
              />
            </div>
          </div>

          <!-- Confirm Password (Register only) -->
          <div v-if="isRegisterMode" class="space-y-1.5">
            <label class="block text-sm font-medium text-slate-300">确认密码</label>
            <div class="relative">
              <Lock class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500" />
              <input
                v-model="confirmPassword"
                type="password"
                placeholder="请再次输入密码"
                required
                :class="[
                  'w-full pl-10 pr-4 py-2.5 bg-slate-700/50 border rounded-lg text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-primary-500/50 focus:border-primary-500',
                  confirmPassword && password !== confirmPassword ? 'border-red-500' : 'border-slate-600'
                ]"
              />
            </div>
            <p v-if="confirmPassword && password !== confirmPassword" class="text-xs text-red-400">
              两次输入的密码不一致
            </p>
          </div>

          <!-- Submit Button -->
          <Button
            type="submit"
            fullWidth
            :loading="authStore.loading"
            :disabled="!username.trim() || !password.trim() || (isRegisterMode && password !== confirmPassword)"
          >
            <UserPlus v-if="isRegisterMode" class="w-4 h-4" />
            <LogIn v-else class="w-4 h-4" />
            {{ isRegisterMode ? '创建账号' : '登录' }}
          </Button>
        </form>

        <!-- Toggle Mode (only show when no admin exists) -->
        <!-- No toggle needed - if admin exists, only show login -->

        <!-- First time message -->
        <div v-if="!authStore.hasAdmin && isRegisterMode" class="mt-4 p-3 bg-amber-500/10 border border-amber-500/30 rounded-lg">
          <p class="text-sm text-amber-400 text-center">
            首次使用，请创建管理员账号
          </p>
        </div>
      </div>

      <!-- Footer -->
      <p class="text-center text-slate-500 text-xs mt-6">
        忘记密码？请使用终端命令: <code class="bg-slate-800 px-1 py-0.5 rounded">./sqlite-manager reset-password</code>
      </p>
    </div>
  </div>
</template>
