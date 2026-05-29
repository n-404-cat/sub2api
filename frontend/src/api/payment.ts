/**
 * User Payment API endpoints
 * Handles payment operations for regular users
 */

import { apiClient } from './client'
import type {
  PaymentConfig,
  SubscriptionPlan,
  PaymentChannel,
  MethodLimitsResponse,
  CheckoutInfoResponse,
  CreateOrderRequest,
  CreateOrderResult,
  PaymentOrder,
  SupportConversation,
  SupportConversationDetail
} from '@/types/payment'
import type { BasePaginationResponse } from '@/types'

export const paymentAPI = {
  /** Get payment configuration (enabled types, limits, etc.) */
  getConfig() {
    return apiClient.get<PaymentConfig>('/payment/config')
  },

  /** Get available subscription plans */
  getPlans() {
    return apiClient.get<SubscriptionPlan[]>('/payment/plans')
  },

  /** Get available payment channels */
  getChannels() {
    return apiClient.get<PaymentChannel[]>('/payment/channels')
  },

  /** Get all checkout page data in a single call */
  getCheckoutInfo() {
    return apiClient.get<CheckoutInfoResponse>('/payment/checkout-info')
  },

  /** Get payment method limits and fee rates */
  getLimits() {
    return apiClient.get<MethodLimitsResponse>('/payment/limits')
  },

  /** Create a new payment order */
  createOrder(data: CreateOrderRequest) {
    return apiClient.post<CreateOrderResult>('/payment/orders', data)
  },

  /** Get current user's orders */
  getMyOrders(params?: { page?: number; page_size?: number; status?: string }) {
    return apiClient.get<BasePaginationResponse<PaymentOrder>>('/payment/orders/my', { params })
  },

  /** Get a specific order by ID */
  getOrder(id: number) {
    return apiClient.get<PaymentOrder>(`/payment/orders/${id}`)
  },

  /** Cancel a pending order */
  cancelOrder(id: number) {
    return apiClient.post(`/payment/orders/${id}/cancel`)
  },

  /** Submit proof for a manual QR payment order */
  submitManualProof(id: number, data: { proof_image_url: string; proof_note?: string }) {
    return apiClient.post<PaymentOrder>(`/payment/orders/${id}/manual-proof`, data)
  },

  /** Switch manual payment source between Alipay and WeChat */
  updateManualPaymentSource(id: number, data: { payment_source: 'manual_alipay' | 'manual_wxpay' }) {
    return apiClient.post<PaymentOrder>(`/payment/orders/${id}/manual-source`, data)
  },

  /** Verify order payment status with upstream provider */
  verifyOrder(outTradeNo: string) {
    return apiClient.post<PaymentOrder>('/payment/orders/verify', { out_trade_no: outTradeNo })
  },

  /** Legacy-compatible public order lookup by out_trade_no */
  verifyOrderPublic(outTradeNo: string) {
    return apiClient.post<PaymentOrder>('/payment/public/orders/verify', { out_trade_no: outTradeNo })
  },

  /** Resolve an order from a signed resume token without auth */
  resolveOrderPublicByResumeToken(resumeToken: string) {
    return apiClient.post<PaymentOrder>('/payment/public/orders/resolve', { resume_token: resumeToken })
  },

  /** Request a refund for a completed order */
  requestRefund(id: number, data: { reason: string }) {
    return apiClient.post(`/payment/orders/${id}/refund-request`, data)
  },

  /** Create or append an order consultation conversation */
  createSupportConversation(data: { order_id?: number; subject?: string; message: string }) {
    return apiClient.post<SupportConversationDetail>('/payment/support/conversations', data)
  },

  /** List current user's support conversations */
  getMySupportConversations() {
    return apiClient.get<SupportConversation[]>('/payment/support/conversations')
  },

  /** Get support conversation detail */
  getMySupportConversation(id: number) {
    return apiClient.get<SupportConversationDetail>(`/payment/support/conversations/${id}`)
  },

  /** Mark support conversation as read */
  markMySupportConversationRead(id: number) {
    return apiClient.post<{ message: string }>(`/payment/support/conversations/${id}/read`)
  },

  /** Reply in a support conversation */
  replyMySupportConversation(id: number, data: { message: string }) {
    return apiClient.post<SupportConversationDetail>(`/payment/support/conversations/${id}/messages`, data)
  },

  /** Bind an order to an existing support conversation */
  bindSupportConversationOrder(id: number, data: { order_id: number }) {
    return apiClient.post<SupportConversationDetail>(`/payment/support/conversations/${id}/bind-order`, data)
  },

  /** Get provider instance IDs that allow user refund */
  getRefundEligibleProviders() {
    return apiClient.get<{ provider_instance_ids: string[] }>('/payment/orders/refund-eligible-providers')
  }
}
