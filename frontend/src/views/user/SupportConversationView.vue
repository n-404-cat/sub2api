<template>
  <AppLayout>
    <div class="mx-auto max-w-4xl space-y-4">
      <div class="card p-5" v-if="detail?.conversation?.order">
        <div class="flex items-center justify-between gap-4">
          <div>
            <p class="text-xs text-gray-400 dark:text-gray-500">{{ localText('订单咨询', 'Order consultation') }}</p>
            <h2 class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">
              {{ detail.conversation.subject || `${localText('订单', 'Order')} #${detail.conversation.order_id}` }}
            </h2>
            <p class="mt-2 text-sm text-gray-600 dark:text-gray-300">
              {{ localText('订单号', 'Order No') }}: {{ detail.conversation.order.out_trade_no }}
            </p>
          </div>
          <div class="flex gap-3">
            <button class="btn btn-secondary" @click="router.push('/orders')">
              {{ localText('返回我的订单', 'Back to my orders') }}
            </button>
            <button class="btn btn-secondary" @click="router.push('/orders')">
              {{ localText('返回订单列表', 'Back to order list') }}
            </button>
          </div>
        </div>
      </div>

      <div class="card p-5">
        <div class="space-y-3">
          <div
            v-for="message in detail?.messages || []"
            :key="message.id"
            :class="[
              'rounded-2xl px-4 py-3 text-sm',
              message.sender_type === 'user'
                ? 'ml-auto max-w-[85%] bg-primary-50 text-primary-900 dark:bg-primary-900/25 dark:text-primary-100'
                : 'mr-auto max-w-[85%] bg-gray-100 text-gray-800 dark:bg-dark-700 dark:text-gray-100'
            ]"
          >
            <div class="mb-1 text-xs opacity-70">
              {{ message.sender_type === 'user' ? localText('你', 'You') : localText('客服', 'Support') }}
              ·
              {{ formatDateTime(message.created_at) }}
            </div>
            <div class="whitespace-pre-wrap break-words">{{ message.message }}</div>
          </div>
        </div>
      </div>

      <div class="card p-5">
        <div class="space-y-3">
          <label class="input-label">{{ localText('继续咨询', 'Continue conversation') }}</label>
          <textarea
            v-model="replyMessage"
            rows="4"
            class="input"
            :placeholder="localText('请输入你的问题或补充说明', 'Enter your message or additional details')"
          ></textarea>
          <div class="flex justify-end">
            <button class="btn btn-primary" :disabled="sending || !replyMessage.trim()" @click="sendReply">
              {{ sending ? t('common.processing') : localText('发送消息', 'Send message') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import { paymentAPI } from '@/api/payment'
import { useAppStore } from '@/stores'
import { extractI18nErrorMessage } from '@/utils/apiError'
import type { SupportConversationDetail } from '@/types/payment'

const { t, locale } = useI18n()
const route = useRoute()
const router = useRouter()
const appStore = useAppStore()

const detail = ref<SupportConversationDetail | null>(null)
const replyMessage = ref('')
const sending = ref(false)

function localText(zh: string, en: string): string {
  return String(locale.value || '').startsWith('zh') ? zh : en
}

function formatDateTime(value: string): string {
  return new Date(value).toLocaleString()
}

async function loadDetail() {
  const id = Number(route.params.id)
  if (!id) return
  try {
    const res = await paymentAPI.getMySupportConversation(id)
    detail.value = res.data
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  }
}

async function sendReply() {
  const id = Number(route.params.id)
  if (!id || !replyMessage.value.trim()) return
  sending.value = true
  try {
    const res = await paymentAPI.replyMySupportConversation(id, {
      message: replyMessage.value.trim(),
    })
    detail.value = res.data
    replyMessage.value = ''
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    sending.value = false
  }
}

onMounted(() => {
  loadDetail()
})
</script>
