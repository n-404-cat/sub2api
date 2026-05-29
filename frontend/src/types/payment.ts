/**
 * Payment System Type Definitions
 */

// ==================== Enums / Union Types ====================

export type OrderStatus =
  | 'PENDING'
  | 'PAID'
  | 'RECHARGING'
  | 'COMPLETED'
  | 'EXPIRED'
  | 'CANCELLED'
  | 'FAILED'
  | 'REFUND_REQUESTED'
  | 'REFUNDING'
  | 'PARTIALLY_REFUNDED'
  | 'REFUNDED'
  | 'REFUND_FAILED'

export type PaymentType = 'alipay' | 'wxpay' | 'alipay_direct' | 'wxpay_direct' | 'stripe' | 'easypay' | 'airwallex'

export type ManualReviewStatus =
  | 'PENDING_USER_PROOF'
  | 'PENDING_ADMIN_REVIEW'
  | 'APPROVED'
  | 'REJECTED'

export interface ManualPaymentConfig {
  enabled: boolean
  require_proof: boolean
  alipay_enabled: boolean
  wechat_enabled: boolean
  alipay_qr_code_image_url: string
  wechat_qr_code_image_url: string
  help_text: string
  review_timeout_minutes: number
}

export interface ManualPaymentOrderMeta {
  enabled: boolean
  payment_source?: 'manual_alipay' | 'manual_wxpay'
  review_status?: ManualReviewStatus
  require_proof?: boolean
  qr_code_image_url?: string
  proof_image_url?: string
  proof_note?: string
  proof_submitted_at?: string
  reviewed_at?: string
  reviewed_by?: string
  review_note?: string
  review_timeout_minutes?: number
}

export type OrderType = 'balance' | 'subscription'

// ==================== Configuration ====================

export interface PaymentConfig {
  payment_enabled: boolean
  min_amount: number
  max_amount: number
  daily_limit: number
  max_pending_orders: number
  order_timeout_minutes: number
  balance_disabled: boolean
  balance_recharge_multiplier: number
  enabled_payment_types: PaymentType[]
  help_image_url: string
  help_text: string
  stripe_publishable_key: string
  manual_payment: ManualPaymentConfig
  support_quick_replies?: string[]
}

export interface MethodLimit {
  currency?: string
  daily_limit: number
  daily_used: number
  daily_remaining: number
  single_min: number
  single_max: number
  fee_rate: number
  available: boolean
}

/** Response from /payment/limits API */
export interface MethodLimitsResponse {
  methods: Record<string, MethodLimit>
  global_min: number  // widest min across all methods; 0 = no minimum
  global_max: number  // widest max across all methods; 0 = no maximum
}

/** Response from /payment/checkout-info API — single call for the payment page */
export interface CheckoutInfoResponse {
  methods: Record<string, MethodLimit>
  global_min: number
  global_max: number
  plans: SubscriptionPlan[]
  balance_disabled: boolean
  balance_recharge_multiplier: number
  recharge_fee_rate: number
  help_text: string
  help_image_url: string
  stripe_publishable_key: string
  manual_payment: ManualPaymentConfig
  /** When true, Alipay payments on mobile always show the QR code instead of redirecting */
  alipay_force_qrcode?: boolean
}

// ==================== Orders ====================

export interface PaymentOrder {
  id: number
  user_id: number
  amount: number
  pay_amount: number
  currency?: string
  fee_rate: number
  payment_type: string
  out_trade_no: string
  status: OrderStatus
  order_type: OrderType
  created_at: string
  expires_at: string
  paid_at?: string
  completed_at?: string
  refund_amount: number
  refund_reason?: string
  refund_requested_at?: string
  refund_requested_by?: number
  refund_request_reason?: string
  plan_id?: number
  provider_instance_id?: string
  manual_payment?: ManualPaymentOrderMeta
}

// ==================== Plans & Channels ====================

export interface SubscriptionPlan {
  id: number
  group_id: number
  group_platform?: string
  group_name?: string
  rate_multiplier?: number
  daily_limit_usd?: number | null
  weekly_limit_usd?: number | null
  monthly_limit_usd?: number | null
  supported_model_scopes?: string[]
  name: string
  description: string
  price: number
  original_price?: number
  validity_days: number
  validity_unit: string
  /** Stored as JSON string in backend; API layer should parse before use */
  features: string[]
  for_sale: boolean
  sort_order: number
}

export interface PaymentChannel {
  id: number
  group_id?: number
  name: string
  platform: string
  rate_multiplier: number
  description: string
  models: string[]
  features: string[]
  enabled: boolean
}

// ==================== Providers ====================

export interface ProviderInstance {
  id: number
  provider_key: string
  name: string
  config: Record<string, string>
  supported_types: string[]
  enabled: boolean
  payment_mode: string
  refund_enabled: boolean
  allow_user_refund: boolean
  limits: string
  sort_order: number
}

// ==================== Request / Response ====================

export interface CreateOrderRequest {
  amount: number
  payment_type: string
  order_type: string
  plan_id?: number
  return_url?: string
  payment_source?: string
  openid?: string
  wechat_resume_token?: string
  is_mobile?: boolean
}

export type CreateOrderResultType = 'order_created' | 'oauth_required' | 'jsapi_ready'

export interface WechatOAuthInfo {
  authorize_url?: string
  appid?: string
  openid?: string
  scope?: string
  state?: string
  redirect_url?: string
}

export interface WechatJSAPIPayload {
  appId?: string
  timeStamp?: string
  nonceStr?: string
  package?: string
  signType?: string
  paySign?: string
}

export interface CreateOrderResult {
  order_id: number
  amount: number
  pay_url?: string
  qr_code?: string
  client_secret?: string
  intent_id?: string
  currency?: string
  country_code?: string
  payment_env?: string
  pay_amount: number
  fee_rate: number
  expires_at: string
  result_type?: CreateOrderResultType
  payment_type?: string
  out_trade_no?: string
  payment_mode?: string
  resume_token?: string
  manual_payment?: ManualPaymentOrderMeta
  oauth?: WechatOAuthInfo
  jsapi?: WechatJSAPIPayload
  jsapi_payload?: WechatJSAPIPayload
}

export interface SupportConversation {
  id: number
  user_id: number
  order_id?: number | null
  subject: string
  status: 'open' | 'closed' | string
  created_at: string
  updated_at: string
  last_message_at: string
  last_user_message_at?: string
  last_admin_message_at?: string
  last_user_read_at?: string
  last_admin_read_at?: string
  unread_count?: number
  last_message_preview?: string
  last_message_sender_type?: string
  user_has_unread?: boolean
  admin_has_unread?: boolean
  order?: PaymentOrder
}

export interface SupportMessage {
  id: number
  conversation_id: number
  sender_type: 'user' | 'admin' | string
  sender_user_id?: number | null
  message: string
  created_at: string
}

export interface SupportConversationDetail {
  conversation: SupportConversation
  messages: SupportMessage[]
}

export interface DashboardStats {
  today_amount: number
  total_amount: number
  today_count: number
  total_count: number
  avg_amount: number
  daily_series: { date: string; amount: number; count: number }[]
  payment_methods: { type: string; amount: number; count: number }[]
  top_users: { user_id: number; email: string; amount: number }[]
}
