package kingclub

import (
	"context"
	"fmt"
	"time"

	"github.com/UTC-Six/brain/internal/milvus"
	"github.com/UTC-Six/brain/internal/nebula"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"go.uber.org/zap"
)

// SchemaService Schema 初始化服务
type SchemaService struct {
	nebulaClient      *nebula.Client
	milvusClient      *milvus.Client
	logger            *zap.Logger
	space             string
	vertexService     *nebula.VertexService
	edgeService       *nebula.EdgeService
	collectionService *milvus.CollectionService
	indexService      *milvus.IndexService
}

// NewSchemaService 创建 Schema 服务
func NewSchemaService(
	nebulaClient *nebula.Client,
	milvusClient *milvus.Client,
	logger *zap.Logger,
	space string,
) *SchemaService {
	return &SchemaService{
		nebulaClient:      nebulaClient,
		milvusClient:      milvusClient,
		logger:            logger,
		space:             space,
		vertexService:     nebula.NewVertexService(nebulaClient, logger),
		edgeService:       nebula.NewEdgeService(nebulaClient, logger),
		collectionService: milvus.NewCollectionService(milvusClient, logger),
		indexService:      milvus.NewIndexService(milvusClient, logger),
	}
}

// InitializeNebulaSchema 初始化 Nebula Graph Schema
func (s *SchemaService) InitializeNebulaSchema(ctx context.Context) error {
	if err := s.ensureSpace(ctx); err != nil {
		return fmt.Errorf("ensure space failed: %w", err)
	}

	// 创建 Tag（顶点类型）
	tags := []string{
		`CREATE TAG IF NOT EXISTS student(
			id string, name string, age int, grade string, gender string,
			avatar_url string, phone string, email string, school string,
			class_name string, enrollment_date timestamp, create_at timestamp, update_at timestamp
		)`,
		`CREATE TAG IF NOT EXISTS course(
			id string, name string, subject string, level string,
			description string, duration int, create_at timestamp
		)`,
		`CREATE TAG IF NOT EXISTS lesson(
			id string, course_id string, title string, lesson_type string,
			start_time timestamp, end_time timestamp, create_at timestamp
		)`,
		`CREATE TAG IF NOT EXISTS knowledge_point(
			id string, name string, subject string, difficulty int,
			description string, create_at timestamp
		)`,
		`CREATE TAG IF NOT EXISTS article(
			id string, student_id string, lesson_id string, title string,
			content string, word_count int, score double, feedback string, create_at timestamp
		)`,
		`CREATE TAG IF NOT EXISTS question(
			id string, lesson_id string, question_type string, content string,
			correct_answer string, difficulty int, create_at timestamp
		)`,
		`CREATE TAG IF NOT EXISTS conversation(
			id string, student_id string, lesson_id string, conversation_type string,
			summary string, message_count int, start_time timestamp, end_time timestamp, create_at timestamp
		)`,
		`CREATE TAG IF NOT EXISTS family_member(
			id string, name string, relationship string,
			education_level string, occupation string, create_at timestamp
		)`,
		`CREATE TAG IF NOT EXISTS teacher(
			id string, name string, subject string, experience_years int, create_at timestamp
		)`,
		`CREATE TAG IF NOT EXISTS hobby(
			id string, name string, category string, create_at timestamp
		)`,
		`CREATE TAG IF NOT EXISTS mbti_test(
			id string, student_id string, mbti_type string,
			e_score int, i_score int, s_score int, n_score int,
			t_score int, f_score int, j_score int, p_score int,
			test_date timestamp, create_at timestamp
		)`,
	}

	for _, tag := range tags {
		if _, err := s.nebulaClient.Execute(ctx, tag); err != nil {
			return fmt.Errorf("create tag failed: %w", err)
		}
		s.logger.Info("tag created", zap.String("tag", tag))
	}

	// 创建 Edge Type（边类型）
	edges := []string{
		`CREATE EDGE IF NOT EXISTS attends(
			student_id string, course_id string, enroll_time timestamp,
			progress double, status string
		)`,
		`CREATE EDGE IF NOT EXISTS studies(
			student_id string, knowledge_id string, mastery_level int,
			study_count int, last_study_time timestamp, total_study_duration int
		)`,
		`CREATE EDGE IF NOT EXISTS contains(
			lesson_id string, knowledge_id string, order_index int
		)`,
		`CREATE EDGE IF NOT EXISTS prerequisite(
			knowledge_id1 string, knowledge_id2 string
		)`,
		`CREATE EDGE IF NOT EXISTS writes(
			student_id string, article_id string, lesson_id string,
			write_time timestamp, word_count int, score double
		)`,
		`CREATE EDGE IF NOT EXISTS answers(
			student_id string, question_id string, lesson_id string,
			answer_content string, is_correct bool, answer_time timestamp, time_spent int
		)`,
		`CREATE EDGE IF NOT EXISTS chats(
			student_id string, conversation_id string, lesson_id string,
			message_count int, avg_response_time double, engagement_score double
		)`,
		`CREATE EDGE IF NOT EXISTS belongs_to(
			student_id string, family_member_id string, relationship string, closeness int
		)`,
		`CREATE EDGE IF NOT EXISTS friends_with(
			student_id1 string, student_id2 string, friendship_level int, since timestamp
		)`,
		`CREATE EDGE IF NOT EXISTS likes(
			student_id string, hobby_id string, interest_level int, since timestamp
		)`,
		`CREATE EDGE IF NOT EXISTS has_mbti(
			student_id string, mbti_test_id string, mbti_type string, test_date timestamp
		)`,
		`CREATE EDGE IF NOT EXISTS previews(
			student_id string, lesson_id string, preview_time timestamp,
			preview_duration int, completion_rate double
		)`,
		`CREATE EDGE IF NOT EXISTS reviews(
			student_id string, lesson_id string, review_time timestamp,
			review_duration int, review_count int
		)`,
		`CREATE EDGE IF NOT EXISTS completes_homework(
			student_id string, lesson_id string, homework_id string,
			submit_time timestamp, completion_time int, score double, is_on_time bool
		)`,
		`CREATE EDGE IF NOT EXISTS collects_mistake(
			student_id string, question_id string, collect_time timestamp,
			review_count int, is_mastered bool
		)`,
		`CREATE EDGE IF NOT EXISTS listens(
			student_id string, lesson_id string, attention_score double,
			interaction_count int, focus_duration int
		)`,
		`CREATE EDGE IF NOT EXISTS interacts(
			student_id string, lesson_id string, interaction_type string,
			interaction_time timestamp, quality_score double
		)`,
	}

	for _, edge := range edges {
		if _, err := s.nebulaClient.Execute(ctx, edge); err != nil {
			return fmt.Errorf("create edge failed: %w", err)
		}
		s.logger.Info("edge created", zap.String("edge", edge))
	}

	s.logger.Info("nebula schema initialized successfully")
	return nil
}

// InitializeMilvusSchema 初始化 Milvus 集合 Schema
func (s *SchemaService) InitializeMilvusSchema(ctx context.Context) error {
	// 1. 会话向量集合
	if err := s.createConversationVectorsCollection(ctx); err != nil {
		return fmt.Errorf("create conversation vectors collection failed: %w", err)
	}

	// 2. 文章向量集合
	if err := s.createArticleVectorsCollection(ctx); err != nil {
		return fmt.Errorf("create article vectors collection failed: %w", err)
	}

	// 3. 问答向量集合
	if err := s.createQuestionAnswerVectorsCollection(ctx); err != nil {
		return fmt.Errorf("create question answer vectors collection failed: %w", err)
	}

	// 4. 学习行为向量集合
	if err := s.createLearningBehaviorVectorsCollection(ctx); err != nil {
		return fmt.Errorf("create learning behavior vectors collection failed: %w", err)
	}

	s.logger.Info("milvus schema initialized successfully")
	return nil
}

// createConversationVectorsCollection 创建会话向量集合
func (s *SchemaService) createConversationVectorsCollection(ctx context.Context) error {
	collectionName := "conversation_vectors"
	dim := 768 // 中文 embedding 模型维度

	schema := &entity.Schema{
		CollectionName: collectionName,
		Description:    "学生与大模型会话向量",
		Fields: []*entity.Field{
			{
				Name:       "id",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "100"},
				PrimaryKey: true,
			},
			{
				Name:     "vector",
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					"dim": fmt.Sprintf("%d", dim),
				},
			},
			{
				Name:       "conversation_id",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "100"},
			},
			{
				Name:       "student_id",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "100"},
			},
			{
				Name:       "lesson_id",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "100"},
			},
			{
				Name:       "message_type",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "20"},
			},
			{
				Name:       "content",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "2000"},
			},
			{
				Name:     "timestamp",
				DataType: entity.FieldTypeInt64,
			},
			{
				Name:       "conversation_type",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "20"},
			},
		},
		EnableDynamicField: true,
	}

	if err := s.collectionService.CreateCollection(ctx, schema); err != nil {
		return err
	}

	// 创建索引
	index := milvus.NewHNSWIndex(entity.L2, 16, 200)
	if err := s.indexService.CreateIndex(ctx, collectionName, "vector", index); err != nil {
		return err
	}

	// 加载集合
	if err := s.collectionService.LoadCollection(ctx, collectionName); err != nil {
		return err
	}

	return nil
}

// createArticleVectorsCollection 创建文章向量集合
func (s *SchemaService) createArticleVectorsCollection(ctx context.Context) error {
	collectionName := "article_vectors"
	dim := 768

	schema := &entity.Schema{
		CollectionName: collectionName,
		Description:    "学生写作文章向量",
		Fields: []*entity.Field{
			{
				Name:       "id",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "100"},
				PrimaryKey: true,
			},
			{
				Name:     "vector",
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					"dim": fmt.Sprintf("%d", dim),
				},
			},
			{
				Name:       "article_id",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "100"},
			},
			{
				Name:       "student_id",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "100"},
			},
			{
				Name:       "lesson_id",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "100"},
			},
			{
				Name:       "title",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "200"},
			},
			{
				Name:       "content",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "10000"},
			},
			{
				Name:     "word_count",
				DataType: entity.FieldTypeInt32,
			},
			{
				Name:     "score",
				DataType: entity.FieldTypeDouble,
			},
			{
				Name:     "timestamp",
				DataType: entity.FieldTypeInt64,
			},
		},
		EnableDynamicField: true,
	}

	if err := s.collectionService.CreateCollection(ctx, schema); err != nil {
		return err
	}

	index := milvus.NewHNSWIndex(entity.L2, 16, 200)
	if err := s.indexService.CreateIndex(ctx, collectionName, "vector", index); err != nil {
		return err
	}

	if err := s.collectionService.LoadCollection(ctx, collectionName); err != nil {
		return err
	}

	return nil
}

// createQuestionAnswerVectorsCollection 创建问答向量集合
func (s *SchemaService) createQuestionAnswerVectorsCollection(ctx context.Context) error {
	collectionName := "question_answer_vectors"
	dim := 768

	schema := &entity.Schema{
		CollectionName: collectionName,
		Description:    "学生问答向量",
		Fields: []*entity.Field{
			{
				Name:       "id",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "100"},
				PrimaryKey: true,
			},
			{
				Name:     "vector",
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					"dim": fmt.Sprintf("%d", dim),
				},
			},
			{
				Name:       "question_id",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "100"},
			},
			{
				Name:       "student_id",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "100"},
			},
			{
				Name:       "lesson_id",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "100"},
			},
			{
				Name:       "question_content",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "1000"},
			},
			{
				Name:       "answer_content",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "2000"},
			},
			{
				Name:     "is_correct",
				DataType: entity.FieldTypeBool,
			},
			{
				Name:     "time_spent",
				DataType: entity.FieldTypeInt32,
			},
			{
				Name:     "timestamp",
				DataType: entity.FieldTypeInt64,
			},
		},
		EnableDynamicField: true,
	}

	if err := s.collectionService.CreateCollection(ctx, schema); err != nil {
		return err
	}

	index := milvus.NewHNSWIndex(entity.L2, 16, 200)
	if err := s.indexService.CreateIndex(ctx, collectionName, "vector", index); err != nil {
		return err
	}

	if err := s.collectionService.LoadCollection(ctx, collectionName); err != nil {
		return err
	}

	return nil
}

// createLearningBehaviorVectorsCollection 创建学习行为向量集合
func (s *SchemaService) createLearningBehaviorVectorsCollection(ctx context.Context) error {
	collectionName := "learning_behavior_vectors"
	dim := 128 // 行为特征向量维度较小

	schema := &entity.Schema{
		CollectionName: collectionName,
		Description:    "学生学习行为向量",
		Fields: []*entity.Field{
			{
				Name:       "id",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "100"},
				PrimaryKey: true,
			},
			{
				Name:     "vector",
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					"dim": fmt.Sprintf("%d", dim),
				},
			},
			{
				Name:       "student_id",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "100"},
			},
			{
				Name:       "behavior_type",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "50"},
			},
			{
				Name:       "lesson_id",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "100"},
			},
			{
				Name:     "duration",
				DataType: entity.FieldTypeInt32,
			},
			{
				Name:     "completion_rate",
				DataType: entity.FieldTypeDouble,
			},
			{
				Name:     "quality_score",
				DataType: entity.FieldTypeDouble,
			},
			{
				Name:     "timestamp",
				DataType: entity.FieldTypeInt64,
			},
		},
		EnableDynamicField: true,
	}

	if err := s.collectionService.CreateCollection(ctx, schema); err != nil {
		return err
	}

	index := milvus.NewHNSWIndex(entity.L2, 16, 200)
	if err := s.indexService.CreateIndex(ctx, collectionName, "vector", index); err != nil {
		return err
	}

	if err := s.collectionService.LoadCollection(ctx, collectionName); err != nil {
		return err
	}

	return nil
}

// ensureSpace 确保空间存在
func (s *SchemaService) ensureSpace(ctx context.Context) error {
	if s.space == "" {
		return nil
	}

	// 检查空间是否存在
	result, err := s.nebulaClient.Execute(ctx, "SHOW SPACES")
	if err != nil {
		return fmt.Errorf("show spaces failed: %w", err)
	}

	for i := 0; i < result.GetRowSize(); i++ {
		row, err := result.GetRowValuesByIndex(i)
		if err != nil {
			continue
		}
		name, err := row.GetValueByColName("Name")
		if err != nil {
			continue
		}
		if n, err := name.AsString(); err == nil && n == s.space {
			// 空间已存在，切换
			if _, err := s.nebulaClient.Execute(ctx, fmt.Sprintf("USE %s", s.space)); err != nil {
				return fmt.Errorf("use space failed: %w", err)
			}
			return nil
		}
	}

	// 创建空间
	createStmt := fmt.Sprintf(
		`CREATE SPACE IF NOT EXISTS %s (partition_num=1, replica_factor=1, vid_type=FIXED_STRING(128))`,
		s.space,
	)
	if _, err := s.nebulaClient.Execute(ctx, createStmt); err != nil {
		return fmt.Errorf("create space failed: %w", err)
	}

	// 等待空间创建完成
	time.Sleep(2 * time.Second)

	// 切换空间
	if _, err := s.nebulaClient.Execute(ctx, fmt.Sprintf("USE %s", s.space)); err != nil {
		return fmt.Errorf("use space failed: %w", err)
	}

	return nil
}
