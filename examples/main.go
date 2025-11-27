package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/UTC-Six/brain/examples/education"
	"github.com/UTC-Six/brain/internal/milvus"
	"github.com/UTC-Six/brain/internal/nebula"
	"github.com/UTC-Six/brain/pkg/config"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 创建 Logger
	logger, err := config.NewLogger(cfg.Logger)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// 初始化 Nebula Graph 客户端
	nebulaConfig := nebula.Config{
		Addresses: cfg.Nebula.Addresses,
		Username:  cfg.Nebula.Username,
		Password:  cfg.Nebula.Password,
		Space:     cfg.Nebula.Space,
		Timeout:   cfg.Nebula.Timeout,
		MaxConn:   cfg.Nebula.MaxConn,
		MinConn:   cfg.Nebula.MinConn,
	}

	nebulaClient, err := nebula.NewClient(nebulaConfig, logger)
	if err != nil {
		log.Fatalf("Failed to create nebula client: %v", err)
	}
	defer nebulaClient.Close()

	// 初始化 Milvus 客户端
	milvusConfig := milvus.Config{
		Host:           cfg.Milvus.Host,
		Port:           cfg.Milvus.Port,
		Username:       cfg.Milvus.Username,
		Password:       cfg.Milvus.Password,
		Database:       cfg.Milvus.Database,
		ConnectTimeout: cfg.Milvus.ConnectTimeout,
		EnableTLS:      cfg.Milvus.EnableTLS,
		TLSCert:        cfg.Milvus.TLSCert,
	}

	milvusClient, err := milvus.NewClient(milvusConfig, logger)
	if err != nil {
		log.Fatalf("Failed to create milvus client: %v", err)
	}
	defer milvusClient.Close()

	// 创建教育服务
	eduService := education.NewEducationService(nebulaClient, milvusClient, logger, cfg.Nebula.Space)

	ctx := context.Background()

	// 示例 1: 初始化 Schema
	fmt.Println("=== 初始化 Schema ===")
	if err := eduService.InitializeSchema(ctx); err != nil {
		logger.Error("Failed to initialize schema", zap.Error(err))
		return
	}
	logger.Info("Schema initialized successfully")

	// 示例 2: 创建学生、课程、知识点
	fmt.Println("\n=== 创建学生、课程、知识点 ===")
	student := education.Student{
		ID:       "student_001",
		Name:     "张三",
		Age:      15,
		Grade:    "高一",
		CreateAt: time.Now(),
	}
	if err := eduService.CreateStudent(ctx, student); err != nil {
		logger.Error("Failed to create student", zap.Error(err))
		return
	}
	logger.Info(fmt.Sprintf("Student created: %s", student.Name))

	course := education.Course{
		ID:       "course_001",
		Name:     "高中数学",
		Subject:  "数学",
		Level:    "高中",
		CreateAt: time.Now(),
	}
	if err := eduService.CreateCourse(ctx, course); err != nil {
		logger.Error("Failed to create course", zap.Error(err))
		return
	}
	logger.Info(fmt.Sprintf("Course created: %s\n", course.Name))

	knowledgePoint := education.KnowledgePoint{
		ID:          "kp_001",
		Name:        "二次函数",
		Description: "二次函数的图像和性质",
		Difficulty:  3,
		CreateAt:    time.Now(),
	}
	if err := eduService.CreateKnowledgePoint(ctx, knowledgePoint); err != nil {
		logger.Error("Failed to create knowledge point", zap.Error(err))
		return
	}
	logger.Info(fmt.Sprintf("Knowledge point created: %s\n", knowledgePoint.Name))

	// 示例 3: 记录学习情况
	fmt.Println("\n=== 记录学习情况 ===")
	learningRecord := education.LearningRecord{
		StudentID:     "student_001",
		KnowledgeID:   "kp_001",
		Score:         85.5,
		MasteryLevel:  4,
		StudyDuration: 120,
		LastStudyTime: time.Now(),
	}
	if err := eduService.RecordLearning(ctx, learningRecord); err != nil {
		logger.Error("Failed to record learning", zap.Error(err))
		return
	}
	logger.Info("Learning record created")

	// 示例 4: 查询学生学习情况
	fmt.Println("\n=== 查询学生学习情况 ===")
	records, err := eduService.GetStudentKnowledgeMastery(ctx, "student_001")
	if err != nil {
		logger.Error("Failed to get student knowledge mastery", zap.Error(err))
		return
	} else {
		fmt.Printf("Found %d learning records\n", len(records))
		for _, record := range records {
			fmt.Printf("  Knowledge: %s, Score: %.2f, Mastery: %d\n",
				record.KnowledgeID, record.Score, record.MasteryLevel)
		}
	}

	// 示例 5: 分析学生学习能力
	fmt.Println("\n=== 分析学生学习能力 ===")
	ability, err := eduService.GetLearningAbility(ctx, "student_001")
	if err != nil {
		logger.Error("Failed to get learning ability", zap.Error(err))
		return
	} else {
		fmt.Printf("Student: %s\n", ability.StudentID)
		fmt.Printf("  Total Knowledge: %d\n", ability.TotalKnowledge)
		fmt.Printf("  Average Score: %.2f\n", ability.AverageScore)
		fmt.Printf("  Average Mastery: %.2f\n", ability.AverageMastery)
		fmt.Printf("  Total Study Time: %d minutes\n", ability.TotalStudyTime)
		fmt.Printf("  Mastered Count: %d\n", ability.MasteredCount)
		fmt.Printf("  Weak Knowledge IDs: %v\n", ability.WeakKnowledgeIDs)
	}

	// 示例 6: 初始化向量集合并插入学习内容
	fmt.Println("\n=== 初始化向量集合 ===")
	collectionName := "learning_content"
	dim := 128 // 向量维度（实际使用时应该与 embedding 模型维度一致）
	if err := eduService.InitializeVectorCollection(ctx, collectionName, dim); err != nil {
		logger.Error("Failed to initialize vector collection", zap.Error(err))
		return
	} else {
		fmt.Printf("Vector collection initialized: %s (dim=%d)\n", collectionName, dim)
	}

	// 示例 7: 插入学习内容向量（模拟）
	fmt.Println("\n=== 插入学习内容向量 ===")
	contents := []education.LearningContent{
		{
			ID:          "content_001",
			Content:     "什么是二次函数？",
			Vector:      make([]float32, dim), // 实际使用时应该通过 embedding 模型生成
			ContentType: "question",
			KnowledgeID: "kp_001",
			Metadata: map[string]interface{}{
				"difficulty": 2,
				"source":     "textbook",
			},
		},
		{
			ID:          "content_002",
			Content:     "二次函数的一般形式是 y=ax²+bx+c",
			Vector:      make([]float32, dim),
			ContentType: "answer",
			KnowledgeID: "kp_001",
			Metadata: map[string]interface{}{
				"difficulty": 1,
				"source":     "textbook",
			},
		},
	}

	// 注意：实际使用时，Vector 应该通过 embedding 模型生成，这里只是示例
	// 例如：vector, err := embeddingModel.Encode(content.Content)
	ids, err := eduService.InsertLearningContent(ctx, collectionName, contents)
	if err != nil {
		logger.Error("Failed to insert learning content", zap.Error(err))
		return
	} else {
		fmt.Printf("Inserted %d learning contents, IDs: %v\n", ids.Len(), columnValues(ids))
	}

	// 示例 8: 向量相似度搜索（内容召回）
	fmt.Println("\n=== 向量相似度搜索 ===")
	queryVector := make([]float32, dim) // 实际使用时应该通过 embedding 模型生成查询向量
	filters := map[string]interface{}{
		"knowledge_id": "kp_001",
		"content_type": "question",
	}

	results, err := eduService.SearchSimilarContent(ctx, collectionName, queryVector, 5, filters)
	if err != nil {
		logger.Error("Failed to search similar content", zap.Error(err))
		return
	} else {
		fmt.Printf("Found %d similar contents\n", len(results))
		for i, result := range results {
			var firstID interface{} = "N/A"
			if result.IDs != nil && result.IDs.Len() > 0 {
				if str, err := result.IDs.GetAsString(0); err == nil {
					firstID = str
				} else if v, err := result.IDs.Get(0); err == nil {
					firstID = v
				}
			}
			score := float32(0)
			if len(result.Scores) > 0 {
				score = result.Scores[0]
			}
			fmt.Printf("  Result %d: ID=%v, Score=%.4f\n", i+1, firstID, score)
		}
	}

	fmt.Println("\n=== 示例完成 ===")
}

func columnValues(col entity.Column) []interface{} {
	if col == nil {
		return nil
	}
	values := make([]interface{}, 0, col.Len())
	for i := 0; i < col.Len(); i++ {
		if str, err := col.GetAsString(i); err == nil {
			values = append(values, str)
			continue
		}
		if v, err := col.Get(i); err == nil {
			values = append(values, v)
		}
	}
	return values
}
