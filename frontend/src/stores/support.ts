import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { paymentAPI } from '@/api/payment'
import { adminSupportAPI } from '@/api/admin'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import type { SupportConversation } from '@/types/payment'

export const useSupportStore = defineStore('support', () => {
  const conversations = ref<SupportConversation[]>([])
  const loading = ref(false)
  const previousUnreadCount = ref(0)
  let timer: ReturnType<typeof setInterval> | null = null

  const unreadCount = computed(() => {
    return conversations.value.reduce((sum, item) => sum + Number(item.unread_count || 0), 0)
  })

  async function fetchConversations(force = false) {
    if (loading.value && !force) return
    loading.value = true
    try {
      const authStore = useAuthStore()
      const appStore = useAppStore()
      if (authStore.isAdmin) {
        const res = await adminSupportAPI.list({ page: 1, page_size: 50 })
        conversations.value = res.data.items || []
      } else {
        const res = await paymentAPI.getMySupportConversations()
        conversations.value = res.data || []
      }
      const nextUnread = conversations.value.reduce((sum, item) => sum + Number(item.unread_count || 0), 0)
      if (previousUnreadCount.value > 0 && nextUnread > previousUnreadCount.value) {
        appStore.showInfo(authStore.isAdmin ? '收到新的咨询消息' : '收到客服新回复')
      }
      previousUnreadCount.value = nextUnread
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

  function startPolling() {
    if (timer) return
    timer = setInterval(() => {
      fetchConversations(true)
    }, 5000)
  }

  function stopPolling() {
    if (!timer) return
    clearInterval(timer)
    timer = null
  }

  return {
    conversations,
    loading,
    unreadCount,
    fetchConversations,
    markConversationSeen,
    startPolling,
    stopPolling,
  }
})
