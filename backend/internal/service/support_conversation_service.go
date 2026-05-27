package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

type SupportConversation struct {
	ID                 int64      `json:"id"`
	UserID             int64      `json:"user_id"`
	OrderID            int64      `json:"order_id"`
	Subject            string     `json:"subject"`
	Status             string     `json:"status"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	LastMessageAt      time.Time  `json:"last_message_at"`
	LastUserMessageAt  *time.Time `json:"last_user_message_at,omitempty"`
	LastAdminMessageAt *time.Time `json:"last_admin_message_at,omitempty"`
	Order              *dbent.PaymentOrder `json:"order,omitempty"`
}

type SupportMessage struct {
	ID             int64     `json:"id"`
	ConversationID int64     `json:"conversation_id"`
	SenderType     string    `json:"sender_type"`
	SenderUserID   *int64    `json:"sender_user_id,omitempty"`
	Message        string    `json:"message"`
	CreatedAt      time.Time `json:"created_at"`
}

type SupportConversationDetail struct {
	Conversation *SupportConversation `json:"conversation"`
	Messages     []*SupportMessage    `json:"messages"`
}

type SupportConversationListParams struct {
	Page      int
	PageSize  int
	Status    string
	UserID    int64
	Keyword   string
}

type CreateSupportConversationRequest struct {
	UserID  int64
	OrderID int64
	Message string
}

type ReplySupportConversationRequest struct {
	ConversationID int64
	SenderType     string
	SenderUserID   *int64
	Message        string
}

type SupportConversationService struct {
	db        *sql.DB
	entClient *dbent.Client
}

func NewSupportConversationService(db *sql.DB, entClient *dbent.Client) *SupportConversationService {
	return &SupportConversationService{db: db, entClient: entClient}
}

func (s *SupportConversationService) CreateOrAppendOrderConversation(ctx context.Context, req CreateSupportConversationRequest) (*SupportConversationDetail, error) {
	if s == nil || s.db == nil || s.entClient == nil {
		return nil, infraerrors.InternalServer("SUPPORT_NOT_READY", "support conversation service is not ready")
	}
	message := strings.TrimSpace(req.Message)
	if message == "" {
		return nil, infraerrors.BadRequest("SUPPORT_MESSAGE_REQUIRED", "message is required")
	}
	order, err := s.entClient.PaymentOrder.Get(ctx, req.OrderID)
	if err != nil {
		return nil, infraerrors.NotFound("ORDER_NOT_FOUND", "order not found")
	}
	if order.UserID != req.UserID {
		return nil, infraerrors.Forbidden("FORBIDDEN", "no permission for this order")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin support conversation tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now()
	var conversationID int64
	query := `SELECT id FROM support_conversations WHERE user_id = $1 AND order_id = $2`
	if err := tx.QueryRowContext(ctx, query, req.UserID, req.OrderID).Scan(&conversationID); err != nil {
		if err == sql.ErrNoRows {
			insert := `
				INSERT INTO support_conversations (user_id, order_id, subject, status, created_at, updated_at, last_message_at, last_user_message_at)
				VALUES ($1, $2, $3, 'open', $4, $4, $4, $4)
				RETURNING id
			`
			subject := fmt.Sprintf("订单咨询 #%d", req.OrderID)
			if err := tx.QueryRowContext(ctx, insert, req.UserID, req.OrderID, subject, now).Scan(&conversationID); err != nil {
				return nil, fmt.Errorf("create support conversation: %w", err)
			}
		} else {
			return nil, fmt.Errorf("query support conversation: %w", err)
		}
	} else {
		update := `
			UPDATE support_conversations
			SET status = 'open', updated_at = $2, last_message_at = $2, last_user_message_at = $2
			WHERE id = $1
		`
		if _, err := tx.ExecContext(ctx, update, conversationID, now); err != nil {
			return nil, fmt.Errorf("update support conversation: %w", err)
		}
	}

	insertMessage := `
		INSERT INTO support_messages (conversation_id, sender_type, sender_user_id, message, created_at)
		VALUES ($1, 'user', $2, $3, $4)
	`
	if _, err := tx.ExecContext(ctx, insertMessage, conversationID, req.UserID, message, now); err != nil {
		return nil, fmt.Errorf("insert support message: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit support conversation tx: %w", err)
	}
	return s.GetConversationDetailForUser(ctx, conversationID, req.UserID)
}

func (s *SupportConversationService) ReplyToConversation(ctx context.Context, req ReplySupportConversationRequest) (*SupportConversationDetail, error) {
	if s == nil || s.db == nil {
		return nil, infraerrors.InternalServer("SUPPORT_NOT_READY", "support conversation service is not ready")
	}
	message := strings.TrimSpace(req.Message)
	if message == "" {
		return nil, infraerrors.BadRequest("SUPPORT_MESSAGE_REQUIRED", "message is required")
	}
	if req.SenderType != "user" && req.SenderType != "admin" {
		return nil, infraerrors.BadRequest("SUPPORT_SENDER_INVALID", "invalid sender type")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin support reply tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now()
	insertMessage := `
		INSERT INTO support_messages (conversation_id, sender_type, sender_user_id, message, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	if _, err := tx.ExecContext(ctx, insertMessage, req.ConversationID, req.SenderType, req.SenderUserID, message, now); err != nil {
		return nil, fmt.Errorf("insert support reply: %w", err)
	}

	lastUser := "last_user_message_at"
	if req.SenderType == "admin" {
		lastUser = "last_admin_message_at"
	}
	updateConversation := fmt.Sprintf(`
		UPDATE support_conversations
		SET status = 'open', updated_at = $2, last_message_at = $2, %s = $2
		WHERE id = $1
	`, lastUser)
	if _, err := tx.ExecContext(ctx, updateConversation, req.ConversationID, now); err != nil {
		return nil, fmt.Errorf("update support conversation reply metadata: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit support reply tx: %w", err)
	}
	return s.GetConversationDetailForAdmin(ctx, req.ConversationID)
}

func (s *SupportConversationService) ListUserConversations(ctx context.Context, userID int64) ([]*SupportConversation, error) {
	query := `
		SELECT id, user_id, order_id, subject, status, created_at, updated_at, last_message_at, last_user_message_at, last_admin_message_at
		FROM support_conversations
		WHERE user_id = $1
		ORDER BY last_message_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list user support conversations: %w", err)
	}
	defer rows.Close()
	return s.scanConversationRows(ctx, rows)
}

func (s *SupportConversationService) ListAdminConversations(ctx context.Context, p SupportConversationListParams) ([]*SupportConversation, int64, error) {
	pageSize, page := applyPagination(p.PageSize, p.Page)
	args := []any{}
	clauses := []string{"1=1"}
	argIdx := 1
	if p.Status != "" {
		clauses = append(clauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, p.Status)
		argIdx++
	}
	if p.UserID > 0 {
		clauses = append(clauses, fmt.Sprintf("user_id = $%d", argIdx))
		args = append(args, p.UserID)
		argIdx++
	}
	if kw := strings.TrimSpace(p.Keyword); kw != "" {
		clauses = append(clauses, fmt.Sprintf("(subject ILIKE $%d OR CAST(order_id AS TEXT) ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+kw+"%")
		argIdx++
	}
	where := strings.Join(clauses, " AND ")

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM support_conversations WHERE %s`, where)
	var total int64
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count admin support conversations: %w", err)
	}

	args = append(args, pageSize, (page-1)*pageSize)
	query := fmt.Sprintf(`
		SELECT id, user_id, order_id, subject, status, created_at, updated_at, last_message_at, last_user_message_at, last_admin_message_at
		FROM support_conversations
		WHERE %s
		ORDER BY last_message_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list admin support conversations: %w", err)
	}
	defer rows.Close()
	items, err := s.scanConversationRows(ctx, rows)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (s *SupportConversationService) GetConversationDetailForUser(ctx context.Context, conversationID, userID int64) (*SupportConversationDetail, error) {
	conversation, err := s.getConversation(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	if conversation.UserID != userID {
		return nil, infraerrors.Forbidden("FORBIDDEN", "no permission for this conversation")
	}
	messages, err := s.getConversationMessages(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	return &SupportConversationDetail{Conversation: conversation, Messages: messages}, nil
}

func (s *SupportConversationService) GetConversationDetailForAdmin(ctx context.Context, conversationID int64) (*SupportConversationDetail, error) {
	conversation, err := s.getConversation(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	messages, err := s.getConversationMessages(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	return &SupportConversationDetail{Conversation: conversation, Messages: messages}, nil
}

func (s *SupportConversationService) getConversation(ctx context.Context, conversationID int64) (*SupportConversation, error) {
	query := `
		SELECT id, user_id, order_id, subject, status, created_at, updated_at, last_message_at, last_user_message_at, last_admin_message_at
		FROM support_conversations
		WHERE id = $1
	`
	row := s.db.QueryRowContext(ctx, query, conversationID)
	var conv SupportConversation
	if err := row.Scan(&conv.ID, &conv.UserID, &conv.OrderID, &conv.Subject, &conv.Status, &conv.CreatedAt, &conv.UpdatedAt, &conv.LastMessageAt, &conv.LastUserMessageAt, &conv.LastAdminMessageAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, infraerrors.NotFound("SUPPORT_CONVERSATION_NOT_FOUND", "support conversation not found")
		}
		return nil, fmt.Errorf("get support conversation: %w", err)
	}
	if s.entClient != nil {
		if order, err := s.entClient.PaymentOrder.Get(ctx, conv.OrderID); err == nil {
			conv.Order = order
		}
	}
	return &conv, nil
}

func (s *SupportConversationService) getConversationMessages(ctx context.Context, conversationID int64) ([]*SupportMessage, error) {
	query := `
		SELECT id, conversation_id, sender_type, sender_user_id, message, created_at
		FROM support_messages
		WHERE conversation_id = $1
		ORDER BY created_at ASC, id ASC
	`
	rows, err := s.db.QueryContext(ctx, query, conversationID)
	if err != nil {
		return nil, fmt.Errorf("get support messages: %w", err)
	}
	defer rows.Close()
	var messages []*SupportMessage
	for rows.Next() {
		item := &SupportMessage{}
		if err := rows.Scan(&item.ID, &item.ConversationID, &item.SenderType, &item.SenderUserID, &item.Message, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan support message: %w", err)
		}
		messages = append(messages, item)
	}
	return messages, rows.Err()
}

func (s *SupportConversationService) scanConversationRows(ctx context.Context, rows *sql.Rows) ([]*SupportConversation, error) {
	var items []*SupportConversation
	for rows.Next() {
		item := &SupportConversation{}
		if err := rows.Scan(&item.ID, &item.UserID, &item.OrderID, &item.Subject, &item.Status, &item.CreatedAt, &item.UpdatedAt, &item.LastMessageAt, &item.LastUserMessageAt, &item.LastAdminMessageAt); err != nil {
			return nil, fmt.Errorf("scan support conversation: %w", err)
		}
		if s.entClient != nil {
			if order, err := s.entClient.PaymentOrder.Get(ctx, item.OrderID); err == nil {
				item.Order = order
			}
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
