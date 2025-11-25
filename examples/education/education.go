package education

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/UTC-Six/brain/internal/milvus"
	"github.com/UTC-Six/brain/internal/nebula"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"go.uber.org/zap"
)

// EducationService AI 教育场景的业务封装示例
type EducationService struct {
	nebulaClient *nebula.Client
	milvusClient *milvus.Client
	logger       *zap.Logger

	// Nebula 服务
	vertexService *nebula.VertexService
	edgeService   *nebula.EdgeService

	// Milvus 服务
	collectionService *milvus.CollectionService
	vectorService     *milvus.VectorService
	indexService      *milvus.IndexService
}

// NewEducationService 创建教育服务
func NewEducationService(
	nebulaClient *nebula.Client,
	milvusClient *milvus.Client,
	logger *zap.Logger,
) *EducationService {
	return &EducationService{
		nebulaClient:      nebulaClient,
		milvusClient:      milvusClient,
		logger:            logger,
		vertexService:     nebula.NewVertexService(nebulaClient, logger),
		edgeService:       nebula.NewEdgeService(nebulaClient, logger),
		collectionService: milvus.NewCollectionService(milvusClient, logger),
		vectorService:     milvus.NewVectorService(milvusClient, logger),
		indexService:      milvus.NewIndexService(milvusClient, logger),
	}
}

// ==================== Nebula Graph 操作示例 ====================

// Student 学生实体
type Student struct {
	ID       string
	Name     string
	Age      int
	Grade    string
	CreateAt time.Time
}

// Course 课程实体
type Course struct {
	ID       string
	Name     string
	Subject  string
	Level    string
	CreateAt time.Time
}

// KnowledgePoint 知识点实体
type KnowledgePoint struct {
	ID          string
	Name        string
	Description string
	Difficulty  int // 1-5
	CreateAt    time.Time
}

// LearningRecord 学习记录（边）
type LearningRecord struct {
	StudentID     string
	KnowledgeID   string
	Score         float64
	MasteryLevel  int // 1-5
	StudyDuration int // 分钟
	LastStudyTime time.Time
}

// InitializeSchema 初始化图数据库 Schema
func (s *EducationService) InitializeSchema(ctx context.Context) error {
	// 创建 Tag（顶点类型）
	tags := []string{
		"CREATE TAG IF NOT EXISTS student(id string, name string, age int, grade string, create_at timestamp)",
		"CREATE TAG IF NOT EXISTS course(id string, name string, subject string, level string, create_at timestamp)",
		"CREATE TAG IF NOT EXISTS knowledge_point(id string, name string, description string, difficulty int, create_at timestamp)",
	}

	for _, tag := range tags {
		if _, err := s.nebulaClient.Execute(ctx, tag); err != nil {
			return fmt.Errorf("create tag failed: %w", err)
		}
	}

	// 创建 Edge Type（边类型）
	edges := []string{
		"CREATE EDGE IF NOT EXISTS studies(student_id string, course_id string, enroll_time timestamp, progress double)",
		"CREATE EDGE IF NOT EXISTS learns(student_id string, knowledge_id string, score double, mastery_level int, study_duration int, last_study_time timestamp)",
		"CREATE EDGE IF NOT EXISTS contains(course_id string, knowledge_id string, order int)",
		"CREATE EDGE IF NOT EXISTS prerequisite(knowledge_id1 string, knowledge_id2 string)",
	}

	for _, edge := range edges {
		if _, err := s.nebulaClient.Execute(ctx, edge); err != nil {
			return fmt.Errorf("create edge failed: %w", err)
		}
	}

	s.logger.Info("education schema initialized")
	return nil
}

// CreateStudent 创建学生
func (s *EducationService) CreateStudent(ctx context.Context, student Student) error {
	vertex := nebula.Vertex{
		VID: student.ID,
		Tag: "student",
		Props: map[string]interface{}{
			"id":        student.ID,
			"name":      student.Name,
			"age":       student.Age,
			"grade":     student.Grade,
			"create_at": student.CreateAt.Unix(),
		},
	}
	return s.vertexService.Create(ctx, vertex)
}

// CreateCourse 创建课程
func (s *EducationService) CreateCourse(ctx context.Context, course Course) error {
	vertex := nebula.Vertex{
		VID: course.ID,
		Tag: "course",
		Props: map[string]interface{}{
			"id":        course.ID,
			"name":      course.Name,
			"subject":   course.Subject,
			"level":     course.Level,
			"create_at": course.CreateAt.Unix(),
		},
	}
	return s.vertexService.Create(ctx, vertex)
}

// CreateKnowledgePoint 创建知识点
func (s *EducationService) CreateKnowledgePoint(ctx context.Context, kp KnowledgePoint) error {
	vertex := nebula.Vertex{
		VID: kp.ID,
		Tag: "knowledge_point",
		Props: map[string]interface{}{
			"id":          kp.ID,
			"name":        kp.Name,
			"description": kp.Description,
			"difficulty":  kp.Difficulty,
			"create_at":   kp.CreateAt.Unix(),
		},
	}
	return s.vertexService.Create(ctx, vertex)
}

// RecordLearning 记录学习情况
func (s *EducationService) RecordLearning(ctx context.Context, record LearningRecord) error {
	edge := nebula.Edge{
		SrcID: record.StudentID,
		DstID: record.KnowledgeID,
		Type:  "learns",
		Props: map[string]interface{}{
			"student_id":      record.StudentID,
			"knowledge_id":    record.KnowledgeID,
			"score":           record.Score,
			"mastery_level":   record.MasteryLevel,
			"study_duration":  record.StudyDuration,
			"last_study_time": record.LastStudyTime.Unix(),
		},
	}
	return s.edgeService.Create(ctx, edge)
}

// GetStudentKnowledgeMastery 获取学生对知识点的掌握情况
func (s *EducationService) GetStudentKnowledgeMastery(ctx context.Context, studentID string) ([]LearningRecord, error) {
	// 查询学生学习的知识点及其掌握情况
	query := fmt.Sprintf(`
		MATCH (s:student)-[l:learns]->(k:knowledge_point)
		WHERE id(s) == "%s"
		RETURN id(k) as knowledge_id, 
		       l.score as score,
		       l.mastery_level as mastery_level,
		       l.study_duration as study_duration,
		       l.last_study_time as last_study_time
	`, studentID)

	resultSet, err := s.nebulaClient.Execute(ctx, query)
	if err != nil {
		return nil, err
	}

	var records []LearningRecord
	for i := 0; i < resultSet.GetRowSize(); i++ {
		row, err := resultSet.GetRowValuesByIndex(i)
		if err != nil {
			continue
		}

		record := LearningRecord{
			StudentID: studentID,
		}

		if knowledgeID, err := row.GetValueByColName("knowledge_id"); err == nil {
			if str, err := knowledgeID.AsString(); err == nil {
				record.KnowledgeID = str
			}
		}
		if score, err := row.GetValueByColName("score"); err == nil {
			if f, err := score.AsFloat(); err == nil {
				record.Score = f
			}
		}
		if masteryLevel, err := row.GetValueByColName("mastery_level"); err == nil {
			if i, err := masteryLevel.AsInt(); err == nil {
				record.MasteryLevel = int(i)
			}
		}
		if studyDuration, err := row.GetValueByColName("study_duration"); err == nil {
			if i, err := studyDuration.AsInt(); err == nil {
				record.StudyDuration = int(i)
			}
		}
		if lastStudyTime, err := row.GetValueByColName("last_study_time"); err == nil {
			if ts, err := lastStudyTime.AsInt(); err == nil {
				record.LastStudyTime = time.Unix(ts, 0)
			}
		}

		records = append(records, record)
	}

	return records, nil
}

// GetRelatedKnowledgePoints 获取相关知识点（基于先修关系）
func (s *EducationService) GetRelatedKnowledgePoints(ctx context.Context, knowledgeID string) ([]string, error) {
	query := fmt.Sprintf(`
		MATCH (k1:knowledge_point)-[p:prerequisite]->(k2:knowledge_point)
		WHERE id(k1) == "%s" OR id(k2) == "%s"
		RETURN DISTINCT id(k1) as k1_id, id(k2) as k2_id
	`, knowledgeID, knowledgeID)

	resultSet, err := s.nebulaClient.Execute(ctx, query)
	if err != nil {
		return nil, err
	}

	knowledgeIDs := make(map[string]bool)
	for i := 0; i < resultSet.GetRowSize(); i++ {
		row, err := resultSet.GetRowValuesByIndex(i)
		if err != nil {
			continue
		}

		if k1ID, err := row.GetValueByColName("k1_id"); err == nil {
			if str, err := k1ID.AsString(); err == nil {
				knowledgeIDs[str] = true
			}
		}
		if k2ID, err := row.GetValueByColName("k2_id"); err == nil {
			if str, err := k2ID.AsString(); err == nil {
				knowledgeIDs[str] = true
			}
		}
	}

	var result []string
	for id := range knowledgeIDs {
		if id != knowledgeID {
			result = append(result, id)
		}
	}

	return result, nil
}

// ==================== Milvus 操作示例 ====================

// LearningContent 学习内容向量
type LearningContent struct {
	ID          string
	Content     string
	Vector      []float32 // 向量表示（通常由 embedding 模型生成）
	ContentType string    // 类型：question, answer, material, etc.
	KnowledgeID string    // 关联的知识点ID
	Metadata    map[string]interface{}
}

// InitializeVectorCollection 初始化向量集合
func (s *EducationService) InitializeVectorCollection(ctx context.Context, collectionName string, dim int) error {
	// 检查集合是否存在
	exists, err := s.collectionService.HasCollection(ctx, collectionName)
	if err != nil {
		return err
	}

	if exists {
		s.logger.Info("collection already exists", zap.String("collection", collectionName))
		return nil
	}

	// 创建 Schema
	schema := &entity.Schema{
		CollectionName: collectionName,
		Description:    "Learning content vectors for education",
		Fields: []*entity.Field{
			{
				Name:       "id",
				DataType:   entity.FieldTypeVarChar,
				PrimaryKey: true,
				TypeParams: map[string]string{
					"max_length": "100",
				},
			},
			{
				Name:     "vector",
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					"dim": fmt.Sprintf("%d", dim),
				},
			},
			{
				Name:     "content_type",
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					"max_length": "50",
				},
			},
			{
				Name:     "knowledge_id",
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					"max_length": "100",
				},
			},
			{
				Name:     "score",
				DataType: entity.FieldTypeDouble,
			},
		},
		EnableDynamicField: true, // 支持动态字段存储 metadata
	}

	// 创建集合
	if err := s.collectionService.CreateCollection(ctx, schema); err != nil {
		return err
	}

	// 创建索引（HNSW 索引，适合向量搜索）
	index := milvus.NewHNSWIndex(entity.L2, 16, 200)
	if err := s.indexService.CreateIndex(ctx, collectionName, "vector", index); err != nil {
		return err
	}

	// 加载集合到内存
	if err := s.collectionService.LoadCollection(ctx, collectionName); err != nil {
		return err
	}

	s.logger.Info("vector collection initialized",
		zap.String("collection", collectionName),
		zap.Int("dim", dim),
	)
	return nil
}

// InsertLearningContent 插入学习内容向量
func (s *EducationService) InsertLearningContent(ctx context.Context, collectionName string, contents []LearningContent) (entity.Column, error) {
	data := make([]map[string]interface{}, 0, len(contents))
	for _, content := range contents {
		item := map[string]interface{}{
			"id":           content.ID,
			"vector":       content.Vector,
			"content_type": content.ContentType,
			"knowledge_id": content.KnowledgeID,
			"score":        0.0, // 默认分数
		}

		// 添加 metadata 到动态字段
		for k, v := range content.Metadata {
			item[k] = v
		}

		data = append(data, item)
	}

	return s.vectorService.Insert(ctx, collectionName, data)
}

// SearchSimilarContent 搜索相似的学习内容（用于内容召回）
func (s *EducationService) SearchSimilarContent(
	ctx context.Context,
	collectionName string,
	queryVector []float32,
	topK int,
	filters map[string]interface{},
) ([]*milvus.SearchResult, error) {
	// 构建过滤表达式
	expr := "1 == 1" // 默认不过滤
	if len(filters) > 0 {
		exprParts := make([]string, 0)
		for k, v := range filters {
			switch val := v.(type) {
			case string:
				exprParts = append(exprParts, fmt.Sprintf("%s == \"%s\"", k, val))
			case int, int64:
				exprParts = append(exprParts, fmt.Sprintf("%s == %d", k, val))
			case float64:
				exprParts = append(exprParts, fmt.Sprintf("%s == %f", k, val))
			}
		}
		if len(exprParts) > 0 {
			expr = strings.Join(exprParts, " && ")
		}
	}

	req := &milvus.SearchRequest{
		CollectionName: collectionName,
		VectorField:    "vector",
		Vectors:        [][]float32{queryVector},
		TopK:           topK,
		MetricType:     entity.L2,
		OutputFields:   []string{"id", "content_type", "knowledge_id", "score"},
		Expr:           expr,
		SearchParams: map[string]interface{}{
			"ef": 100, // HNSW 搜索参数
		},
	}

	return s.vectorService.Search(ctx, req)
}

// GetLearningAbility 分析学生学习能力（结合图数据和向量数据）
func (s *EducationService) GetLearningAbility(ctx context.Context, studentID string) (*LearningAbility, error) {
	// 1. 从 Nebula Graph 获取学习记录
	records, err := s.GetStudentKnowledgeMastery(ctx, studentID)
	if err != nil {
		return nil, err
	}

	// 2. 计算学习能力指标
	ability := &LearningAbility{
		StudentID:        studentID,
		TotalKnowledge:   len(records),
		AverageScore:     0,
		AverageMastery:   0,
		TotalStudyTime:   0,
		MasteredCount:    0,
		WeakKnowledgeIDs: make([]string, 0),
	}

	if len(records) == 0 {
		return ability, nil
	}

	var totalScore float64
	var totalMastery int
	for _, record := range records {
		totalScore += record.Score
		totalMastery += record.MasteryLevel
		ability.TotalStudyTime += record.StudyDuration

		if record.MasteryLevel >= 4 {
			ability.MasteredCount++
		} else if record.MasteryLevel <= 2 {
			ability.WeakKnowledgeIDs = append(ability.WeakKnowledgeIDs, record.KnowledgeID)
		}
	}

	ability.AverageScore = totalScore / float64(len(records))
	ability.AverageMastery = float64(totalMastery) / float64(len(records))

	return ability, nil
}

// LearningAbility 学习能力分析结果
type LearningAbility struct {
	StudentID        string
	TotalKnowledge   int
	AverageScore     float64
	AverageMastery   float64
	TotalStudyTime   int // 分钟
	MasteredCount    int
	WeakKnowledgeIDs []string
}
