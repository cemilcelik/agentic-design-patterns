// Multi-Agent Collaboration Pattern - Çoklu Ajan İşbirliği Sistemi (Go)
// Farklı uzmanlıklara sahip ajanların koordineli çalışarak karmaşık görevleri çözmesini sağlar

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

// Agent temel ajan yapısı
type Agent struct {
	Name         string
	Role         string
	SystemPrompt string
	Client       *genai.Client
}

// Message ajanlar arası mesaj yapısı
type Message struct {
	From    string
	To      string
	Content string
	Type    string
}

// ResearchResult araştırma sonucu
type ResearchResult struct {
	Query    string   `json:"query"`
	Findings []string `json:"findings"`
	Sources  []string `json:"sources"`
}

// AnalysisResult analiz sonucu
type AnalysisResult struct {
	KeyPoints  []string `json:"key_points"`
	Insights   []string `json:"insights"`
	Confidence string   `json:"confidence"`
}

// Coordinator ajanları koordine eden yapı
type Coordinator struct {
	Agents  map[string]*Agent
	Results map[string]interface{}
}

// NewAgent yeni bir ajan oluşturur
func NewAgent(name, role, systemPrompt string, apiKey string) (*Agent, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("client oluşturulamadı: %v", err)
	}

	return &Agent{
		Name:         name,
		Role:         role,
		SystemPrompt: systemPrompt,
		Client:       client,
	}, nil
}

// Execute ajanın görevini çalıştırır
func (a *Agent) Execute(ctx context.Context, input string) (string, error) {
	model := a.Client.GenerativeModel(os.Getenv("GEMINI_MODEL"))

	// System prompt'u yapılandır
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(a.SystemPrompt)},
	}

	// Prompt oluştur
	prompt := fmt.Sprintf("Görev: %s\n\nGirdi:\n%s", a.Role, input)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("içerik üretilirken hata: %v", err)
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("cevap alınamadı")
	}

	var result strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		result.WriteString(fmt.Sprintf("%v", part))
	}

	return result.String(), nil
}

// Close ajanı kapatır
func (a *Agent) Close() {
	if a.Client != nil {
		a.Client.Close()
	}
}

// NewCoordinator yeni bir koordinatör oluşturur
func NewCoordinator() *Coordinator {
	return &Coordinator{
		Agents:  make(map[string]*Agent),
		Results: make(map[string]interface{}),
	}
}

// AddAgent koordinatöre ajan ekler
func (c *Coordinator) AddAgent(agent *Agent) {
	c.Agents[agent.Name] = agent
}

// ExecuteWorkflow iş akışını çalıştırır
func (c *Coordinator) ExecuteWorkflow(ctx context.Context, topic string) (string, error) {
	fmt.Printf("\n🚀 Multi-Agent Workflow Başlatılıyor: '%s'\n", topic)
	fmt.Println(strings.Repeat("=", 70))

	// 1. Research Agent - Araştırma yap
	fmt.Println("\n📚 Adım 1: Research Agent çalışıyor...")
	researchAgent, exists := c.Agents["researcher"]
	if !exists {
		return "", fmt.Errorf("research agent bulunamadı")
	}

	researchInput := fmt.Sprintf("Konu: %s\n\nBu konu hakkında kapsamlı araştırma yap. Ana bulguları, önemli noktaları ve kaynakları JSON formatında listele.", topic)
	researchOutput, err := researchAgent.Execute(ctx, researchInput)
	if err != nil {
		return "", fmt.Errorf("research agent hatası: %v", err)
	}
	c.Results["research"] = researchOutput
	fmt.Printf("✅ Araştırma tamamlandı (%d karakter)\n", len(researchOutput))

	// 2. Analysis Agent - Analiz yap
	fmt.Println("\n🔍 Adım 2: Analysis Agent çalışıyor...")
	analysisAgent, exists := c.Agents["analyst"]
	if !exists {
		return "", fmt.Errorf("analysis agent bulunamadı")
	}

	analysisInput := fmt.Sprintf("Araştırma Bulguları:\n%s\n\nBu bulguları analiz et, ana noktaları çıkar ve içgörüler sun.", researchOutput)
	analysisOutput, err := analysisAgent.Execute(ctx, analysisInput)
	if err != nil {
		return "", fmt.Errorf("analysis agent hatası: %v", err)
	}
	c.Results["analysis"] = analysisOutput
	fmt.Printf("✅ Analiz tamamlandı (%d karakter)\n", len(analysisOutput))

	// 3. Summary Agent - Özet oluştur
	fmt.Println("\n📝 Adım 3: Summary Agent çalışıyor...")
	summaryAgent, exists := c.Agents["summarizer"]
	if !exists {
		return "", fmt.Errorf("summary agent bulunamadı")
	}

	summaryInput := fmt.Sprintf("Konu: %s\n\nAraştırma:\n%s\n\nAnaliz:\n%s\n\nTüm bu bilgileri birleştirerek kapsamlı, yapılandırılmış bir rapor oluştur.",
		topic, researchOutput, analysisOutput)
	finalReport, err := summaryAgent.Execute(ctx, summaryInput)
	if err != nil {
		return "", fmt.Errorf("summary agent hatası: %v", err)
	}
	c.Results["summary"] = finalReport
	fmt.Printf("✅ Özet rapor tamamlandı (%d karakter)\n", len(finalReport))

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("✨ Workflow başarıyla tamamlandı!")

	return finalReport, nil
}

// Close tüm ajanları kapatır
func (c *Coordinator) Close() {
	for _, agent := range c.Agents {
		agent.Close()
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env dosyası bulunamadı, sistem environment variable'ları kullanılacak")
	}

	// API anahtarını çevre değişkeninden al
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		log.Fatal("GOOGLE_API_KEY çevre değişkeni ayarlanmalı")
	}

	ctx := context.Background()

	// Ajanları oluştur
	fmt.Println("🔧 Ajanlar oluşturuluyor...")

	researchAgent, err := NewAgent(
		"researcher",
		"Araştırmacı",
		"Sen bir araştırma uzmanısın. Verilen konular hakkında detaylı, doğru ve güncel bilgiler toplarsın. Bulgularını yapılandırılmış şekilde sunarsın.",
		apiKey,
	)
	if err != nil {
		log.Fatalf("Research agent oluşturulamadı: %v", err)
	}
	defer researchAgent.Close()

	analysisAgent, err := NewAgent(
		"analyst",
		"Analist",
		"Sen bir veri analisti ve strateji uzmanısın. Araştırma verilerini analiz eder, önemli örüntüleri bulur ve içgörüler sunarsın.",
		apiKey,
	)
	if err != nil {
		log.Fatalf("Analysis agent oluşturulamadı: %v", err)
	}
	defer analysisAgent.Close()

	summaryAgent, err := NewAgent(
		"summarizer",
		"Özetleyici",
		"Sen bir teknik yazar ve özetleme uzmanısın. Karmaşık bilgileri anlaşılır, yapılandırılmış ve profesyonel raporlara dönüştürürsün.",
		apiKey,
	)
	if err != nil {
		log.Fatalf("Summary agent oluşturulamadı: %v", err)
	}
	defer summaryAgent.Close()

	// Koordinatörü oluştur ve ajanları ekle
	coordinator := NewCoordinator()
	coordinator.AddAgent(researchAgent)
	coordinator.AddAgent(analysisAgent)
	coordinator.AddAgent(summaryAgent)
	defer coordinator.Close()

	fmt.Println("✅ Tüm ajanlar hazır!")

	// Örnek araştırma konusu
	topic := "Yapay Zeka'nın sağlık sektöründeki güncel uygulamaları ve gelecek potansiyeli"

	// Workflow'u çalıştır
	finalReport, err := coordinator.ExecuteWorkflow(ctx, topic)
	if err != nil {
		log.Fatalf("Workflow hatası: %v", err)
	}

	// Sonucu göster
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("📊 NİHAİ RAPOR")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println(finalReport)
	fmt.Println(strings.Repeat("=", 70))
}
