package milvus

import (
	"context"
	"fmt"
	"strconv"

	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"go.uber.org/zap"
)

// IndexService 索引操作服务
type IndexService struct {
	client *Client
	logger *zap.Logger
}

// NewIndexService 创建索引服务
func NewIndexService(client *Client, logger *zap.Logger) *IndexService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &IndexService{
		client: client,
		logger: logger,
	}
}

// CreateIndex 创建索引
func (s *IndexService) CreateIndex(ctx context.Context, collectionName string, fieldName string, index entity.Index) error {
	milvusClient := s.client.GetClient()

	err := milvusClient.CreateIndex(ctx, collectionName, fieldName, index, false)
	if err != nil {
		s.logger.Error("create index failed",
			zap.String("collection", collectionName),
			zap.String("field", fieldName),
			zap.Error(err),
		)
		return fmt.Errorf("create index failed: %w", err)
	}

	s.logger.Info("index created",
		zap.String("collection", collectionName),
		zap.String("field", fieldName),
		zap.String("index_type", string(index.IndexType())),
	)
	return nil
}

// DropIndex 删除索引
func (s *IndexService) DropIndex(ctx context.Context, collectionName string, fieldName string) error {
	milvusClient := s.client.GetClient()

	err := milvusClient.DropIndex(ctx, collectionName, fieldName)
	if err != nil {
		s.logger.Error("drop index failed",
			zap.String("collection", collectionName),
			zap.String("field", fieldName),
			zap.Error(err),
		)
		return fmt.Errorf("drop index failed: %w", err)
	}

	s.logger.Info("index dropped",
		zap.String("collection", collectionName),
		zap.String("field", fieldName),
	)
	return nil
}

// DescribeIndex 描述索引
func (s *IndexService) DescribeIndex(ctx context.Context, collectionName string, fieldName string) ([]entity.Index, error) {
	milvusClient := s.client.GetClient()

	indexes, err := milvusClient.DescribeIndex(ctx, collectionName, fieldName)
	if err != nil {
		s.logger.Error("describe index failed",
			zap.String("collection", collectionName),
			zap.String("field", fieldName),
			zap.Error(err),
		)
		return nil, fmt.Errorf("describe index failed: %w", err)
	}

	return indexes, nil
}

// NewHNSWIndex 创建 HNSW 索引（推荐用于向量搜索）
func NewHNSWIndex(metricType entity.MetricType, M int, efConstruction int) entity.Index {
	params := map[string]string{
		"metric_type":    string(metricType),
		"M":              strconv.Itoa(M),
		"efConstruction": strconv.Itoa(efConstruction),
	}
	return entity.NewGenericIndex("", entity.HNSW, params)
}

// NewIVFFlatIndex 创建 IVF_FLAT 索引
func NewIVFFlatIndex(metricType entity.MetricType, nlist int) entity.Index {
	params := map[string]string{
		"metric_type": string(metricType),
		"nlist":       strconv.Itoa(nlist),
	}
	return entity.NewGenericIndex("", entity.IvfFlat, params)
}

// NewFLATIndex 创建 FLAT 索引（精确搜索，小数据集）
func NewFLATIndex(metricType entity.MetricType) entity.Index {
	return entity.NewGenericIndex("", entity.Flat, map[string]string{
		"metric_type": string(metricType),
	})
}
