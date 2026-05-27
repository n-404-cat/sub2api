import { apiClient } from '../client'
import type { SupportConversation, SupportConversationDetail } from '@/types/payment'
import type { PaginatedResponse } from '@/types'

export interface ListSupportConversationParams {
  page?: number
  page_size?: number
  status?: string
  user_id?: number
  keyword?: string
}

export const adminSupportAPI = {
  list(params?: ListSupportConversationParams) {
    return apiClient.get<PaginatedResponse<SupportConversation>>('/admin/payment/support/conversations', { params })
  },
  get(id: number) {
    return apiClient.get<SupportConversationDetail>(`/admin/payment/support/conversations/${id}`)
  },
  reply(id: number, data: { message: string }) {
    return apiClient.post<SupportConversationDetail>(`/admin/payment/support/conversations/${id}/messages`, data)
  },
}

export default adminSupportAPI
