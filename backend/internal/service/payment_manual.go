package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	PaymentSourceManualAlipay = "manual_alipay"
	PaymentSourceManualWxpay  = "manual_wxpay"

	ManualReviewStatusPendingUser  = "PENDING_USER_PROOF"
	ManualReviewStatusPendingAdmin = "PENDING_ADMIN_REVIEW"
	ManualReviewStatusApproved     = "APPROVED"
	ManualReviewStatusRejected     = "REJECTED"

	SettingManualPaymentEnabled            = "MANUAL_PAYMENT_ENABLED"
	SettingManualPaymentRequireProof       = "MANUAL_PAYMENT_REQUIRE_PROOF"
	SettingManualPaymentAlipayEnabled      = "MANUAL_PAYMENT_ALIPAY_ENABLED"
	SettingManualPaymentWechatEnabled      = "MANUAL_PAYMENT_WECHAT_ENABLED"
	SettingManualPaymentAlipayQRImageURL   = "MANUAL_PAYMENT_ALIPAY_QR_IMAGE_URL"
	SettingManualPaymentWechatQRImageURL   = "MANUAL_PAYMENT_WECHAT_QR_IMAGE_URL"
	SettingManualPaymentHelpText           = "MANUAL_PAYMENT_HELP_TEXT"
	SettingManualPaymentReviewTimeoutMins  = "MANUAL_PAYMENT_REVIEW_TIMEOUT_MINUTES"
	defaultManualPaymentReviewTimeoutMins  = 1440
	defaultManualPaymentRequireProof       = true
	defaultManualPaymentEnabled            = false
)

type ManualPaymentConfig struct {
	Enabled              bool   `json:"enabled"`
	RequireProof         bool   `json:"require_proof"`
	AlipayEnabled        bool   `json:"alipay_enabled"`
	WechatEnabled        bool   `json:"wechat_enabled"`
	AlipayQRCodeImageURL string `json:"alipay_qr_code_image_url"`
	WechatQRCodeImageURL string `json:"wechat_qr_code_image_url"`
	HelpText             string `json:"help_text"`
	ReviewTimeoutMinutes int    `json:"review_timeout_minutes"`
}

type ManualPaymentOrderMeta struct {
	PaymentSource         string     `json:"payment_source,omitempty"`
	ReviewStatus          string     `json:"review_status,omitempty"`
	RequireProof          bool       `json:"require_proof,omitempty"`
	QRCodeImageURL        string     `json:"qr_code_image_url,omitempty"`
	ProofImageURL         string     `json:"proof_image_url,omitempty"`
	ProofNote             string     `json:"proof_note,omitempty"`
	ProofSubmittedAt      *time.Time `json:"proof_submitted_at,omitempty"`
	ReviewedAt            *time.Time `json:"reviewed_at,omitempty"`
	ReviewedBy            string     `json:"reviewed_by,omitempty"`
	ReviewNote            string     `json:"review_note,omitempty"`
	ReviewTimeoutMinutes  int        `json:"review_timeout_minutes,omitempty"`
}

type ManualPaymentOrderMetaResponse struct {
	Enabled               bool       `json:"enabled"`
	PaymentSource         string     `json:"payment_source,omitempty"`
	ReviewStatus          string     `json:"review_status,omitempty"`
	RequireProof          bool       `json:"require_proof,omitempty"`
	QRCodeImageURL        string     `json:"qr_code_image_url,omitempty"`
	ProofImageURL         string     `json:"proof_image_url,omitempty"`
	ProofNote             string     `json:"proof_note,omitempty"`
	ProofSubmittedAt      *time.Time `json:"proof_submitted_at,omitempty"`
	ReviewedAt            *time.Time `json:"reviewed_at,omitempty"`
	ReviewedBy            string     `json:"reviewed_by,omitempty"`
	ReviewNote            string     `json:"review_note,omitempty"`
	ReviewTimeoutMinutes  int        `json:"review_timeout_minutes,omitempty"`
}

type SubmitManualProofRequest struct {
	ProofImageURL string
	ProofNote     string
}

type ReviewManualPaymentRequest struct {
	Approved bool
	Note     string
	Operator string
}

func isManualPaymentSource(source string) bool {
	switch NormalizePaymentSource(source) {
	case PaymentSourceManualAlipay, PaymentSourceManualWxpay:
		return true
	default:
		return false
	}
}

func manualPaymentTypeFromSource(source string) string {
	switch NormalizePaymentSource(source) {
	case PaymentSourceManualAlipay:
		return payment.TypeAlipay
	case PaymentSourceManualWxpay:
		return payment.TypeWxpay
	default:
		return ""
	}
}

func normalizeManualReviewStatus(status string) string {
	switch strings.TrimSpace(strings.ToUpper(status)) {
	case ManualReviewStatusPendingUser,
		ManualReviewStatusPendingAdmin,
		ManualReviewStatusApproved,
		ManualReviewStatusRejected:
		return strings.TrimSpace(strings.ToUpper(status))
	default:
		return ""
	}
}

func (s *PaymentConfigService) GetManualPaymentConfig(ctx context.Context) (*ManualPaymentConfig, error) {
	if s == nil || s.settingRepo == nil {
		return &ManualPaymentConfig{
			Enabled:              defaultManualPaymentEnabled,
			RequireProof:         defaultManualPaymentRequireProof,
			ReviewTimeoutMinutes: defaultManualPaymentReviewTimeoutMins,
		}, nil
	}
	keys := []string{
		SettingManualPaymentEnabled,
		SettingManualPaymentRequireProof,
		SettingManualPaymentAlipayEnabled,
		SettingManualPaymentWechatEnabled,
		SettingManualPaymentAlipayQRImageURL,
		SettingManualPaymentWechatQRImageURL,
		SettingManualPaymentHelpText,
		SettingManualPaymentReviewTimeoutMins,
	}
	vals, err := s.settingRepo.GetMultiple(ctx, keys)
	if err != nil {
		return nil, fmt.Errorf("get manual payment config settings: %w", err)
	}
	return &ManualPaymentConfig{
		Enabled:              parseBoolWithDefault(vals[SettingManualPaymentEnabled], defaultManualPaymentEnabled),
		RequireProof:         parseBoolWithDefault(vals[SettingManualPaymentRequireProof], defaultManualPaymentRequireProof),
		AlipayEnabled:        vals[SettingManualPaymentAlipayEnabled] == "true",
		WechatEnabled:        vals[SettingManualPaymentWechatEnabled] == "true",
		AlipayQRCodeImageURL: strings.TrimSpace(vals[SettingManualPaymentAlipayQRImageURL]),
		WechatQRCodeImageURL: strings.TrimSpace(vals[SettingManualPaymentWechatQRImageURL]),
		HelpText:             vals[SettingManualPaymentHelpText],
		ReviewTimeoutMinutes: pcParseInt(vals[SettingManualPaymentReviewTimeoutMins], defaultManualPaymentReviewTimeoutMins),
	}, nil
}

func parseBoolWithDefault(raw string, fallback bool) bool {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "true":
		return true
	case "false":
		return false
	default:
		return fallback
	}
}

func (s *PaymentService) validateManualPaymentRequest(ctx context.Context, req CreateOrderRequest, cfg *PaymentConfig) (*ManualPaymentConfig, error) {
	source := NormalizePaymentSource(req.PaymentSource)
	if !isManualPaymentSource(source) {
		return nil, nil
	}
	manualCfg, err := s.configService.GetManualPaymentConfig(ctx)
	if err != nil {
		return nil, err
	}
	if manualCfg == nil || !manualCfg.Enabled {
		return nil, infraerrors.Forbidden("MANUAL_PAYMENT_DISABLED", "manual qr payment is disabled")
	}
	if req.OrderType == payment.OrderTypeBalance && cfg.BalanceDisabled {
		return nil, infraerrors.Forbidden("BALANCE_PAYMENT_DISABLED", "balance recharge has been disabled")
	}
	switch source {
	case PaymentSourceManualAlipay:
		if !manualCfg.AlipayEnabled || strings.TrimSpace(manualCfg.AlipayQRCodeImageURL) == "" {
			return nil, infraerrors.Forbidden("MANUAL_PAYMENT_METHOD_DISABLED", "manual alipay payment is unavailable")
		}
	case PaymentSourceManualWxpay:
		if !manualCfg.WechatEnabled || strings.TrimSpace(manualCfg.WechatQRCodeImageURL) == "" {
			return nil, infraerrors.Forbidden("MANUAL_PAYMENT_METHOD_DISABLED", "manual wechat payment is unavailable")
		}
	default:
		return nil, infraerrors.BadRequest("INVALID_PAYMENT_SOURCE", "unsupported manual payment source")
	}
	if expected := manualPaymentTypeFromSource(source); expected != "" && NormalizeVisibleMethod(req.PaymentType) != expected {
		return nil, infraerrors.BadRequest("INVALID_PAYMENT_SOURCE", "payment source does not match payment type")
	}
	return manualCfg, nil
}

func (s *PaymentService) resolveManualPaymentSource(ctx context.Context, req CreateOrderRequest) (string, *ManualPaymentConfig, error) {
	source := NormalizePaymentSource(req.PaymentSource)
	if isManualPaymentSource(source) {
		cfg, err := s.configService.GetManualPaymentConfig(ctx)
		if err != nil {
			return source, nil, err
		}
		return source, cfg, nil
	}

	method := NormalizeVisibleMethod(req.PaymentType)
	if method != payment.TypeAlipay && method != payment.TypeWxpay {
		return source, nil, nil
	}
	if s == nil || s.configService == nil || s.configService.settingRepo == nil {
		return source, nil, nil
	}
	sourceKey := visibleMethodSourceSettingKey(method)
	if sourceKey == "" {
		return source, nil, nil
	}
	value, err := s.configService.settingRepo.GetValue(ctx, sourceKey)
	if err != nil {
		return source, nil, nil
	}
	switch NormalizeVisibleMethodSource(method, value) {
	case VisibleMethodSourceManualAlipay:
		cfg, cfgErr := s.configService.GetManualPaymentConfig(ctx)
		return PaymentSourceManualAlipay, cfg, cfgErr
	case VisibleMethodSourceManualWechat:
		cfg, cfgErr := s.configService.GetManualPaymentConfig(ctx)
		return PaymentSourceManualWxpay, cfg, cfgErr
	default:
		return source, nil, nil
	}
}

func manualPaymentQRCodeForSource(cfg *ManualPaymentConfig, source string) string {
	if cfg == nil {
		return ""
	}
	switch NormalizePaymentSource(source) {
	case PaymentSourceManualAlipay:
		return strings.TrimSpace(cfg.AlipayQRCodeImageURL)
	case PaymentSourceManualWxpay:
		return strings.TrimSpace(cfg.WechatQRCodeImageURL)
	default:
		return ""
	}
}

func manualPaymentProviderSnapshot(cfg *ManualPaymentConfig, source string) map[string]any {
	qrCodeImageURL := manualPaymentQRCodeForSource(cfg, source)
	reviewTimeoutMinutes := defaultManualPaymentReviewTimeoutMins
	requireProof := defaultManualPaymentRequireProof
	if cfg != nil {
		requireProof = cfg.RequireProof
		if cfg.ReviewTimeoutMinutes > 0 {
			reviewTimeoutMinutes = cfg.ReviewTimeoutMinutes
		}
	}
	return map[string]any{
		"schema_version": 2,
		"manual_payment": map[string]any{
			"payment_source":          NormalizePaymentSource(source),
			"review_status":           initialManualReviewStatus(requireProof),
			"require_proof":           requireProof,
			"qr_code_image_url":       qrCodeImageURL,
			"review_timeout_minutes":  reviewTimeoutMinutes,
		},
	}
}

func initialManualReviewStatus(requireProof bool) string {
	if requireProof {
		return ManualReviewStatusPendingUser
	}
	return ManualReviewStatusPendingAdmin
}

func extractManualPaymentOrderMeta(order *dbent.PaymentOrder) ManualPaymentOrderMeta {
	if order == nil || order.ProviderSnapshot == nil {
		return ManualPaymentOrderMeta{}
	}
	raw, ok := order.ProviderSnapshot["manual_payment"]
	if !ok {
		return ManualPaymentOrderMeta{}
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return ManualPaymentOrderMeta{}
	}
	var meta ManualPaymentOrderMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return ManualPaymentOrderMeta{}
	}
	meta.PaymentSource = NormalizePaymentSource(meta.PaymentSource)
	meta.ReviewStatus = normalizeManualReviewStatus(meta.ReviewStatus)
	return meta
}

func ExtractManualPaymentOrderMetaForResponse(order *dbent.PaymentOrder) ManualPaymentOrderMetaResponse {
	meta := extractManualPaymentOrderMeta(order)
	if !isManualPaymentSource(meta.PaymentSource) {
		return ManualPaymentOrderMetaResponse{}
	}
	return ManualPaymentOrderMetaResponse{
		Enabled:              true,
		PaymentSource:        meta.PaymentSource,
		ReviewStatus:         meta.ReviewStatus,
		RequireProof:         meta.RequireProof,
		QRCodeImageURL:       meta.QRCodeImageURL,
		ProofImageURL:        meta.ProofImageURL,
		ProofNote:            meta.ProofNote,
		ProofSubmittedAt:     meta.ProofSubmittedAt,
		ReviewedAt:           meta.ReviewedAt,
		ReviewedBy:           meta.ReviewedBy,
		ReviewNote:           meta.ReviewNote,
		ReviewTimeoutMinutes: meta.ReviewTimeoutMinutes,
	}
}

func mergeManualPaymentOrderMeta(snapshot map[string]any, meta ManualPaymentOrderMeta) map[string]any {
	next := cloneSnapshotMap(snapshot)
	if next == nil {
		next = map[string]any{}
	}
	if _, ok := next["schema_version"]; !ok {
		next["schema_version"] = 2
	}
	data, _ := json.Marshal(meta)
	var payload map[string]any
	_ = json.Unmarshal(data, &payload)
	next["manual_payment"] = payload
	return next
}

func cloneSnapshotMap(input map[string]any) map[string]any {
	if len(input) == 0 {
		return map[string]any{}
	}
	data, err := json.Marshal(input)
	if err != nil {
		return map[string]any{}
	}
	var cloned map[string]any
	if err := json.Unmarshal(data, &cloned); err != nil {
		return map[string]any{}
	}
	return cloned
}

func validateImageDataURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return infraerrors.BadRequest("INVALID_MANUAL_PAYMENT_IMAGE", "image is required")
	}
	if len(raw) > 4*1024*1024 {
		return infraerrors.BadRequest("INVALID_MANUAL_PAYMENT_IMAGE", "image is too large")
	}
	if !strings.HasPrefix(strings.ToLower(raw), "data:image/") {
		return infraerrors.BadRequest("INVALID_MANUAL_PAYMENT_IMAGE", "only data:image URLs are supported")
	}
	return nil
}

func (s *PaymentService) SubmitManualPaymentProof(ctx context.Context, orderID, userID int64, req SubmitManualProofRequest) (*dbent.PaymentOrder, error) {
	order, err := s.GetOrder(ctx, orderID, userID)
	if err != nil {
		return nil, err
	}
	meta := extractManualPaymentOrderMeta(order)
	if !isManualPaymentSource(meta.PaymentSource) {
		return nil, infraerrors.BadRequest("INVALID_MANUAL_PAYMENT_ORDER", "order is not a manual payment order")
	}
	if order.Status != OrderStatusPending {
		return nil, infraerrors.BadRequest("INVALID_STATUS", "manual payment proof can only be submitted for pending orders")
	}
	if err := validateImageDataURL(req.ProofImageURL); err != nil {
		return nil, err
	}
	now := time.Now()
	meta.ProofImageURL = strings.TrimSpace(req.ProofImageURL)
	meta.ProofNote = strings.TrimSpace(req.ProofNote)
	meta.ProofSubmittedAt = &now
	meta.ReviewStatus = ManualReviewStatusPendingAdmin
	meta.ReviewedAt = nil
	meta.ReviewedBy = ""
	meta.ReviewNote = ""

	updated, err := s.entClient.PaymentOrder.UpdateOneID(order.ID).
		SetProviderSnapshot(mergeManualPaymentOrderMeta(order.ProviderSnapshot, meta)).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update manual payment proof: %w", err)
	}
	s.writeAuditLog(ctx, order.ID, "MANUAL_PAYMENT_PROOF_SUBMITTED", fmt.Sprintf("user:%d", userID), map[string]any{
		"paymentSource": meta.PaymentSource,
		"proofNote":     meta.ProofNote,
	})
	return updated, nil
}

func (s *PaymentService) ReviewManualPayment(ctx context.Context, orderID int64, req ReviewManualPaymentRequest) (*dbent.PaymentOrder, error) {
	order, err := s.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	meta := extractManualPaymentOrderMeta(order)
	if !isManualPaymentSource(meta.PaymentSource) {
		return nil, infraerrors.BadRequest("INVALID_MANUAL_PAYMENT_ORDER", "order is not a manual payment order")
	}
	if order.Status != OrderStatusPending {
		return nil, infraerrors.BadRequest("INVALID_STATUS", "manual payment order cannot be reviewed in current status")
	}
	if meta.RequireProof && strings.TrimSpace(meta.ProofImageURL) == "" {
		return nil, infraerrors.BadRequest("MANUAL_PAYMENT_PROOF_REQUIRED", "payment proof has not been submitted yet")
	}

	now := time.Now()
	meta.ReviewedAt = &now
	meta.ReviewedBy = strings.TrimSpace(req.Operator)
	meta.ReviewNote = strings.TrimSpace(req.Note)

	if req.Approved {
		meta.ReviewStatus = ManualReviewStatusApproved
		tradeNo := order.PaymentTradeNo
		if strings.TrimSpace(tradeNo) == "" {
			tradeNo = fmt.Sprintf("manual-%d", order.ID)
		}
		updated, err := s.entClient.PaymentOrder.UpdateOneID(order.ID).
			SetStatus(OrderStatusPaid).
			SetPaymentTradeNo(tradeNo).
			SetPaidAt(now).
			SetProviderSnapshot(mergeManualPaymentOrderMeta(order.ProviderSnapshot, meta)).
			Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("approve manual payment order: %w", err)
		}
		s.writeAuditLog(ctx, order.ID, "MANUAL_PAYMENT_APPROVED", req.Operator, map[string]any{
			"paymentSource": meta.PaymentSource,
			"reviewNote":    meta.ReviewNote,
		})
		if err := s.executeFulfillment(ctx, updated.ID); err != nil {
			return nil, err
		}
		return s.GetOrderByID(ctx, updated.ID)
	}

	meta.ReviewStatus = ManualReviewStatusRejected
	updated, err := s.entClient.PaymentOrder.UpdateOneID(order.ID).
		SetProviderSnapshot(mergeManualPaymentOrderMeta(order.ProviderSnapshot, meta)).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("reject manual payment order: %w", err)
	}
	s.writeAuditLog(ctx, order.ID, "MANUAL_PAYMENT_REJECTED", req.Operator, map[string]any{
		"paymentSource": meta.PaymentSource,
		"reviewNote":    meta.ReviewNote,
	})
	return updated, nil
}
