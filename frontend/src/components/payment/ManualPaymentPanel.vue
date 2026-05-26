<template>
  <div class="space-y-4">
    <div class="card p-6">
      <div class="flex flex-col gap-6 lg:flex-row">
        <div class="flex-1 space-y-4">
          <div class="flex flex-wrap gap-2">
            <button
              v-if="availableMethods.alipay"
              type="button"
              :class="[
                'rounded-lg border px-3 py-2 text-sm font-medium transition-colors',
                currentSource === 'manual_alipay'
                  ? 'border-[#02A9F1] bg-blue-50 text-[#02A9F1] dark:bg-blue-950/20'
                  : 'border-gray-300 text-gray-600 hover:border-gray-400 dark:border-dark-600 dark:text-gray-300'
              ]"
              :disabled="switching || currentSource === 'manual_alipay' || !canSwitchSource"
              @click="switchSource('manual_alipay')"
            >
              {{ t('payment.methods.alipay') }}
            </button>
            <button
              v-if="availableMethods.wechat"
              type="button"
              :class="[
                'rounded-lg border px-3 py-2 text-sm font-medium transition-colors',
                currentSource === 'manual_wxpay'
                  ? 'border-[#09BB07] bg-green-50 text-[#09BB07] dark:bg-green-950/20'
                  : 'border-gray-300 text-gray-600 hover:border-gray-400 dark:border-dark-600 dark:text-gray-300'
              ]"
              :disabled="switching || currentSource === 'manual_wxpay' || !canSwitchSource"
              @click="switchSource('manual_wxpay')"
            >
              {{ localText('微信支付', 'WeChat Pay') }}
            </button>
          </div>

          <div>
            <p class="text-xs font-medium uppercase tracking-wide text-gray-400 dark:text-gray-500">
              {{ paymentMethodLabel }}
            </p>
            <h3 class="mt-1 text-xl font-semibold text-gray-900 dark:text-white">
              {{ title }}
            </h3>
            <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">
              {{ description }}
            </p>
          </div>

          <div class="rounded-xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-800">
            <div class="grid grid-cols-2 gap-3 text-sm">
              <div>
                <p class="text-xs text-gray-400 dark:text-gray-500">{{ t('payment.orders.orderId') }}</p>
                <p class="mt-1 font-medium text-gray-900 dark:text-white">#{{ order.id }}</p>
              </div>
              <div>
                <p class="text-xs text-gray-400 dark:text-gray-500">{{ t('payment.orders.orderNo') }}</p>
                <p class="mt-1 break-all font-medium text-gray-900 dark:text-white">{{ order.out_trade_no }}</p>
              </div>
              <div>
                <p class="text-xs text-gray-400 dark:text-gray-500">{{ t('payment.orders.amount') }}</p>
                <p class="mt-1 font-medium text-gray-900 dark:text-white">{{ formatMoney(order.pay_amount, order.currency) }}</p>
              </div>
              <div>
                <p class="text-xs text-gray-400 dark:text-gray-500">{{ localText('审核状态', 'Review status') }}</p>
                <p class="mt-1 font-medium" :class="statusClass">{{ statusText }}</p>
              </div>
            </div>
          </div>

          <div v-if="helpText" class="rounded-xl border border-dashed border-gray-300 p-4 text-sm text-gray-600 dark:border-dark-600 dark:text-gray-300">
            {{ helpText }}
          </div>

          <div v-if="status === 'PENDING_ADMIN_REVIEW'" class="rounded-xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-700 dark:border-amber-900/50 dark:bg-amber-900/20 dark:text-amber-300">
            {{ localText('付款凭证已提交，等待管理员审核入账。', 'Proof submitted. Waiting for admin approval.') }}
          </div>

          <div v-if="status === 'REJECTED'" class="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-700 dark:border-red-900/50 dark:bg-red-900/20 dark:text-red-300">
            <p>{{ localText('管理员已拒绝本次凭证，请根据备注重新提交。', 'The admin rejected this proof. Please review the note and resubmit.') }}</p>
            <p v-if="order.manual_payment?.review_note" class="mt-2 break-words">
              {{ localText('备注：', 'Note: ') }}{{ order.manual_payment.review_note }}
            </p>
          </div>
        </div>

        <div class="w-full shrink-0 lg:w-[280px]">
          <div class="rounded-2xl border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-900">
            <img
              v-if="order.manual_payment?.qr_code_image_url"
              :src="order.manual_payment.qr_code_image_url"
              :alt="paymentMethodLabel"
              class="mx-auto aspect-square w-full rounded-xl object-contain"
            />
            <div v-else class="flex aspect-square items-center justify-center rounded-xl bg-gray-100 text-sm text-gray-400 dark:bg-dark-800">
              {{ localText('未配置收款码', 'QR code not configured') }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <div
      v-if="showProofForm"
      class="card p-6"
    >
      <div class="space-y-4">
        <div>
          <label class="input-label">{{ localText('付款截图', 'Payment proof image') }}</label>
          <ImageUpload
            v-model="proofImage"
            :upload-label="t('admin.settings.site.uploadImage')"
            :remove-label="t('admin.settings.site.remove')"
            :placeholder="localText('上传转账截图', 'Upload transfer screenshot')"
          />
        </div>
        <div>
          <label class="input-label">{{ localText('备注', 'Note') }}</label>
          <textarea
            v-model="proofNote"
            rows="3"
            class="input"
            :placeholder="localText('可填写付款时间、付款账号后四位等信息', 'Optional: payment time, last 4 digits, or other proof details')"
          ></textarea>
        </div>
        <div class="flex justify-end">
          <button
            class="btn btn-primary"
            :disabled="submitting || !proofImage.trim()"
            @click="submitProof"
          >
            <span v-if="submitting">{{ t('common.processing') }}</span>
            <span v-else>{{ localText('提交付款凭证', 'Submit payment proof') }}</span>
          </button>
        </div>
      </div>
    </div>

    <div v-if="order.manual_payment?.proof_image_url" class="card p-6">
      <div class="space-y-4">
        <div class="flex items-center justify-between">
          <h4 class="text-sm font-semibold text-gray-900 dark:text-white">
            {{ localText('已提交凭证', 'Submitted proof') }}
          </h4>
          <p v-if="order.manual_payment?.proof_submitted_at" class="text-xs text-gray-400 dark:text-gray-500">
            {{ formatDateTime(order.manual_payment.proof_submitted_at) }}
          </p>
        </div>
        <img
          :src="order.manual_payment.proof_image_url"
          :alt="localText('付款凭证', 'Payment proof')"
          class="max-h-[360px] rounded-xl border border-gray-200 object-contain dark:border-dark-700"
        />
        <p v-if="order.manual_payment?.proof_note" class="text-sm text-gray-600 dark:text-gray-300">
          {{ order.manual_payment.proof_note }}
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import ImageUpload from '@/components/common/ImageUpload.vue'
import { paymentAPI } from '@/api/payment'
import { formatPaymentAmount, normalizePaymentCurrency } from '@/components/payment/currency'
import type { PaymentOrder } from '@/types/payment'
import { useAppStore } from '@/stores'
import { extractI18nErrorMessage } from '@/utils/apiError'

const props = defineProps<{
  order: PaymentOrder
  helpText?: string
  availableMethods?: {
    alipay?: boolean
    wechat?: boolean
  }
}>()

const emit = defineEmits<{
  (e: 'updated', order: PaymentOrder): void
}>()

const { t, locale } = useI18n()
const appStore = useAppStore()

const proofImage = ref('')
const proofNote = ref('')
const submitting = ref(false)
const switching = ref(false)

watch(
  () => props.order.manual_payment?.proof_note,
  (value) => {
    if (!proofNote.value && value) {
      proofNote.value = value
    }
  },
  { immediate: true },
)

const manualPayment = computed(() => props.order.manual_payment ?? {
  enabled: false,
  payment_source: '',
  review_status: '',
  require_proof: true,
  qr_code_image_url: '',
  proof_image_url: '',
  proof_note: '',
})

const status = computed(() => manualPayment.value.review_status || '')
const currentSource = computed(() => manualPayment.value.payment_source || 'manual_alipay')
const canSwitchSource = computed(() => status.value !== 'PENDING_ADMIN_REVIEW')
const showProofForm = computed(() =>
  props.order.status === 'PENDING'
  && manualPayment.value.require_proof !== false
  && (status.value === 'PENDING_USER_PROOF' || status.value === 'REJECTED')
)

const paymentMethodLabel = computed(() => {
  if (manualPayment.value.payment_source === 'manual_wxpay') {
    return t('payment.methods.wxpay')
  }
  return t('payment.methods.alipay')
})

const title = computed(() => {
  if (status.value === 'PENDING_ADMIN_REVIEW') {
    return localText('等待审核', 'Waiting for review')
  }
  if (status.value === 'APPROVED') {
    return localText('审核通过', 'Approved')
  }
  if (status.value === 'REJECTED') {
    return localText('凭证被拒绝', 'Proof rejected')
  }
  return localText('扫码付款并提交凭证', 'Scan to pay and submit proof')
})

const description = computed(() => {
  if (status.value === 'PENDING_ADMIN_REVIEW') {
    return localText('管理员确认到账后会自动为当前账号充值。', 'The balance will be credited automatically after admin approval.')
  }
  if (status.value === 'APPROVED') {
    return localText('管理员已确认到账，余额会自动更新。', 'The admin confirmed payment. Your balance will update automatically.')
  }
  if (status.value === 'REJECTED') {
    return localText('请根据备注重新上传正确的付款凭证。', 'Please upload corrected payment proof based on the review note.')
  }
  return localText('使用右侧收款码转账，完成后上传付款截图等待审核。', 'Transfer using the QR code on the right, then upload proof for review.')
})

const statusText = computed(() => {
  switch (status.value) {
    case 'PENDING_ADMIN_REVIEW':
      return localText('待管理员审核', 'Pending admin review')
    case 'APPROVED':
      return localText('已审核通过', 'Approved')
    case 'REJECTED':
      return localText('审核已拒绝', 'Rejected')
    default:
      return localText('待提交凭证', 'Pending proof')
  }
})

const statusClass = computed(() => {
  switch (status.value) {
    case 'PENDING_ADMIN_REVIEW':
      return 'text-amber-600 dark:text-amber-300'
    case 'APPROVED':
      return 'text-green-600 dark:text-green-300'
    case 'REJECTED':
      return 'text-red-600 dark:text-red-300'
    default:
      return 'text-gray-700 dark:text-gray-200'
  }
})

function localText(zh: string, en: string): string {
  return String(locale.value).startsWith('zh') ? zh : en
}

function formatMoney(value: number, currency?: string) {
  return formatPaymentAmount(value, normalizePaymentCurrency(currency), locale.value)
}

function formatDateTime(value?: string) {
  if (!value) return ''
  return new Date(value).toLocaleString()
}

async function submitProof() {
  if (!proofImage.value.trim() || submitting.value) return
  submitting.value = true
  try {
    const res = await paymentAPI.submitManualProof(props.order.id, {
      proof_image_url: proofImage.value.trim(),
      proof_note: proofNote.value.trim() || undefined,
    })
    emit('updated', res.data)
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    submitting.value = false
  }
}

async function switchSource(source: 'manual_alipay' | 'manual_wxpay') {
  if (switching.value || currentSource.value === source || !canSwitchSource.value) return
  switching.value = true
  try {
    const res = await paymentAPI.updateManualPaymentSource(props.order.id, {
      payment_source: source,
    })
    emit('updated', res.data)
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    switching.value = false
  }
}
</script>
