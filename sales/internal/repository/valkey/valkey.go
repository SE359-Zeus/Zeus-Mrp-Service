package valkey

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"zeus-sales-service/internal/models"
	rootrepo "zeus-sales-service/internal/repository"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Repository struct {
	client        *redis.Client
	queueKey      string
	payloadKey    string
	atpPrefix     string
	reservePrefix string
}

func (repo *Repository) Close() error {
	if repo == nil || repo.client == nil {
		return nil
	}
	return repo.client.Close()
}

func New(client *redis.Client) *Repository {
	return &Repository{
		client:        client,
		queueKey:      "sales:allocation_queue",
		payloadKey:    "sales:allocation_queue:payload",
		atpPrefix:     "sales:atp:",
		reservePrefix: "sales:reservation:",
	}
}

func (repo *Repository) EnqueueOrder(ctx context.Context, entry models.AllocationQueueEntry) error {
	payload, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	score := float64(entry.PriorityScore)*1e16 + float64(entry.IngestedAt.UnixMicro())
	member := entry.OrderID.String()
	pipeline := repo.client.TxPipeline()
	pipeline.HSet(ctx, repo.payloadKey, member, payload)
	pipeline.ZAdd(ctx, repo.queueKey, redis.Z{Score: score, Member: member})
	_, err = pipeline.Exec(ctx)
	return err
}

func (repo *Repository) DequeueOrder(ctx context.Context) (*models.AllocationQueueEntry, error) {
	result, err := repo.client.ZPopMin(ctx, repo.queueKey, 1).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	if len(result) == 0 {
		return nil, nil
	}
	member := result[0].Member.(string)
	payload, err := repo.client.HGet(ctx, repo.payloadKey, member).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	if err := repo.client.HDel(ctx, repo.payloadKey, member).Err(); err != nil {
		return nil, err
	}
	var entry models.AllocationQueueEntry
	if err := json.Unmarshal(payload, &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

func (repo *Repository) ListQueue(ctx context.Context) ([]models.AllocationQueueEntry, error) {
	members, err := repo.client.ZRange(ctx, repo.queueKey, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	entries := make([]models.AllocationQueueEntry, 0, len(members))
	for _, member := range members {
		payload, err := repo.client.HGet(ctx, repo.payloadKey, member).Bytes()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			return nil, err
		}
		var entry models.AllocationQueueEntry
		if err := json.Unmarshal(payload, &entry); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (repo *Repository) ClearQueue(ctx context.Context) error {
	return repo.client.Del(ctx, repo.queueKey, repo.payloadKey).Err()
}

func (repo *Repository) GetATP(ctx context.Context, sku string) (int, error) {
	value, err := repo.client.Get(ctx, repo.atpPrefix+strings.ToUpper(strings.TrimSpace(sku))).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, err
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	return parsed, nil
}

func (repo *Repository) SetATP(ctx context.Context, sku string, quantity int) error {
	return repo.client.Set(ctx, repo.atpPrefix+strings.ToUpper(strings.TrimSpace(sku)), quantity, 0).Err()
}

func (repo *Repository) ReserveInventory(ctx context.Context, orderID uuid.UUID, items []models.ReservationItem) error {
	pipe := repo.client.TxPipeline()
	reservationKey := repo.reservePrefix + orderID.String()
	for _, item := range items {
		key := repo.atpPrefix + strings.ToUpper(strings.TrimSpace(item.SKU))
		current, err := repo.GetATP(ctx, item.SKU)
		if err != nil {
			return err
		}
		if current < item.Quantity {
			return fmt.Errorf("insufficient inventory for %s", item.SKU)
		}
		pipe.DecrBy(ctx, key, int64(item.Quantity))
		pipe.HSet(ctx, reservationKey, strings.ToUpper(strings.TrimSpace(item.SKU)), item.Quantity)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (repo *Repository) ReleaseInventory(ctx context.Context, orderID uuid.UUID) error {
	reservationKey := repo.reservePrefix + orderID.String()
	entries, err := repo.client.HGetAll(ctx, reservationKey).Result()
	if err != nil {
		return err
	}
	pipe := repo.client.TxPipeline()
	for sku, quantityText := range entries {
		quantity, err := strconv.Atoi(quantityText)
		if err != nil {
			return err
		}
		pipe.IncrBy(ctx, repo.atpPrefix+sku, int64(quantity))
	}
	pipe.Del(ctx, reservationKey)
	_, err = pipe.Exec(ctx)
	return err
}

var _ rootrepo.ValkeyRepository = (*Repository)(nil)
