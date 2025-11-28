package kingclub

import (
	"context"
	"fmt"

	"github.com/UTC-Six/brain/internal/milvus"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"go.uber.org/zap"
)

// RAGService RAG 召回服务
type RAGService struct {
	milvusClient  *milvus.Client
	logger        *zap.Logger
	embeddingFunc func(text string) ([]float32, error) // 向量生成函数
}

// NewRAGService 创建 RAG 服务
func NewRAGService(
	milvusClient *milvus.Client,
	logger *zap.Logger,
	embeddingFunc func(text string) ([]float32, error),
) *RAGService {
	return &RAGService{
		milvusClient:  milvusClient,
		logger:        logger,
		embeddingFunc: embeddingFunc,
	}
}

// RecallConversations 召回相关会话内容
func (r *RAGService) RecallConversations(ctx context.Context, query string, studentID string, topK int) ([]*ConversationVector, error) {
	if r.embeddingFunc == nil {
		return nil, fmt.Errorf("embedding function not set")
	}

	// 生成查询向量
	queryVector, err := r.embeddingFunc(query)
	if err != nil {
		return nil, fmt.Errorf("generate query vector failed: %w", err)
	}

	// 构建过滤表达式
	expr := fmt.Sprintf("student_id == \"%s\"", studentID)

	// 搜索
	req := &milvus.SearchRequest{
		CollectionName: "conversation_vectors",
		VectorField:    "vector",
		Vectors:        [][]float32{queryVector},
		TopK:           topK,
		MetricType:     entity.L2,
		OutputFields:   []string{"id", "conversation_id", "student_id", "lesson_id", "message_type", "content", "timestamp", "conversation_type"},
		Expr:           expr,
		SearchParams: map[string]interface{}{
			"ef": 100,
		},
	}

	vectorService := milvus.NewVectorService(r.milvusClient, r.logger)
	results, err := vectorService.Search(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("search conversations failed: %w", err)
	}

	// 转换为结果（简化实现，实际需要根据 Milvus SDK 的实际返回结构解析）
	conversations := make([]*ConversationVector, 0)
	for _, result := range results {
		if result.IDs == nil || result.IDs.Len() == 0 {
			continue
		}

		// 注意：这里需要根据 Milvus SDK 的实际返回结构来解析 Fields
		// Fields 是 []entity.Column，需要根据 OutputFields 的顺序来匹配
		// 简化实现：只返回 ID，实际使用时需要正确解析所有字段
		for i := 0; i < result.IDs.Len(); i++ {
			id, _ := result.IDs.GetAsString(i)
			conv := &ConversationVector{
				ID: id,
				// 其他字段需要根据实际返回的 Fields 来解析
				// 这里暂时使用默认值
			}
			conversations = append(conversations, conv)
		}
	}

	return conversations, nil
}

// RecallArticles 召回相关文章
func (r *RAGService) RecallArticles(ctx context.Context, query string, studentID string, topK int) ([]*ArticleVector, error) {
	if r.embeddingFunc == nil {
		return nil, fmt.Errorf("embedding function not set")
	}

	queryVector, err := r.embeddingFunc(query)
	if err != nil {
		return nil, fmt.Errorf("generate query vector failed: %w", err)
	}

	expr := fmt.Sprintf("student_id == \"%s\"", studentID)

	req := &milvus.SearchRequest{
		CollectionName: "article_vectors",
		VectorField:    "vector",
		Vectors:        [][]float32{queryVector},
		TopK:           topK,
		MetricType:     entity.L2,
		OutputFields:   []string{"id", "article_id", "student_id", "lesson_id", "title", "content", "word_count", "score", "timestamp"},
		Expr:           expr,
		SearchParams: map[string]interface{}{
			"ef": 100,
		},
	}

	vectorService := milvus.NewVectorService(r.milvusClient, r.logger)
	results, err := vectorService.Search(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("search articles failed: %w", err)
	}

	articles := make([]*ArticleVector, 0)
	for _, result := range results {
		if result.IDs == nil || result.IDs.Len() == 0 {
			continue
		}

		for i := 0; i < result.IDs.Len(); i++ {
			id, _ := result.IDs.GetAsString(i)
			article := &ArticleVector{
				ID: id,
				// 其他字段需要根据实际返回的 Fields 来解析
			}
			articles = append(articles, article)
		}
	}

	return articles, nil
}

// RecallQuestionAnswers 召回相关问答
func (r *RAGService) RecallQuestionAnswers(ctx context.Context, query string, studentID string, topK int) ([]*QuestionAnswerVector, error) {
	if r.embeddingFunc == nil {
		return nil, fmt.Errorf("embedding function not set")
	}

	queryVector, err := r.embeddingFunc(query)
	if err != nil {
		return nil, fmt.Errorf("generate query vector failed: %w", err)
	}

	expr := fmt.Sprintf("student_id == \"%s\"", studentID)

	req := &milvus.SearchRequest{
		CollectionName: "question_answer_vectors",
		VectorField:    "vector",
		Vectors:        [][]float32{queryVector},
		TopK:           topK,
		MetricType:     entity.L2,
		OutputFields:   []string{"id", "question_id", "student_id", "lesson_id", "question_content", "answer_content", "is_correct", "time_spent", "timestamp"},
		Expr:           expr,
		SearchParams: map[string]interface{}{
			"ef": 100,
		},
	}

	vectorService := milvus.NewVectorService(r.milvusClient, r.logger)
	results, err := vectorService.Search(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("search question answers failed: %w", err)
	}

	qas := make([]*QuestionAnswerVector, 0)
	for _, result := range results {
		if result.IDs == nil || result.IDs.Len() == 0 {
			continue
		}

		for i := 0; i < result.IDs.Len(); i++ {
			id, _ := result.IDs.GetAsString(i)
			qa := &QuestionAnswerVector{
				ID: id,
				// 其他字段需要根据实际返回的 Fields 来解析
			}
			qas = append(qas, qa)
		}
	}

	return qas, nil
}

// RecallLearningBehaviors 召回学习行为
func (r *RAGService) RecallLearningBehaviors(ctx context.Context, query string, studentID string, topK int) ([]*LearningBehaviorVector, error) {
	if r.embeddingFunc == nil {
		return nil, fmt.Errorf("embedding function not set")
	}

	queryVector, err := r.embeddingFunc(query)
	if err != nil {
		return nil, fmt.Errorf("generate query vector failed: %w", err)
	}

	expr := fmt.Sprintf("student_id == \"%s\"", studentID)

	req := &milvus.SearchRequest{
		CollectionName: "learning_behavior_vectors",
		VectorField:    "vector",
		Vectors:        [][]float32{queryVector},
		TopK:           topK,
		MetricType:     entity.L2,
		OutputFields:   []string{"id", "student_id", "behavior_type", "lesson_id", "duration", "completion_rate", "quality_score", "timestamp"},
		Expr:           expr,
		SearchParams: map[string]interface{}{
			"ef": 100,
		},
	}

	vectorService := milvus.NewVectorService(r.milvusClient, r.logger)
	results, err := vectorService.Search(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("search learning behaviors failed: %w", err)
	}

	behaviors := make([]*LearningBehaviorVector, 0)
	for _, result := range results {
		if result.IDs == nil || result.IDs.Len() == 0 {
			continue
		}

		for i := 0; i < result.IDs.Len(); i++ {
			id, _ := result.IDs.GetAsString(i)
			behavior := &LearningBehaviorVector{
				ID: id,
				// 其他字段需要根据实际返回的 Fields 来解析
			}
			behaviors = append(behaviors, behavior)
		}
	}

	return behaviors, nil
}
