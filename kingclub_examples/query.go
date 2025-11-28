package kingclub

import (
	"context"
	"fmt"
	"time"

	"github.com/UTC-Six/brain/internal/milvus"
	"github.com/UTC-Six/brain/internal/nebula"
	nebula_go "github.com/vesoft-inc/nebula-go/v3"
	"go.uber.org/zap"
)

// QueryService 查询服务
type QueryService struct {
	nebulaClient *nebula.Client
	milvusClient *milvus.Client
	logger       *zap.Logger
}

// NewQueryService 创建查询服务
func NewQueryService(
	nebulaClient *nebula.Client,
	milvusClient *milvus.Client,
	logger *zap.Logger,
) *QueryService {
	return &QueryService{
		nebulaClient: nebulaClient,
		milvusClient: milvusClient,
		logger:       logger,
	}
}

// GetStudentBasicInfo 获取学生基础信息
func (q *QueryService) GetStudentBasicInfo(ctx context.Context, studentID string) (*StudentBasicInfo, error) {
	query := fmt.Sprintf(`
		MATCH (s:student {id: "%s"})
		RETURN s.id as id, s.name as name, s.age as age, s.grade as grade,
		       s.gender as gender, s.avatar_url as avatar_url,
		       s.school as school, s.class_name as class_name
	`, studentID)

	result, err := q.nebulaClient.Execute(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query student basic info failed: %w", err)
	}

	if result.IsEmpty() {
		return nil, fmt.Errorf("student not found: %s", studentID)
	}

	row, err := result.GetRowValuesByIndex(0)
	if err != nil {
		return nil, err
	}

	info := &StudentBasicInfo{}
	if id, err := getStringValue(row, "id"); err == nil {
		info.ID = id
	}
	if name, err := getStringValue(row, "name"); err == nil {
		info.Name = name
	}
	if age, err := getIntValue(row, "age"); err == nil {
		info.Age = age
	}
	if grade, err := getStringValue(row, "grade"); err == nil {
		info.Grade = grade
	}
	if gender, err := getStringValue(row, "gender"); err == nil {
		info.Gender = gender
	}
	if avatar, err := getStringValue(row, "avatar_url"); err == nil {
		info.AvatarURL = avatar
	}
	if school, err := getStringValue(row, "school"); err == nil {
		info.School = school
	}
	if className, err := getStringValue(row, "class_name"); err == nil {
		info.ClassName = className
	}

	return info, nil
}

// GetLearningAbility 获取学生学习能力
func (q *QueryService) GetLearningAbility(ctx context.Context, studentID string) (*LearningAbility, error) {
	// 查询知识点掌握情况
	query := fmt.Sprintf(`
		MATCH (s:student {id: "%s"})-[st:studies]->(kp:knowledge_point)
		RETURN kp.name as name, st.mastery_level as mastery_level,
		       st.study_count as study_count, st.total_study_duration as duration
	`, studentID)

	result, err := q.nebulaClient.Execute(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query learning ability failed: %w", err)
	}

	ability := &LearningAbility{
		TotalKnowledge: result.GetRowSize(),
	}

	if ability.TotalKnowledge == 0 {
		return ability, nil
	}

	var totalMastery int
	for i := 0; i < result.GetRowSize(); i++ {
		row, err := result.GetRowValuesByIndex(i)
		if err != nil {
			continue
		}
		if mastery, err := getIntValue(row, "mastery_level"); err == nil {
			totalMastery += mastery
		}
	}

	ability.AverageMastery = float64(totalMastery) / float64(ability.TotalKnowledge)

	return ability, nil
}

// GetSubjectAbility 获取学生学科能力
func (q *QueryService) GetSubjectAbility(ctx context.Context, studentID string) (*SubjectAbility, error) {
	query := fmt.Sprintf(`
		MATCH (s:student {id: "%s"})-[a:attends]->(c:course)
		MATCH (s)-[st:studies]->(kp:knowledge_point)
		WHERE kp.subject == c.subject
		RETURN c.subject as subject, AVG(st.mastery_level) as avg_mastery,
		       COUNT(DISTINCT kp.id) as knowledge_count
		GROUP BY c.subject
	`, studentID)

	result, err := q.nebulaClient.Execute(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query subject ability failed: %w", err)
	}

	ability := &SubjectAbility{
		Subjects: make(map[string]*SubjectPerformance),
	}

	for i := 0; i < result.GetRowSize(); i++ {
		row, err := result.GetRowValuesByIndex(i)
		if err != nil {
			continue
		}

		subject, _ := getStringValue(row, "subject")
		if subject == "" {
			continue
		}

		perf := &SubjectPerformance{Subject: subject}
		if avgMastery, err := getDoubleValue(row, "avg_mastery"); err == nil {
			perf.AverageMastery = avgMastery
		}
		if count, err := getIntValue(row, "knowledge_count"); err == nil {
			perf.KnowledgeCount = count
		}

		// 查询文章评分
		articleQuery := fmt.Sprintf(`
			MATCH (s:student {id: "%s"})-[w:writes]->(art:article)
			WHERE art.lesson_id IN (
				MATCH (l:lesson)-[:contains]->(kp:knowledge_point)
				WHERE kp.subject == "%s"
				RETURN l.id
			)
			RETURN w.score as score
			ORDER BY w.write_time
		`, studentID, subject)

		articleResult, err := q.nebulaClient.Execute(ctx, articleQuery)
		if err == nil {
			scores := make([]float64, 0)
			for j := 0; j < articleResult.GetRowSize(); j++ {
				articleRow, err := articleResult.GetRowValuesByIndex(j)
				if err != nil {
					continue
				}
				if score, err := getDoubleValue(articleRow, "score"); err == nil {
					scores = append(scores, score)
				}
			}
			perf.ArticleScores = scores
		}

		ability.Subjects[subject] = perf
	}

	return ability, nil
}

// GetPersonality 获取学生性格特征
func (q *QueryService) GetPersonality(ctx context.Context, studentID string) (*Personality, error) {
	query := fmt.Sprintf(`
		MATCH (s:student {id: "%s"})-[h:has_mbti]->(m:mbti_test)
		RETURN m.mbti_type as mbti_type, m.e_score as e_score, m.i_score as i_score,
		       m.s_score as s_score, m.n_score as n_score, m.t_score as t_score,
		       m.f_score as f_score, m.j_score as j_score, m.p_score as p_score
		ORDER BY m.test_date DESC
		LIMIT 1
	`, studentID)

	result, err := q.nebulaClient.Execute(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query personality failed: %w", err)
	}

	if result.IsEmpty() {
		return &Personality{}, nil
	}

	row, err := result.GetRowValuesByIndex(0)
	if err != nil {
		return nil, err
	}

	personality := &Personality{}
	if mbtiType, err := getStringValue(row, "mbti_type"); err == nil {
		personality.MBTIType = mbtiType
	}
	if score, err := getIntValue(row, "e_score"); err == nil {
		personality.EScore = score
	}
	if score, err := getIntValue(row, "i_score"); err == nil {
		personality.IScore = score
	}
	if score, err := getIntValue(row, "s_score"); err == nil {
		personality.SScore = score
	}
	if score, err := getIntValue(row, "n_score"); err == nil {
		personality.NScore = score
	}
	if score, err := getIntValue(row, "t_score"); err == nil {
		personality.TScore = score
	}
	if score, err := getIntValue(row, "f_score"); err == nil {
		personality.FScore = score
	}
	if score, err := getIntValue(row, "j_score"); err == nil {
		personality.JScore = score
	}
	if score, err := getIntValue(row, "p_score"); err == nil {
		personality.PScore = score
	}

	return personality, nil
}

// GetLearningHabits 获取学生学习习惯
func (q *QueryService) GetLearningHabits(ctx context.Context, studentID string) (*LearningHabits, error) {
	habits := &LearningHabits{}

	// 查询预习情况
	previewQuery := fmt.Sprintf(`
		MATCH (s:student {id: "%s"})-[p:previews]->(l:lesson)
		RETURN COUNT(*) as count, AVG(p.completion_rate) as avg_rate,
		       AVG(p.preview_duration) as avg_duration
	`, studentID)

	result, err := q.nebulaClient.Execute(ctx, previewQuery)
	if err == nil && !result.IsEmpty() {
		row, _ := result.GetRowValuesByIndex(0)
		habits.Preview = &HabitStats{}
		if count, err := getIntValue(row, "count"); err == nil {
			habits.Preview.Count = count
		}
		if rate, err := getDoubleValue(row, "avg_rate"); err == nil {
			habits.Preview.AverageRate = rate
		}
		if duration, err := getDoubleValue(row, "avg_duration"); err == nil {
			habits.Preview.AverageDuration = int(duration)
		}
	}

	// 查询复习情况
	reviewQuery := fmt.Sprintf(`
		MATCH (s:student {id: "%s"})-[r:reviews]->(l:lesson)
		RETURN COUNT(*) as count, AVG(r.review_duration) as avg_duration,
		       AVG(r.review_count) as avg_count
	`, studentID)

	result, err = q.nebulaClient.Execute(ctx, reviewQuery)
	if err == nil && !result.IsEmpty() {
		row, _ := result.GetRowValuesByIndex(0)
		habits.Review = &HabitStats{}
		if count, err := getIntValue(row, "count"); err == nil {
			habits.Review.Count = count
		}
		if duration, err := getDoubleValue(row, "avg_duration"); err == nil {
			habits.Review.AverageDuration = int(duration)
		}
	}

	// 查询作业情况
	homeworkQuery := fmt.Sprintf(`
		MATCH (s:student {id: "%s"})-[h:completes_homework]->(l:lesson)
		RETURN COUNT(*) as count, 
		       SUM(CASE WHEN h.is_on_time THEN 1 ELSE 0 END) as on_time_count,
		       AVG(h.score) as avg_score, AVG(h.completion_time) as avg_time
	`, studentID)

	result, err = q.nebulaClient.Execute(ctx, homeworkQuery)
	if err == nil && !result.IsEmpty() {
		row, _ := result.GetRowValuesByIndex(0)
		habits.Homework = &HabitStats{}
		if count, err := getIntValue(row, "count"); err == nil {
			habits.Homework.Count = count
		}
		if score, err := getDoubleValue(row, "avg_score"); err == nil {
			habits.Homework.AverageRate = score
		}
		if time, err := getDoubleValue(row, "avg_time"); err == nil {
			habits.Homework.AverageDuration = int(time)
		}
	}

	// 查询错题整理
	mistakeQuery := fmt.Sprintf(`
		MATCH (s:student {id: "%s"})-[cm:collects_mistake]->(q:question)
		RETURN COUNT(*) as count,
		       SUM(CASE WHEN cm.is_mastered THEN 1 ELSE 0 END) as mastered_count,
		       AVG(cm.review_count) as avg_review
	`, studentID)

	result, err = q.nebulaClient.Execute(ctx, mistakeQuery)
	if err == nil && !result.IsEmpty() {
		row, _ := result.GetRowValuesByIndex(0)
		habits.MistakeCollection = &HabitStats{}
		if count, err := getIntValue(row, "count"); err == nil {
			habits.MistakeCollection.Count = count
		}
		if review, err := getDoubleValue(row, "avg_review"); err == nil {
			habits.MistakeCollection.AverageRate = review
		}
	}

	return habits, nil
}

// GetSocialBackground 获取学生社会关系和背景
func (q *QueryService) GetSocialBackground(ctx context.Context, studentID string) (*SocialBackground, error) {
	background := &SocialBackground{
		FamilyMembers: make([]*FamilyMemberInfo, 0),
		Friends:       make([]*FriendInfo, 0),
	}

	// 查询家庭成员
	familyQuery := fmt.Sprintf(`
		MATCH (s:student {id: "%s"})-[b:belongs_to]->(fm:family_member)
		RETURN fm.name as name, fm.relationship as relationship,
		       fm.education_level as education_level, fm.occupation as occupation,
		       b.closeness as closeness
	`, studentID)

	result, err := q.nebulaClient.Execute(ctx, familyQuery)
	if err == nil {
		for i := 0; i < result.GetRowSize(); i++ {
			row, err := result.GetRowValuesByIndex(i)
			if err != nil {
				continue
			}

			info := &FamilyMemberInfo{}
			if name, err := getStringValue(row, "name"); err == nil {
				info.Name = name
			}
			if rel, err := getStringValue(row, "relationship"); err == nil {
				info.Relationship = rel
			}
			if edu, err := getStringValue(row, "education_level"); err == nil {
				info.EducationLevel = edu
			}
			if occ, err := getStringValue(row, "occupation"); err == nil {
				info.Occupation = occ
			}
			if close, err := getIntValue(row, "closeness"); err == nil {
				info.Closeness = close
			}
			background.FamilyMembers = append(background.FamilyMembers, info)
		}
	}

	// 查询朋友关系
	friendQuery := fmt.Sprintf(`
		MATCH (s:student {id: "%s"})-[f:friends_with]->(s2:student)
		RETURN s2.name as name, s2.grade as grade,
		       f.friendship_level as level, f.since as since
	`, studentID)

	result, err = q.nebulaClient.Execute(ctx, friendQuery)
	if err == nil {
		for i := 0; i < result.GetRowSize(); i++ {
			row, err := result.GetRowValuesByIndex(i)
			if err != nil {
				continue
			}

			info := &FriendInfo{}
			if name, err := getStringValue(row, "name"); err == nil {
				info.Name = name
			}
			if grade, err := getStringValue(row, "grade"); err == nil {
				info.Grade = grade
			}
			if level, err := getIntValue(row, "level"); err == nil {
				info.FriendshipLevel = level
			}
			if since, err := getInt64Value(row, "since"); err == nil {
				info.Since = time.Unix(since, 0)
			}
			background.Friends = append(background.Friends, info)
		}
	}

	return background, nil
}

// GetInterests 获取学生兴趣爱好
func (q *QueryService) GetInterests(ctx context.Context, studentID string) ([]*Interest, error) {
	query := fmt.Sprintf(`
		MATCH (s:student {id: "%s"})-[l:likes]->(h:hobby)
		RETURN h.name as name, h.category as category,
		       l.interest_level as level, l.since as since
		ORDER BY l.interest_level DESC
	`, studentID)

	result, err := q.nebulaClient.Execute(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query interests failed: %w", err)
	}

	interests := make([]*Interest, 0, result.GetRowSize())
	for i := 0; i < result.GetRowSize(); i++ {
		row, err := result.GetRowValuesByIndex(i)
		if err != nil {
			continue
		}

		interest := &Interest{}
		if name, err := getStringValue(row, "name"); err == nil {
			interest.Name = name
		}
		if category, err := getStringValue(row, "category"); err == nil {
			interest.Category = category
		}
		if level, err := getIntValue(row, "level"); err == nil {
			interest.InterestLevel = level
		}
		if since, err := getInt64Value(row, "since"); err == nil {
			interest.Since = time.Unix(since, 0)
		}
		interests = append(interests, interest)
	}

	return interests, nil
}

// 辅助函数：从结果行获取值
func getStringValue(row *nebula_go.Record, colName string) (string, error) {
	val, err := row.GetValueByColName(colName)
	if err != nil {
		return "", err
	}
	return val.AsString()
}

func getIntValue(row *nebula_go.Record, colName string) (int, error) {
	val, err := row.GetValueByColName(colName)
	if err != nil {
		return 0, err
	}
	i, err := val.AsInt()
	if err != nil {
		return 0, err
	}
	return int(i), nil
}

func getInt64Value(row *nebula_go.Record, colName string) (int64, error) {
	val, err := row.GetValueByColName(colName)
	if err != nil {
		return 0, err
	}
	return val.AsInt()
}

func getDoubleValue(row *nebula_go.Record, colName string) (float64, error) {
	val, err := row.GetValueByColName(colName)
	if err != nil {
		return 0, err
	}
	return val.AsFloat()
}
