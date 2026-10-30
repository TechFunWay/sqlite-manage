<script setup>
import { X, Heart } from 'lucide-vue-next'
import { useSupportStore } from '../stores/support'

const support = useSupportStore()
</script>

<template>
  <teleport to="body">
    <transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="support.show"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
        @click.self="support.dismiss()"
      >
        <!-- 精简为一屏展示（mobile-ui 规范 2026-10-01） -->
        <div class="bg-slate-800/95 backdrop-blur-xl rounded-2xl shadow-2xl border border-slate-700/50 w-full max-w-[320px] relative text-center">
          <button
            class="absolute top-2 right-2.5 text-slate-400 hover:text-slate-200 hover:bg-slate-700 transition-all px-2 py-1 rounded-md text-base leading-none z-10"
            @click="support.dismiss()"
          >
            ✕
          </button>

          <div class="px-5 pt-5 pb-4">
            <div class="text-base font-bold text-primary-400 mb-1.5">☕ 请站长喝杯咖啡？</div>

            <p class="text-slate-400 text-xs mb-3 leading-relaxed">
              应用免费无广告、数据保存在本机。若对你有帮助，欢迎赞赏——<strong class="text-slate-200">金额随意，1 元也是心意</strong>。
            </p>

            <div class="flex justify-center my-2">
              <div class="text-center">
                <img
                  src="/img/donate-wechat.png"
                  alt="微信收款码"
                  class="w-[150px] h-auto rounded-lg border border-slate-700 bg-white"
                />
                <span class="block mt-1 text-[11px] text-slate-500">微信扫码</span>
              </div>
            </div>

            <p class="text-[11px] text-slate-500 mt-2 leading-4">
              已支付可点「已支持」发送一次匿名计数（与支付记录无关）
            </p>
            <p v-if="support.errorText" class="text-xs text-rose-400 mt-1.5">{{ support.errorText }}</p>

            <div class="flex justify-center gap-2.5 mt-3.5">
              <button
                class="flex-1 px-3 py-2 text-sm text-slate-300 bg-slate-700/50 hover:bg-slate-700 rounded-lg transition-colors"
                @click="support.dismiss()"
              >
                暂不支持
              </button>
              <button
                class="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-sm text-white bg-gradient-to-r from-indigo-500 to-violet-500 hover:opacity-90 rounded-lg transition-opacity disabled:opacity-50"
                :disabled="support.sending"
                @click="support.confirmSupported()"
              >
                <Heart class="w-3.5 h-3.5" />
                {{ support.sending ? '发送中…' : '已支持' }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </transition>
  </teleport>
</template>
