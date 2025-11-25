package milvus

import (
	"context"
	"fmt"

	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"go.uber.org/zap"
)

// CollectionService Collection 操作服务
type CollectionService struct {
	client *Client
	logger *zap.Logger
}

// NewCollectionService 创建 Collection 服务
func NewCollectionService(client *Client, logger *zap.Logger) *CollectionService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &CollectionService{
		client: client,
		logger: logger,
	}
}

// CreateCollection 创建集合
func (s *CollectionService) CreateCollection(ctx context.Context, schema *entity.Schema) error {
	milvusClient := s.client.GetClient()

	err := milvusClient.CreateCollection(ctx, schema, entity.DefaultShardNumber)
	if err != nil {
		s.logger.Error("create collection failed",
			zap.String("collection", schema.CollectionName),
			zap.Error(err),
		)
		return fmt.Errorf("create collection failed: %w", err)
	}

	s.logger.Info("collection created",
		zap.String("collection", schema.CollectionName),
	)
	return nil
}

// DropCollection 删除集合
func (s *CollectionService) DropCollection(ctx context.Context, collectionName string) error {
	milvusClient := s.client.GetClient()

	err := milvusClient.DropCollection(ctx, collectionName)
	if err != nil {
		s.logger.Error("drop collection failed",
			zap.String("collection", collectionName),
			zap.Error(err),
		)
		return fmt.Errorf("drop collection failed: %w", err)
	}

	s.logger.Info("collection dropped",
		zap.String("collection", collectionName),
	)
	return nil
}

// HasCollection 检查集合是否存在
func (s *CollectionService) HasCollection(ctx context.Context, collectionName string) (bool, error) {
	milvusClient := s.client.GetClient()

	exists, err := milvusClient.HasCollection(ctx, collectionName)
	if err != nil {
		s.logger.Error("check collection existence failed",
			zap.String("collection", collectionName),
			zap.Error(err),
		)
		return false, fmt.Errorf("check collection existence failed: %w", err)
	}

	return exists, nil
}

// ListCollections 列出所有集合
func (s *CollectionService) ListCollections(ctx context.Context) ([]*entity.Collection, error) {
	milvusClient := s.client.GetClient()

	collections, err := milvusClient.ListCollections(ctx)
	if err != nil {
		s.logger.Error("list collections failed", zap.Error(err))
		return nil, fmt.Errorf("list collections failed: %w", err)
	}

	return collections, nil
}

// LoadCollection 加载集合到内存
func (s *CollectionService) LoadCollection(ctx context.Context, collectionName string) error {
	milvusClient := s.client.GetClient()

	err := milvusClient.LoadCollection(ctx, collectionName, false)
	if err != nil {
		s.logger.Error("load collection failed",
			zap.String("collection", collectionName),
			zap.Error(err),
		)
		return fmt.Errorf("load collection failed: %w", err)
	}

	s.logger.Info("collection loaded",
		zap.String("collection", collectionName),
	)
	return nil
}

// ReleaseCollection 释放集合内存
func (s *CollectionService) ReleaseCollection(ctx context.Context, collectionName string) error {
	milvusClient := s.client.GetClient()

	err := milvusClient.ReleaseCollection(ctx, collectionName)
	if err != nil {
		s.logger.Error("release collection failed",
			zap.String("collection", collectionName),
			zap.Error(err),
		)
		return fmt.Errorf("release collection failed: %w", err)
	}

	s.logger.Info("collection released",
		zap.String("collection", collectionName),
	)
	return nil
}

// GetCollectionSchema 获取集合 Schema
func (s *CollectionService) GetCollectionSchema(ctx context.Context, collectionName string) (*entity.Collection, error) {
	milvusClient := s.client.GetClient()

	schema, err := milvusClient.DescribeCollection(ctx, collectionName)
	if err != nil {
		s.logger.Error("get collection schema failed",
			zap.String("collection", collectionName),
			zap.Error(err),
		)
		return nil, fmt.Errorf("get collection schema failed: %w", err)
	}

	return schema, nil
}

// GetCollectionStats 获取集合统计信息
func (s *CollectionService) GetCollectionStats(ctx context.Context, collectionName string) (map[string]interface{}, error) {
	milvusClient := s.client.GetClient()

	stats, err := milvusClient.GetCollectionStatistics(ctx, collectionName)
	if err != nil {
		s.logger.Error("get collection stats failed",
			zap.String("collection", collectionName),
			zap.Error(err),
		)
		return nil, fmt.Errorf("get collection stats failed: %w", err)
	}

	result := map[string]interface{}{
		"row_count": stats["row_count"],
	}

	return result, nil
}
