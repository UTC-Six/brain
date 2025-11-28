package kingclub

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

// ProfileService 学生画像生成服务
type ProfileService struct {
	queryService *QueryService
	ragService   *RAGService
	logger       *zap.Logger
	llmFunc      func(prompt string) (string, error) // 大模型调用函数（需要外部提供）
}

// NewProfileService 创建画像服务
func NewProfileService(
	queryService *QueryService,
	ragService *RAGService,
	logger *zap.Logger,
	llmFunc func(prompt string) (string, error),
) *ProfileService {
	return &ProfileService{
		queryService: queryService,
		ragService:   ragService,
		logger:       logger,
		llmFunc:      llmFunc,
	}
}

// GenerateStudentProfile 生成学生完整画像
func (p *ProfileService) GenerateStudentProfile(ctx context.Context, studentID string) (*StudentProfile, error) {
	profile := &StudentProfile{}

	// 1. 获取基础信息
	basicInfo, err := p.queryService.GetStudentBasicInfo(ctx, studentID)
	if err != nil {
		return nil, fmt.Errorf("get basic info failed: %w", err)
	}
	profile.BasicInfo = basicInfo

	// 2. 获取身体状况（通过 RAG 召回相关会话）
	physicalCondition, err := p.generatePhysicalCondition(ctx, studentID)
	if err != nil {
		p.logger.Warn("generate physical condition failed", zap.Error(err))
	} else {
		profile.PhysicalCondition = physicalCondition
	}

	// 3. 获取学习能力
	learningAbility, err := p.queryService.GetLearningAbility(ctx, studentID)
	if err != nil {
		p.logger.Warn("get learning ability failed", zap.Error(err))
	} else {
		profile.LearningAbility = learningAbility
	}

	// 4. 获取学科能力
	subjectAbility, err := p.queryService.GetSubjectAbility(ctx, studentID)
	if err != nil {
		p.logger.Warn("get subject ability failed", zap.Error(err))
	} else {
		profile.SubjectAbility = subjectAbility
	}

	// 5. 获取性格特征
	personality, err := p.queryService.GetPersonality(ctx, studentID)
	if err != nil {
		p.logger.Warn("get personality failed", zap.Error(err))
	} else {
		profile.Personality = personality
	}

	// 6. 获取学习习惯
	learningHabits, err := p.queryService.GetLearningHabits(ctx, studentID)
	if err != nil {
		p.logger.Warn("get learning habits failed", zap.Error(err))
	} else {
		profile.LearningHabits = learningHabits
	}

	// 7. 获取社会关系和背景
	socialBackground, err := p.queryService.GetSocialBackground(ctx, studentID)
	if err != nil {
		p.logger.Warn("get social background failed", zap.Error(err))
	} else {
		profile.SocialBackground = socialBackground
	}

	// 8. 获取兴趣爱好
	interests, err := p.queryService.GetInterests(ctx, studentID)
	if err != nil {
		p.logger.Warn("get interests failed", zap.Error(err))
	} else {
		profile.Interests = interests
	}

	// 9. 使用大模型生成综合分析（可选）
	if p.llmFunc != nil {
		if err := p.enhanceProfileWithLLM(ctx, profile, studentID); err != nil {
			p.logger.Warn("enhance profile with LLM failed", zap.Error(err))
		}
	}

	return profile, nil
}

// generatePhysicalCondition 生成身体状况信息
func (p *ProfileService) generatePhysicalCondition(ctx context.Context, studentID string) (*PhysicalCondition, error) {
	// 通过 RAG 召回相关会话
	query := "学生的身体状况、身体能力、生活习惯、健康状况"
	conversations, err := p.ragService.RecallConversations(ctx, query, studentID, 10)
	if err != nil {
		return nil, err
	}

	if len(conversations) == 0 {
		return &PhysicalCondition{}, nil
	}

	// 提取会话内容
	contents := make([]string, 0, len(conversations))
	for _, conv := range conversations {
		if conv.Content != "" {
			contents = append(contents, conv.Content)
		}
	}

	// 使用大模型分析（如果有）
	condition := &PhysicalCondition{}
	if p.llmFunc != nil {
		prompt := fmt.Sprintf(`
请分析以下学生会话内容，提取学生的身体状况、身体能力、生活习惯和健康状况信息：

%s

请以JSON格式返回，包含以下字段：
- physical_ability: 身体能力描述
- living_habits: 生活习惯描述
- health_status: 健康状况描述
		`, strings.Join(contents, "\n\n"))

		result, err := p.llmFunc(prompt)
		if err == nil {
			// 解析 JSON 结果（简化处理，实际应该使用 JSON 解析）
			condition.PhysicalAbility = extractField(result, "physical_ability")
			condition.LivingHabits = extractField(result, "living_habits")
			condition.HealthStatus = extractField(result, "health_status")
		}
	}

	return condition, nil
}

// enhanceProfileWithLLM 使用大模型增强画像
func (p *ProfileService) enhanceProfileWithLLM(ctx context.Context, profile *StudentProfile, studentID string) error {
	if p.llmFunc == nil {
		return nil
	}

	// 召回相关数据
	conversations, _ := p.ragService.RecallConversations(ctx, "学生的学习状态和表现", studentID, 5)
	articles, _ := p.ragService.RecallArticles(ctx, "学生的写作能力和表达", studentID, 3)
	behaviors, _ := p.ragService.RecallLearningBehaviors(ctx, "学生的学习习惯和时间管理", studentID, 10)

	// 构建提示词
	prompt := fmt.Sprintf(`
基于以下学生画像数据和相关信息，生成一份综合的学生画像分析报告：

## 基础信息
%s

## 学习能力
%s

## 学科能力
%s

## 性格特征
%s

## 学习习惯
%s

## 社会关系和背景
%s

## 兴趣爱好
%s

## 相关会话内容
%s

## 相关文章
%s

## 学习行为
%s

请生成一份详细的学生画像分析报告，包括：
1. 学生整体特点总结
2. 学习能力评估
3. 学科优势与劣势
4. 性格特征分析
5. 学习习惯评价
6. 改进建议
	`,
		p.formatBasicInfo(profile.BasicInfo),
		p.formatLearningAbility(profile.LearningAbility),
		p.formatSubjectAbility(profile.SubjectAbility),
		p.formatPersonality(profile.Personality),
		p.formatLearningHabits(profile.LearningHabits),
		p.formatSocialBackground(profile.SocialBackground),
		p.formatInterests(profile.Interests),
		p.formatConversations(conversations),
		p.formatArticles(articles),
		p.formatBehaviors(behaviors),
	)

	result, err := p.llmFunc(prompt)
	if err != nil {
		return err
	}

	// 将结果存储到画像中（可以添加一个新字段存储 LLM 分析结果）
	p.logger.Info("LLM profile analysis generated", zap.String("student_id", studentID), zap.String("result", result))

	return nil
}

// 辅助函数：构建画像摘要
func (p *ProfileService) buildProfileSummary(profile *StudentProfile) string {
	var parts []string

	if profile.BasicInfo != nil {
		parts = append(parts, fmt.Sprintf("学生：%s，%d岁，%s年级", profile.BasicInfo.Name, profile.BasicInfo.Age, profile.BasicInfo.Grade))
	}

	if profile.LearningAbility != nil {
		parts = append(parts, fmt.Sprintf("学习能力：掌握%d个知识点，平均掌握度%.2f", profile.LearningAbility.TotalKnowledge, profile.LearningAbility.AverageMastery))
	}

	if profile.Personality != nil && profile.Personality.MBTIType != "" {
		parts = append(parts, fmt.Sprintf("性格类型：%s", profile.Personality.MBTIType))
	}

	return strings.Join(parts, "；")
}

// 格式化函数
func (p *ProfileService) formatBasicInfo(info *StudentBasicInfo) string {
	if info == nil {
		return "无"
	}
	return fmt.Sprintf("姓名：%s，年龄：%d，年级：%s，性别：%s，学校：%s，班级：%s", info.Name, info.Age, info.Grade, info.Gender, info.School, info.ClassName)
}

func (p *ProfileService) formatLearningAbility(ability *LearningAbility) string {
	if ability == nil {
		return "无"
	}
	return fmt.Sprintf("总知识点：%d，平均掌握度：%.2f", ability.TotalKnowledge, ability.AverageMastery)
}

func (p *ProfileService) formatSubjectAbility(ability *SubjectAbility) string {
	if ability == nil || len(ability.Subjects) == 0 {
		return "无"
	}
	var parts []string
	for subject, perf := range ability.Subjects {
		parts = append(parts, fmt.Sprintf("%s：掌握%d个知识点，平均掌握度%.2f", subject, perf.KnowledgeCount, perf.AverageMastery))
	}
	return strings.Join(parts, "；")
}

func (p *ProfileService) formatPersonality(personality *Personality) string {
	if personality == nil || personality.MBTIType == "" {
		return "无"
	}
	return fmt.Sprintf("MBTI类型：%s，E/I：%d/%d，S/N：%d/%d，T/F：%d/%d，J/P：%d/%d",
		personality.MBTIType, personality.EScore, personality.IScore,
		personality.SScore, personality.NScore, personality.TScore, personality.FScore,
		personality.JScore, personality.PScore)
}

func (p *ProfileService) formatLearningHabits(habits *LearningHabits) string {
	if habits == nil {
		return "无"
	}
	var parts []string
	if habits.Preview != nil {
		parts = append(parts, fmt.Sprintf("预习：%d次，平均完成率%.2f", habits.Preview.Count, habits.Preview.AverageRate))
	}
	if habits.Review != nil {
		parts = append(parts, fmt.Sprintf("复习：%d次，平均时长%d分钟", habits.Review.Count, habits.Review.AverageDuration))
	}
	if habits.Homework != nil {
		parts = append(parts, fmt.Sprintf("作业：%d次，平均得分%.2f", habits.Homework.Count, habits.Homework.AverageRate))
	}
	return strings.Join(parts, "；")
}

func (p *ProfileService) formatSocialBackground(background *SocialBackground) string {
	if background == nil {
		return "无"
	}
	var parts []string
	if len(background.FamilyMembers) > 0 {
		parts = append(parts, fmt.Sprintf("家庭成员：%d人", len(background.FamilyMembers)))
	}
	if len(background.Friends) > 0 {
		parts = append(parts, fmt.Sprintf("朋友：%d人", len(background.Friends)))
	}
	return strings.Join(parts, "；")
}

func (p *ProfileService) formatInterests(interests []*Interest) string {
	if len(interests) == 0 {
		return "无"
	}
	var parts []string
	for _, interest := range interests {
		parts = append(parts, fmt.Sprintf("%s（%s）", interest.Name, interest.Category))
	}
	return strings.Join(parts, "、")
}

func (p *ProfileService) formatConversations(conversations []*ConversationVector) string {
	if len(conversations) == 0 {
		return "无"
	}
	var parts []string
	for i, conv := range conversations {
		if i >= 3 { // 只显示前3条
			break
		}
		content := conv.Content
		if len(content) > 100 {
			content = content[:100] + "..."
		}
		parts = append(parts, content)
	}
	return strings.Join(parts, "\n\n")
}

func (p *ProfileService) formatArticles(articles []*ArticleVector) string {
	if len(articles) == 0 {
		return "无"
	}
	var parts []string
	for i, article := range articles {
		if i >= 2 { // 只显示前2篇
			break
		}
		parts = append(parts, fmt.Sprintf("《%s》（%d字，得分%.2f）", article.Title, article.WordCount, article.Score))
	}
	return strings.Join(parts, "\n")
}

func (p *ProfileService) formatBehaviors(behaviors []*LearningBehaviorVector) string {
	if len(behaviors) == 0 {
		return "无"
	}
	var parts []string
	for i, behavior := range behaviors {
		if i >= 5 { // 只显示前5条
			break
		}
		parts = append(parts, fmt.Sprintf("%s：时长%d分钟，完成率%.2f", behavior.BehaviorType, behavior.Duration, behavior.CompletionRate))
	}
	return strings.Join(parts, "；")
}

// extractField 从文本中提取字段（简化实现）
func extractField(text, field string) string {
	// 实际应该使用 JSON 解析
	// 这里只是简化实现
	prefix := fmt.Sprintf(`"%s":`, field)
	idx := strings.Index(text, prefix)
	if idx == -1 {
		return ""
	}
	start := idx + len(prefix)
	// 跳过引号和空格
	for start < len(text) && (text[start] == '"' || text[start] == ' ' || text[start] == ':') {
		start++
	}
	end := start
	for end < len(text) && text[end] != '"' && text[end] != ',' && text[end] != '}' {
		end++
	}
	return text[start:end]
}
