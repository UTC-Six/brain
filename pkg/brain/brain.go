package brain

import (
	"context"
	"fmt"

	"github.com/UTC-Six/brain/internal/milvus"
	"github.com/UTC-Six/brain/internal/nebula"
	"go.uber.org/zap"
)

// Brain 统一的数据访问层，封装 Nebula Graph 和 Milvus
type Brain struct {
	Nebula *nebula.Client `json:"nebula,omitempty" :"nebula"`
	Milvus *milvus.Client `json:"milvus,omitempty" :"milvus"`
	logger *zap.Logger    `json:"logger,omitempty" :"logger"`

	// Nebula 服务
	VertexService *nebula.VertexService `json:"vertex_service,omitempty" :"vertex_service"`
	EdgeService   *nebula.EdgeService   `json:"edge_service,omitempty" :"edge_service"`

	// Milvus 服务
	CollectionService *milvus.CollectionService `json:"collection_service,omitempty" :"collection_service"`
	VectorService     *milvus.VectorService     `json:"vector_service,omitempty" :"vector_service"`
	IndexService      *milvus.IndexService      `json:"index_service,omitempty" :"index_service"`
}

// Config 配置
type Config struct {
	Nebula nebula.Config
	Milvus milvus.Config
	Logger *zap.Logger
}

// New 创建新的 Brain 实例
func New(config Config) (*Brain, error) {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}

	// 初始化 Nebula Graph 客户端
	nebulaClient, err := nebula.NewClient(config.Nebula, config.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create nebula client: %w", err)
	}

	// 初始化 Milvus 客户端
	milvusClient, err := milvus.NewClient(config.Milvus, config.Logger)
	if err != nil {
		nebulaClient.Close()
		return nil, fmt.Errorf("failed to create milvus client: %w", err)
	}

	// 创建服务
	brain := &Brain{
		Nebula:            nebulaClient,
		Milvus:            milvusClient,
		logger:            config.Logger,
		VertexService:     nebula.NewVertexService(nebulaClient, config.Logger),
		EdgeService:       nebula.NewEdgeService(nebulaClient, config.Logger),
		CollectionService: milvus.NewCollectionService(milvusClient, config.Logger),
		VectorService:     milvus.NewVectorService(milvusClient, config.Logger),
		IndexService:      milvus.NewIndexService(milvusClient, config.Logger),
	}

	config.Logger.Info("brain initialized successfully")
	return brain, nil
}

// Close 关闭所有连接
func (b *Brain) Close() error {
	var errs []error

	if b.Nebula != nil {
		if err := b.Nebula.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close nebula failed: %w", err))
		}
	}

	if b.Milvus != nil {
		if err := b.Milvus.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close milvus failed: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing brain: %v", errs)
	}

	b.logger.Info("brain closed successfully")
	return nil
}

// Ping 检查所有连接
func (b *Brain) Ping(ctx context.Context) error {
	if err := b.Nebula.Ping(ctx); err != nil {
		return fmt.Errorf("nebula ping failed: %w", err)
	}

	if err := b.Milvus.Ping(ctx); err != nil {
		return fmt.Errorf("milvus ping failed: %w", err)
	}

	return nil
}
