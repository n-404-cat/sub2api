<template>
  <AppLayout>
    <div class="mx-auto max-w-6xl space-y-4">
      <div class="card p-5">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ localText('联系客服', 'Contact support') }}</h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ localText('可以先直接聊天，也可以先选订单再开始。', 'Start chatting directly or choose an order first.') }}</p>
          </div>
          <div class="flex gap-2">
            <button class="btn btn-secondary" @click="startDirectConversation">{{ localText('直接咨询', 'Chat now') }}</button>
            <button class="btn btn-primary" :disabled="!selectedOrder || starting" @click="startOrderConversation">
              {{ starting ? t('common.processing') : localText('选单咨询', 'Order support') }}
            </button>
          </div>
        </div>
      </div>

      <div class="grid grid-cols-1 gap-4 lg:grid-cols-[360px_minmax(0,1fr)]">
        <div class="card p-4">
          <div class="space-y-3">
            <label class="input-label">{{ localText('最近订单', 'Recent orders') }}</label>
            <div class="space-y-2 max-h-[58vh] overflow-y-auto pr-1">
              <button
                v-for="item in orders"
                :key="item.id"
                type="button"
                :class="[
                  'w-full rounded-xl border px-4 py-3 text-left transition-colors',
                  selectedOrder?.id === item.id
                    ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20'
                    : 'border-gray-200 hover:bg-gray-50 dark:border-dark-700 dark:hover:bg-dark-800'
                ]"
                @click="selectedOrder = item"
              >
                <div class="flex items-center justify-between gap-3">
                  <p class="truncate text-sm font-medium text-gray-900 dark:text-white">{{ item.out_trade_no }}</p>
                  <span class="text-xs text-gray-400">#{{ item.id }}</span>
                </div>
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ item.payment_type }}</p>
              </button>
            </div>
          </div>
        </div>

        <div class="card p-4">
          <div class="flex items-center justify-between">
            <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ localText('会话列表', 'Conversations') }}</h3>
            <button class="btn btn-secondary btn-sm" :disabled="loading" @click="loadData">{{ t('common.refresh') }}</button>
          </div>
          <div class="mt-4 space-y-3">
            <div v-if="loading" class="py-10 text-center text-sm text-gray-500 dark:text-gray-400">{{ t('common.loading') }}</div>
            <div v-else-if="conversations.length === 0" class="py-10 text-center text-sm text-gray-500 dark:text-gray-400">{{ localText('暂无会话', 'No conversations yet') }}</div>
            <div v-else class="space-y-3">
              <div
                v-for="conv in conversations"
                :key="conv.id"
                class="cursor-pointer rounded-xl border border-gray-200 px-4 py-3 hover:bg-gray-50 dark:border-dark-700 dark:hover:bg-dark-800"
                @click="router.push(`/support/${conv.id}`)"
              >
                <div class="flex items-center justify-between gap-3">
                  <p class="text-sm font-semibold text-gray-900 dark:text-white">{{ conv.subject || localText('客服咨询', 'Support') }}</p>
                  <span class="text-xs text-gray-400">{{ formatTime(conv.last_message_at) }}</span>
                </div>
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                  {{ conv.order?.out_trade_no ? `${localText('订单号', 'Order No')}: ${conv.order.out_trade_no}` : localText('未关联订单', 'No order linked') }}
                </p>
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
import AppLayout from '@/components/layout/AppLayout.vue'
import { paymentAPI } from '@/api/payment'
import { useAppStore } from '@/stores'
import { extractI18nErrorMessage } from '@/utils/apiError'
import type { PaymentOrder, SupportConversation } from '@/types/payment'

const { t, locale } = useI18n()
const router = useRouter()
const appStore = useAppStore()

const loading = ref(false)
const starting = ref(false)
const orders = ref<PaymentOrder[]>([])
const conversations = ref<SupportConversation[]>([])
const selectedOrder = ref<PaymentOrder | null>(null)

function localText(zh: string, en: string): string {
  return String(locale.value || '').startsWith('zh') ? zh : en
}

function formatTime(value: string): string {
  return new Date(value).toLocaleString()
}

async function loadData() {
  loading.value = true
  try {
    const [ordersRes, convRes] = await Promise.all([
      paymentAPI.getMyOrders({ page: 1, page_size: 20 }),
      paymentAPI.getMySupportConversations(),
    ])
    orders.value = ordersRes.data.items || []
    conversations.value = convRes.data || []
    selectedOrder.value = selectedOrder.value || orders.value[0] || null
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

async function startDirectConversation() {
  starting.value = true
  try {
    const res = await paymentAPI.createSupportConversation({
      subject: localText('客服咨询', 'Support'),
      message: localText('你好，我想咨询一下。', 'Hi, I need help.'),
    })
    await router.push(`/support/${res.data.conversation.id}`)
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    starting.value = false
  }
}

async function startOrderConversation() {
  if (!selectedOrder.value) return
  starting.value = true
  try {
    const res = await paymentAPI.createSupportConversation({
      order_id: selectedOrder.value.id,
      subject: `${localText('订单咨询', 'Order support')} #${selectedOrder.value.id}`,
      message: localText('订单咨询', 'Order support'),
    })
    await router.push(`/support/${res.data.conversation.id}`)
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    starting.value = false
  }
}

onMounted(loadData)
</script>
