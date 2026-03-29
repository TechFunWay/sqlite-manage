<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useDatabaseStore } from '../../stores/database'
import { useAuthStore } from '../../stores/auth'
import { systemApi } from '../../api'
import { Database, LogOut, Plus, ChevronDown, X, Key, User, Info } from 'lucide-vue-next'
import Button from '../common/Button.vue'
import Modal from '../common/Modal.vue'
import Input from '../common/Input.vue'

const router = useRouter()
const store = useDatabaseStore()
const authStore = useAuthStore()

const showDropdown = ref(false)
const showUserMenu = ref(false)
const showChangePasswordModal = ref(false)
const appVersion = ref('')

const oldPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')

onMounted(async () => {
  try {
    const response = await systemApi.getVersion()
    appVersion.value = response.data.appVersion
  } catch (error) {
    console.error('Failed to get version', error)
  }
})

function toggleDropdown() {
  showDropdown.value = !showDropdown.value
  showUserMenu.value = false
}

function toggleUserMenu() {
  showUserMenu.value = !showUserMenu.value
  showDropdown.value = false
}

async function selectDb(db) {
  if (!db.active) {
    await store.activateDatabase(db.id)
  }
  showDropdown.value = false
}

async function closeDb(db, event) {
  event.stopPropagation()
  await store.closeDatabase(db.id)
}

function goHome() {
  showDropdown.value = false
  store.showDatabaseSelector()
}

function openChangePasswordModal() {
  oldPassword.value = ''
  newPassword.value = ''
  confirmPassword.value = ''
  showUserMenu.value = false
  showChangePasswordModal.value = true
}

async function handleChangePassword() {
  if (!oldPassword.value || !newPassword.value || newPassword.value !== confirmPassword.value) {
    return
  }
  
  const success = await authStore.changePassword(oldPassword.value, newPassword.value)
  if (success) {
    showChangePasswordModal.value = false
  }
}

async function handleLogout() {
  await authStore.logout()
  router.push('/login')
}
</script>

<template>
  <header class="h-14 bg-slate-800/80 backdrop-blur-xl border-b border-slate-700 flex items-center justify-between px-4 relative z-40">
    <div class="flex items-center gap-4">
      <div class="flex items-center gap-2">
        <div class="w-8 h-8 bg-primary-500/10 rounded-lg flex items-center justify-center">
          <Database class="w-4 h-4 text-primary-400" />
        </div>
        <div class="flex flex-col">
          <span class="font-semibold text-slate-100 text-sm">SQLite Manager</span>
          <span v-if="appVersion" class="text-xs text-slate-500">v{{ appVersion }}</span>
        </div>
      </div>
      
      <div class="relative database-dropdown">
        <button
          @click="toggleDropdown"
          class="flex items-center gap-2 px-3 py-1.5 bg-slate-700/50 hover:bg-slate-700 rounded-lg text-sm transition-colors max-w-xs sm:max-w-md"
        >
          <div v-if="store.databaseInfo" class="w-2 h-2 bg-emerald-400 rounded-full flex-shrink-0"></div>
          <span class="text-slate-300 truncate" style="max-width: 250px;" :title="store.databaseInfo?.path">
            {{ store.databaseInfo?.name || '选择数据库' }}
          </span>
          <ChevronDown class="w-4 h-4 text-slate-400 flex-shrink-0" />
        </button>
        
        <Transition name="dropdown">
          <div
            v-if="showDropdown"
            class="absolute top-full left-0 mt-2 w-96 max-w-[90vw] bg-slate-800 rounded-lg border border-slate-700 shadow-2xl overflow-hidden z-50"
          >
            <div class="p-2 border-b border-slate-700">
              <button
                @click="goHome"
                class="w-full flex items-center gap-2 px-3 py-2 text-sm text-primary-400 hover:bg-slate-700 rounded-md transition-colors"
              >
                <Plus class="w-4 h-4" />
                打开新数据库
              </button>
            </div>
            
            <div class="max-h-64 overflow-auto">
              <div
                v-for="db in store.databases"
                :key="db.id"
                @click="selectDb(db)"
                class="group flex items-center gap-3 px-4 py-3 hover:bg-slate-700/50 cursor-pointer transition-colors"
                :class="{ 'bg-primary-500/10': db.active }"
              >
                <Database class="w-4 h-4 text-slate-500 flex-shrink-0" />
                <div class="flex-1 min-w-0">
                  <p class="text-sm text-slate-200 truncate">{{ db.name }}</p>
                  <p class="text-xs text-slate-500 truncate" :title="db.path">{{ db.path }}</p>
                </div>
                <div class="flex items-center gap-2 flex-shrink-0">
                  <span v-if="db.active" class="px-1.5 py-0.5 bg-emerald-500/20 text-emerald-400 text-xs rounded">
                    活跃
                  </span>
                  <button
                    @click="closeDb(db, $event)"
                    class="p-1 rounded text-slate-500 hover:text-red-400 hover:bg-red-500/10 opacity-0 group-hover:opacity-100 transition-all"
                    title="关闭数据库"
                  >
                    <X class="w-3.5 h-3.5" />
                  </button>
                </div>
              </div>
              
              <div v-if="store.databases.length === 0" class="px-4 py-8 text-center text-slate-500 text-sm">
                暂无数据库
              </div>
            </div>
          </div>
        </Transition>
      </div>
    </div>
    
    <div class="flex items-center gap-3">
      <div v-if="store.databaseInfo" class="hidden lg:flex items-center gap-4 text-xs text-slate-500">
        <span>{{ store.databaseInfo.path }}</span>
      </div>
      
      <!-- User Menu -->
      <div class="relative">
        <button
          @click="toggleUserMenu"
          class="flex items-center gap-2 px-3 py-1.5 bg-slate-700/50 hover:bg-slate-700 rounded-lg text-sm transition-colors"
        >
          <User class="w-4 h-4 text-slate-400" />
          <span class="text-slate-300 hidden sm:inline">{{ authStore.user?.username }}</span>
          <ChevronDown class="w-4 h-4 text-slate-400" />
        </button>
        
        <Transition name="dropdown">
          <div
            v-if="showUserMenu"
            class="absolute top-full right-0 mt-2 w-48 bg-slate-800 rounded-lg border border-slate-700 shadow-2xl overflow-hidden z-50"
          >
            <div class="px-4 py-2 border-b border-slate-700">
              <p class="text-xs text-slate-500">版本 v{{ appVersion }}</p>
            </div>
            <button
              @click="openChangePasswordModal"
              class="w-full flex items-center gap-2 px-4 py-3 text-sm text-slate-300 hover:bg-slate-700/50 transition-colors"
            >
              <Key class="w-4 h-4" />
              修改密码
            </button>
            <button
              @click="handleLogout"
              class="w-full flex items-center gap-2 px-4 py-3 text-sm text-red-400 hover:bg-red-500/10 transition-colors"
            >
              <LogOut class="w-4 h-4" />
              退出登录
            </button>
          </div>
        </Transition>
      </div>
    </div>
    
    <!-- Change Password Modal -->
    <Modal :show="showChangePasswordModal" title="修改密码" @close="showChangePasswordModal = false">
      <div class="space-y-4">
        <Input
          v-model="oldPassword"
          type="password"
          label="当前密码"
          placeholder="请输入当前密码"
          required
        />
        <Input
          v-model="newPassword"
          type="password"
          label="新密码"
          placeholder="请输入新密码"
          required
        />
        <Input
          v-model="confirmPassword"
          type="password"
          label="确认新密码"
          placeholder="请再次输入新密码"
          :error="confirmPassword && newPassword !== confirmPassword ? '两次输入的密码不一致' : ''"
          required
        />
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <Button variant="secondary" @click="showChangePasswordModal = false">取消</Button>
          <Button 
            @click="handleChangePassword" 
            :disabled="!oldPassword || !newPassword || newPassword !== confirmPassword"
          >
            确认修改
          </Button>
        </div>
      </template>
    </Modal>
  </header>
</template>

<style scoped>
.dropdown-enter-active,
.dropdown-leave-active {
  transition: all 0.2s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}
</style>
