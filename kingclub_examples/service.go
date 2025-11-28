package kingclub

import (
	"context"
	"fmt"

	"github.com/UTC-Six/brain/internal/milvus"
	"github.com/UTC-Six/brain/internal/nebula"
	"go.uber.org/zap"
)

// KingClubService KingClub 主服务，整合所有功能
type KingClubService struct {
	// 基础服务
	schemaService  *SchemaService
	writerService  *WriterService
	queryService   *QueryService
	ragService     *RAGService
	profileService *ProfileService

	// 客户端
	nebulaClient *nebula.Client
	milvusClient *milvus.Client
	logger       *zap.Logger
	space        string
}

// ServiceConfig 服务配置
type ServiceConfig struct {
	NebulaClient  *nebula.Client
	MilvusClient  *milvus.Client
	Logger        *zap.Logger
	Space         string
	EmbeddingFunc func(text string) ([]float32, error) // 向量生成函数
	LLMFunc       func(prompt string) (string, error)  // 大模型调用函数
}

// NewKingClubService 创建 KingClub 服务
func NewKingClubService(config ServiceConfig) *KingClubService {
	schemaService := NewSchemaService(
		config.NebulaClient,
		config.MilvusClient,
		config.Logger,
		config.Space,
	)

	writerService := NewWriterService(
		config.NebulaClient,
		config.MilvusClient,
		config.Logger,
		config.EmbeddingFunc,
	)

	queryService := NewQueryService(
		config.NebulaClient,
		config.MilvusClient,
		config.Logger,
	)

	ragService := NewRAGService(
		config.MilvusClient,
		config.Logger,
		config.EmbeddingFunc,
	)

	profileService := NewProfileService(
		queryService,
		ragService,
		config.Logger,
		config.LLMFunc,
	)

	return &KingClubService{
		schemaService:  schemaService,
		writerService:  writerService,
		queryService:   queryService,
		ragService:     ragService,
		profileService: profileService,
		nebulaClient:   config.NebulaClient,
		milvusClient:   config.MilvusClient,
		logger:         config.Logger,
		space:          config.Space,
	}
}

// InitializeSchema 初始化所有 Schema
func (s *KingClubService) InitializeSchema(ctx context.Context) error {
	// 初始化 Nebula Schema
	if err := s.schemaService.InitializeNebulaSchema(ctx); err != nil {
		return err
	}

	// 初始化 Milvus Schema
	if err := s.schemaService.InitializeMilvusSchema(ctx); err != nil {
		return err
	}

	return nil
}

// GetSchemaService 获取 Schema 服务
func (s *KingClubService) GetSchemaService() *SchemaService {
	return s.schemaService
}

// GetWriterService 获取写入服务
func (s *KingClubService) GetWriterService() *WriterService {
	return s.writerService
}

// GetQueryService 获取查询服务
func (s *KingClubService) GetQueryService() *QueryService {
	return s.queryService
}

// GetRAGService 获取 RAG 服务
func (s *KingClubService) GetRAGService() *RAGService {
	return s.ragService
}

// GetProfileService 获取画像服务
func (s *KingClubService) GetProfileService() *ProfileService {
	return s.profileService
}

// Close 关闭服务
func (s *KingClubService) Close() error {
	var errs []error

	if s.nebulaClient != nil {
		if err := s.nebulaClient.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if s.milvusClient != nil {
		if err := s.milvusClient.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing service: %v", errs)
	}

	return nil
}
