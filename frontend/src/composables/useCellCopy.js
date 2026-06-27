import { useToastStore } from '../stores/toast'

// 单元格快捷复制：
// - 手机/触屏：长按约 500ms 复制
// - 电脑/鼠标：单击复制（双击仍用于编辑，因此与双击区分开）
// 复制内容为单元格原始值，NULL 值不复制只提示。
//
// 同一时刻只会有一次交互，故交互状态在 composable 实例内共享，
// 避免把状态放在 getCellHandlers 的闭包里——后者会随模板重渲染被重建。
export function useCellCopy() {
  const toast = useToastStore()

  const LONG_PRESS_MS = 500
  const MOVE_CANCEL_PX = 10
  const DBLCLICK_GAP_MS = 250

  let longPressTimer = null
  let longPressed = false
  let startX = 0
  let startY = 0
  let clickTimer = null

  function clearLongPress() {
    if (longPressTimer) {
      clearTimeout(longPressTimer)
      longPressTimer = null
    }
  }

  async function writeClipboard(text) {
    // 优先使用现代剪贴板 API，不可用时降级到 execCommand。
    if (navigator.clipboard && window.isSecureContext) {
      try {
        await navigator.clipboard.writeText(text)
        return true
      } catch (e) {
        // 继续走降级方案
      }
    }
    try {
      const ta = document.createElement('textarea')
      ta.value = text
      ta.style.position = 'fixed'
      ta.style.left = '-9999px'
      ta.style.top = '0'
      document.body.appendChild(ta)
      ta.focus()
      ta.select()
      const ok = document.execCommand('copy')
      document.body.removeChild(ta)
      return ok
    } catch (e) {
      return false
    }
  }

  async function copyValue(value) {
    if (value === null || value === undefined) {
      toast.warning('NULL 值，无内容可复制')
      return
    }
    const ok = await writeClipboard(String(value))
    if (ok) {
      toast.success('已复制')
    } else {
      toast.error('复制失败')
    }
  }

  // 为某个单元格内容元素返回一组 pointer 事件处理器（用于 v-on 对象绑定）。
  function getCellHandlers(value) {
    return {
      pointerdown(e) {
        longPressed = false
        if (e.pointerType === 'mouse') return
        // 触屏 / 触控笔：启动长按计时
        startX = e.clientX
        startY = e.clientY
        clearLongPress()
        longPressTimer = setTimeout(() => {
          longPressed = true
          longPressTimer = null
          if (navigator.vibrate) navigator.vibrate(15)
          copyValue(value)
        }, LONG_PRESS_MS)
      },
      pointermove(e) {
        if (!longPressTimer) return
        if (
          Math.abs(e.clientX - startX) > MOVE_CANCEL_PX ||
          Math.abs(e.clientY - startY) > MOVE_CANCEL_PX
        ) {
          clearLongPress() // 视为滚动，取消长按
        }
      },
      pointerup() {
        clearLongPress()
      },
      pointerleave() {
        clearLongPress()
      },
      // 鼠标单击复制；通过延时与双击（编辑）区分开。
      click(e) {
        if (e.pointerType !== undefined && e.pointerType !== 'mouse') return
        if (e.detail >= 2) {
          // 双击：取消待执行的单击复制，交给 dblclick 处理
          if (clickTimer) {
            clearTimeout(clickTimer)
            clickTimer = null
          }
          return
        }
        if (clickTimer) clearTimeout(clickTimer)
        clickTimer = setTimeout(() => {
          clickTimer = null
          copyValue(value)
        }, DBLCLICK_GAP_MS)
      },
      contextmenu(e) {
        // 长按触发后阻止原生选词/右键菜单
        if (longPressed) e.preventDefault()
      },
    }
  }

  return { copyValue, getCellHandlers }
}
