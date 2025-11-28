package kingclub

import (
	"context"
	"fmt"

	"github.com/UTC-Six/brain/internal/milvus"
	"github.com/UTC-Six/brain/internal/nebula"
	"go.uber.org/zap"
)

// WriterService 数据写入服务
type WriterService struct {
	nebulaClient  *nebula.Client
	milvusClient  *milvus.Client
	logger        *zap.Logger
	vertexService *nebula.VertexService
	edgeService   *nebula.EdgeService
	vectorService *milvus.VectorService
	embeddingFunc func(text string) ([]float32, error) // 向量生成函数（需要外部提供）
}

// NewWriterService 创建写入服务
func NewWriterService(
	nebulaClient *nebula.Client,
	milvusClient *milvus.Client,
	logger *zap.Logger,
	embeddingFunc func(text string) ([]float32, error),
) *WriterService {
	return &WriterService{
		nebulaClient:  nebulaClient,
		milvusClient:  milvusClient,
		logger:        logger,
		vertexService: nebula.NewVertexService(nebulaClient, logger),
		edgeService:   nebula.NewEdgeService(nebulaClient, logger),
		vectorService: milvus.NewVectorService(milvusClient, logger),
		embeddingFunc: embeddingFunc,
	}
}

// WriteStudent 写入学生信息
func (w *WriterService) WriteStudent(ctx context.Context, student *Student) error {
	vertex := nebula.Vertex{
		VID: student.ID,
		Tag: "student",
		Props: map[string]interface{}{
			"id":              student.ID,
			"name":            student.Name,
			"age":             student.Age,
			"grade":           student.Grade,
			"gender":          student.Gender,
			"avatar_url":      student.AvatarURL,
			"phone":           student.Phone,
			"email":           student.Email,
			"school":          student.School,
			"class_name":      student.ClassName,
			"enrollment_date": student.EnrollmentDate.Unix(),
			"create_at":       student.CreateAt.Unix(),
			"update_at":       student.UpdateAt.Unix(),
		},
	}
	return w.vertexService.Create(ctx, vertex)
}

// WriteConversation 写入会话数据（Nebula + Milvus）
func (w *WriterService) WriteConversation(ctx context.Context, conv *Conversation, messages []ConversationVector) error {
	// 1. 写入 Nebula 顶点
	vertex := nebula.Vertex{
		VID: conv.ID,
		Tag: "conversation",
		Props: map[string]interface{}{
			"id":                conv.ID,
			"student_id":        conv.StudentID,
			"lesson_id":         conv.LessonID,
			"conversation_type": conv.ConversationType,
			"summary":           conv.Summary,
			"message_count":     conv.MessageCount,
			"start_time":        conv.StartTime.Unix(),
			"end_time":          conv.EndTime.Unix(),
			"create_at":         conv.CreateAt.Unix(),
		},
	}
	if err := w.vertexService.Create(ctx, vertex); err != nil {
		return fmt.Errorf("create conversation vertex failed: %w", err)
	}

	// 2. 创建边
	if conv.StudentID != "" {
		edge := nebula.Edge{
			SrcID: conv.StudentID,
			DstID: conv.ID,
			Type:  "chats",
			Props: map[string]interface{}{
				"student_id":        conv.StudentID,
				"conversation_id":   conv.ID,
				"lesson_id":         conv.LessonID,
				"message_count":     conv.MessageCount,
				"avg_response_time": 0.0, // 需要计算
				"engagement_score":  0.0, // 需要计算
			},
		}
		if err := w.edgeService.Create(ctx, edge); err != nil {
			w.logger.Warn("create chats edge failed", zap.Error(err))
		}
	}

	// 3. 写入 Milvus 向量
	if len(messages) > 0 {
		data := make([]map[string]interface{}, 0, len(messages))
		for _, msg := range messages {
			item := map[string]interface{}{
				"id":                msg.ID,
				"vector":            msg.Vector,
				"conversation_id":   msg.ConversationID,
				"student_id":        msg.StudentID,
				"lesson_id":         msg.LessonID,
				"message_type":      msg.MessageType,
				"content":           msg.Content,
				"timestamp":         msg.Timestamp,
				"conversation_type": msg.ConversationType,
			}
			// 添加 metadata
			for k, v := range msg.Metadata {
				item[k] = normalizeMetadataValue(v)
			}
			data = append(data, item)
		}

		if _, err := w.vectorService.Insert(ctx, "conversation_vectors", data); err != nil {
			return fmt.Errorf("insert conversation vectors failed: %w", err)
		}
	}

	return nil
}

// WriteArticle 写入文章数据（Nebula + Milvus）
func (w *WriterService) WriteArticle(ctx context.Context, article *Article) error {
	// 1. 写入 Nebula 顶点
	vertex := nebula.Vertex{
		VID: article.ID,
		Tag: "article",
		Props: map[string]interface{}{
			"id":         article.ID,
			"student_id": article.StudentID,
			"lesson_id":  article.LessonID,
			"title":      article.Title,
			"content":    article.Content,
			"word_count": article.WordCount,
			"score":      article.Score,
			"feedback":   article.Feedback,
			"create_at":  article.CreateAt.Unix(),
		},
	}
	if err := w.vertexService.Create(ctx, vertex); err != nil {
		return fmt.Errorf("create article vertex failed: %w", err)
	}

	// 2. 创建边
	edge := nebula.Edge{
		SrcID: article.StudentID,
		DstID: article.ID,
		Type:  "writes",
		Props: map[string]interface{}{
			"student_id": article.StudentID,
			"article_id": article.ID,
			"lesson_id":  article.LessonID,
			"write_time": article.CreateAt.Unix(),
			"word_count": article.WordCount,
			"score":      article.Score,
		},
	}
	if err := w.edgeService.Create(ctx, edge); err != nil {
		w.logger.Warn("create writes edge failed", zap.Error(err))
	}

	// 3. 生成向量并写入 Milvus
	if w.embeddingFunc != nil {
		// 使用标题+内容生成向量
		text := article.Title + "\n" + article.Content
		vector, err := w.embeddingFunc(text)
		if err != nil {
			w.logger.Warn("generate article vector failed", zap.Error(err))
		} else {
			articleVec := ArticleVector{
				ID:        fmt.Sprintf("article_vec_%s", article.ID),
				Vector:    vector,
				ArticleID: article.ID,
				StudentID: article.StudentID,
				LessonID:  article.LessonID,
				Title:     article.Title,
				Content:   article.Content,
				WordCount: article.WordCount,
				Score:     article.Score,
				Timestamp: article.CreateAt.Unix(),
			}

			data := []map[string]interface{}{
				{
					"id":         articleVec.ID,
					"vector":     articleVec.Vector,
					"article_id": articleVec.ArticleID,
					"student_id": articleVec.StudentID,
					"lesson_id":  articleVec.LessonID,
					"title":      articleVec.Title,
					"content":    articleVec.Content,
					"word_count": articleVec.WordCount,
					"score":      articleVec.Score,
					"timestamp":  articleVec.Timestamp,
				},
			}

			if _, err := w.vectorService.Insert(ctx, "article_vectors", data); err != nil {
				w.logger.Warn("insert article vector failed", zap.Error(err))
			}
		}
	}

	return nil
}

// WriteQuestionAnswer 写入问答数据（Nebula + Milvus）
func (w *WriterService) WriteQuestionAnswer(ctx context.Context, question *Question, answer *AnswersEdge) error {
	// 1. 写入问题顶点（如果不存在）
	questionVertex := nebula.Vertex{
		VID: question.ID,
		Tag: "question",
		Props: map[string]interface{}{
			"id":             question.ID,
			"lesson_id":      question.LessonID,
			"question_type":  question.QuestionType,
			"content":        question.Content,
			"correct_answer": question.CorrectAnswer,
			"difficulty":     question.Difficulty,
			"create_at":      question.CreateAt.Unix(),
		},
	}
	if err := w.vertexService.Create(ctx, questionVertex); err != nil {
		w.logger.Warn("create question vertex failed", zap.Error(err))
	}

	// 2. 创建答题边
	edge := nebula.Edge{
		SrcID: answer.StudentID,
		DstID: answer.QuestionID,
		Type:  "answers",
		Props: map[string]interface{}{
			"student_id":     answer.StudentID,
			"question_id":    answer.QuestionID,
			"lesson_id":      answer.LessonID,
			"answer_content": answer.AnswerContent,
			"is_correct":     answer.IsCorrect,
			"answer_time":    answer.AnswerTime.Unix(),
			"time_spent":     answer.TimeSpent,
		},
	}
	if err := w.edgeService.Create(ctx, edge); err != nil {
		w.logger.Warn("create answers edge failed", zap.Error(err))
	}

	// 3. 生成向量并写入 Milvus
	if w.embeddingFunc != nil {
		text := question.Content + "\n" + answer.AnswerContent
		vector, err := w.embeddingFunc(text)
		if err != nil {
			w.logger.Warn("generate qa vector failed", zap.Error(err))
		} else {
			qaVec := QuestionAnswerVector{
				ID:              fmt.Sprintf("qa_vec_%s_%s", answer.QuestionID, answer.StudentID),
				Vector:          vector,
				QuestionID:      answer.QuestionID,
				StudentID:       answer.StudentID,
				LessonID:        answer.LessonID,
				QuestionContent: question.Content,
				AnswerContent:   answer.AnswerContent,
				IsCorrect:       answer.IsCorrect,
				TimeSpent:       answer.TimeSpent,
				Timestamp:       answer.AnswerTime.Unix(),
			}

			data := []map[string]interface{}{
				{
					"id":               qaVec.ID,
					"vector":           qaVec.Vector,
					"question_id":      qaVec.QuestionID,
					"student_id":       qaVec.StudentID,
					"lesson_id":        qaVec.LessonID,
					"question_content": qaVec.QuestionContent,
					"answer_content":   qaVec.AnswerContent,
					"is_correct":       qaVec.IsCorrect,
					"time_spent":       qaVec.TimeSpent,
					"timestamp":        qaVec.Timestamp,
				},
			}

			if _, err := w.vectorService.Insert(ctx, "question_answer_vectors", data); err != nil {
				w.logger.Warn("insert qa vector failed", zap.Error(err))
			}
		}
	}

	return nil
}

// WriteLearningBehavior 写入学习行为（Nebula + Milvus）
func (w *WriterService) WriteLearningBehavior(ctx context.Context, behavior *LearningBehaviorVector) error {
	// 1. 写入 Milvus 向量
	data := []map[string]interface{}{
		{
			"id":              behavior.ID,
			"vector":          behavior.Vector,
			"student_id":      behavior.StudentID,
			"behavior_type":   behavior.BehaviorType,
			"lesson_id":       behavior.LessonID,
			"duration":        behavior.Duration,
			"completion_rate": behavior.CompletionRate,
			"quality_score":   behavior.QualityScore,
			"timestamp":       behavior.Timestamp,
		},
	}

	// 添加 metadata
	for k, v := range behavior.Metadata {
		data[0][k] = normalizeMetadataValue(v)
	}

	if _, err := w.vectorService.Insert(ctx, "learning_behavior_vectors", data); err != nil {
		return fmt.Errorf("insert learning behavior vector failed: %w", err)
	}

	// 2. 根据行为类型创建 Nebula 边
	switch behavior.BehaviorType {
	case "preview":
		edge := nebula.Edge{
			SrcID: behavior.StudentID,
			DstID: behavior.LessonID,
			Type:  "previews",
			Props: map[string]interface{}{
				"student_id":       behavior.StudentID,
				"lesson_id":        behavior.LessonID,
				"preview_time":     behavior.Timestamp,
				"preview_duration": behavior.Duration,
				"completion_rate":  behavior.CompletionRate,
			},
		}
		if err := w.edgeService.Create(ctx, edge); err != nil {
			w.logger.Warn("create previews edge failed", zap.Error(err))
		}
	case "review":
		edge := nebula.Edge{
			SrcID: behavior.StudentID,
			DstID: behavior.LessonID,
			Type:  "reviews",
			Props: map[string]interface{}{
				"student_id":      behavior.StudentID,
				"lesson_id":       behavior.LessonID,
				"review_time":     behavior.Timestamp,
				"review_duration": behavior.Duration,
				"review_count":    1, // 需要从其他地方获取
			},
		}
		if err := w.edgeService.Create(ctx, edge); err != nil {
			w.logger.Warn("create reviews edge failed", zap.Error(err))
		}
	case "homework":
		edge := nebula.Edge{
			SrcID: behavior.StudentID,
			DstID: behavior.LessonID,
			Type:  "completes_homework",
			Props: map[string]interface{}{
				"student_id":      behavior.StudentID,
				"lesson_id":       behavior.LessonID,
				"homework_id":     "", // 需要从 metadata 获取
				"submit_time":     behavior.Timestamp,
				"completion_time": behavior.Duration,
				"score":           behavior.QualityScore,
				"is_on_time":      true, // 需要判断
			},
		}
		if err := w.edgeService.Create(ctx, edge); err != nil {
			w.logger.Warn("create completes_homework edge failed", zap.Error(err))
		}
	case "mistake":
		// 需要 question_id
		if questionID, ok := behavior.Metadata["question_id"].(string); ok {
			edge := nebula.Edge{
				SrcID: behavior.StudentID,
				DstID: questionID,
				Type:  "collects_mistake",
				Props: map[string]interface{}{
					"student_id":   behavior.StudentID,
					"question_id":  questionID,
					"collect_time": behavior.Timestamp,
					"review_count": 1,
					"is_mastered":  false,
				},
			}
			if err := w.edgeService.Create(ctx, edge); err != nil {
				w.logger.Warn("create collects_mistake edge failed", zap.Error(err))
			}
		}
	}

	return nil
}

// WriteMBTITest 写入MBTI测试结果
func (w *WriterService) WriteMBTITest(ctx context.Context, test *MBTITest) error {
	// 1. 写入顶点
	vertex := nebula.Vertex{
		VID: test.ID,
		Tag: "mbti_test",
		Props: map[string]interface{}{
			"id":         test.ID,
			"student_id": test.StudentID,
			"mbti_type":  test.MBTIType,
			"e_score":    test.EScore,
			"i_score":    test.IScore,
			"s_score":    test.SScore,
			"n_score":    test.NScore,
			"t_score":    test.TScore,
			"f_score":    test.FScore,
			"j_score":    test.JScore,
			"p_score":    test.PScore,
			"test_date":  test.TestDate.Unix(),
			"create_at":  test.CreateAt.Unix(),
		},
	}
	if err := w.vertexService.Create(ctx, vertex); err != nil {
		return fmt.Errorf("create mbti_test vertex failed: %w", err)
	}

	// 2. 创建边
	edge := nebula.Edge{
		SrcID: test.StudentID,
		DstID: test.ID,
		Type:  "has_mbti",
		Props: map[string]interface{}{
			"student_id":   test.StudentID,
			"mbti_test_id": test.ID,
			"mbti_type":    test.MBTIType,
			"test_date":    test.TestDate.Unix(),
		},
	}
	return w.edgeService.Create(ctx, edge)
}

// normalizeMetadataValue 规范化 metadata 值类型
func normalizeMetadataValue(v interface{}) interface{} {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	default:
		return v
	}
}
