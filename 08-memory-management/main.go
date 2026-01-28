// Memory Management Pattern - Temel Bellek Yönetimi (Go)
// Müşteri etkileşimlerini saklayarak bağlam farkındalığı sağlayan sistem

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

// CustomerMemory müşteri hafıza yapısı
type CustomerMemory struct {
	CustomerID       string                 `json:"customer_id"`
	Name             string                 `json:"name"`
	Preferences      map[string]interface{} `json:"preferences"`
	PastInteractions []Interaction          `json:"past_interactions"`
	LastContactDate  time.Time              `json:"last_contact_date"`
}

// Interaction geçmiş etkileşim yapısı
type Interaction struct {
	Timestamp  time.Time `json:"timestamp"`
	Issue      string    `json:"issue"`
	Resolution string    `json:"resolution"`
	Sentiment  string    `json:"sentiment"`
}

// MemoryManager bellek yönetim sistemi
type MemoryManager struct {
	memories map[string]*CustomerMemory
	client   *genai.Client
	model    *genai.GenerativeModel
}

// NewMemoryManager yeni bellek yöneticisi oluşturur
func NewMemoryManager(apiKey string) (*MemoryManager, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("client oluşturulamadı: %v", err)
	}

	model := client.GenerativeModel(os.Getenv("GEMINI_MODEL"))
	model.SetTemperature(0.7)
	model.SetTopP(0.95)
	model.SetMaxOutputTokens(1024)

	return &MemoryManager{
		memories: make(map[string]*CustomerMemory),
		client:   client,
		model:    model,
	}, nil
}

// SaveMemory müşteri hafızasını kaydeder
func (mm *MemoryManager) SaveMemory(customerID string, memory *CustomerMemory) {
	mm.memories[customerID] = memory
	memory.LastContactDate = time.Now()
}

// GetMemory müşteri hafızasını getirir
func (mm *MemoryManager) GetMemory(customerID string) *CustomerMemory {
	return mm.memories[customerID]
}

// AddInteraction yeni etkileşim ekler
func (mm *MemoryManager) AddInteraction(customerID string, issue, resolution, sentiment string) {
	memory := mm.GetMemory(customerID)
	if memory == nil {
		memory = &CustomerMemory{
			CustomerID:       customerID,
			Preferences:      make(map[string]interface{}),
			PastInteractions: []Interaction{},
		}
	}

	interaction := Interaction{
		Timestamp:  time.Now(),
		Issue:      issue,
		Resolution: resolution,
		Sentiment:  sentiment,
	}

	memory.PastInteractions = append(memory.PastInteractions, interaction)
	mm.SaveMemory(customerID, memory)
}

// BuildContextPrompt bağlam içeren prompt oluşturur
func (mm *MemoryManager) BuildContextPrompt(customerID, currentQuery string) string {
	memory := mm.GetMemory(customerID)
	if memory == nil {
		return fmt.Sprintf("Müşteri Sorusu: %s", currentQuery)
	}

	contextPrompt := fmt.Sprintf(`Sen bir müşteri hizmetleri asistanısın. Aşağıda müşteri hakkında bilgiler var:

Müşteri: %s (ID: %s)
Son İletişim: %s

Geçmiş Etkileşimler:
`, memory.Name, memory.CustomerID, memory.LastContactDate.Format("2006-01-02 15:04"))

	// Son 3 etkileşimi ekle
	start := len(memory.PastInteractions) - 3
	if start < 0 {
		start = 0
	}

	for _, interaction := range memory.PastInteractions[start:] {
		contextPrompt += fmt.Sprintf("- [%s] Sorun: %s | Çözüm: %s | Duygu: %s\n",
			interaction.Timestamp.Format("2006-01-02"),
			interaction.Issue,
			interaction.Resolution,
			interaction.Sentiment)
	}

	if len(memory.Preferences) > 0 {
		prefsJSON, _ := json.MarshalIndent(memory.Preferences, "", "  ")
		contextPrompt += fmt.Sprintf("\nMüşteri Tercihleri:\n%s\n", string(prefsJSON))
	}

	contextPrompt += fmt.Sprintf("\nGüncel Soru: %s\n\nLütfen müşterinin geçmişini dikkate alarak yardımcı ol.", currentQuery)

	return contextPrompt
}

// ProcessQuery sorguyu bellek ile işler
func (mm *MemoryManager) ProcessQuery(ctx context.Context, customerID, query string) (string, error) {
	prompt := mm.BuildContextPrompt(customerID, query)

	resp, err := mm.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("içerik üretilemedi: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("yanıt alınamadı")
	}

	response := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	return response, nil
}

// Close kaynakları temizler
func (mm *MemoryManager) Close() {
	if mm.client != nil {
		mm.client.Close()
	}
}

func main() {
	// .env dosyasını yükle
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env dosyası bulunamadı, sistem environment variable'ları kullanılacak")
	}

	// API anahtarını al
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		log.Fatal("GOOGLE_API_KEY environment variable tanımlanmamış")
	}

	mm, err := NewMemoryManager(apiKey)
	if err != nil {
		log.Fatalf("MemoryManager oluşturulamadı: %v", err)
	}
	defer mm.Close()

	// Örnek müşteri verisi oluştur
	customer := &CustomerMemory{
		CustomerID: "CUST_12345",
		Name:       "Ahmet Yılmaz",
		Preferences: map[string]interface{}{
			"preferred_language": "Türkçe",
			"communication_time": "Sabah saatleri",
			"product_interest":   "Elektronik",
		},
		PastInteractions: []Interaction{
			{
				Timestamp:  time.Now().AddDate(0, 0, -7),
				Issue:      "Ürün geç geldi",
				Resolution: "Kargo takibi yapıldı, özür dilendi",
				Sentiment:  "Olumsuz",
			},
			{
				Timestamp:  time.Now().AddDate(0, 0, -3),
				Issue:      "İndirim kodu çalışmıyor",
				Resolution: "Yeni kod gönderildi, %20 ekstra indirim verildi",
				Sentiment:  "Olumlu",
			},
		},
	}

	mm.SaveMemory(customer.CustomerID, customer)

	// Yeni bir sorgu işle
	ctx := context.Background()
	query := "Merhaba, yeni aldığım kulaklığın garantisi ne kadar süreyle geçerli?"

	fmt.Println("=== Müşteri Destek Ajanı ===")
	fmt.Printf("\nMüşteri: %s\n", query)
	fmt.Println("\nAjan yanıtı hazırlanıyor...")

	response, err := mm.ProcessQuery(ctx, customer.CustomerID, query)
	if err != nil {
		log.Fatalf("Sorgu işlenemedi: %v", err)
	}

	fmt.Printf("\nAjan: %s\n", response)

	// Etkileşimi kaydet
	mm.AddInteraction(
		customer.CustomerID,
		"Garanti süresi sorgusu",
		"Garanti bilgisi verildi",
		"Nötr",
	)

	fmt.Println("\n=== Hafıza Güncellendi ===")
	updatedMemory := mm.GetMemory(customer.CustomerID)
	fmt.Printf("Toplam etkileşim sayısı: %d\n", len(updatedMemory.PastInteractions))
	fmt.Printf("Son iletişim: %s\n", updatedMemory.LastContactDate.Format("2006-01-02 15:04:05"))
}
