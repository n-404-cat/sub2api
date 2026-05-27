<template>
  <AppLayout>
    <div class="mx-auto max-w-4xl space-y-4">
      <div class="card p-5">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div>
            <p class="text-xs text-gray-400 dark:text-gray-500">{{ localText('客服会话', 'Support conversation') }}</p>
            <h2 class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">
              {{ detail?.conversation.subject || localText('客服咨询', 'Support') }}
            </h2>
            <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">
              {{ detail?.conversation.order?.out_trade_no ? `${localText('已关联订单', 'Linked order')}: ${detail.conversation.order.out_trade_no}` : localText('当前未关联订单', 'No order linked yet') }}
            </p>
          </div>
          <div class="flex gap-2">
            <button class="btn btn-secondary" @click="showOrderPicker = true">{{ localText('选择订单', 'Choose order') }}</button>
            <button class="btn btn-secondary" @click="router.push('/support')">{{ localText('返回列表', 'Back') }}</button>
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

      <div class="card p-5 space-y-3">
        <div class="flex flex-wrap gap-2">
          <button v-for="emoji in emojis" :key="emoji" class="rounded-full border border-gray-200 px-3 py-1 text-sm hover:bg-gray-50 dark:border-dark-700 dark:hover:bg-dark-800" @click="appendEmoji(emoji)">
            {{ emoji }}
          </button>
        </div>

        <div v-if="quickReplies.length" class="flex flex-wrap gap-2">
          <button v-for="reply in quickReplies" :key="reply" class="rounded-full bg-primary-50 px-3 py-1 text-sm text-primary-700 hover:bg-primary-100 dark:bg-primary-900/20 dark:text-primary-200" @click="insertQuickReply(reply)">
            {{ reply }}
          </button>
        </div>

        <textarea v-model="replyMessage" rows="4" class="input" :placeholder="localText('输入消息', 'Type a message')"></textarea>
        <div class="flex items-center justify-between gap-3">
          <span class="text-xs text-gray-400">{{ localText('表情和快捷回复可直接插入', 'Emoji and quick replies insert text directly') }}</span>
          <button class="btn btn-primary" :disabled="sending || !replyMessage.trim()" @click="sendReply">
            {{ sending ? t('common.processing') : localText('发送消息', 'Send message') }}
          </button>
        </div>
      </div>
    </div>
  </AppLayout>

  <BaseDialog :show="showOrderPicker" :title="localText('选择订单', 'Choose order')" width="normal" @close="showOrderPicker = false">
    <div class="space-y-2">
      <button
        v-for="order in orders"
        :key="order.id"
        class="w-full rounded-xl border px-4 py-3 text-left hover:bg-gray-50 dark:border-dark-700 dark:hover:bg-dark-800"
        @click="bindOrder(order.id)"
      >
        <div class="flex justify-between">
          <span class="font-medium text-gray-900 dark:text-white">{{ order.out_trade_no }}</span>
          <span class="text-xs text-gray-400">#{{ order.id }}</span>
        </div>
      </button>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { paymentAPI } from '@/api/payment'
import { useAppStore } from '@/stores'
import { extractI18nErrorMessage } from '@/utils/apiError'
import type { PaymentOrder, SupportConversationDetail } from '@/types/payment'

const { t, locale } = useI18n()
const route = useRoute()
const router = useRouter()
const appStore = useAppStore()

const detail = ref<SupportConversationDetail | null>(null)
const replyMessage = ref('')
const sending = ref(false)
const showOrderPicker = ref(false)
const orders = ref<PaymentOrder[]>([])
const emojis = ['😀', '😍', '🙏', '👍', '🎉', '💬', '📦', '✅']

const quickReplies = ref<string[]>([])

function localText(zh: string, en: string): string {
  return String(locale.value || '').startsWith('zh') ? zh : en
}

function formatDateTime(value: string): string {
  return new Date(value).toLocaleString()
}

function appendEmoji(emoji: string) {
  replyMessage.value += emoji
}

function insertQuickReply(text: string) {
  replyMessage.value = replyMessage.value ? `${replyMessage.value}\n${text}` : text
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

async function loadOrders() {
  try {
    const res = await paymentAPI.getMyOrders({ page: 1, page_size: 20 })
    orders.value = res.data.items || []
  } catch {}
}

async function loadQuickReplies() {
  try {
    const res = await paymentAPI.getConfig()
    quickReplies.value = res.data.support_quick_replies || []
  } catch {}
}

async function bindOrder(orderId: number) {
  const id = Number(route.params.id)
  try {
    const res = await paymentAPI.bindSupportConversationOrder(id, { order_id: orderId })
    detail.value = res.data
    showOrderPicker.value = false
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  }
}

async function sendReply() {
  const id = Number(route.params.id)
  if (!id || !replyMessage.value.trim()) return
  sending.value = true
  try {
    const res = await paymentAPI.replyMySupportConversation(id, { message: replyMessage.value.trim() })
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
  loadOrders()
  loadQuickReplies()
})
</script>
