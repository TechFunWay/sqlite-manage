import { defineStore } from 'pinia'
import { ref } from 'vue'
import axios from 'axios'

// 赞赏支持状态（与家族其他应用一致）：
// - 应用每次启动后首次进入页面自动弹一次；点【暂不支持】同版本同次运行内
//   不再弹（刷新不弹），应用重启后再次提示；【已支持】上报成功后本版本内不再提示
// - "本次运行"由后端 /api/version 的 startedAt 区分
//
// 关键约束：关闭抑制记录必须同时绑定版本号。升级换版本后，旧版本里记下的
// 关闭记录一律失效——弹窗与横幅必定再次弹出；点【暂不支持】只表示
// 「本版本本次运行内不再打扰」。
//
// 按版本提示：支持记录按版本号记忆（donate_supported_version）。
// 应用升级到新版本后，即使旧版本支持过，也会重新出现横幅并再次弹窗提醒，
// 每个新版本都重来一次；「已支持」上报成功后本版本内停止提示。
//
// 赞赏横幅常驻展示（关闭不持久化）：刷新浏览器后再次出现，直到用户
// 支持过当前版本才隐藏。赞赏完全自愿：不赞赏不影响任何功能。

const KEY_SUPPORTED_VERSION = 'donate_supported_version'
const KEY_DISMISSED_START = 'donate_dismissed_start'
const KEY_DISMISSED_VERSION = 'donate_dismissed_version'

function lsGet(key) {
  try {
    return localStorage.getItem(key)
  } catch {
    return null
  }
}

function lsSet(key, value) {
  try {
    localStorage.setItem(key, value)
  } catch {
    /* 隐私模式下不可用，忽略 */
  }
}

/** 当前版本是否已支持过（支持记录按版本号记忆）。 */
function isSupportedFor(appVersion, storedVersion) {
  if (!storedVersion || !appVersion) return false
  return storedVersion === appVersion
}

export const useSupportStore = defineStore('support', () => {
  const show = ref(false)
  const sending = ref(false)
  const errorText = ref('')

  const appVersion = ref('')
  const serverStartedAt = ref('')

  // 记忆的版本号：最近一次「已支持」上报成功时的应用版本
  const supportedVersion = ref(lsGet(KEY_SUPPORTED_VERSION) || '')
  // 当前版本是否已支持
  const donateSupported = ref(false)

  // 顶部横幅：管理员可见；关闭不持久化，刷新后再次出现；
  // 支持过当前版本后隐藏，升级到新版本后再次出现
  const bannerVisible = ref(false)

  async function init(isAdmin) {
    if (!isAdmin) return
    try {
      const res = await axios.get(`${import.meta.env.BASE_URL}api/version`)
      appVersion.value = res.data?.appVersion || ''
      serverStartedAt.value = res.data?.startedAt || ''
    } catch { /* 版本接口失败不影响使用 */ }

    refreshSupported()

    if (!donateSupported.value) {
      setTimeout(() => schedule(), 1500)
      maybeShowBanner()
    }
  }

  /** 版本或支持记录变化后重新计算「当前版本是否已支持」。 */
  function refreshSupported() {
    donateSupported.value = isSupportedFor(appVersion.value, supportedVersion.value || null)
  }

  function schedule() {
    if (donateSupported.value) return
    // 抑制记录必须「同版本 + 同次运行」才有效：升级换版本后旧记录失效，
    // 弹窗重新出现；未拿到 startedAt 时按「本次运行」降级处理。
    const dismissedStart = lsGet(KEY_DISMISSED_START)
    const dismissedVersion = lsGet(KEY_DISMISSED_VERSION)
    const sameRun = !!serverStartedAt.value && dismissedStart === serverStartedAt.value
    const sameVersion = !!appVersion.value && dismissedVersion === appVersion.value
    if (sameRun && sameVersion) return
    if (sameRun && !appVersion.value) return
    show.value = true
  }

  function maybeShowBanner() {
    // 横幅常驻展示：未支持当前版本前每次进入都显示；关闭不记忆
    if (donateSupported.value) return
    bannerVisible.value = true
  }

  /** 关闭横幅：本次会话隐藏，刷新浏览器后再次出现。 */
  function dismissBanner() {
    bannerVisible.value = false
  }

  // 【暂不支持】/ 关闭：同版本同次运行内不再弹出
  function dismiss() {
    show.value = false
    errorText.value = ''
    if (serverStartedAt.value) lsSet(KEY_DISMISSED_START, serverStartedAt.value)
    if (appVersion.value) lsSet(KEY_DISMISSED_VERSION, appVersion.value)
  }

  // 【已支持】上报匿名支持计数，成功才记住当前版本；升级后重新提示
  async function confirmSupported() {
    sending.value = true
    errorText.value = ''
    let ok = false
    try {
      const token = localStorage.getItem('token')
      const res = await axios.post(
        `${import.meta.env.BASE_URL}api/donate/support`,
        null,
        { headers: { Authorization: `Bearer ${token}` } }
      )
      ok = res.status === 200 && res.data?.ok === true
    } catch {
      ok = false
    }
    sending.value = false
    if (ok) {
      supportedVersion.value = appVersion.value
      lsSet(KEY_SUPPORTED_VERSION, appVersion.value)
      refreshSupported()
      show.value = false
      bannerVisible.value = false
    } else {
      errorText.value = '发送失败，请稍后重试（不影响你的支持 ❤）'
    }
    return ok
  }

  // 入口按钮：手动打开弹窗
  function open() {
    errorText.value = ''
    show.value = true
  }

  return {
    show, sending, errorText,
    appVersion, serverStartedAt, donateSupported, bannerVisible,
    init, dismiss, dismissBanner, confirmSupported, open,
  }
})
