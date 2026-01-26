// Parallelization Pattern - Ürün İnceleme Analizi (Go)
// Birden fazla görevi eş zamanlı olarak paralel işler

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

// AnalysisTask paralel çalışacak analiz görevlerini temsil eder
type AnalysisTask string

const (
	SentimentAnalysis AnalysisTask = "sentiment"
	KeywordExtraction AnalysisTask = "keywords"
	SummaryGeneration AnalysisTask = "summary"
	CompetitorCompare AnalysisTask = "competitor"
	TrendAnalysis     AnalysisTask = "trends"
)

// TaskResult her görevin sonucunu tutar
type TaskResult struct {
	Task      AnalysisTask
	Result    string
	Duration  time.Duration
	Error     error
	Timestamp time.Time
}

// ProductReview ürün incelemesini temsil eder
type ProductReview struct {
	ProductName string
	Reviews     []string
	Competitor  string
}

// ParallelAnalyzer paralel analiz işlemlerini yönetir
type ParallelAnalyzer struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

// NewParallelAnalyzer yeni bir analyzer instance oluşturur
func NewParallelAnalyzer(ctx context.Context, apiKey string) (*ParallelAnalyzer, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("client oluşturulamadı: %w", err)
	}

	model := client.GenerativeModel(os.Getenv("GEMINI_MODEL"))

	// Paralel istekler için timeout ayarla
	temperature := float32(0.7)
	model.Temperature = &temperature

	return &ParallelAnalyzer{
		client: client,
		model:  model,
	}, nil
}

// Close client'ı kapatır
func (pa *ParallelAnalyzer) Close() error {
	return pa.client.Close()
}

// AnalyzeSentiment duygu analizi yapar
func (pa *ParallelAnalyzer) AnalyzeSentiment(ctx context.Context, reviews []string) (string, error) {
	reviewText := joinReviews(reviews)

	prompt := fmt.Sprintf(`
Aşağıdaki ürün yorumlarının duygu analizini yap:

%s

Lütfen şunları belirt:
1. Genel duygu durumu (Pozitif/Nötr/Negatif) ve yüzde dağılımı
2. En çok bahsedilen pozitif yönler
3. En çok bahsedilen negatif yönler
4. Genel müşteri memnuniyeti skoru (1-10)

Kısa ve öz yanıt ver.
`, reviewText)

	resp, err := pa.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}

	return extractText(resp), nil
}

// ExtractKeywords anahtar kelimeleri çıkarır
func (pa *ParallelAnalyzer) ExtractKeywords(ctx context.Context, reviews []string) (string, error) {
	reviewText := joinReviews(reviews)

	prompt := fmt.Sprintf(`
Aşağıdaki ürün yorumlarından en önemli anahtar kelimeleri ve konuları çıkar:

%s

Lütfen şunları listele:
1. En sık bahsedilen 10 özellik/kelime
2. Müşterilerin en çok önemsediği 5 konu
3. Teknik terimler ve jargon

Sadece liste halinde ver.
`, reviewText)

	resp, err := pa.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}

	return extractText(resp), nil
}

// GenerateSummary özet oluşturur
func (pa *ParallelAnalyzer) GenerateSummary(ctx context.Context, productName string, reviews []string) (string, error) {
	reviewText := joinReviews(reviews)

	prompt := fmt.Sprintf(`
"%s" ürünü için aşağıdaki yorumların kapsamlı bir özetini oluştur:

%s

Özet şunları içermeli:
- Ürünün genel değerlendirmesi (2-3 cümle)
- Ana güçlü yönler
- Ana zayıf yönler
- Satın alma önerisi

Maksimum 150 kelime.
`, productName, reviewText)

	resp, err := pa.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}

	return extractText(resp), nil
}

// CompareWithCompetitor rakip ürün karşılaştırması yapar
func (pa *ParallelAnalyzer) CompareWithCompetitor(ctx context.Context, productName, competitor string, reviews []string) (string, error) {
	reviewText := joinReviews(reviews)

	prompt := fmt.Sprintf(`
"%s" ürününün müşteri yorumlarına dayanarak, "%s" ile karşılaştırma yap:

Yorumlar:
%s

Lütfen şunları belirt:
1. Avantajları (bizim ürünün üstün olduğu noktalar)
2. Dezavantajları (rakibin üstün olduğu noktalar)
3. Fiyat-performans değerlendirmesi
4. Rekabet önerisi

Kısa ve stratejik bir analiz yap.
`, productName, competitor, reviewText)

	resp, err := pa.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}

	return extractText(resp), nil
}

// AnalyzeTrends trend analizi yapar
func (pa *ParallelAnalyzer) AnalyzeTrends(ctx context.Context, reviews []string) (string, error) {
	reviewText := joinReviews(reviews)

	prompt := fmt.Sprintf(`
Aşağıdaki yorumlardaki trendleri ve örüntüleri analiz et:

%s

Lütfen şunları belirt:
1. Ortaya çıkan müşteri beklentileri
2. Değişen kullanım senaryoları
3. Gelecek için öngörüler
4. Ürün geliştirme önerileri

Stratejik bir bakış açısıyla yanıtla.
`, reviewText)

	resp, err := pa.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}

	return extractText(resp), nil
}

// AnalyzeParallel tüm analizleri paralel olarak çalıştırır
func (pa *ParallelAnalyzer) AnalyzeParallel(ctx context.Context, review ProductReview) map[AnalysisTask]TaskResult {
	fmt.Printf("\n🚀 Paralel analiz başlıyor: %s\n", review.ProductName)
	fmt.Printf("📊 %d yorum analiz edilecek\n", len(review.Reviews))
	fmt.Printf("⏱️  Başlangıç zamanı: %s\n\n", time.Now().Format("15:04:05"))

	startTime := time.Now()

	// Results channel ve WaitGroup
	resultsChan := make(chan TaskResult, 5)
	var wg sync.WaitGroup

	// Her analiz görevini paralel olarak başlat
	tasks := []struct {
		task AnalysisTask
		fn   func() (string, error)
	}{
		{
			task: SentimentAnalysis,
			fn: func() (string, error) {
				return pa.AnalyzeSentiment(ctx, review.Reviews)
			},
		},
		{
			task: KeywordExtraction,
			fn: func() (string, error) {
				return pa.ExtractKeywords(ctx, review.Reviews)
			},
		},
		{
			task: SummaryGeneration,
			fn: func() (string, error) {
				return pa.GenerateSummary(ctx, review.ProductName, review.Reviews)
			},
		},
		{
			task: CompetitorCompare,
			fn: func() (string, error) {
				return pa.CompareWithCompetitor(ctx, review.ProductName, review.Competitor, review.Reviews)
			},
		},
		{
			task: TrendAnalysis,
			fn: func() (string, error) {
				return pa.AnalyzeTrends(ctx, review.Reviews)
			},
		},
	}

	// Her görevi ayrı bir goroutine'de çalıştır
	for _, t := range tasks {
		wg.Add(1)
		go func(task AnalysisTask, taskFn func() (string, error)) {
			defer wg.Done()

			taskStart := time.Now()
			fmt.Printf("🔄 %s başladı...\n", getTaskName(task))

			result, err := taskFn()
			duration := time.Since(taskStart)

			resultsChan <- TaskResult{
				Task:      task,
				Result:    result,
				Duration:  duration,
				Error:     err,
				Timestamp: time.Now(),
			}

			if err != nil {
				fmt.Printf("❌ %s HATA: %v (%.2fs)\n", getTaskName(task), err, duration.Seconds())
			} else {
				fmt.Printf("✅ %s tamamlandı (%.2fs)\n", getTaskName(task), duration.Seconds())
			}
		}(t.task, t.fn)
	}

	// Goroutine'lerin bitmesini bekle ve channel'ı kapat
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Sonuçları topla
	results := make(map[AnalysisTask]TaskResult)
	for result := range resultsChan {
		results[result.Task] = result
	}

	totalDuration := time.Since(startTime)
	fmt.Printf("\n⏱️  Toplam süre: %.2f saniye\n", totalDuration.Seconds())
	fmt.Printf("✨ Tüm analizler tamamlandı!\n")
	fmt.Println(strings.Repeat("=", 70))

	return results
}

// AnalyzeSequential tüm analizleri sıralı olarak çalıştırır (karşılaştırma için)
func (pa *ParallelAnalyzer) AnalyzeSequential(ctx context.Context, review ProductReview) map[AnalysisTask]TaskResult {
	fmt.Printf("\n🐌 Sıralı analiz başlıyor: %s\n", review.ProductName)
	fmt.Printf("📊 %d yorum analiz edilecek\n", len(review.Reviews))
	fmt.Printf("⏱️  Başlangıç zamanı: %s\n\n", time.Now().Format("15:04:05"))

	startTime := time.Now()
	results := make(map[AnalysisTask]TaskResult)

	// Sıralı olarak çalıştır
	tasks := []struct {
		task AnalysisTask
		fn   func() (string, error)
	}{
		{SentimentAnalysis, func() (string, error) { return pa.AnalyzeSentiment(ctx, review.Reviews) }},
		{KeywordExtraction, func() (string, error) { return pa.ExtractKeywords(ctx, review.Reviews) }},
		{SummaryGeneration, func() (string, error) { return pa.GenerateSummary(ctx, review.ProductName, review.Reviews) }},
		{CompetitorCompare, func() (string, error) {
			return pa.CompareWithCompetitor(ctx, review.ProductName, review.Competitor, review.Reviews)
		}},
		{TrendAnalysis, func() (string, error) { return pa.AnalyzeTrends(ctx, review.Reviews) }},
	}

	for _, t := range tasks {
		taskStart := time.Now()
		fmt.Printf("🔄 %s başladı...\n", getTaskName(t.task))

		result, err := t.fn()
		duration := time.Since(taskStart)

		results[t.task] = TaskResult{
			Task:      t.task,
			Result:    result,
			Duration:  duration,
			Error:     err,
			Timestamp: time.Now(),
		}

		if err != nil {
			fmt.Printf("❌ %s HATA: %v (%.2fs)\n", getTaskName(t.task), err, duration.Seconds())
		} else {
			fmt.Printf("✅ %s tamamlandı (%.2fs)\n", getTaskName(t.task), duration.Seconds())
		}
	}

	totalDuration := time.Since(startTime)
	fmt.Printf("\n⏱️  Toplam süre: %.2f saniye\n", totalDuration.Seconds())
	fmt.Printf("✨ Tüm analizler tamamlandı!\n")
	fmt.Println(strings.Repeat("=", 70))

	return results
}

// DisplayResults sonuçları formatlı şekilde gösterir
func DisplayResults(results map[AnalysisTask]TaskResult) {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("📋 ANALİZ SONUÇLARI")
	fmt.Println(strings.Repeat("=", 70))

	tasks := []AnalysisTask{SentimentAnalysis, KeywordExtraction, SummaryGeneration, CompetitorCompare, TrendAnalysis}

	for _, task := range tasks {
		result, ok := results[task]
		if !ok {
			continue
		}

		fmt.Printf("\n🔹 %s\n", getTaskName(task))
		fmt.Println(strings.Repeat("-", 70))

		if result.Error != nil {
			fmt.Printf("❌ HATA: %v\n", result.Error)
		} else {
			fmt.Printf("%s\n", result.Result)
		}

		fmt.Printf("\n⏱️  Süre: %.2f saniye\n", result.Duration.Seconds())
	}

	fmt.Println("\n" + strings.Repeat("=", 70))
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

func joinReviews(reviews []string) string {
	result := ""
	for i, review := range reviews {
		result += fmt.Sprintf("Yorum %d: %s\n", i+1, review)
	}
	return result
}

func getTaskName(task AnalysisTask) string {
	names := map[AnalysisTask]string{
		SentimentAnalysis: "Duygu Analizi",
		KeywordExtraction: "Anahtar Kelime Çıkarımı",
		SummaryGeneration: "Özet Oluşturma",
		CompetitorCompare: "Rakip Karşılaştırma",
		TrendAnalysis:     "Trend Analizi",
	}
	return names[task]
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

	// Analyzer oluştur
	analyzer, err := NewParallelAnalyzer(ctx, apiKey)
	if err != nil {
		log.Fatalf("Analyzer oluşturulamadı: %v", err)
	}
	defer analyzer.Close()

	// Örnek ürün incelemeleri
	review := ProductReview{
		ProductName: "TechPro X1 Kablosuz Kulaklık",
		Competitor:  "SoundMax Pro",
		Reviews: []string{
			"Ses kalitesi harika! Bass performansı çok iyi. Ancak batarya ömrü biraz kısa.",
			"Rahat kullanım, ergonomik tasarım. Gürültü engelleme özelliği ortalama seviyede.",
			"Fiyatına göre çok iyi bir ürün. Bluetooth bağlantısı stabil. Mikrofondan memnun kalmadım.",
			"Şarj süresi çok uzun. Ses kalitesi için ödediğim paraya değer. Kullanım kılavuzu yetersiz.",
			"Aktif gürültü engelleme beklediğim kadar etkili değil. Ses kalitesi mükemmel.",
			"Kullanışlı taşıma kutusu. Batarya göstergesi çok kullanışlı. Çağrı kalitesi iyi.",
			"Spor yaparken kullanmak için ideal. Su geçirmez özellik harika. Fiyat biraz yüksek.",
		},
	}

	// PARALEL ANALİZ
	parallelResults := analyzer.AnalyzeParallel(ctx, review)
	DisplayResults(parallelResults)

	fmt.Println("\n\n" + strings.Repeat("█", 70))
	fmt.Println("KARŞILAŞTIRMA İÇİN SIRALI ANALİZ BAŞLIYOR")
	fmt.Println(strings.Repeat("█", 70))

	// SIRALI ANALİZ (Karşılaştırma için)
	sequentialResults := analyzer.AnalyzeSequential(ctx, review)
	DisplayResults(sequentialResults)

	// Performans karşılaştırması
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("⚡ PERFORMANS KARŞILAŞTIRMASI")
	fmt.Println(strings.Repeat("=", 70))

	var parallelTotal, sequentialTotal time.Duration
	for _, result := range parallelResults {
		parallelTotal += result.Duration
	}
	for _, result := range sequentialResults {
		sequentialTotal += result.Duration
	}

	// Paralel işlemde en uzun görev toplam süreyi belirler
	var parallelMax time.Duration
	for _, result := range parallelResults {
		if result.Duration > parallelMax {
			parallelMax = result.Duration
		}
	}

	fmt.Printf("\n🔄 Paralel Analiz (Gerçek Süre): %.2f saniye\n", parallelMax.Seconds())
	fmt.Printf("🐌 Sıralı Analiz (Gerçek Süre): %.2f saniye\n", sequentialTotal.Seconds())
	speedup := sequentialTotal.Seconds() / parallelMax.Seconds()
	fmt.Printf("⚡ Hız Kazancı: %.2fx daha hızlı\n", speedup)
	fmt.Printf("💾 Zaman Tasarrufu: %.2f saniye\n\n", sequentialTotal.Seconds()-parallelMax.Seconds())
}
