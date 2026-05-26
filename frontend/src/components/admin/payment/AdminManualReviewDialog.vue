<template>
  <BaseDialog
    :show="show"
    :title="approved ? localText('通过人工充值', 'Approve manual payment') : localText('拒绝人工充值', 'Reject manual payment')"
    width="normal"
    @close="emit('cancel')"
  >
    <form id="manual-review-form" @submit.prevent="handleSubmit" class="space-y-4">
      <div class="rounded-lg bg-gray-50 p-3 dark:bg-dark-700">
        <div class="flex justify-between text-sm">
          <span class="text-gray-500 dark:text-gray-400">{{ t('payment.orders.orderId') }}</span>
          <span class="font-mono text-gray-900 dark:text-white">#{{ order?.id }}</span>
        </div>
        <div class="mt-1 flex justify-between text-sm">
          <span class="text-gray-500 dark:text-gray-400">{{ localText('支付来源', 'Payment source') }}</span>
          <span class="text-gray-900 dark:text-white">{{ order?.manual_payment?.payment_source || '-' }}</span>
        </div>
        <div class="mt-1 flex justify-between text-sm">
          <span class="text-gray-500 dark:text-gray-400">{{ t('payment.orders.payAmount') }}</span>
          <span class="text-gray-900 dark:text-white">¥{{ order?.pay_amount?.toFixed(2) }}</span>
        </div>
      </div>

      <div v-if="order?.manual_payment?.proof_image_url" class="space-y-2">
        <label class="input-label">{{ localText('付款凭证', 'Payment proof') }}</label>
        <img
          :src="order.manual_payment.proof_image_url"
          alt=""
          class="max-h-72 rounded-lg border border-gray-200 object-contain dark:border-dark-700"
        />
      </div>

      <div>
        <label class="input-label">
          {{ approved ? localText('审核备注（可选）', 'Review note (optional)') : localText('拒绝原因', 'Rejection reason') }}
        </label>
        <textarea
          v-model="note"
          rows="4"
          class="input"
          :placeholder="
            approved
              ? localText('例如：已核实到账', 'For example: payment verified')
              : localText('请填写拒绝原因，便于用户重新提交', 'Please provide a reason so the user can resubmit')
          "
        ></textarea>
      </div>
    </form>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button type="button" @click="emit('cancel')" class="btn btn-secondary">
          {{ t('common.cancel') }}
        </button>
        <button
          type="submit"
          form="manual-review-form"
          :disabled="submitting || (!approved && !note.trim())"
          :class="approved ? 'btn btn-primary' : 'rounded-md bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700 disabled:opacity-50'"
        >
          {{ submitting ? t('common.processing') : (approved ? localText('确认通过', 'Confirm approval') : localText('确认拒绝', 'Confirm rejection')) }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import type { PaymentOrder } from '@/types/payment'

const props = defineProps<{
  show: boolean
  order: PaymentOrder | null
  approved: boolean
  submitting?: boolean
}>()

const emit = defineEmits<{
  (e: 'confirm', payload: { note?: string }): void
  (e: 'cancel'): void
}>()

const { t, locale } = useI18n()
const note = ref('')

watch(() => props.show, (visible) => {
  if (visible) {
    note.value = ''
  }
})

function localText(zh: string, en: string): string {
  return String(locale.value || '').startsWith('zh') ? zh : en
}

function handleSubmit() {
  if (!props.approved && !note.value.trim()) return
  emit('confirm', { note: note.value.trim() || undefined })
}
</script>
