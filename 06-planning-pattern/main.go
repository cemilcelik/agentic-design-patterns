// Planning Pattern - Dinamik Planlama ve Yürütme Sistemi (Go)
// Karmaşık hedefleri analiz eder, plan oluşturur ve adımları icra eder

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

// Plan bir görevin adımlarını temsil eder
type Plan struct {
	Goal  string   `json:"goal"`
	Steps []string `json:"steps"`
}

// StepResult bir adımın sonucunu tutar
type StepResult struct {
	StepNumber int
	StepName   string
	Content    string
	Success    bool
}

// PlanningAgent planlama desenini uygulayan ajan
type PlanningAgent struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

// NewPlanningAgent yeni bir planlama ajanı oluşturur
func NewPlanningAgent(apiKey string) (*PlanningAgent, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("client oluşturulamadı: %w", err)
	}

	model := client.GenerativeModel(os.Getenv("GEMINI_MODEL"))
	model.SetTemperature(0.7)

	return &PlanningAgent{
		client: client,
		model:  model,
	}, nil
}

// Close kaynakları temizler
func (pa *PlanningAgent) Close() {
	pa.client.Close()
}

// CreatePlan verilen görev için bir plan oluşturur
func (pa *PlanningAgent) CreatePlan(ctx context.Context, task string) (*Plan, error) {
	prompt := fmt.Sprintf(`Aşağıdaki görevi tamamlamak için bir plan oluştur. Minimum 2, maksimum 5 step olmalı.
Planı JSON formatında ver, sadece JSON döndür, başka açıklama ekleme.

Format:
{
  "goal": "görev açıklaması",
  "steps": ["adım 1", "adım 2", "adım 3", ...]
}

Görev: %s`, task)

	resp, err := pa.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("plan oluşturulamadı: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("boş yanıt alındı")
	}

	content := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var plan Plan
	if err := json.Unmarshal([]byte(content), &plan); err != nil {
		return nil, fmt.Errorf("JSON parse hatası: %w\nContent: %s", err, content)
	}

	return &plan, nil
}

// ExecuteStep tek bir adımı yürütür
func (pa *PlanningAgent) ExecuteStep(ctx context.Context, step string, stepNumber int, previousResults []StepResult) (*StepResult, error) {
	context := ""
	if len(previousResults) > 0 {
		context = "\n\nÖnceki adımların sonuçları:\n"
		for _, result := range previousResults {
			context += fmt.Sprintf("\n%d. %s:\n%s\n", result.StepNumber, result.StepName, result.Content)
		}
	}

	prompt := fmt.Sprintf(`Bu adımı tamamla ve detaylı sonuç üret:%s

Adım: %s

Not: Bu adımı tamamlarken önceki adımların sonuçlarını dikkate al ve üzerine inşa et.`, context, step)

	resp, err := pa.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return &StepResult{
			StepNumber: stepNumber,
			StepName:   step,
			Content:    "",
			Success:    false,
		}, fmt.Errorf("adım yürütülemedi: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return &StepResult{
			StepNumber: stepNumber,
			StepName:   step,
			Content:    "",
			Success:    false,
		}, fmt.Errorf("boş yanıt alındı")
	}

	content := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

	return &StepResult{
		StepNumber: stepNumber,
		StepName:   step,
		Content:    content,
		Success:    true,
	}, nil
}

// ExecutePlan tüm planı yürütür
func (pa *PlanningAgent) ExecutePlan(ctx context.Context, plan *Plan) ([]StepResult, error) {
	results := make([]StepResult, 0, len(plan.Steps))

	fmt.Printf("\n🎯 Hedef: %s\n", plan.Goal)
	fmt.Printf("📋 Toplam %d adım planlandı\n\n", len(plan.Steps))

	for i, step := range plan.Steps {
		fmt.Printf("⚙️  Adım %d/%d: %s\n", i+1, len(plan.Steps), step)

		result, err := pa.ExecuteStep(ctx, step, i+1, results)
		if err != nil {
			log.Printf("❌ Adım %d başarısız: %v\n", i+1, err)
			return results, err
		}

		results = append(results, *result)
		fmt.Printf("✅ Adım %d tamamlandı\n\n", i+1)
	}

	return results, nil
}

// CombineResults tüm adımların sonuçlarını birleştirir
func (pa *PlanningAgent) CombineResults(ctx context.Context, plan *Plan, results []StepResult) (string, error) {
	allContent := fmt.Sprintf("Hedef: %s\n\n", plan.Goal)

	for _, result := range results {
		allContent += fmt.Sprintf("## %s\n\n%s\n\n", result.StepName, result.Content)
	}

	prompt := fmt.Sprintf(`Aşağıdaki içerikleri tutarlı ve profesyonel bir blog yazısı haline getir.
Her bölümü doğal bir akışla birleştir ve geçişleri yumuşat.

%s`, allContent)

	resp, err := pa.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("sonuçlar birleştirilemedi: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("boş yanıt alındı")
	}

	finalContent := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	return finalContent, nil
}

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env dosyası bulunamadı, sistem environment variable'ları kullanılacak")
	}

	apiKey := os.Getenv("GOOGLE_API_KEY")

	agent, err := NewPlanningAgent(apiKey)
	if err != nil {
		log.Fatalf("Agent oluşturulamadı: %v", err)
	}
	defer agent.Close()

	ctx := context.Background()

	// Görev tanımı
	task := "Go programlama dilinde microservices mimarisi hakkında kapsamlı bir blog yazısı yaz. Best practice'leri, design pattern'leri ve pratik örnekleri içersin."

	fmt.Println("🚀 Planning Pattern Demo - Blog Yazısı Oluşturucu")
	fmt.Println("================================================")

	// 1. Plan Oluşturma
	fmt.Println("📝 Plan oluşturuluyor...")
	plan, err := agent.CreatePlan(ctx, task)
	if err != nil {
		log.Fatalf("Plan oluşturulamadı: %v", err)
	}

	// 2. Planı Yürütme
	results, err := agent.ExecutePlan(ctx, plan)
	if err != nil {
		log.Fatalf("Plan yürütülemedi: %v", err)
	}

	// 3. Sonuçları Birleştirme
	fmt.Println("🔄 Sonuçlar birleştiriliyor...")
	finalBlogPost, err := agent.CombineResults(ctx, plan, results)
	if err != nil {
		log.Fatalf("Sonuçlar birleştirilemedi: %v", err)
	}

	// 4. Sonucu Göster
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("📄 TAMAMLANAN BLOG YAZISI")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println(finalBlogPost)
	fmt.Println(strings.Repeat("=", 80))

	fmt.Println("\n✨ İşlem başarıyla tamamlandı!")
}
