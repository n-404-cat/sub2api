<template>
  <BaseDialog :show="show" :title="localText('订单咨询', 'Order consultation')" width="normal" @close="emit('cancel')">
    <div v-if="order" class="space-y-4">
      <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-800">
        <div class="flex justify-between text-sm">
          <span class="text-gray-500 dark:text-gray-400">{{ localText('订单号', 'Order No') }}</span>
          <span class="font-medium text-gray-900 dark:text-white">{{ order.out_trade_no }}</span>
        </div>
        <div class="mt-2 flex justify-between text-sm">
          <span class="text-gray-500 dark:text-gray-400">{{ localText('订单ID', 'Order ID') }}</span>
          <span class="font-mono text-gray-900 dark:text-white">#{{ order.id }}</span>
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
        <button class="btn btn-primary" :disabled="submitting || !message.trim()" @click="emit('confirm', { message: message.trim() })">
          {{ submitting ? t('common.processing') : localText('发起咨询', 'Start conversation') }}
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
  submitting?: boolean
}>()

const emit = defineEmits<{
  (e: 'confirm', payload: { message: string }): void
  (e: 'cancel'): void
}>()

const { t, locale } = useI18n()
const message = ref('')

watch(() => props.show, (visible) => {
  if (visible) {
    message.value = ''
  }
})

function localText(zh: string, en: string): string {
  return String(locale.value || '').startsWith('zh') ? zh : en
}
</script>
