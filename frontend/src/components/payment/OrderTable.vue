<template>
  <DataTable :columns="columns" :data="orders" :loading="loading">
    <template #cell-id="{ value }">
      <span class="font-mono text-sm">#{{ value }}</span>
    </template>
    <template #cell-out_trade_no="{ value }">
      <span class="text-sm text-gray-900 dark:text-white">{{ value }}</span>
    </template>
    <template v-if="showUser" #cell-user_email="{ value, row }">
      <div class="text-sm">
        <span class="text-gray-900 dark:text-white">{{ value || row.user_name || '#' + row.user_id }}</span>
        <span v-if="row.user_notes" class="ml-1 text-xs text-gray-400">({{ row.user_notes }})</span>
      </div>
    </template>
    <template #cell-pay_amount="{ value, row }">
      <div class="text-sm">
        <span class="font-medium text-gray-900 dark:text-white">¥{{ value.toFixed(2) }}</span>
        <span v-if="row.fee_rate > 0" class="ml-1 text-xs text-gray-400" :title="t('payment.orders.fee') + ': ' + row.fee_rate + '%'">
          ({{ t('payment.orders.fee') }} {{ row.fee_rate }}%)
        </span>
        <div v-if="row.amount !== row.pay_amount" class="text-xs text-gray-500">
          {{ t('payment.orders.creditedAmount') }}: {{ row.order_type === 'balance' ? '$' : '¥' }}{{ row.amount.toFixed(2) }}
        </div>
      </div>
    </template>
    <template #cell-payment_type="{ value, row }">
      <div class="flex items-center gap-2 text-sm">
        <span class="text-gray-700 dark:text-gray-300">{{ t('payment.methods.' + value, value) }}</span>
        <span
          v-if="row.manual_payment?.enabled"
          class="rounded-full bg-amber-100 px-2 py-0.5 text-[11px] font-medium text-amber-700 dark:bg-amber-900/30 dark:text-amber-300"
        >
          {{ localeText('人工', 'Manual') }}
        </span>
      </div>
    </template>
    <template #cell-status="{ value, row }">
      <div class="space-y-1">
        <OrderStatusBadge :status="value" />
        <div
          v-if="row.manual_payment?.enabled"
          class="text-[11px] text-gray-500 dark:text-gray-400"
        >
          {{ manualReviewStatusLabel(row.manual_payment.review_status) }}
        </div>
      </div>
    </template>
    <template #cell-created_at="{ value }">
      <span class="text-xs text-gray-500 dark:text-gray-400">{{ formatDate(value) }}</span>
    </template>
    <template #cell-actions="{ row }">
      <slot name="actions" :row="row" />
    </template>
  </DataTable>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { PaymentOrder } from '@/types/payment'
import type { Column } from '@/components/common/types'
import DataTable from '@/components/common/DataTable.vue'
import OrderStatusBadge from '@/components/payment/OrderStatusBadge.vue'

const { t, locale } = useI18n()

const props = defineProps<{
  orders: PaymentOrder[]
  loading: boolean
  showUser?: boolean
}>()

function formatDate(dateStr: string) { return new Date(dateStr).toLocaleString() }

function localeText(zh: string, en: string): string {
  return String(locale.value || '').startsWith('zh') ? zh : en
}

function manualReviewStatusLabel(status?: string): string {
  switch (status) {
    case 'PENDING_USER_PROOF':
      return localeText('待用户提交凭证', 'Waiting for proof')
    case 'PENDING_ADMIN_REVIEW':
      return localeText('待管理员审核', 'Pending admin review')
    case 'APPROVED':
      return localeText('人工审核已通过', 'Manual review approved')
    case 'REJECTED':
      return localeText('人工审核已拒绝', 'Manual review rejected')
    default:
      return localeText('人工充值', 'Manual top-up')
  }
}

const columns = computed((): Column[] => {
  const cols: Column[] = [
    { key: 'id', label: t('payment.orders.orderId') },
    { key: 'out_trade_no', label: t('payment.orders.orderNo') },
  ]
  if (props.showUser) {
    cols.push({ key: 'user_email', label: t('payment.admin.colUser') })
  }
  cols.push(
    { key: 'pay_amount', label: t('payment.orders.payAmount') },
    { key: 'payment_type', label: t('payment.orders.paymentMethod') },
    { key: 'status', label: t('payment.orders.status') },
    { key: 'created_at', label: t('payment.orders.createdAt') },
    { key: 'actions', label: t('common.actions') },
  )
  return cols
})
</script>
