package service

import (
	"context"
	"fmt"
	"time"

	"zeus-scm-service/internal/messaging"
	"zeus-scm-service/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IPOService interface {
	CreateDraft(ctx context.Context, vendorID uuid.UUID, targetBuild string) (*models.PurchaseOrder, error)
	AddLineItemWithLock(ctx context.Context, poID string, sku string, qty int) error
	ApprovePO(ctx context.Context, poID string) error
	TransitionState(ctx context.Context, poID string, newState models.POStatus) error
}

type poService struct {
	db *gorm.DB
	mq *messaging.RabbitMQ
}

func NewPOService(db *gorm.DB, mq *messaging.RabbitMQ) IPOService {
	return &poService{db: db, mq: mq}
}

func (s *poService) CreateDraft(ctx context.Context, vendorID uuid.UUID, targetBuild string) (*models.PurchaseOrder, error) {
	var existingPO models.PurchaseOrder
	if err := s.db.WithContext(ctx).
		Where("vendor_id = ? AND status IN ?", vendorID, []models.POStatus{models.POStatusDraft, models.POStatusApproved, models.POStatusInTransit}).
		First(&existingPO).Error; err == nil {
		return nil, ErrMonoVendorViolation
	}

	var count int64
	year := time.Now().Year()
	s.db.WithContext(ctx).Model(&models.PurchaseOrder{}).
		Where("id LIKE ?", fmt.Sprintf("PO-%d-%%", year)).
		Count(&count)

	po := &models.PurchaseOrder{
		ID:          fmt.Sprintf("PO-%d-%d", year, count+1),
		VendorID:    vendorID,
		TargetBuild: targetBuild,
		Status:      models.POStatusDraft,
		TotalValue:  0,
	}
	if err := s.db.WithContext(ctx).Create(po).Error; err != nil {
		return nil, err
	}
	return po, nil
}

func (s *poService) AddLineItemWithLock(ctx context.Context, poID string, sku string, qty int) error {
	var po models.PurchaseOrder
	if err := s.db.WithContext(ctx).First(&po, "id = ?", poID).Error; err != nil {
		return ErrNotFound
	}
	if po.Status != models.POStatusDraft {
		return ErrInvalidTransition
	}

	msgs, err := s.mq.ConsumeFromPool()
	if err != nil {
		return err
	}
	select {
	case msg, ok := <-msgs:
		if !ok {
			return ErrInsufficientDeficit
		}
		var deficit messaging.DeficitMessage
		if err := deficit.FromDelivery(msg); err != nil {
			s.mq.Nack(msg.DeliveryTag, true)
			return err
		}
		if deficit.SKU != sku || deficit.Qty < qty {
			s.mq.Nack(msg.DeliveryTag, true)
			return ErrInsufficientDeficit
		}
		reservedMsg := messaging.DeficitMessage{
			SKU: sku,
			Qty: qty,
		}
		if err := s.mq.PublishToReserved(reservedMsg); err != nil {
			s.mq.Nack(msg.DeliveryTag, true)
			return err
		}
		_ = s.mq.Ack(msg.DeliveryTag)
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		return ErrInsufficientDeficit
	}

	var catalog models.ComponentStock
	if err := s.db.WithContext(ctx).First(&catalog, "sku = ?", sku).Error; err != nil {
		return err
	}

	lineItem := &models.POLineItem{
		ID:         uuid.New(),
		POID:       poID,
		SKU:        sku,
		OrderedQty: qty,
		UnitPrice:  catalog.UnitCost,
	}
	return s.db.WithContext(ctx).Create(lineItem).Error
}

func (s *poService) ApprovePO(ctx context.Context, poID string) error {
	var po models.PurchaseOrder
	if err := s.db.WithContext(ctx).First(&po, "id = ?", poID).Error; err != nil {
		return ErrNotFound
	}
	if po.Status != models.POStatusDraft {
		return ErrInvalidTransition
	}

	po.Status = models.POStatusApproved
	var totalValue float64
	var lineItems []models.POLineItem
	s.db.WithContext(ctx).Where("po_id = ?", poID).Find(&lineItems)
	for _, item := range lineItems {
		totalValue += float64(item.OrderedQty) * item.UnitPrice
	}
	po.TotalValue = totalValue

	return s.db.WithContext(ctx).Save(&po).Error
}

func (s *poService) TransitionState(ctx context.Context, poID string, newState models.POStatus) error {
	var po models.PurchaseOrder
	if err := s.db.WithContext(ctx).First(&po, "id = ?", poID).Error; err != nil {
		return ErrNotFound
	}

	valid := validTransition(po.Status, newState)
	if !valid {
		return ErrStateRegression
	}

	return s.db.WithContext(ctx).Model(&po).Update("status", newState).Error
}

func validTransition(current, new models.POStatus) bool {
	order := []models.POStatus{
		models.POStatusDraft,
		models.POStatusApproved,
		models.POStatusInTransit,
		models.POStatusReceived,
		models.POStatusPartial,
		models.POStatusVoid,
	}
	currentIdx := -1
	newIdx := -1
	for i, s := range order {
		if s == current {
			currentIdx = i
		}
		if s == new {
			newIdx = i
		}
	}
	if currentIdx == -1 || newIdx == -1 {
		return false
	}
	if new == models.POStatusVoid && current == models.POStatusDraft {
		return true
	}
	return newIdx > currentIdx
}
