<template>
  <AppLayout>
    <div class="space-y-4">
      <div class="card p-4">
        <div class="flex flex-wrap items-center gap-3">
          <input v-model="keyword" class="input w-full sm:w-72" :placeholder="localText('搜索订单号或会话标题', 'Search')" @input="debounceLoad" />
          <Select v-model="status" :options="statusOptions" class="w-40" @change="loadConversations" />
          <div class="flex flex-1 justify-end gap-2">
            <button class="btn btn-secondary" @click="loadConversations">{{ t('common.refresh') }}</button>
          </div>
        </div>
      </div>

      <div class="grid grid-cols-1 gap-4 xl:grid-cols-[360px_minmax(0,1fr)]">
        <div class="card max-h-[70vh] overflow-y-auto p-3">
          <div
            v-for="item in conversations"
            :key="item.id"
            class="cursor-pointer rounded-xl border px-4 py-3 transition-colors"
            :class="selectedConversation?.id === item.id ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20' : 'border-gray-200 hover:bg-gray-50 dark:border-dark-700 dark:hover:bg-dark-800'"
            @click="openConversation(item.id)"
          >
            <div class="flex items-center justify-between gap-3">
              <p class="truncate text-sm font-semibold text-gray-900 dark:text-white">{{ item.subject || localText('客服咨询', 'Support') }}</p>
              <span class="text-xs text-gray-400">{{ formatDateTime(item.last_message_at) }}</span>
            </div>
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
              {{ item.order?.out_trade_no ? `${localText('订单号', 'Order No')}: ${item.order.out_trade_no}` : localText('未关联订单', 'No order linked') }}
            </p>
          </div>
        </div>

        <div class="card p-5" v-if="selectedDetail?.conversation">
          <div class="space-y-4">
            <div class="rounded-xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-800">
              <div class="flex items-center justify-between gap-4">
                <div>
                  <p class="text-xs text-gray-400 dark:text-gray-500">{{ localText('会话', 'Conversation') }}</p>
                  <p class="mt-1 text-sm font-medium text-gray-900 dark:text-white">{{ selectedDetail.conversation.subject || localText('客服咨询', 'Support') }}</p>
                  <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ localText('用户ID', 'User ID') }} #{{ selectedDetail.conversation.user_id }}</p>
                </div>
                <button class="btn btn-secondary" @click="showOrderPreview = true">{{ localText('查看订单详情', 'Order detail') }}</button>
              </div>
            </div>

            <div class="max-h-[45vh] space-y-3 overflow-y-auto">
              <div
                v-for="message in selectedDetail.messages"
                :key="message.id"
                :class="[
                  'rounded-2xl px-4 py-3 text-sm',
                  message.sender_type === 'admin'
                    ? 'ml-auto max-w-[85%] bg-primary-50 text-primary-900 dark:bg-primary-900/25 dark:text-primary-100'
                    : 'mr-auto max-w-[85%] bg-gray-100 text-gray-800 dark:bg-dark-700 dark:text-gray-100'
                ]"
              >
                <div class="mb-1 text-xs opacity-70">
                  {{ message.sender_type === 'admin' ? localText('客服', 'Support') : localText('用户', 'User') }} · {{ formatDateTime(message.created_at) }}
                </div>
                <div class="whitespace-pre-wrap break-words">{{ message.message }}</div>
              </div>
            </div>

            <div class="space-y-3">
              <div class="flex flex-wrap gap-2">
                <button v-for="emoji in emojis" :key="emoji" class="rounded-full border border-gray-200 px-3 py-1 text-sm hover:bg-gray-50 dark:border-dark-700 dark:hover:bg-dark-800" @click="replyMessage += emoji">{{ emoji }}</button>
              </div>
              <div v-if="quickReplies.length" class="flex flex-wrap gap-2">
                <button v-for="reply in quickReplies" :key="reply" class="rounded-full bg-primary-50 px-3 py-1 text-sm text-primary-700 hover:bg-primary-100 dark:bg-primary-900/20 dark:text-primary-200" @click="insertQuickReply(reply)">{{ reply }}</button>
              </div>
              <textarea v-model="replyMessage" rows="4" class="input" :placeholder="localText('输入回复内容', 'Type your reply')"></textarea>
              <div class="flex justify-end">
                <button class="btn btn-primary" :disabled="sending || !replyMessage.trim()" @click="sendReply">
                  {{ sending ? t('common.processing') : localText('发送回复', 'Send reply') }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>

  <BaseDialog :show="showOrderPreview" :title="localText('订单详情', 'Order detail')" width="wide" @close="showOrderPreview = false">
    <div v-if="selectedDetail?.conversation?.order" class="space-y-4">
      <div class="grid grid-cols-2 gap-4">
        <div><p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.orders.orderId') }}</p><p class="font-mono text-sm font-medium text-gray-900 dark:text-white">#{{ selectedDetail.conversation.order.id }}</p></div>
        <div><p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.orders.orderNo') }}</p><p class="text-sm font-medium text-gray-900 dark:text-white">{{ selectedDetail.conversation.order.out_trade_no }}</p></div>
      </div>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminSupportAPI } from '@/api/admin'
import { usePaymentStore } from '@/stores/payment'
import { useAppStore } from '@/stores/app'
import { extractI18nErrorMessage } from '@/utils/apiError'
import type { SupportConversation, SupportConversationDetail } from '@/types/payment'
import AppLayout from '@/components/layout/AppLayout.vue'
import Select from '@/components/common/Select.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'

const { t, locale } = useI18n()
const appStore = useAppStore()
const paymentStore = usePaymentStore()

const loading = ref(false)
const sending = ref(false)
const conversations = ref<SupportConversation[]>([])
const selectedConversation = ref<SupportConversation | null>(null)
const selectedDetail = ref<SupportConversationDetail | null>(null)
const keyword = ref('')
const status = ref('')
const replyMessage = ref('')
const showOrderPreview = ref(false)
const emojis = ['😀', '😍', '🙏', '👍', '🎉', '💬', '📦', '✅']
let debounceTimer: ReturnType<typeof setTimeout> | null = null

const quickReplies = computed(() => paymentStore.config?.support_quick_replies || [])
const statusOptions = [
  { value: '', label: '全部状态' },
  { value: 'open', label: '进行中' },
  { value: 'closed', label: '已关闭' },
]

function localText(zh: string, en: string): string {
  return String(locale.value || '').startsWith('zh') ? zh : en
}

function formatDateTime(value: string): string {
  return new Date(value).toLocaleString()
}

function debounceLoad() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => loadConversations(), 300)
}

function insertQuickReply(text: string) {
  replyMessage.value = replyMessage.value ? `${replyMessage.value}\n${text}` : text
}

async function loadConversations() {
  loading.value = true
  try {
    const res = await adminSupportAPI.list({ keyword: keyword.value || undefined, status: status.value || undefined })
    conversations.value = res.data.items || []
    if (!selectedConversation.value && conversations.value.length > 0) {
      await openConversation(conversations.value[0].id)
    }
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

async function openConversation(id: number) {
  selectedConversation.value = conversations.value.find(item => item.id === id) || null
  try {
    const res = await adminSupportAPI.get(id)
    selectedDetail.value = res.data
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  }
}

async function sendReply() {
  if (!selectedConversation.value || !replyMessage.value.trim()) return
  sending.value = true
  try {
    const res = await adminSupportAPI.reply(selectedConversation.value.id, { message: replyMessage.value.trim() })
    selectedDetail.value = res.data
    replyMessage.value = ''
    await loadConversations()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    sending.value = false
  }
}

onMounted(() => {
  paymentStore.fetchConfig()
  loadConversations()
})
</script>
