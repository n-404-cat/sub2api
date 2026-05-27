<template>
  <BaseDialog :show="show" :title="localText('订单咨询', 'Order consultation')" width="normal" @close="emit('cancel')">
    <div v-if="selectedOrder" class="space-y-4">
      <div v-if="orders.length > 1" class="space-y-2">
        <label class="input-label">{{ localText('选择订单', 'Select order') }}</label>
        <div class="grid grid-cols-1 gap-2">
          <button
            v-for="item in orders"
            :key="item.id"
            type="button"
            :class="[
              'rounded-xl border px-4 py-3 text-left transition-colors',
              selectedOrder?.id === item.id
                ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20'
                : 'border-gray-200 hover:bg-gray-50 dark:border-dark-700 dark:hover:bg-dark-800'
            ]"
            @click="selectedOrderId = item.id"
          >
            <div class="flex justify-between text-sm">
              <span class="font-medium text-gray-900 dark:text-white">{{ item.out_trade_no }}</span>
              <span class="text-gray-500 dark:text-gray-400">#{{ item.id }}</span>
            </div>
          </button>
        </div>
      </div>

      <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-800">
        <div class="flex justify-between text-sm">
          <span class="text-gray-500 dark:text-gray-400">{{ localText('订单号', 'Order No') }}</span>
          <span class="font-medium text-gray-900 dark:text-white">{{ selectedOrder.out_trade_no }}</span>
        </div>
        <div class="mt-2 flex justify-between text-sm">
          <span class="text-gray-500 dark:text-gray-400">{{ localText('订单ID', 'Order ID') }}</span>
          <span class="font-mono text-gray-900 dark:text-white">#{{ selectedOrder.id }}</span>
        </div>
      </div>

      <div>
        <label class="input-label">{{ localText('咨询内容', 'Message') }}</label>
        <textarea
          v-model="message"
          rows="4"
          class="input"
          :placeholder="localText('请描述你的问题，例如付款已完成但未到账。', 'Describe your issue, for example: payment completed but balance not updated.')"
        ></textarea>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button class="btn btn-secondary" @click="emit('cancel')">{{ t('common.cancel') }}</button>
        <button class="btn btn-primary" :disabled="submitting || !message.trim() || !selectedOrder" @click="emit('confirm', { orderId: selectedOrder!.id, message: message.trim() })">
          {{ submitting ? t('common.processing') : localText('发起咨询', 'Start conversation') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import type { PaymentOrder } from '@/types/payment'

const props = defineProps<{
  show: boolean
  order: PaymentOrder | null
  orders?: PaymentOrder[]
  submitting?: boolean
}>()

const emit = defineEmits<{
  (e: 'confirm', payload: { orderId: number; message: string }): void
  (e: 'cancel'): void
}>()

const { t, locale } = useI18n()
const message = ref('')
const selectedOrderId = ref<number | null>(null)

const orders = computed(() => {
  const list = props.orders || []
  return list.length > 0 ? list : (props.order ? [props.order] : [])
})

const selectedOrder = computed(() => {
  if (selectedOrderId.value != null) {
    return orders.value.find(item => item.id === selectedOrderId.value) || null
  }
  return props.order || orders.value[0] || null
})

watch(() => props.show, (visible) => {
  if (visible) {
    message.value = ''
    selectedOrderId.value = props.order?.id ?? props.orders?.[0]?.id ?? null
  }
})

function localText(zh: string, en: string): string {
  return String(locale.value || '').startsWith('zh') ? zh : en
}
</script>
