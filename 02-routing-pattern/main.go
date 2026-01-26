// Routing Pattern - Müşteri Destek Sistemi (Go)
// Gelen sorguları otomatik olarak doğru departmana yönlendirir

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

// Department destek departmanlarını temsil eder
type Department string

const (
	Technical Department = "technical"
	Billing   Department = "billing"
	Sales     Department = "sales"
	General   Department = "general"
	Complaint Department = "complaint"
	Unknown   Department = "unknown"
)

// RoutingDecision yönlendirme kararını tutar
type RoutingDecision struct {
	Department Department
	Confidence string
	Reasoning  string
	Priority   string
}

// CustomerQuery müşteri sorgusunu temsil eder
type CustomerQuery struct {
	ID      string
	Content string
	User    string
}

// QueryRouter sorguları ilgili departmanlara yönlendirir
type QueryRouter struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

// NewQueryRouter yeni bir router instance oluşturur
func NewQueryRouter(ctx context.Context, apiKey string) (*QueryRouter, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("client oluşturulamadı: %w", err)
	}

	model := client.GenerativeModel(os.Getenv("GEMINI_MODEL"))

	// Temperature düşük tutuyoruz çünkü tutarlı routing istiyoruz
	temperature := float32(0.1)
	model.Temperature = &temperature

	return &QueryRouter{
		client: client,
		model:  model,
	}, nil
}

// Close client'ı kapatır
func (r *QueryRouter) Close() error {
	return r.client.Close()
}

// RouteQuery sorguyu analiz edip doğru departmana yönlendirir
func (r *QueryRouter) RouteQuery(ctx context.Context, query CustomerQuery) (*RoutingDecision, error) {
	prompt := fmt.Sprintf(`
Sen bir müşteri destek yönlendirme asistanısın. Gelen sorguları analiz edip doğru departmana yönlendiriyorsun.

Müşteri Sorgusu:
%s

Mevcut Departmanlar:
1. TECHNICAL - Teknik sorunlar, hata mesajları, sistem çalışmıyor, bağlantı problemleri
2. BILLING - Fatura, ödeme, fiyatlandırma, ücret iadesi, abonelik değişiklikleri
3. SALES - Yeni ürün satın alma, ürün özellikleri, paket karşılaştırma
4. COMPLAINT - Şikayet, memnuniyetsizlik, kötü deneyim, hizmet kalitesi
5. GENERAL - Genel bilgi, kullanım kılavuzu, hesap ayarları

Lütfen aşağıdaki formatı AYNEN kullanarak cevap ver:
DEPARTMENT: [technical/billing/sales/complaint/general]
CONFIDENCE: [high/medium/low]
PRIORITY: [urgent/high/normal/low]
REASONING: [Kısa açıklama]

Örnek:
DEPARTMENT: technical
CONFIDENCE: high
PRIORITY: urgent
REASONING: Kullanıcı sisteme giriş yapamıyor ve hata mesajı alıyor.
`, query.Content)

	resp, err := r.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("routing analizi başarısız: %w", err)
	}

	result := extractText(resp)
	decision := parseRoutingDecision(result)

	return decision, nil
}

// HandleTechnical teknik destek işlemlerini yapar
func (r *QueryRouter) HandleTechnical(ctx context.Context, query CustomerQuery) (string, error) {
	prompt := fmt.Sprintf(`
Sen bir teknik destek uzmanısın. Aşağıdaki teknik sorunu analiz et ve çözüm öner:

Sorun: %s

Lütfen şunları sağla:
1. Sorunun muhtemel nedenleri
2. Adım adım çözüm önerileri
3. Alternatif çözümler
4. Gerekirse yükseltme önerisi

Profesyonel ama samimi bir dille yanıt ver.
`, query.Content)

	resp, err := r.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("teknik yanıt oluşturulamadı: %w", err)
	}

	return extractText(resp), nil
}

// HandleBilling fatura ve ödeme işlemlerini yapar
func (r *QueryRouter) HandleBilling(ctx context.Context, query CustomerQuery) (string, error) {
	prompt := fmt.Sprintf(`
Sen bir faturalama uzmanısın. Aşağıdaki fatura/ödeme sorusunu yanıtla:

Soru: %s

Lütfen şunları içer:
1. Açık ve net açıklama
2. Gerekli adımlar
3. İlgili politikalar
4. İletişim bilgileri (gerekirse)

Profesyonel ve güvenilir bir dille yanıt ver.
`, query.Content)

	resp, err := r.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("fatura yanıtı oluşturulamadı: %w", err)
	}

	return extractText(resp), nil
}

// HandleSales satış sorgularını işler
func (r *QueryRouter) HandleSales(ctx context.Context, query CustomerQuery) (string, error) {
	prompt := fmt.Sprintf(`
Sen bir satış danışmanısın. Aşağıdaki satış sorusunu yanıtla:

Soru: %s

Lütfen şunları içer:
1. Ürün/hizmet özellikleri
2. Fiyatlandırma bilgisi
3. Müşteriye değer önerisi
4. Sonraki adımlar

Dostça ve ikna edici bir dille yanıt ver.
`, query.Content)

	resp, err := r.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("satış yanıtı oluşturulamadı: %w", err)
	}

	return extractText(resp), nil
}

// HandleComplaint şikayet yönetimi yapar
func (r *QueryRouter) HandleComplaint(ctx context.Context, query CustomerQuery) (string, error) {
	prompt := fmt.Sprintf(`
Sen bir müşteri ilişkileri uzmanısın. Aşağıdaki şikayeti empatik bir şekilde ele al:

Şikayet: %s

Lütfen şunları içer:
1. Empatik bir açılış
2. Sorunu kabul etme
3. Çözüm önerisi veya telafi
4. Takip süreci

Samimi, anlayışlı ve çözüm odaklı bir dille yanıt ver.
`, query.Content)

	resp, err := r.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("şikayet yanıtı oluşturulamadı: %w", err)
	}

	return extractText(resp), nil
}

// HandleGeneral genel sorguları işler
func (r *QueryRouter) HandleGeneral(ctx context.Context, query CustomerQuery) (string, error) {
	prompt := fmt.Sprintf(`
Sen bir genel destek asistanısın. Aşağıdaki genel soruyu yanıtla:

Soru: %s

Lütfen net, yararlı ve kullanıcı dostu bir yanıt ver.
`, query.Content)

	resp, err := r.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("genel yanıt oluşturulamadı: %w", err)
	}

	return extractText(resp), nil
}

// ProcessQuery tam routing ve işlem sürecini gerçekleştirir
func (r *QueryRouter) ProcessQuery(ctx context.Context, query CustomerQuery) error {
	fmt.Printf("\n📨 Yeni Sorgu (#%s) - Kullanıcı: %s\n", query.ID, query.User)
	fmt.Printf("İçerik: %s\n\n", query.Content)

	// 1. Route edilecek departmanı belirle
	fmt.Println("🔍 Sorgu analiz ediliyor...")
	decision, err := r.RouteQuery(ctx, query)
	if err != nil {
		return fmt.Errorf("routing hatası: %w", err)
	}

	// Routing kararını göster
	fmt.Printf("\n📍 Routing Kararı:\n")
	fmt.Printf("   Departman: %s\n", strings.ToUpper(string(decision.Department)))
	fmt.Printf("   Güven: %s\n", decision.Confidence)
	fmt.Printf("   Öncelik: %s\n", decision.Priority)
	fmt.Printf("   Sebep: %s\n", decision.Reasoning)

	fmt.Println("\n" + strings.Repeat("─", 70))

	// 2. İlgili departman handler'ını çağır
	var response string

	fmt.Printf("\n🔄 %s departmanına yönlendiriliyor...\n\n", strings.ToUpper(string(decision.Department)))

	switch decision.Department {
	case Technical:
		response, err = r.HandleTechnical(ctx, query)
	case Billing:
		response, err = r.HandleBilling(ctx, query)
	case Sales:
		response, err = r.HandleSales(ctx, query)
	case Complaint:
		response, err = r.HandleComplaint(ctx, query)
	case General:
		response, err = r.HandleGeneral(ctx, query)
	default:
		response, err = r.HandleGeneral(ctx, query)
	}

	if err != nil {
		return fmt.Errorf("handler hatası: %w", err)
	}

	// 3. Yanıtı göster
	fmt.Printf("✅ Yanıt:\n%s\n", response)
	fmt.Println("\n" + strings.Repeat("=", 60))

	return nil
}

// parseRoutingDecision AI yanıtını parse eder
func parseRoutingDecision(response string) *RoutingDecision {
	decision := &RoutingDecision{
		Department: Unknown,
		Confidence: "medium",
		Priority:   "normal",
		Reasoning:  "Analiz tamamlandı",
	}

	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "DEPARTMENT:") {
			dept := strings.TrimSpace(strings.TrimPrefix(line, "DEPARTMENT:"))
			decision.Department = Department(strings.ToLower(dept))
		} else if strings.HasPrefix(line, "CONFIDENCE:") {
			decision.Confidence = strings.TrimSpace(strings.TrimPrefix(line, "CONFIDENCE:"))
		} else if strings.HasPrefix(line, "PRIORITY:") {
			decision.Priority = strings.TrimSpace(strings.TrimPrefix(line, "PRIORITY:"))
		} else if strings.HasPrefix(line, "REASONING:") {
			decision.Reasoning = strings.TrimSpace(strings.TrimPrefix(line, "REASONING:"))
		}
	}

	return decision
}

// extractText response'dan text içeriği çıkarır
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

	// Router oluştur
	router, err := NewQueryRouter(ctx, apiKey)
	if err != nil {
		log.Fatalf("Router oluşturulamadı: %v", err)
	}
	defer router.Close()

	// Test sorguları
	queries := []CustomerQuery{
		{
			ID:      "001",
			User:    "ahmet.yilmaz@email.com",
			Content: "Sistemde sürekli 'Connection timeout' hatası alıyorum. Giriş yapamıyorum.",
		},
		{
			ID:      "002",
			User:    "ayse.kaya@email.com",
			Content: "Geçen ay iki kez fatura kesilmiş. Fazla ödeme iadesi nasıl alırım?",
		},
		{
			ID:      "003",
			User:    "mehmet.demir@email.com",
			Content: "Enterprise paketi ile Professional paket arasındaki farklar nelerdir?",
		},
		{
			ID:      "004",
			User:    "zeynep.ozturk@email.com",
			Content: "3 gündür destek talebime yanıt alamadım. Çok memnuniyetsizim!",
		},
		{
			ID:      "005",
			User:    "can.arslan@email.com",
			Content: "Hesap ayarlarımı nasıl değiştirebilirim?",
		},
	}

	// Her sorguyu işle
	for _, query := range queries {
		if err := router.ProcessQuery(ctx, query); err != nil {
			log.Printf("Sorgu işleme hatası (#%s): %v", query.ID, err)
		}
		fmt.Println()
	}
}
