<script setup>
import { RouterView } from 'vue-router'
import { useRoute } from 'vue-router'
import { computed, watch } from 'vue'
import { X } from 'lucide-vue-next'
import Toast from './components/common/Toast.vue'
import SupportModal from './components/SupportModal.vue'
import { useAuthStore } from './stores/auth'
import { useSupportStore } from './stores/support'

const route = useRoute()
const authStore = useAuthStore()
const support = useSupportStore()

const isLoginPage = computed(() => route.path === '/login')

// 顶部横幅：管理员可见；未支持当前版本前常驻展示（关闭不记忆，
// 刷新后再次出现，升级到新版本后也会再次出现）
const showDonateBanner = computed(() => {
  return !isLoginPage.value
    && authStore.isAdmin
    && support.bannerVisible
    && !support.donateSupported
})

const dismissDonateBanner = () => {
  support.dismissBanner()
}

// 赞赏弹窗调度：应用每次启动后首次进入页面时弹一次；
// 点【暂不支持】同版本同次运行内不再弹，应用重启后再次提示
const initDonate = () => {
  if (!authStore.isLoggedIn || !authStore.isAdmin) return
  support.init(true)
}

// SPA 内登录后不会重新挂载 App：离开登录页时补一次初始化（仅一次）
const donateInited = { value: false }
watch(isLoginPage, (loginPage, wasLoginPage) => {
  if (wasLoginPage && !loginPage && !donateInited.value) {
    donateInited.value = true
    initDonate()
  }
})
</script>

<template>
  <div class="min-h-screen bg-slate-900">
    <!-- 赞赏横幅：管理员可见，未支持当前版本前常驻展示；文档流内展示 -->
    <div
      v-if="!isLoginPage && showDonateBanner"
      class="flex items-center justify-between gap-3 px-4 py-2.5 bg-gradient-to-r from-indigo-500 to-violet-500 text-white text-sm shadow-md"
    >
      <span>☕ SQLite Manager 免费、无广告、数据都在你自己的设备上。如果它帮到了你，欢迎请站长喝杯咖啡——1 元也是心意，完全自愿，不影响任何功能。</span>
      <span class="flex items-center gap-2 shrink-0">
        <button
          class="underline underline-offset-2 hover:opacity-80"
          @click="support.show = true"
        >去赞赏</button>
        <button class="hover:opacity-80" title="关闭" @click="dismissDonateBanner">
          <X class="w-4 h-4" />
        </button>
      </span>
    </div>

    <RouterView />
    <SupportModal />
    <Toast />
  </div>
</template>
