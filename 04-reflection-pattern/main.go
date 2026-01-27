// Reflection Pattern - Kod İnceleme ve İyileştirme Sistemi (Go)
// AI'ın kendi çıktısını değerlendirip iteratif olarak iyileştirmesi

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

// ReflectionResult her iterasyonun sonucunu tutar
type ReflectionResult struct {
	Iteration    int
	Output       string
	Critique     string
	Score        int // 1-10 arası
	Issues       []string
	Improvements []string
	Duration     time.Duration
	Timestamp    time.Time
}

// CodeReviewRequest kod review isteğini temsil eder
type CodeReviewRequest struct {
	Code          string
	Language      string
	Purpose       string
	MaxIterations int
	TargetScore   int // Hedef kalite skoru
}

// ReflectiveAgent self-reflection yapabilen agent
type ReflectiveAgent struct {
	client  *genai.Client
	model   *genai.GenerativeModel
	history []ReflectionResult
}

// NewReflectiveAgent yeni bir reflective agent oluşturur
func NewReflectiveAgent(ctx context.Context, apiKey string) (*ReflectiveAgent, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("client oluşturulamadı: %w", err)
	}

	model := client.GenerativeModel(os.Getenv("GEMINI_MODEL"))

	// Reflection için daha düşük temperature
	temperature := float32(0.3)
	model.Temperature = &temperature

	return &ReflectiveAgent{
		client:  client,
		model:   model,
		history: []ReflectionResult{},
	}, nil
}

// Close client'ı kapatır
func (ra *ReflectiveAgent) Close() error {
	return ra.client.Close()
}

// GenerateCode kod üretir
func (ra *ReflectiveAgent) GenerateCode(ctx context.Context, request CodeReviewRequest) (string, error) {
	prompt := fmt.Sprintf(`
Sen uzman bir %s geliştiricisisin. Aşağıdaki amaca uygun, temiz ve profesyonel bir kod yaz:

Amaç: %s

Lütfen şunlara dikkat et:
- Clean code prensipleri
- Best practices
- Hata yönetimi
- Kod yorumları
- Okunabilirlik
- Performans

Sadece kodu yaz, açıklama ekleme.
`, request.Language, request.Purpose)

	resp, err := ra.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}

	return extractText(resp), nil
}

// ReflectOnCode kodu analiz eder ve eleştirir
func (ra *ReflectiveAgent) ReflectOnCode(ctx context.Context, code string, language string) (string, int, []string, error) {
	prompt := fmt.Sprintf(`
Sen deneyimli bir kod reviewer'sın. Aşağıdaki %s kodunu kritik bir gözle incele:

%s

Lütfen aşağıdaki formatı AYNEN kullanarak analiz yap:

SCORE: [1-10 arası sayı]
ISSUES:
- [Sorun 1]
- [Sorun 2]
- [Sorun 3]

IMPROVEMENTS:
- [İyileştirme önerisi 1]
- [İyileştirme önerisi 2]
- [İyileştirme önerisi 3]

CRITIQUE:
[Detaylı değerlendirme - 2-3 paragraf]

Değerlendirme Kriterleri:
- Kod kalitesi ve okunabilirlik (2 puan)
- Best practices'e uyum (2 puan)
- Hata yönetimi (2 puan)
- Performans (2 puan)
- Güvenlik (2 puan)

Sert ve objektif ol. Gerçek problemlere odaklan.
`, language, code)

	resp, err := ra.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", 0, nil, err
	}

	result := extractText(resp)
	critique, score, issues := parseReflectionResult(result)

	return critique, score, issues, nil
}

// ImproveCode kodu iyileştirir
func (ra *ReflectiveAgent) ImproveCode(ctx context.Context, code string, critique string, issues []string, language string) (string, error) {
	issuesText := strings.Join(issues, "\n- ")

	prompt := fmt.Sprintf(`
Aşağıdaki %s kodunu, verilen eleştiri ve sorunlara göre iyileştir:

MEVCUT KOD:
%s

ELEŞTİRİ:
%s

SORUNLAR:
- %s

Lütfen kodu şu şekilde iyileştir:
1. Belirtilen tüm sorunları çöz
2. Best practices'i uygula
3. Kod kalitesini artır
4. Hata yönetimini geliştir
5. Performansı optimize et

Sadece iyileştirilmiş kodu ver, açıklama ekleme.
`, language, code, critique, issuesText)

	resp, err := ra.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}

	return extractText(resp), nil
}

// ReflectAndImprove reflection pattern ile kodu iteratif olarak iyileştirir
func (ra *ReflectiveAgent) ReflectAndImprove(ctx context.Context, request CodeReviewRequest) (*ReflectionResult, error) {
	fmt.Printf("\n🤖 Reflection Pattern ile Kod İyileştirme Başlıyor\n")
	fmt.Printf("🎯 Hedef: %s (%s)\n", request.Purpose, request.Language)
	fmt.Printf("🔄 Maksimum İterasyon: %d\n", request.MaxIterations)
	fmt.Printf("⭐ Hedef Skor: %d/10\n", request.TargetScore)
	fmt.Println(strings.Repeat("=", 70))

	// İlk kod üretimi
	fmt.Println("\n📝 İlk kod üretiliyor...")
	startTime := time.Now()

	currentCode, err := ra.GenerateCode(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("ilk kod üretimi başarısız: %w", err)
	}

	fmt.Println("✅ İlk kod üretildi")
	fmt.Printf("\n📄 BAŞLANGIÇ KODU:\n%s\n%s\n%s\n",
		strings.Repeat("-", 70), currentCode, strings.Repeat("-", 70))

	// Reflection döngüsü
	for iteration := 1; iteration <= request.MaxIterations; iteration++ {
		iterationStart := time.Now()

		fmt.Printf("\n🔍 İTERASYON %d\n", iteration)
		fmt.Println()

		// 1. Reflection: Kodu değerlendir
		fmt.Println("🔎 Kod analiz ediliyor...")
		critique, score, issues, err := ra.ReflectOnCode(ctx, currentCode, request.Language)
		if err != nil {
			return nil, fmt.Errorf("reflection hatası (iterasyon %d): %w", iteration, err)
		}

		// Sonuçları kaydet
		result := ReflectionResult{
			Iteration: iteration,
			Output:    currentCode,
			Critique:  critique,
			Score:     score,
			Issues:    issues,
			Duration:  time.Since(iterationStart),
			Timestamp: time.Now(),
		}
		ra.history = append(ra.history, result)

		// Sonuçları göster
		fmt.Printf("\n📊 Analiz Sonuçları:\n")
		fmt.Printf("   Skor: %d/10 ", score)
		if score >= request.TargetScore {
			fmt.Printf("✨ (Hedef skor ulaşıldı!)\n")
		} else {
			fmt.Printf("(Hedef: %d/10)\n", request.TargetScore)
		}

		if len(issues) > 0 {
			fmt.Printf("   Sorunlar (%d adet):\n", len(issues))
			for i, issue := range issues {
				fmt.Printf("      %d. %s\n", i+1, issue)
			}
		}

		fmt.Printf("\n💭 Eleştiri:\n%s\n", critique)
		fmt.Printf("\n⏱️  İterasyon süresi: %.2f saniye\n", result.Duration.Seconds())

		// Hedef skora ulaşıldı mı?
		if score >= request.TargetScore {
			fmt.Printf("\n🎉 Başarı! Hedef kalite skoruna ulaşıldı! (%d/10)\n", score)
			fmt.Printf("⏱️  Toplam süre: %.2f saniye\n", time.Since(startTime).Seconds())
			fmt.Printf("🔄 Toplam iterasyon: %d\n", iteration)
			return &result, nil
		}

		// Son iterasyonda iyileştirme yapma
		if iteration == request.MaxIterations {
			fmt.Printf("\n⚠️  Maksimum iterasyona ulaşıldı (%d)\n", request.MaxIterations)
			fmt.Printf("📊 Final skor: %d/10 (Hedef: %d/10)\n", score, request.TargetScore)
			fmt.Printf("⏱️  Toplam süre: %.2f saniye\n", time.Since(startTime).Seconds())
			return &result, nil
		}

		// 2. Improvement: Kodu iyileştir
		fmt.Printf("\n🔧 Kod iyileştiriliyor (İterasyon %d → %d)...\n", iteration, iteration+1)
		improvedCode, err := ra.ImproveCode(ctx, currentCode, critique, issues, request.Language)
		if err != nil {
			return nil, fmt.Errorf("iyileştirme hatası (iterasyon %d): %w", iteration, err)
		}

		currentCode = improvedCode
		fmt.Println("✅ Kod iyileştirildi, bir sonraki iterasyona geçiliyor...")
	}

	// Son sonucu döndür
	if len(ra.history) > 0 {
		return &ra.history[len(ra.history)-1], nil
	}

	return nil, fmt.Errorf("beklenmedik durum: hiç sonuç üretilmedi")
}

// DisplayHistory tüm iterasyon geçmişini gösterir
func (ra *ReflectiveAgent) DisplayHistory() {
	if len(ra.history) == 0 {
		fmt.Println("⚠️  Henüz iterasyon geçmişi yok")
		return
	}

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("📚 İTERASYON GEÇMİŞİ")
	fmt.Println()

	for _, result := range ra.history {
		fmt.Printf("\n🔄 İterasyon %d\n", result.Iteration)
		fmt.Println(strings.Repeat("-", 70))
		fmt.Printf("⭐ Skor: %d/10\n", result.Score)
		fmt.Printf("⏱️  Süre: %.2f saniye\n", result.Duration.Seconds())
		fmt.Printf("🕐 Zaman: %s\n", result.Timestamp.Format("15:04:05"))

		if len(result.Issues) > 0 {
			fmt.Printf("❌ Sorunlar: %d adet\n", len(result.Issues))
		}
	}

	// İlerleme grafiği
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("📈 SKOR İLERLEMESİ")
	fmt.Println()

	for _, result := range ra.history {
		bar := strings.Repeat("█", result.Score)
		empty := strings.Repeat("░", result.Score)
		fmt.Printf("İterasyon %d: %s%s %d/10\n", result.Iteration, bar, empty, result.Score)
	}

	// İstatistikler
	if len(ra.history) > 0 {
		firstScore := ra.history[0].Score
		lastScore := ra.history[len(ra.history)-1].Score
		improvement := lastScore - firstScore

		fmt.Println("\n" + strings.Repeat("=", 70))
		fmt.Println("📊 İSTATİSTİKLER")
		fmt.Println()
		fmt.Printf("Başlangıç Skoru: %d/10\n", firstScore)
		fmt.Printf("Final Skoru: %d/10\n", lastScore)

		if improvement > 0 {
			fmt.Printf("İyileşme: +%d puan ✨\n", improvement)
		} else if improvement < 0 {
			fmt.Printf("Değişim: %d puan ⚠️\n", improvement)
		} else {
			fmt.Printf("Değişim: Sabit\n")
		}

		var totalDuration time.Duration
		for _, result := range ra.history {
			totalDuration += result.Duration
		}
		fmt.Printf("Toplam Süre: %.2f saniye\n", totalDuration.Seconds())
		fmt.Printf("Ortalama İterasyon: %.2f saniye\n", totalDuration.Seconds()/float64(len(ra.history)))
	}
}

// parseReflectionResult AI yanıtını parse eder
func parseReflectionResult(response string) (string, int, []string) {
	score := 5 // varsayılan
	var issues []string
	var critique string

	lines := strings.Split(response, "\n")
	inIssues := false
	inCritique := false
	var critiqueLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "SCORE:") {
			scoreStr := strings.TrimSpace(strings.TrimPrefix(line, "SCORE:"))
			fmt.Sscanf(scoreStr, "%d", &score)
		} else if strings.HasPrefix(line, "ISSUES:") {
			inIssues = true
			inCritique = false
		} else if strings.HasPrefix(line, "IMPROVEMENTS:") || strings.HasPrefix(line, "CRITIQUE:") {
			inIssues = false
			if strings.HasPrefix(line, "CRITIQUE:") {
				inCritique = true
			}
		} else if inIssues && strings.HasPrefix(line, "-") {
			issue := strings.TrimPrefix(line, "-")
			issue = strings.TrimSpace(issue)
			if issue != "" {
				issues = append(issues, issue)
			}
		} else if inCritique && line != "" {
			critiqueLines = append(critiqueLines, line)
		}
	}

	critique = strings.Join(critiqueLines, " ")
	if critique == "" {
		critique = "Analiz tamamlandı"
	}

	return critique, score, issues
}

// Yardımcı fonksiyonlar

func extractText(resp *genai.GenerateContentResponse) string {
	if len(resp.Candidates) == 0 {
		return ""
	}

	var result string
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			result += string(text)
		}
	}
	return result
}

func main() {
	// .env dosyasını yükle
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env dosyası bulunamadı")
	}

	// API anahtarını al
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		log.Fatal("GOOGLE_API_KEY environment variable tanımlanmamış")
	}

	ctx := context.Background()

	// Reflective agent oluştur
	agent, err := NewReflectiveAgent(ctx, apiKey)
	if err != nil {
		log.Fatalf("Agent oluşturulamadı: %v", err)
	}
	defer agent.Close()

	// Kod review isteği
	request := CodeReviewRequest{
		Code:          "", // Yeni kod üretilecek
		Language:      "Go",
		Purpose:       "Bir HTTP API endpoint'i yaz. User verilerini JSON olarak döndürsün. Database'den veri çeksin, error handling ve logging olsun.",
		MaxIterations: 3,
		TargetScore:   8, // 8/10 hedef kalite
	}

	// Reflection pattern ile iyileştirme
	finalResult, err := agent.ReflectAndImprove(ctx, request)
	if err != nil {
		log.Fatalf("Reflection süreci başarısız: %v", err)
	}

	// Final kodu göster
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("🎯 FİNAL KOD")
	fmt.Println()
	fmt.Printf("\n%s\n", finalResult.Output)

	// Geçmişi göster
	agent.DisplayHistory()

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("✨ Reflection Pattern Tamamlandı!")
	fmt.Println()
}
