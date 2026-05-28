import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { paymentAPI } from '@/api/payment'
import { adminSupportAPI } from '@/api/admin'
import { useAuthStore } from '@/stores/auth'
import type { SupportConversation } from '@/types/payment'

export const useSupportStore = defineStore('support', () => {
  const conversations = ref<SupportConversation[]>([])
  const loading = ref(false)

  const unreadCount = computed(() => {
    return conversations.value.reduce((sum, item) => sum + Number(item.unread_count || 0), 0)
  })

  async function fetchConversations(force = false) {
    if (loading.value && !force) return
    loading.value = true
    try {
      const authStore = useAuthStore()
      if (authStore.isAdmin) {
        const res = await adminSupportAPI.list({ page: 1, page_size: 50 })
        conversations.value = res.data.items || []
      } else {
        const res = await paymentAPI.getMySupportConversations()
        conversations.value = res.data || []
      }
    } finally {
      loading.value = false
    }
  }

  function markConversationSeen(id: number, at?: string) {
    const target = conversations.value.find(item => item.id === id)
    if (!target) return
    target.unread_count = 0
    if (at) {
      target.last_message_at = at
    }
  }

  return {
    conversations,
    loading,
    unreadCount,
    fetchConversations,
    markConversationSeen,
  }
})
