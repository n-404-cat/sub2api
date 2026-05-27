<template>
  <AppLayout>
    <div class="space-y-4">
      <div class="card p-4">
        <div class="flex flex-wrap items-center gap-3">
          <input
            v-model="keyword"
            type="text"
            class="input w-full sm:w-72"
            :placeholder="localText('搜索订单号或会话标题', 'Search order no or conversation subject')"
            @input="debounceLoad"
          />
          <Select v-model="status" :options="statusOptions" class="w-40" @change="loadConversations" />
          <div class="flex flex-1 justify-end">
            <button class="btn btn-secondary" :disabled="loading" @click="loadConversations">
              {{ t('common.refresh') }}
            </button>
          </div>
        </div>
      </div>

      <div class="grid grid-cols-1 gap-4 xl:grid-cols-[360px_minmax(0,1fr)]">
        <div class="card max-h-[70vh] overflow-y-auto p-3">
          <div
            v-for="item in conversations"
            :key="item.id"
            :class="[
              'cursor-pointer rounded-xl border px-4 py-3 transition-colors',
              selectedConversation?.id === item.id
                ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20'
                : 'border-gray-200 hover:bg-gray-50 dark:border-dark-700 dark:hover:bg-dark-800'
            ]"
            @click="openConversation(item.id)"
          >
            <div class="flex items-center justify-between gap-3">
              <p class="truncate text-sm font-semibold text-gray-900 dark:text-white">{{ item.subject || `${localText('订单咨询', 'Order support')} #${item.order_id}` }}</p>
              <span class="text-xs text-gray-400">{{ formatDateTime(item.last_message_at) }}</span>
            </div>
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ localText('订单号', 'Order No') }}: {{ item.order?.out_trade_no || item.order_id }}</p>
          </div>
        </div>

        <div class="card p-5" v-if="selectedDetail?.conversation">
          <div class="space-y-4">
            <div class="rounded-xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-800">
              <div class="flex items-center justify-between gap-4">
                <div>
                  <p class="text-xs text-gray-400 dark:text-gray-500">{{ localText('关联订单', 'Linked order') }}</p>
                  <p class="mt-1 text-sm font-medium text-gray-900 dark:text-white">{{ selectedDetail.conversation.order?.out_trade_no || selectedDetail.conversation.order_id }}</p>
                </div>
                <button class="btn btn-secondary" @click="goOrderDetail(selectedDetail.conversation.order_id)">
                  {{ localText('前往订单管理', 'Go to order management') }}
                </button>
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
                  {{ message.sender_type === 'admin' ? localText('客服', 'Support') : localText('用户', 'User') }}
                  ·
                  {{ formatDateTime(message.created_at) }}
                </div>
                <div class="whitespace-pre-wrap break-words">{{ message.message }}</div>
              </div>
            </div>

            <div class="space-y-3">
              <textarea
                v-model="replyMessage"
                rows="4"
                class="input"
                :placeholder="localText('输入回复内容', 'Type your reply')"
              ></textarea>
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
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { adminSupportAPI } from '@/api/admin'
import type { SupportConversation, SupportConversationDetail } from '@/types/payment'
import AppLayout from '@/components/layout/AppLayout.vue'
import Select from '@/components/common/Select.vue'
import { useAppStore } from '@/stores/app'
import { extractI18nErrorMessage } from '@/utils/apiError'

const { t, locale } = useI18n()
const router = useRouter()
const appStore = useAppStore()

const loading = ref(false)
const sending = ref(false)
const conversations = ref<SupportConversation[]>([])
const selectedConversation = ref<SupportConversation | null>(null)
const selectedDetail = ref<SupportConversationDetail | null>(null)
const keyword = ref('')
const status = ref('')
const replyMessage = ref('')
let debounceTimer: ReturnType<typeof setTimeout> | null = null

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

async function loadConversations() {
  loading.value = true
  try {
    const res = await adminSupportAPI.list({
      keyword: keyword.value || undefined,
      status: status.value || undefined,
    })
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
  const conversation = conversations.value.find(item => item.id === id) || null
  selectedConversation.value = conversation
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
    const res = await adminSupportAPI.reply(selectedConversation.value.id, {
      message: replyMessage.value.trim(),
    })
    selectedDetail.value = res.data
    replyMessage.value = ''
    await loadConversations()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    sending.value = false
  }
}

function goOrderDetail(orderID: number) {
  router.push('/admin/orders')
}

onMounted(() => {
  loadConversations()
})
</script>
