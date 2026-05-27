<template>
  <AppLayout>
    <div class="space-y-4">
      <!-- Filters -->
      <div class="card p-4">
        <div class="flex flex-wrap items-center gap-3">
          <div class="flex-1 sm:max-w-64">
            <input v-model="orderSearch" type="text" :placeholder="t('payment.admin.searchOrders')" class="input" @input="debounceLoadOrders" />
          </div>
          <Select v-model="orderFilters.status" :options="statusFilterOptions" class="w-36" @change="loadOrders" />
          <Select v-model="orderFilters.payment_type" :options="paymentTypeFilterOptions" class="w-40" @change="loadOrders" />
          <Select v-model="orderFilters.order_type" :options="orderTypeFilterOptions" class="w-52" @change="loadOrders" />
          <div class="flex flex-1 flex-wrap items-center justify-end gap-2">
            <button @click="loadOrders" :disabled="ordersLoading" class="btn btn-secondary" :title="t('common.refresh')">
              <Icon name="refresh" size="md" :class="ordersLoading ? 'animate-spin' : ''" />
            </button>
          </div>
        </div>
        <p class="mt-3 text-xs text-gray-500 dark:text-gray-400">
          {{ localText('人工充值订单会归类在支付宝/微信支付方式下，并带有“人工”标记。', 'Manual top-up orders are listed under Alipay or WeChat Pay and marked as Manual.') }}
        </p>
      </div>

      <!-- Table -->
      <OrderTable :orders="orders" :loading="ordersLoading" show-user>
        <template #actions="{ row }">
          <div class="flex items-center gap-1">
            <button @click="showOrderDetail(row)" class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-dark-600">
              <Icon name="eye" size="sm" />
              {{ t('common.view') }}
            </button>
            <button v-if="row.status === 'PENDING'" @click="handleCancelOrder(row)" class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-yellow-600 hover:bg-yellow-50 dark:text-yellow-400 dark:hover:bg-yellow-900/20">
              <Icon name="x" size="sm" />
              {{ t('payment.orders.cancel') }}
            </button>
            <button v-if="row.status === 'FAILED'" @click="handleRetryOrder(row)" class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-blue-600 hover:bg-blue-50 dark:text-blue-400 dark:hover:bg-blue-900/20">
              <Icon name="refresh" size="sm" />
              {{ t('payment.admin.retry') }}
            </button>
            <button
              v-if="row.status === 'PENDING' && row.manual_payment?.enabled && row.manual_payment.review_status === 'PENDING_ADMIN_REVIEW'"
              @click="reviewManualPayment(row, true)"
              class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-green-600 hover:bg-green-50 dark:text-green-400 dark:hover:bg-green-900/20"
            >
              <Icon name="check" size="sm" />
              {{ localText('通过', 'Approve') }}
            </button>
            <button
              v-if="row.status === 'PENDING' && row.manual_payment?.enabled && row.manual_payment.review_status === 'PENDING_ADMIN_REVIEW'"
              @click="reviewManualPayment(row, false)"
              class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-900/20"
            >
              <Icon name="x" size="sm" />
              {{ localText('拒绝', 'Reject') }}
            </button>
            <template v-if="row.status === 'REFUND_REQUESTED'">
              <span v-if="row.refund_amount" class="rounded-full bg-purple-100 px-1.5 py-0.5 text-xs font-medium text-purple-700 dark:bg-purple-900/30 dark:text-purple-300">{{ row.order_type === 'balance' ? '$' : '¥' }}{{ row.refund_amount.toFixed(2) }}</span>
              <button @click="openRefundDialog(row)" class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-purple-600 hover:bg-purple-50 dark:text-purple-400 dark:hover:bg-purple-900/20">
                <Icon name="check" size="sm" />
                {{ t('payment.admin.approveRefund') }}
              </button>
            </template>
            <button v-else-if="row.status === 'REFUND_FAILED'" @click="openRefundDialog(row)" class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-purple-600 hover:bg-purple-50 dark:text-purple-400 dark:hover:bg-purple-900/20">
              <Icon name="refresh" size="sm" />
              {{ t('payment.admin.retryRefund') }}
            </button>
            <button v-else-if="row.status === 'COMPLETED' || row.status === 'PARTIALLY_REFUNDED'" @click="openRefundDialog(row)" class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-900/20">
              <Icon name="dollar" size="sm" />
              {{ t('payment.admin.refund') }}
            </button>
          </div>
        </template>
      </OrderTable>
      <Pagination v-if="orderPagination.total > 0" :page="orderPagination.page" :total="orderPagination.total" :page-size="orderPagination.page_size" @update:page="handleOrderPageChange" @update:pageSize="handleOrderPageSizeChange" />
    </div>

    <!-- Order Detail Dialog -->
    <BaseDialog :show="showDetailDialog" :title="t('payment.admin.orderDetail')" width="wide" @close="showDetailDialog = false">
      <div v-if="selectedOrder" class="space-y-4">
        <div class="grid grid-cols-2 gap-4">
          <div><p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.orders.orderId') }}</p><p class="font-mono text-sm font-medium text-gray-900 dark:text-white">#{{ selectedOrder.id }}</p></div>
          <div><p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.orders.orderNo') }}</p><p class="text-sm font-medium text-gray-900 dark:text-white">{{ selectedOrder.out_trade_no }}</p></div>
          <div><p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.orders.status') }}</p><OrderStatusBadge :status="selectedOrder.status" /></div>
          <div><p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.orders.amount') }}</p><p class="text-sm font-medium text-gray-900 dark:text-white">{{ selectedOrder.order_type === 'balance' ? '$' : '¥' }}{{ selectedOrder.amount.toFixed(2) }}</p></div>
          <div><p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.orders.payAmount') }}</p><p class="text-sm font-medium text-gray-900 dark:text-white">{{ formatMoneyWithUnit(selectedOrder.pay_amount, selectedOrder.order_type) }}</p></div>
          <div><p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.orders.paymentMethod') }}</p><p class="text-sm text-gray-700 dark:text-gray-300">{{ t('payment.methods.' + selectedOrder.payment_type, selectedOrder.payment_type) }}<span v-if="selectedOrder.manual_payment?.enabled" class="ml-2 rounded-full bg-amber-100 px-2 py-0.5 text-[11px] font-medium text-amber-700 dark:bg-amber-900/30 dark:text-amber-300">{{ localText('人工', 'Manual') }}</span></p></div>
          <div><p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.feeRate') }}</p><p class="text-sm text-gray-700 dark:text-gray-300">{{ selectedOrder.fee_rate }}%</p></div>
          <div><p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.orders.createdAt') }}</p><p class="text-sm text-gray-700 dark:text-gray-300">{{ formatDateTime(selectedOrder.created_at) }}</p></div>
          <div><p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.expiresAt') }}</p><p class="text-sm text-gray-700 dark:text-gray-300">{{ formatDateTime(selectedOrder.expires_at) }}</p></div>
          <div v-if="selectedOrder.paid_at"><p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.paidAt') }}</p><p class="text-sm text-gray-700 dark:text-gray-300">{{ formatDateTime(selectedOrder.paid_at) }}</p></div>
          <div v-if="selectedOrder.refund_amount"><p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.refundAmount') }}</p><p class="text-sm font-medium text-red-600 dark:text-red-400">{{ selectedOrder.order_type === 'balance' ? '$' : '¥' }}{{ selectedOrder.refund_amount.toFixed(2) }}</p></div>
          <div v-if="selectedOrder.refund_reason" class="col-span-2"><p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.refundReason') }}</p><p class="text-sm text-gray-700 dark:text-gray-300">{{ selectedOrder.refund_reason }}</p></div>
          <div v-if="selectedOrder.manual_payment?.enabled" class="col-span-2 border-t border-gray-200 pt-3 dark:border-dark-600">
            <p class="mb-2 text-xs font-medium text-gray-500 dark:text-gray-400">{{ localText('人工充值信息', 'Manual payment details') }}</p>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ localText('支付来源', 'Payment source') }}</p>
                <p class="text-sm text-gray-700 dark:text-gray-300">{{ formatManualPaymentSource(selectedOrder.manual_payment.payment_source) }}</p>
              </div>
              <div>
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ localText('审核状态', 'Review status') }}</p>
                <p class="text-sm text-gray-700 dark:text-gray-300">{{ formatManualReviewStatus(selectedOrder.manual_payment.review_status) }}</p>
              </div>
              <div v-if="selectedOrder.manual_payment.proof_note" class="col-span-2">
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ localText('付款备注', 'Proof note') }}</p>
                <p class="text-sm text-gray-700 dark:text-gray-300">{{ selectedOrder.manual_payment.proof_note }}</p>
              </div>
              <div v-if="selectedOrder.manual_payment.review_note" class="col-span-2">
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ localText('审核备注', 'Review note') }}</p>
                <p class="text-sm text-gray-700 dark:text-gray-300">{{ selectedOrder.manual_payment.review_note }}</p>
              </div>
              <div v-if="selectedOrder.manual_payment.proof_image_url" class="col-span-2">
                <p class="mb-2 text-xs text-gray-500 dark:text-gray-400">{{ localText('付款凭证', 'Payment proof') }}</p>
                <img :src="selectedOrder.manual_payment.proof_image_url" alt="" class="max-h-72 cursor-zoom-in rounded-lg border border-gray-200 object-contain dark:border-dark-700" @click="previewImage = selectedOrder.manual_payment?.proof_image_url || ''" />
              </div>
            </div>
          </div>
          <!-- Refund request info -->
          <div v-if="selectedOrder.refund_requested_at" class="col-span-2 border-t border-gray-200 pt-3 dark:border-dark-600">
            <p class="mb-2 text-xs font-medium text-purple-600 dark:text-purple-400">{{ t('payment.admin.refundRequestInfo') }}</p>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.refundRequestedAt') }}</p>
                <p class="text-sm text-gray-700 dark:text-gray-300">{{ formatDateTime(selectedOrder.refund_requested_at) }}</p>
              </div>
              <div>
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.refundRequestedBy') }}</p>
                <p class="text-sm text-gray-700 dark:text-gray-300">#{{ selectedOrder.refund_requested_by }}</p>
              </div>
              <div class="col-span-2">
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.refundRequestReason') }}</p>
                <p class="text-sm text-gray-700 dark:text-gray-300">{{ selectedOrder.refund_request_reason }}</p>
              </div>
            </div>
          </div>
        </div>
        <!-- Audit Logs -->
        <div v-if="orderAuditLogs.length > 0" class="border-t border-gray-200 pt-4 dark:border-dark-600">
          <p class="mb-2 text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('payment.admin.auditLogs') }}</p>
          <div class="max-h-48 space-y-2 overflow-y-auto">
            <div v-for="log in orderAuditLogs" :key="log.id" class="rounded-lg border border-gray-100 bg-gray-50 p-2.5 dark:border-dark-600 dark:bg-dark-800">
              <div class="flex items-center justify-between">
                <span class="text-xs font-medium text-gray-700 dark:text-gray-300">{{ formatAuditAction(log.action) }}</span>
                <span class="text-xs text-gray-400">{{ formatDateTime(log.created_at) }}</span>
              </div>
              <div v-if="log.detail" class="mt-1 whitespace-pre-wrap break-all text-xs text-gray-500 dark:text-gray-400">{{ formatAuditDetail(log) }}</div>
              <div v-if="log.operator" class="mt-1 text-xs text-gray-400">{{ localText('操作人', 'Operator') }}: {{ formatOperator(log.operator) }}</div>
            </div>
          </div>
        </div>
      </div>
    </BaseDialog>

    <AdminRefundDialog :show="showRefundDialog" :order="selectedOrder" :submitting="refundSubmitting" @confirm="handleRefund" @cancel="showRefundDialog = false" />
    <AdminManualReviewDialog
      :show="showManualReviewDialog"
      :order="selectedOrder"
      :approved="manualReviewApproved"
      :submitting="manualReviewSubmitting"
      @confirm="handleManualReviewConfirm"
      @cancel="showManualReviewDialog = false"
    />
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-if="previewImage"
          class="fixed inset-0 z-[70] flex items-center justify-center bg-black/75 p-4"
          @click="previewImage = ''"
        >
          <img
            :src="previewImage"
            alt=""
            class="max-h-[92vh] max-w-[92vw] rounded-2xl bg-white object-contain shadow-2xl"
          />
        </div>
      </Transition>
    </Teleport>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminPaymentAPI } from '@/api/admin/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { formatOrderDateTime } from '@/components/payment/orderUtils'
import type { PaymentOrder } from '@/types/payment'
import AppLayout from '@/components/layout/AppLayout.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import AdminRefundDialog from '@/components/admin/payment/AdminRefundDialog.vue'
import AdminManualReviewDialog from '@/components/admin/payment/AdminManualReviewDialog.vue'
import OrderStatusBadge from '@/components/payment/OrderStatusBadge.vue'
import OrderTable from '@/components/payment/OrderTable.vue'

interface AuditLog {
  id: number
  action: string
  detail: string | null
  operator: string | null
  created_at: string
}

function normalizeAdminOrder(order: PaymentOrder & Record<string, unknown>): PaymentOrder {
  const providerSnapshot = order.provider_snapshot as Record<string, unknown> | undefined
  if (!order.manual_payment && providerSnapshot?.manual_payment) {
    return {
      ...order,
      manual_payment: providerSnapshot.manual_payment as PaymentOrder['manual_payment'],
    }
  }
  return order
}

const { t, locale } = useI18n()
const appStore = useAppStore()

const ordersLoading = ref(false)
const orders = ref<PaymentOrder[]>([])
const orderSearch = ref('')
const orderFilters = reactive({ status: '', payment_type: '', order_type: '' })
const orderPagination = reactive({ page: 1, page_size: 20, total: 0 })
const selectedOrder = ref<PaymentOrder | null>(null)
const showDetailDialog = ref(false)
const showRefundDialog = ref(false)
const refundSubmitting = ref(false)
const showManualReviewDialog = ref(false)
const manualReviewApproved = ref(true)
const manualReviewSubmitting = ref(false)
const orderAuditLogs = ref<AuditLog[]>([])
const previewImage = ref('')

function localText(zh: string, en: string): string {
  return String(locale.value || '').startsWith('zh') ? zh : en
}

let debounceTimer: ReturnType<typeof setTimeout> | null = null
function debounceLoadOrders() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => loadOrders(), 300)
}

async function loadOrders() {
  ordersLoading.value = true
  try {
    const res = await adminPaymentAPI.getOrders({
      page: orderPagination.page, page_size: orderPagination.page_size,
      keyword: orderSearch.value || undefined, status: orderFilters.status || undefined,
      payment_type: orderFilters.payment_type || undefined, order_type: orderFilters.order_type || undefined,
    })
    orders.value = (res.data.items || []).map((item) => normalizeAdminOrder(item as PaymentOrder & Record<string, unknown>))
    orderPagination.total = res.data.total || 0
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally { ordersLoading.value = false }
}

function handleOrderPageChange(page: number) { orderPagination.page = page; loadOrders() }
function handleOrderPageSizeChange(size: number) { orderPagination.page_size = size; orderPagination.page = 1; loadOrders() }

const statusFilterOptions = computed(() => [
  { value: '', label: t('payment.admin.allStatuses') },
  { value: 'PENDING', label: t('payment.status.pending') },
  { value: 'PAID', label: t('payment.status.paid') },
  { value: 'COMPLETED', label: t('payment.status.completed') },
  { value: 'EXPIRED', label: t('payment.status.expired') },
  { value: 'CANCELLED', label: t('payment.status.cancelled') },
  { value: 'FAILED', label: t('payment.status.failed') },
  { value: 'REFUNDED', label: t('payment.status.refunded') },
  { value: 'REFUND_REQUESTED', label: t('payment.status.refund_requested') },
  { value: 'REFUND_FAILED', label: t('payment.status.refund_failed') },
])

const paymentTypeFilterOptions = computed(() => [
  { value: '', label: t('payment.admin.allPaymentTypes') },
  { value: 'alipay', label: t('payment.methods.alipay') },
  { value: 'wxpay', label: t('payment.methods.wxpay') },
  { value: 'manual_alipay', label: localText('人工支付宝', 'Manual Alipay') },
  { value: 'manual_wxpay', label: localText('人工微信', 'Manual WeChat') },
  { value: 'stripe', label: t('payment.methods.stripe') },
  { value: 'airwallex', label: t('payment.methods.airwallex') },
])

const orderTypeFilterOptions = computed(() => [
  { value: '', label: t('payment.admin.allOrderTypes') },
  { value: 'balance', label: t('payment.admin.balanceOrder') },
  { value: 'subscription', label: t('payment.admin.subscriptionOrder') },
])

async function showOrderDetail(order: PaymentOrder) {
  selectedOrder.value = order
  orderAuditLogs.value = []
  showDetailDialog.value = true
  try {
    const res = await adminPaymentAPI.getOrder(order.id)
    const data = res.data as unknown as Record<string, unknown>
    if (data.order) selectedOrder.value = normalizeAdminOrder(data.order as PaymentOrder & Record<string, unknown>)
    orderAuditLogs.value = ((data.auditLogs || data.audit_logs || []) as unknown) as AuditLog[]
  } catch (_err: unknown) { /* keep cached order data */ }
}

async function handleCancelOrder(order: PaymentOrder) {
  try { await adminPaymentAPI.cancelOrder(order.id); appStore.showSuccess(t('payment.admin.orderCancelled')); loadOrders() }
  catch (err: unknown) { appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error'))) }
}

async function handleRetryOrder(order: PaymentOrder) {
  try { await adminPaymentAPI.retryRecharge(order.id); appStore.showSuccess(t('payment.admin.retrySuccess')); loadOrders() }
  catch (err: unknown) { appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error'))) }
}

async function reviewManualPayment(order: PaymentOrder, approved: boolean) {
  selectedOrder.value = order
  manualReviewApproved.value = approved
  showManualReviewDialog.value = true
}

async function handleManualReviewConfirm(payload: { note?: string }) {
  if (!selectedOrder.value) return
  manualReviewSubmitting.value = true
  try {
    await adminPaymentAPI.reviewManualPayment(selectedOrder.value.id, {
      approved: manualReviewApproved.value,
      note: payload.note,
    })
    appStore.showSuccess(
      manualReviewApproved.value
        ? localText('已审核通过', 'Approved')
        : localText('已拒绝该凭证', 'Rejected')
    )
    showManualReviewDialog.value = false
    await loadOrders()
    if (selectedOrder.value) {
      await showOrderDetail(selectedOrder.value)
    }
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    manualReviewSubmitting.value = false
  }
}

function openRefundDialog(order: PaymentOrder) { selectedOrder.value = order; showRefundDialog.value = true }

async function handleRefund(data: { amount: number; reason: string; deduct_balance: boolean; force: boolean }) {
  if (!selectedOrder.value) return
  refundSubmitting.value = true
  try {
    await adminPaymentAPI.refundOrder(selectedOrder.value.id, { amount: data.amount, reason: data.reason, deduct_balance: data.deduct_balance, force: data.force })
    appStore.showSuccess(t('payment.admin.refundSuccess')); showRefundDialog.value = false; loadOrders()
  } catch (err: unknown) { appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error'))) }
  finally { refundSubmitting.value = false }
}

function formatDateTime(dateStr: string): string { return formatOrderDateTime(dateStr) }

function formatOperator(operator?: string | null): string {
  if (!operator) return '-'
  if (operator === 'admin') return localText('管理员', 'Admin')
  if (operator.startsWith('user:')) return `${localText('用户', 'User')} ${operator.slice(5)}`
  return operator
}

function formatManualPaymentSource(source?: string): string {
  switch (source) {
    case 'manual_alipay':
      return localText('人工支付宝', 'Manual Alipay')
    case 'manual_wxpay':
      return localText('人工微信', 'Manual WeChat')
    default:
      return source || '-'
  }
}

function formatManualReviewStatus(status?: string): string {
  switch (status) {
    case 'PENDING_USER_PROOF':
      return localText('待提交凭证', 'Pending proof')
    case 'PENDING_ADMIN_REVIEW':
      return localText('待管理员审核', 'Pending admin review')
    case 'APPROVED':
      return localText('已审核通过', 'Approved')
    case 'REJECTED':
      return localText('已拒绝', 'Rejected')
    default:
      return status || '-'
  }
}

function formatMoneyWithUnit(amount?: number, orderType?: string): string {
  const value = Number(amount || 0)
  return `${orderType === 'balance' ? '$' : '¥'}${value.toFixed(2)}`
}

function formatAuditAction(action: string): string {
  const map: Record<string, string> = {
    ORDER_CREATED: localText('创建订单', 'Order created'),
    ORDER_EXPIRED: localText('订单已过期', 'Order expired'),
    ORDER_CANCELLED: localText('订单已取消', 'Order cancelled'),
    MANUAL_PAYMENT_PROOF_SUBMITTED: localText('用户提交付款凭证', 'Payment proof submitted'),
    MANUAL_PAYMENT_SOURCE_CHANGED: localText('切换人工支付方式', 'Manual payment source changed'),
    MANUAL_PAYMENT_APPROVED: localText('人工审核通过', 'Manual payment approved'),
    MANUAL_PAYMENT_REJECTED: localText('人工审核拒绝', 'Manual payment rejected'),
    MANUAL_PAYMENT_APPROVAL_FULFILLMENT_SUCCEEDED: localText('审核通过后已完成入账', 'Fulfillment succeeded after approval'),
    MANUAL_PAYMENT_APPROVAL_FULFILLMENT_FAILED: localText('审核通过后入账失败', 'Fulfillment failed after approval'),
    MANUAL_REFUND_MARKED: localText('已登记线下人工退款', 'Offline manual refund recorded'),
    RECHARGE_SUCCESS: localText('充值成功', 'Recharge successful'),
    AFFILIATE_REBATE_SKIPPED: localText('返利跳过', 'Affiliate rebate skipped'),
    REFUND_REQUESTED: localText('用户申请退款', 'Refund requested'),
    REFUND_SUCCESS: localText('退款成功', 'Refund successful'),
    REFUND_FAILED: localText('退款失败', 'Refund failed'),
  }
  return map[action] || action
}

function formatAuditDetail(log: AuditLog): string {
  if (!log.detail) return ''
  try {
    const data = JSON.parse(log.detail) as Record<string, unknown>
    const lines: string[] = []
    if (data.paymentSource) lines.push(`${localText('支付来源', 'Payment source')}: ${formatManualPaymentSource(String(data.paymentSource))}`)
    if (data.paymentType) lines.push(`${localText('支付方式', 'Payment method')}: ${String(data.paymentType)}`)
    if (typeof data.paymentAmount === 'number') lines.push(`${localText('支付金额', 'Payment amount')}: ¥${Number(data.paymentAmount).toFixed(2)}`)
    if (typeof data.payAmount === 'number') lines.push(`${localText('实付金额', 'Actual paid')}: ¥${Number(data.payAmount).toFixed(2)}`)
    if (typeof data.creditedAmount === 'number') lines.push(`${localText('到账金额', 'Credited amount')}: ${Number(data.creditedAmount).toFixed(2)}`)
    if (data.orderType) lines.push(`${localText('订单类型', 'Order type')}: ${String(data.orderType) === 'subscription' ? localText('订阅套餐', 'Subscription') : localText('余额充值', 'Balance top-up')}`)
    if (data.proofNote) lines.push(`${localText('用户备注', 'User note')}: ${String(data.proofNote)}`)
    if (data.reviewNote) lines.push(`${localText('审核备注', 'Review note')}: ${String(data.reviewNote)}`)
    if (data.reason) lines.push(`${localText('原因', 'Reason')}: ${String(data.reason)}`)
    if (data.detail) {
      const detail = String(data.detail)
      lines.push(`${localText('详情', 'Detail')}: ${detail === 'order expired' ? localText('订单超时未支付，系统已自动关闭', 'The order timed out and was closed automatically') : detail}`)
    }
    if (typeof data.refundAmount === 'number') lines.push(`${localText('退款金额', 'Refund amount')}: ¥${Number(data.refundAmount).toFixed(2)}`)
    if (data.mode) lines.push(`${localText('退款方式', 'Refund mode')}: ${String(data.mode) === 'offline_manual_refund' ? localText('线下人工退款', 'Offline manual refund') : String(data.mode)}`)
    if (data.result) lines.push(`${localText('结果', 'Result')}: ${String(data.result)}`)
    if (String(log.action) === 'AFFILIATE_REBATE_SKIPPED' && lines.length === 0) {
      return localText('原因：未绑定邀请人，或返利金额小于等于 0', 'Reason: no inviter bound or rebate amount <= 0')
    }
    return lines.length > 0 ? lines.join('\n') : log.detail
  } catch {
    if (log.action === 'AFFILIATE_REBATE_SKIPPED' && log.detail.includes('no inviter bound or rebate amount <= 0')) {
      return localText('原因：未绑定邀请人，或返利金额小于等于 0', 'Reason: no inviter bound or rebate amount <= 0')
    }
    if (log.action === 'ORDER_EXPIRED' && log.detail.includes('order expired')) {
      return localText('详情：订单超时未支付，系统已自动关闭', 'Detail: the order timed out and was closed automatically')
    }
    return log.detail
  }
}

onMounted(() => loadOrders())
</script>
