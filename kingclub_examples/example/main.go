package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/UTC-Six/brain/kingclub_examples"
	"github.com/UTC-Six/brain/internal/milvus"
	"github.com/UTC-Six/brain/internal/nebula"
	"github.com/UTC-Six/brain/pkg/config"
	"go.uber.org/zap"
)

// 示例：简单的 embedding 函数（实际使用时应该调用真实的 embedding 模型）
func mockEmbeddingFunc(text string) ([]float32, error) {
	// 这里应该调用真实的 embedding 模型，如 text2vec-chinese
	// 示例：返回随机向量
	dim := 768
	vector := make([]float32, dim)
	for i := range vector {
		vector[i] = 0.1 // 实际应该是模型生成的向量
	}
	return vector, nil
}

// 示例：简单的大模型调用函数（实际使用时应该调用真实的大模型 API）
func mockLLMFunc(prompt string) (string, error) {
	// 这里应该调用真实的大模型 API，如 OpenAI、Claude 等
	// 示例：返回模拟结果
	return fmt.Sprintf("基于提供的数据，分析结果：%s", prompt[:min(50, len(prompt))]), nil
}

func main() {
	// 加载配置
	cfg, err := config.Load("../../config/config.yaml")
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
		Space:      cfg.Nebula.Space,
		Timeout:    cfg.Nebula.Timeout,
		MaxConn:    cfg.Nebula.MaxConn,
		MinConn:    cfg.Nebula.MinConn,
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

	// 创建 KingClub 服务
	service := kingclub.NewKingClubService(kingclub.ServiceConfig{
		NebulaClient:  nebulaClient,
		MilvusClient:  milvusClient,
		Logger:        logger,
		Space:         cfg.Nebula.Space,
		EmbeddingFunc: mockEmbeddingFunc, // 实际使用时替换为真实的 embedding 函数
		LLMFunc:       mockLLMFunc,      // 实际使用时替换为真实的大模型调用函数
	})
	defer service.Close()

	ctx := context.Background()

	// 1. 初始化 Schema
	fmt.Println("=== 初始化 Schema ===")
	if err := service.InitializeSchema(ctx); err != nil {
		logger.Error("Failed to initialize schema", zap.Error(err))
	} else {
		fmt.Println("Schema initialized successfully")
	}

	// 2. 写入示例数据
	fmt.Println("\n=== 写入示例数据 ===")
	writer := service.GetWriterService()

	// 写入学生
	student := &kingclub.Student{
		ID:            "student_001",
		Name:          "张三",
		Age:           15,
		Grade:         "高一",
		Gender:        "男",
		AvatarURL:     "https://example.com/avatar.jpg",
		Phone:         "13800138000",
		Email:         "zhangsan@example.com",
		School:        "XX中学",
		ClassName:     "高一(1)班",
		EnrollmentDate: time.Now().AddDate(-1, 0, 0),
		CreateAt:      time.Now(),
		UpdateAt:      time.Now(),
	}
	if err := writer.WriteStudent(ctx, student); err != nil {
		logger.Error("Failed to write student", zap.Error(err))
	} else {
		fmt.Printf("Student written: %s\n", student.Name)
	}

	// 写入会话
	conversation := &kingclub.Conversation{
		ID:              "conv_001",
		StudentID:       "student_001",
		LessonID:        "",
		ConversationType: "首页",
		Summary:         "学生询问学习计划",
		MessageCount:    10,
		StartTime:       time.Now().Add(-30 * time.Minute),
		EndTime:         time.Now(),
		CreateAt:        time.Now(),
	}

	messages := []kingclub.ConversationVector{
		{
			ID:              "msg_001",
			ConversationID:  "conv_001",
			StudentID:       "student_001",
			LessonID:        "",
			MessageType:     "user",
			Content:         "我想制定一个学习计划",
			Timestamp:       time.Now().Add(-30 * time.Minute).Unix(),
			ConversationType: "首页",
			Vector:          []float32{}, // 实际应该通过 embedding 生成
		},
	}

	// 生成向量
	for i := range messages {
		if vector, err := mockEmbeddingFunc(messages[i].Content); err == nil {
			messages[i].Vector = vector
		}
	}

	if err := writer.WriteConversation(ctx, conversation, messages); err != nil {
		logger.Error("Failed to write conversation", zap.Error(err))
	} else {
		fmt.Println("Conversation written")
	}

	// 3. 查询学生信息
	fmt.Println("\n=== 查询学生信息 ===")
	queryService := service.GetQueryService()

	basicInfo, err := queryService.GetStudentBasicInfo(ctx, "student_001")
	if err != nil {
		logger.Error("Failed to get basic info", zap.Error(err))
	} else {
		fmt.Printf("Basic Info: %+v\n", basicInfo)
	}

	learningAbility, err := queryService.GetLearningAbility(ctx, "student_001")
	if err != nil {
		logger.Error("Failed to get learning ability", zap.Error(err))
	} else {
		fmt.Printf("Learning Ability: Total Knowledge=%d, Average Mastery=%.2f\n",
			learningAbility.TotalKnowledge, learningAbility.AverageMastery)
	}

	// 4. RAG 召回
	fmt.Println("\n=== RAG 召回 ===")
	ragService := service.GetRAGService()

	conversations, err := ragService.RecallConversations(ctx, "学生的学习状态", "student_001", 5)
	if err != nil {
		logger.Error("Failed to recall conversations", zap.Error(err))
	} else {
		fmt.Printf("Recalled %d conversations\n", len(conversations))
		for i, conv := range conversations {
			if i >= 3 {
				break
			}
			content := conv.Content
			if len(content) > 50 {
				content = content[:50]
			}
			fmt.Printf("  Conversation %d: %s\n", i+1, content)
		}
	}

	// 5. 生成学生画像
	fmt.Println("\n=== 生成学生画像 ===")
	profileService := service.GetProfileService()

	profile, err := profileService.GenerateStudentProfile(ctx, "student_001")
	if err != nil {
		logger.Error("Failed to generate profile", zap.Error(err))
	} else {
		fmt.Printf("Profile generated:\n")
		if profile.BasicInfo != nil {
			fmt.Printf("  Name: %s, Age: %d, Grade: %s\n",
				profile.BasicInfo.Name, profile.BasicInfo.Age, profile.BasicInfo.Grade)
		}
		if profile.LearningAbility != nil {
			fmt.Printf("  Learning Ability: %d knowledge points, avg mastery: %.2f\n",
				profile.LearningAbility.TotalKnowledge, profile.LearningAbility.AverageMastery)
		}
		if profile.Personality != nil && profile.Personality.MBTIType != "" {
			fmt.Printf("  Personality: %s\n", profile.Personality.MBTIType)
		}
	}

	fmt.Println("\n=== 示例完成 ===")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

