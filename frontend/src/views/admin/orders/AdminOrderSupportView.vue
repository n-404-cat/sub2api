<template>
  <AppLayout>
    <div class="space-y-4">
      <div class="card p-4">
        <div class="flex flex-wrap items-center gap-3">
          <input
            v-model="keyword"
            class="input w-full sm:w-72"
            :placeholder="localText('搜索订单号、邮箱、会话标题', 'Search order, email or subject')"
            @input="debounceLoad"
          />
          <Select v-model="status" :options="statusOptions" class="w-40" @change="loadConversations" />
          <div class="flex flex-1 items-center justify-end gap-2">
            <span class="text-xs text-gray-400">{{ autoRefresh.enabled.value ? localText('自动刷新中', 'Auto refresh on') : localText('自动刷新关闭', 'Auto refresh off') }}</span>
            <button class="btn btn-secondary" @click="goQuickReplyConfig">{{ localText('配置快捷回复', 'Configure quick replies') }}</button>
            <button class="btn btn-secondary" @click="loadConversations">{{ t('common.refresh') }}</button>
          </div>
        </div>
      </div>

      <div class="grid grid-cols-1 gap-4 xl:grid-cols-[360px_minmax(0,1fr)]">
        <div class="card max-h-[72vh] overflow-y-auto p-3">
          <div
            v-for="item in conversations"
            :key="item.id"
            class="cursor-pointer rounded-xl border px-4 py-3 transition-colors"
            :class="[
              selectedConversation?.id === item.id ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20' : 'border-gray-200 hover:bg-gray-50 dark:border-dark-700 dark:hover:bg-dark-800',
              item.unread_count ? 'ring-1 ring-red-200 dark:ring-red-900/40' : ''
            ]"
            @click="openConversation(item.id)"
          >
            <div class="flex items-center justify-between gap-3">
              <div class="min-w-0 flex-1">
                <p class="truncate text-sm font-semibold text-gray-900 dark:text-white">{{ item.subject || localText('客服咨询', 'Support') }}</p>
                <p class="mt-1 truncate text-xs" :class="item.unread_count ? 'font-medium text-red-600 dark:text-red-300' : 'text-gray-500 dark:text-gray-400'">
                  {{ conversationPreview(item) }}
                </p>
              </div>
              <div class="flex flex-col items-end gap-1">
                <span class="text-xs text-gray-400">{{ formatDateTime(item.last_message_at) }}</span>
                <span
                  v-if="item.unread_count"
                  class="rounded-full bg-red-500 px-1.5 text-[10px] font-semibold leading-[18px] text-white"
                >
                  {{ item.unread_count > 99 ? '99+' : item.unread_count }}
                </span>
              </div>
            </div>
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
              {{ item.order?.out_trade_no ? `${localText('订单号', 'Order No')}: ${item.order.out_trade_no}` : localText('未关联订单', 'No order linked') }}
            </p>
            <p class="mt-1 text-xs text-gray-400 dark:text-gray-500">
              {{ localText('用户ID', 'User ID') }} #{{ item.user_id }}
            </p>
          </div>
        </div>

        <div class="card p-5" v-if="selectedDetail?.conversation">
          <div class="space-y-4">
            <div class="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-4">
              <div class="rounded-xl bg-gray-50 p-3 dark:bg-dark-800">
                <p class="text-xs text-gray-400 dark:text-gray-500">{{ localText('会话ID', 'Conversation ID') }}</p>
                <p class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">#{{ selectedDetail.conversation.id }}</p>
              </div>
              <div class="rounded-xl bg-gray-50 p-3 dark:bg-dark-800">
                <p class="text-xs text-gray-400 dark:text-gray-500">{{ localText('用户ID', 'User ID') }}</p>
                <button class="mt-1 text-left text-sm font-semibold text-primary-700 hover:underline dark:text-primary-300" @click="openUserPreview(selectedDetail.conversation.user_id)">
                  #{{ selectedDetail.conversation.user_id }}
                </button>
              </div>
              <div class="rounded-xl bg-gray-50 p-3 dark:bg-dark-800">
                <p class="text-xs text-gray-400 dark:text-gray-500">{{ localText('会话状态', 'Status') }}</p>
                <p class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">{{ selectedDetail.conversation.status }}</p>
              </div>
              <div class="rounded-xl bg-gray-50 p-3 dark:bg-dark-800">
                <p class="text-xs text-gray-400 dark:text-gray-500">{{ localText('最近消息', 'Last message') }}</p>
                <p class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">{{ formatDateTime(selectedDetail.conversation.last_message_at) }}</p>
              </div>
            </div>

            <div v-if="selectedDetail.conversation.order" class="rounded-2xl border border-gray-200 p-4 dark:border-dark-700">
              <div class="mb-3 flex items-center justify-between">
                <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ localText('订单详情', 'Order detail') }}</h3>
                <button class="btn btn-secondary btn-sm" @click="showOrderPreview = true">{{ localText('展开查看', 'Preview') }}</button>
              </div>
              <div class="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-4">
                <div><p class="text-xs text-gray-400 dark:text-gray-500">{{ localText('订单号', 'Order No') }}</p><p class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">{{ selectedDetail.conversation.order.out_trade_no }}</p></div>
                <div><p class="text-xs text-gray-400 dark:text-gray-500">{{ localText('金额', 'Amount') }}</p><p class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">{{ formatAmount(selectedDetail.conversation.order) }}</p></div>
                <div><p class="text-xs text-gray-400 dark:text-gray-500">{{ localText('支付方式', 'Payment') }}</p><p class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">{{ formatPaymentType(selectedDetail.conversation.order.payment_type) }}</p></div>
                <div><p class="text-xs text-gray-400 dark:text-gray-500">{{ localText('订单状态', 'Order status') }}</p><p class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">{{ formatStatus(selectedDetail.conversation.order.status) }}</p></div>
              </div>
            </div>

            <div ref="messagesContainer" class="max-h-[44vh] space-y-3 overflow-y-auto pr-1">
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
                <div v-if="isOrderCard(message.message)" class="space-y-2">
                  <div class="text-xs font-semibold opacity-80">{{ localText('订单卡片', 'Order card') }}</div>
                  <button
                    class="w-full rounded-xl border border-current/10 bg-white/40 p-3 text-left transition hover:bg-white/60 dark:bg-black/10 dark:hover:bg-black/20"
                    @click="openOrderCardPreview(message.message)"
                  >
                    <div
                      v-for="line in parseOrderCard(message.message)"
                      :key="line"
                      class="whitespace-pre-wrap break-words text-sm"
                    >
                      {{ line }}
                    </div>
                  </button>
                </div>
                <div v-else class="whitespace-pre-wrap break-words">{{ message.message }}</div>
              </div>
            </div>

            <div class="space-y-3">
              <div class="flex items-center gap-3 border-b border-gray-100 pb-3 dark:border-dark-700">
                <button class="btn btn-secondary btn-sm" @click="toggleToolPanel('emoji')">
                  {{ localText('表情', 'Emoji') }}
                </button>
                <button class="btn btn-secondary btn-sm" @click="toggleToolPanel('order')">
                  {{ localText('选择订单', 'Order') }}
                </button>
                <button class="btn btn-secondary btn-sm" @click="toggleToolPanel('quickReply')">
                  {{ localText('快捷回复', 'Quick reply') }}
                </button>
              </div>

              <div v-if="activeToolPanel === 'emoji'" class="flex flex-wrap gap-2">
                <button
                  v-for="emoji in emojis"
                  :key="emoji"
                  class="rounded-full border border-gray-200 px-3 py-2 text-lg hover:bg-gray-50 dark:border-dark-700 dark:hover:bg-dark-800"
                  @click="replyMessage += emoji"
                >
                  {{ emoji }}
                </button>
              </div>

              <div v-else-if="activeToolPanel === 'order'" class="space-y-2">
                <div
                  v-for="order in relatedOrders"
                  :key="order.id"
                  :class="[
                    'cursor-pointer rounded-2xl border p-4 transition-colors',
                    pendingOrderId === order.id
                      ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20'
                      : 'border-gray-200 hover:bg-gray-50 dark:border-dark-700 dark:hover:bg-dark-800'
                  ]"
                  @click="pendingOrderId = order.id"
                >
                  <div class="flex items-center justify-between gap-3">
                    <p class="text-sm font-semibold text-gray-900 dark:text-white">{{ order.out_trade_no }}</p>
                    <span class="text-xs text-gray-400">#{{ order.id }}</span>
                  </div>
                  <div class="mt-2 grid grid-cols-2 gap-2 text-xs text-gray-500 dark:text-gray-400">
                    <span>{{ localText('金额', 'Amount') }}: {{ formatAmount(order) }}</span>
                    <span>{{ localText('状态', 'Status') }}: {{ order.status }}</span>
                    <span>{{ localText('方式', 'Method') }}: {{ order.payment_type }}</span>
                    <span>{{ localText('时间', 'Created') }}: {{ formatShortTime(order.created_at) }}</span>
                  </div>
                </div>
                <div class="flex justify-end">
                  <button class="btn btn-primary" :disabled="!pendingOrderId || bindingOrder" @click="bindOrderForConversation">
                    {{ bindingOrder ? t('common.processing') : localText('确认关联订单', 'Bind order') }}
                  </button>
                </div>
              </div>

              <div v-else-if="activeToolPanel === 'quickReply'" class="flex flex-wrap gap-2">
                <button
                  v-for="reply in quickReplies"
                  :key="reply"
                  class="rounded-full bg-primary-50 px-3 py-2 text-sm text-primary-700 hover:bg-primary-100 dark:bg-primary-900/20 dark:text-primary-200"
                  @click="insertQuickReply(reply)"
                >
                  {{ reply }}
                </button>
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

  <AdminOrderDetail
    :show="showOrderPreview"
    :order="selectedDetail?.conversation?.order || null"
    @close="showOrderPreview = false"
    @cancel="() => {}"
    @retry="() => {}"
    @refund="() => {}"
  />

  <BaseDialog :show="showOrderCardPreview" :title="localText('订单卡片详情', 'Order card detail')" width="wide" @close="showOrderCardPreview = false">
    <div class="space-y-3">
      <div
        v-for="line in selectedOrderCardLines"
        :key="line"
        class="rounded-xl bg-gray-50 px-4 py-3 text-sm text-gray-700 dark:bg-dark-800 dark:text-gray-200"
      >
        {{ line }}
      </div>
    </div>
  </BaseDialog>

  <BaseDialog :show="showUserPreview" :title="localText('用户信息', 'User info')" width="wide" @close="showUserPreview = false">
    <div v-if="selectedUser" class="grid grid-cols-1 gap-3 md:grid-cols-2">
      <div><p class="text-xs text-gray-500 dark:text-gray-400">ID</p><p class="text-sm font-medium text-gray-900 dark:text-white">#{{ selectedUser.id }}</p></div>
      <div><p class="text-xs text-gray-500 dark:text-gray-400">{{ localText('邮箱', 'Email') }}</p><p class="text-sm font-medium text-gray-900 dark:text-white">{{ selectedUser.email || '-' }}</p></div>
      <div><p class="text-xs text-gray-500 dark:text-gray-400">{{ localText('用户名', 'Username') }}</p><p class="text-sm font-medium text-gray-900 dark:text-white">{{ selectedUser.username || '-' }}</p></div>
      <div><p class="text-xs text-gray-500 dark:text-gray-400">{{ localText('角色', 'Role') }}</p><p class="text-sm font-medium text-gray-900 dark:text-white">{{ selectedUser.role || '-' }}</p></div>
      <div><p class="text-xs text-gray-500 dark:text-gray-400">{{ localText('状态', 'Status') }}</p><p class="text-sm font-medium text-gray-900 dark:text-white">{{ selectedUser.status || '-' }}</p></div>
      <div><p class="text-xs text-gray-500 dark:text-gray-400">{{ localText('余额', 'Balance') }}</p><p class="text-sm font-medium text-gray-900 dark:text-white">{{ selectedUser.balance ?? '-' }}</p></div>
      <div><p class="text-xs text-gray-500 dark:text-gray-400">{{ localText('并发', 'Concurrency') }}</p><p class="text-sm font-medium text-gray-900 dark:text-white">{{ selectedUser.concurrency ?? '-' }}</p></div>
      <div><p class="text-xs text-gray-500 dark:text-gray-400">{{ localText('注册时间', 'Created at') }}</p><p class="text-sm font-medium text-gray-900 dark:text-white">{{ selectedUser.created_at || '-' }}</p></div>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminSupportAPI } from '@/api/admin'
import { adminPaymentAPI } from '@/api/admin/payment'
import * as adminUsersAPI from '@/api/admin/users'
import { usePaymentStore } from '@/stores/payment'
import { useAppStore } from '@/stores/app'
import { useSupportStore } from '@/stores/support'
import { extractI18nErrorMessage } from '@/utils/apiError'
import type { PaymentOrder, SupportConversation, SupportConversationDetail } from '@/types/payment'
import AppLayout from '@/components/layout/AppLayout.vue'
import Select from '@/components/common/Select.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import AdminOrderDetail from '@/components/admin/payment/AdminOrderDetail.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

const { t, locale } = useI18n()
const appStore = useAppStore()
const paymentStore = usePaymentStore()
const supportStore = useSupportStore()

const loading = ref(false)
const sending = ref(false)
const conversations = ref<SupportConversation[]>([])
const selectedConversation = ref<SupportConversation | null>(null)
const selectedDetail = ref<SupportConversationDetail | null>(null)
const keyword = ref('')
const status = ref('')
const replyMessage = ref('')
const showOrderPreview = ref(false)
const bindingOrder = ref(false)
const pendingOrderId = ref<number | null>(null)
const relatedOrders = ref<PaymentOrder[]>([])
const activeToolPanel = ref<'emoji' | 'order' | 'quickReply' | null>(null)
const messagesContainer = ref<HTMLElement | null>(null)
const showOrderCardPreview = ref(false)
const selectedOrderCardLines = ref<string[]>([])
const showUserPreview = ref(false)
const selectedUser = ref<Record<string, any> | null>(null)
let debounceTimer: ReturnType<typeof setTimeout> | null = null

const quickReplies = computed(() => paymentStore.config?.support_quick_replies || [])
const emojis = ['😀', '😁', '😂', '😅', '😍', '🙏', '👍', '🎉', '📦', '✅', '⌛', '💬']
const statusOptions = [
  { value: '', label: '全部状态' },
  { value: 'open', label: '进行中' },
  { value: 'closed', label: '已关闭' },
]

const autoRefresh = useAutoRefresh({
  storageKey: 'admin-support-auto-refresh',
  defaultInterval: 5,
  onRefresh: () => selectedConversation.value ? openConversation(selectedConversation.value.id) : loadConversations(),
  shouldPause: () => sending.value,
})

function localText(zh: string, en: string): string {
  return String(locale.value || '').startsWith('zh') ? zh : en
}

function formatDateTime(value: string): string {
  return new Date(value).toLocaleString()
}

function formatAmount(order: PaymentOrder): string {
  const prefix = order.order_type === 'balance' ? '$' : '¥'
  return `${prefix}${order.amount.toFixed(2)}`
}

function formatStatus(status: string): string {
  const map: Record<string, string> = {
    PENDING: '待支付',
    PAID: '已支付',
    RECHARGING: '充值中',
    COMPLETED: '已完成',
    EXPIRED: '已过期',
    CANCELLED: '已取消',
    FAILED: '失败',
    REFUND_REQUESTED: '退款申请中',
    REFUNDING: '退款中',
    PARTIALLY_REFUNDED: '部分退款',
    REFUNDED: '已退款',
    REFUND_FAILED: '退款失败',
  }
  return map[status] || status
}

function formatPaymentType(type: string): string {
  const map: Record<string, string> = {
    alipay: '支付宝',
    wxpay: '微信支付',
    alipay_direct: '支付宝直连',
    wxpay_direct: '微信直连',
    stripe: 'Stripe',
    easypay: 'EasyPay',
    airwallex: 'Airwallex',
  }
  return map[type] || type
}

function formatShortTime(value: string): string {
  return new Date(value).toLocaleString([], { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

function conversationPreview(item: SupportConversation): string {
  const prefix = item.last_message_sender_type === 'admin' ? localText('客服', 'Support') : localText('用户', 'User')
  const preview = item.last_message_preview?.replace('[ORDER_CARD]\n', '').split('\n')[0] || localText('暂无消息', 'No messages')
  return `${prefix}: ${preview}`
}

function debounceLoad() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => loadConversations(), 300)
}

function toggleToolPanel(panel: 'emoji' | 'order' | 'quickReply') {
  activeToolPanel.value = activeToolPanel.value === panel ? null : panel
}

function insertQuickReply(text: string) {
  replyMessage.value = replyMessage.value ? `${replyMessage.value}\n${text}` : text
}

function goQuickReplyConfig() {
  if (typeof window !== 'undefined') {
    window.location.href = '/admin/orders/dashboard'
  }
}

function isOrderCard(message: string): boolean {
  return message.startsWith('[ORDER_CARD]')
}

function parseOrderCard(message: string): string[] {
  return message
    .replace('[ORDER_CARD]\n', '')
    .replace(/状态: (.+)/, (_, status) => `状态: ${formatStatus(String(status))}`)
    .replace(/支付方式: (.+)/, (_, type) => `支付方式: ${formatPaymentType(String(type))}`)
    .split('\n')
    .filter(Boolean)
}

function openOrderCardPreview(message: string) {
  selectedOrderCardLines.value = parseOrderCard(message)
  showOrderCardPreview.value = true
}

async function openUserPreview(userId: number) {
  try {
    selectedUser.value = await adminUsersAPI.getById(userId)
    showUserPreview.value = true
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'common', t('common.error')))
  }
}

async function scrollToBottom() {
  await nextTick()
  if (!messagesContainer.value) return
  messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
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
    supportStore.markConversationSeen(id, res.data.conversation.last_message_at)
    pendingOrderId.value = res.data.conversation.order_id ?? null
    await loadRelatedOrders(res.data.conversation.user_id)
    autoRefresh.resetCountdown()
    await scrollToBottom()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  }
}

async function loadRelatedOrders(userId: number) {
  try {
    const res = await adminPaymentAPI.getOrders({ page: 1, page_size: 50, user_id: userId })
    relatedOrders.value = res.data.items || []
  } catch {
    relatedOrders.value = []
  }
}

async function bindOrderForConversation() {
  if (!selectedConversation.value || !pendingOrderId.value) return
  bindingOrder.value = true
  try {
    const res = await adminSupportAPI.bindOrder(selectedConversation.value.id, { order_id: pendingOrderId.value })
    selectedDetail.value = res.data
    activeToolPanel.value = null
    pendingOrderId.value = null
    await scrollToBottom()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    bindingOrder.value = false
  }
}

async function sendReply() {
  if (!selectedConversation.value || !replyMessage.value.trim()) return
  sending.value = true
  try {
    const res = await adminSupportAPI.reply(selectedConversation.value.id, { message: replyMessage.value.trim() })
    selectedDetail.value = res.data
    replyMessage.value = ''
    await scrollToBottom()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    sending.value = false
  }
}

watch(() => selectedDetail.value?.messages?.length, () => {
  scrollToBottom()
})

onMounted(async () => {
  await paymentStore.fetchConfig()
  await supportStore.fetchConversations(true)
  await loadConversations()
  autoRefresh.setEnabled(true)
  autoRefresh.start()
})
</script>
