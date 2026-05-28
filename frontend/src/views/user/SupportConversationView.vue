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
              {{ linkedOrderSummary }}
            </p>
          </div>
          <div class="flex items-center gap-2">
            <span class="text-xs text-gray-400">{{ autoRefreshEnabled ? localText('自动刷新中', 'Auto refresh on') : localText('自动刷新关闭', 'Auto refresh off') }}</span>
            <button class="btn btn-secondary" @click="router.push('/support')">{{ localText('返回列表', 'Back') }}</button>
          </div>
        </div>
      </div>

      <div v-if="detail?.conversation.order" class="card p-4">
        <div class="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-4">
          <div class="rounded-xl bg-gray-50 p-3 dark:bg-dark-800">
            <p class="text-xs text-gray-400 dark:text-gray-500">{{ localText('订单号', 'Order No') }}</p>
            <p class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">{{ detail.conversation.order.out_trade_no }}</p>
          </div>
          <div class="rounded-xl bg-gray-50 p-3 dark:bg-dark-800">
            <p class="text-xs text-gray-400 dark:text-gray-500">{{ localText('金额', 'Amount') }}</p>
            <p class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">{{ formatAmount(detail.conversation.order) }}</p>
          </div>
          <div class="rounded-xl bg-gray-50 p-3 dark:bg-dark-800">
            <p class="text-xs text-gray-400 dark:text-gray-500">{{ localText('支付方式', 'Payment') }}</p>
            <p class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">{{ detail.conversation.order.payment_type }}</p>
          </div>
          <div class="rounded-xl bg-gray-50 p-3 dark:bg-dark-800">
            <p class="text-xs text-gray-400 dark:text-gray-500">{{ localText('订单状态', 'Status') }}</p>
            <p class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">{{ detail.conversation.order.status }}</p>
          </div>
        </div>
      </div>

      <div class="card p-5">
        <div ref="messagesContainer" class="max-h-[54vh] space-y-3 overflow-y-auto pr-1">
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
            <div v-if="isOrderCard(message.message)" class="space-y-2">
              <div class="text-xs font-semibold opacity-80">{{ localText('已关联订单', 'Linked order') }}</div>
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
      </div>

      <div class="card p-5 space-y-3">
        <div class="flex items-center gap-3 border-b border-gray-100 pb-3 dark:border-dark-700">
          <button class="btn btn-secondary btn-sm" @click="toggleToolPanel('emoji')">
            <Icon name="chatBubble" size="sm" />
            {{ localText('表情', 'Emoji') }}
          </button>
          <button class="btn btn-secondary btn-sm" @click="toggleToolPanel('order')">
            <Icon name="link" size="sm" />
            {{ localText('选择订单', 'Order') }}
          </button>
        </div>

        <div v-if="activeToolPanel === 'emoji'" class="flex flex-wrap gap-2">
          <button
            v-for="emoji in emojis"
            :key="emoji"
            class="rounded-full border border-gray-200 px-3 py-2 text-lg hover:bg-gray-50 dark:border-dark-700 dark:hover:bg-dark-800"
            @click="appendEmoji(emoji)"
          >
            {{ emoji }}
          </button>
        </div>

        <div v-else-if="activeToolPanel === 'order'" class="space-y-2">
          <div
            v-for="order in orders"
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
            <button class="btn btn-primary" :disabled="!pendingOrderId || bindingOrder" @click="confirmBindOrder">
              {{ bindingOrder ? t('common.processing') : localText('确认关联订单', 'Bind order') }}
            </button>
          </div>
        </div>

        <textarea
          v-model="replyMessage"
          rows="4"
          class="input"
          :placeholder="localText('输入消息', 'Type a message')"
        ></textarea>
        <div class="flex items-center justify-between gap-3">
          <button class="btn btn-secondary btn-sm" @click="loadDetail(true)">
            <Icon name="refresh" size="sm" />
            {{ localText('立即刷新', 'Refresh now') }}
          </button>
          <button class="btn btn-primary" :disabled="sending || !replyMessage.trim()" @click="sendReply">
            {{ sending ? t('common.processing') : localText('发送消息', 'Send message') }}
          </button>
        </div>
      </div>
    </div>
  </AppLayout>

  <BaseDialog :show="showOrderCardPreview" :title="localText('订单详情', 'Order detail')" width="wide" @close="showOrderCardPreview = false">
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
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { paymentAPI } from '@/api/payment'
import { useAppStore } from '@/stores'
import { extractI18nErrorMessage } from '@/utils/apiError'
import type { PaymentOrder, SupportConversationDetail } from '@/types/payment'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

const { t, locale } = useI18n()
const route = useRoute()
const router = useRouter()
const appStore = useAppStore()

const detail = ref<SupportConversationDetail | null>(null)
const replyMessage = ref('')
const sending = ref(false)
const bindingOrder = ref(false)
const orders = ref<PaymentOrder[]>([])
const pendingOrderId = ref<number | null>(null)
const messagesContainer = ref<HTMLElement | null>(null)
const activeToolPanel = ref<'emoji' | 'order' | null>(null)
const showOrderCardPreview = ref(false)
const selectedOrderCardLines = ref<string[]>([])
const emojis = ['😀', '😁', '😂', '😅', '😍', '🙏', '👍', '🎉', '📦', '✅', '⌛', '💬']

function localText(zh: string, en: string): string {
  return String(locale.value || '').startsWith('zh') ? zh : en
}

const linkedOrderSummary = computed(() => {
  const order = detail.value?.conversation.order
  if (!order) return localText('当前未关联订单，可在下方工具栏里选择订单。', 'No order linked. You can bind one from the toolbar below.')
  return `${localText('已关联订单', 'Linked order')}: ${order.out_trade_no} · ${formatAmount(order)} · ${order.status}`
})

const autoRefresh = useAutoRefresh({
  storageKey: 'support-conversation-auto-refresh',
  defaultInterval: 5,
  onRefresh: () => loadDetail(false),
  shouldPause: () => sending.value || bindingOrder.value,
})

const autoRefreshEnabled = computed(() => autoRefresh.enabled.value)

function formatDateTime(value: string): string {
  return new Date(value).toLocaleString()
}

function formatShortTime(value: string): string {
  return new Date(value).toLocaleString([], { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
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

function appendEmoji(emoji: string) {
  replyMessage.value += emoji
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

function toggleToolPanel(panel: 'emoji' | 'order') {
  activeToolPanel.value = activeToolPanel.value === panel ? null : panel
}

async function scrollToBottom() {
  await nextTick()
  if (!messagesContainer.value) return
  messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
}

async function loadDetail(resetCountdown = true) {
  const id = Number(route.params.id)
  if (!id) return
  try {
    const res = await paymentAPI.getMySupportConversation(id)
    detail.value = res.data
    if (resetCountdown) autoRefresh.resetCountdown()
    await scrollToBottom()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  }
}

async function loadOrders() {
  try {
    const res = await paymentAPI.getMyOrders({ page: 1, page_size: 50 })
    orders.value = res.data.items || []
  } catch {}
}

async function confirmBindOrder() {
  const id = Number(route.params.id)
  if (!id || !pendingOrderId.value) return
  bindingOrder.value = true
  try {
    const res = await paymentAPI.bindSupportConversationOrder(id, { order_id: pendingOrderId.value })
    detail.value = res.data
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
  const id = Number(route.params.id)
  if (!id || !replyMessage.value.trim()) return
  sending.value = true
  try {
    const res = await paymentAPI.replyMySupportConversation(id, { message: replyMessage.value.trim() })
    detail.value = res.data
    replyMessage.value = ''
    await scrollToBottom()
    autoRefresh.resetCountdown()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    sending.value = false
  }
}

watch(() => detail.value?.messages?.length, () => {
  scrollToBottom()
})

onMounted(async () => {
  await Promise.all([loadDetail(), loadOrders()])
  autoRefresh.setEnabled(true)
  autoRefresh.start()
})
</script>
