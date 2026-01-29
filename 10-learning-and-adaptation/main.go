// Learning Pattern - RAG ve Memory Tabanlı Öğrenen Sistem (Go)
// Yeni bilgilerle bilgi tabanını güncelleyen ve müşteri geri bildirimlerinden öğrenen sistem

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

// Document şirket bilgi tabanındaki bir dökümanı temsil eder
type Document struct {
	ID       string
	Title    string
	Content  string
	Category string
	Vector   []float64 // Embedding vektörü (gerçek uygulamada daha sofistike olur)
}

// KnowledgeBase şirketin bilgi tabanını temsil eder
type KnowledgeBase struct {
	Documents []Document
}

// CustomerInteraction müşteri etkileşimini saklar (Memory-Based Learning için)
type CustomerInteraction struct {
	CustomerID   string
	Query        string
	Response     string
	Satisfaction int // 1-5 arası
	Timestamp    time.Time
}

// RAGAgent RAG tabanlı müşteri destek agent'ı
type RAGAgent struct {
	client        *genai.GenerativeModel
	knowledgeBase *KnowledgeBase
	memory        []CustomerInteraction // Online Learning için geçmiş etkileşimler
}

// NewRAGAgent yeni bir RAG agent oluşturur
func NewRAGAgent(apiKey string) (*RAGAgent, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("client oluşturulamadı: %v", err)
	}

	model := client.GenerativeModel(os.Getenv("GEMINI_MODEL"))

	// Model parametrelerini ayarla
	model.SetTemperature(0.7)
	model.SetTopP(0.9)
	model.SetMaxOutputTokens(1000)

	return &RAGAgent{
		client:        model,
		knowledgeBase: initializeKnowledgeBase(),
		memory:        make([]CustomerInteraction, 0),
	}, nil
}

// initializeKnowledgeBase örnek bilgi tabanını oluşturur
func initializeKnowledgeBase() *KnowledgeBase {
	return &KnowledgeBase{
		Documents: []Document{
			{
				ID:       "doc001",
				Title:    "İade Politikası",
				Content:  "Ürünlerimizi satın aldıktan sonra 30 gün içinde iade edebilirsiniz. İade için ürünün orijinal ambalajında ve kullanılmamış olması gerekir. İade işlemi için müşteri hizmetlerine başvurun. İade onaylandıktan sonra 5-7 iş günü içinde paranız iade edilir.",
				Category: "policy",
			},
			{
				ID:       "doc002",
				Title:    "Kargo Süreleri",
				Content:  "Standart kargoda ürünler 3-5 iş günü içinde teslim edilir. Hızlı kargo seçeneğinde 1-2 iş günü içinde teslim yapılır. Aynı gün kargo İstanbul, Ankara ve İzmir için geçerlidir ve saat 14:00'e kadar verilen siparişler aynı gün kargoya verilir.",
				Category: "shipping",
			},
			{
				ID:       "doc003",
				Title:    "Ürün Garantisi",
				Content:  "Tüm elektronik ürünlerimiz 2 yıl garantilidir. Garanti kapsamında arızalı ürünler ücretsiz tamir edilir veya değiştirilir. Fiziksel hasar garanti kapsamı dışındadır. Garanti için fatura ve garanti belgesi gereklidir.",
				Category: "warranty",
			},
			{
				ID:       "doc004",
				Title:    "Ödeme Yöntemleri",
				Content:  "Kredi kartı, banka kartı, havale/EFT ve kapıda ödeme seçenekleri mevcuttur. Kredi kartına 6 aya varan taksit imkanı sunuyoruz. 500 TL üzeri alışverişlerde kargo ücretsizdir.",
				Category: "payment",
			},
			{
				ID:       "doc005",
				Title:    "Müşteri Destek Saatleri",
				Content:  "Müşteri hizmetlerimiz hafta içi 09:00-18:00 saatleri arasında hizmet vermektedir. Hafta sonu ve resmi tatillerde destek hattımız kapalıdır. Acil durumlar için info@sirket.com adresinden 7/24 mail desteği alabilirsiniz.",
				Category: "support",
			},
		},
	}
}

// simpleEmbedding basit bir embedding fonksiyonu (gerçek uygulamada Gemini Embedding API kullanılmalı)
func simpleEmbedding(text string) []float64 {
	// Basit bir TF-IDF benzeri hesaplama (demo amaçlı)
	words := strings.Fields(strings.ToLower(text))
	wordFreq := make(map[string]float64)

	for _, word := range words {
		wordFreq[word]++
	}

	// 10 boyutlu basit bir vektör oluştur
	vector := make([]float64, 10)
	for i := range vector {
		vector[i] = wordFreq[words[i%len(words)]]
	}

	return vector
}

// cosineSimilarity iki vektör arasındaki cosine similarity'yi hesaplar
func cosineSimilarity(v1, v2 []float64) float64 {
	if len(v1) != len(v2) {
		return 0
	}

	var dotProduct, norm1, norm2 float64
	for i := range v1 {
		dotProduct += v1[i] * v2[i]
		norm1 += v1[i] * v1[i]
		norm2 += v2[i] * v2[i]
	}

	if norm1 == 0 || norm2 == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}

// RetrieveRelevantDocs sorguya en yakın dökümanları bulur (RAG'in R kısmı)
func (agent *RAGAgent) RetrieveRelevantDocs(query string, topK int) []Document {
	queryVector := simpleEmbedding(query)

	type docScore struct {
		doc   Document
		score float64
	}

	scores := make([]docScore, 0, len(agent.knowledgeBase.Documents))

	// Her dökümanın sorguya yakınlığını hesapla
	for _, doc := range agent.knowledgeBase.Documents {
		docVector := simpleEmbedding(doc.Content)
		similarity := cosineSimilarity(queryVector, docVector)
		scores = append(scores, docScore{doc: doc, score: similarity})
	}

	// Skor'a göre sırala (basit bubble sort)
	for i := 0; i < len(scores)-1; i++ {
		for j := 0; j < len(scores)-i-1; j++ {
			if scores[j].score < scores[j+1].score {
				scores[j], scores[j+1] = scores[j+1], scores[j]
			}
		}
	}

	// En iyi topK dökümanı döndür
	result := make([]Document, 0, topK)
	for i := 0; i < topK && i < len(scores); i++ {
		result = append(result, scores[i].doc)
	}

	return result
}

// GetRecentInteractions belirli bir müşterinin son etkileşimlerini getirir (Memory-Based Learning)
func (agent *RAGAgent) GetRecentInteractions(customerID string, limit int) []CustomerInteraction {
	interactions := make([]CustomerInteraction, 0)

	for i := len(agent.memory) - 1; i >= 0 && len(interactions) < limit; i-- {
		if agent.memory[i].CustomerID == customerID {
			interactions = append(interactions, agent.memory[i])
		}
	}

	return interactions
}

// AnswerQuery RAG kullanarak müşteri sorusunu yanıtlar
func (agent *RAGAgent) AnswerQuery(ctx context.Context, customerID, query string) (string, error) {
	// 1. RETRIEVAL: İlgili dökümanları bul
	relevantDocs := agent.RetrieveRelevantDocs(query, 2)

	// 2. MEMORY: Müşterinin geçmiş etkileşimlerini getir (Online Learning)
	recentInteractions := agent.GetRecentInteractions(customerID, 3)

	// 3. CONTEXT BUILDING: Prompt'u oluştur
	contextBuilder := strings.Builder{}
	contextBuilder.WriteString("Sen bir müşteri destek asistanısın. Aşağıdaki bilgilere dayanarak soruyu yanıtla.\n\n")

	// İlgili dökümanları ekle
	contextBuilder.WriteString("=== ŞİRKET BİLGİ BANKASI ===\n")
	for i, doc := range relevantDocs {
		contextBuilder.WriteString(fmt.Sprintf("\nDöküman %d - %s:\n%s\n", i+1, doc.Title, doc.Content))
	}

	// Müşteri geçmişini ekle (kişiselleştirme için)
	if len(recentInteractions) > 0 {
		contextBuilder.WriteString("\n=== MÜŞTERİ GEÇMİŞİ ===\n")
		for i, interaction := range recentInteractions {
			contextBuilder.WriteString(fmt.Sprintf("\nÖnceki Soru %d: %s\n", i+1, interaction.Query))
			contextBuilder.WriteString(fmt.Sprintf("Memnuniyet: %d/5\n", interaction.Satisfaction))
		}
	}

	contextBuilder.WriteString("\n=== KURALLAR ===\n")
	contextBuilder.WriteString("- Sadece yukarıdaki bilgi bankasındaki bilgileri kullan\n")
	contextBuilder.WriteString("- Bilmediğin bir şey varsa dürüstçe söyle\n")
	contextBuilder.WriteString("- Nazik ve profesyonel ol\n")
	contextBuilder.WriteString("- Müşterinin geçmiş etkileşimlerini dikkate al\n\n")

	contextBuilder.WriteString(fmt.Sprintf("Müşteri Sorusu: %s\n\nYanıt:", query))

	prompt := contextBuilder.String()

	// 4. GENERATION: Gemini ile yanıt oluştur
	resp, err := agent.client.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("yanıt oluşturulamadı: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("model yanıt üretemedi")
	}

	response := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

	// 5. ADAPTATION: Etkileşimi kaydet (Online Learning için)
	interaction := CustomerInteraction{
		CustomerID: customerID,
		Query:      query,
		Response:   response,
		Timestamp:  time.Now(),
		// Gerçek uygulamada müşteri geri bildirimi beklenir
	}
	agent.memory = append(agent.memory, interaction)

	return response, nil
}

// UpdateSatisfaction müşteri memnuniyetini günceller (Reinforcement Learning benzeri)
func (agent *RAGAgent) UpdateSatisfaction(customerID string, satisfaction int) {
	// Son etkileşimi bul ve memnuniyeti güncelle
	for i := len(agent.memory) - 1; i >= 0; i-- {
		if agent.memory[i].CustomerID == customerID {
			agent.memory[i].Satisfaction = satisfaction
			fmt.Printf("✅ Müşteri memnuniyeti güncellendi: %d/5\n", satisfaction)
			break
		}
	}
}

// AddDocument yeni döküman ekler (Knowledge Base güncellemesi)
func (agent *RAGAgent) AddDocument(doc Document) {
	agent.knowledgeBase.Documents = append(agent.knowledgeBase.Documents, doc)
	fmt.Printf("✅ Yeni döküman eklendi: %s\n", doc.Title)
}

// GetStatistics agent'ın performans istatistiklerini gösterir
func (agent *RAGAgent) GetStatistics() map[string]interface{} {
	totalInteractions := len(agent.memory)

	var totalSatisfaction int
	satisfiedCount := 0

	for _, interaction := range agent.memory {
		if interaction.Satisfaction > 0 {
			totalSatisfaction += interaction.Satisfaction
			satisfiedCount++
		}
	}

	avgSatisfaction := 0.0
	if satisfiedCount > 0 {
		avgSatisfaction = float64(totalSatisfaction) / float64(satisfiedCount)
	}

	return map[string]interface{}{
		"total_interactions":   totalInteractions,
		"rated_interactions":   satisfiedCount,
		"average_satisfaction": avgSatisfaction,
		"knowledge_base_size":  len(agent.knowledgeBase.Documents),
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env dosyası bulunamadı, sistem environment variable'ları kullanılacak")
	}

	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		log.Fatal("GOOGLE_API_KEY environment variable tanımlanmamış")
	}

	// RAG Agent'ı oluştur
	agent, err := NewRAGAgent(apiKey)
	if err != nil {
		log.Fatalf("Agent oluşturulamadı: %v", err)
	}

	ctx := context.Background()

	fmt.Println("🤖 RAG Tabanlı Müşteri Destek AI Agent Başlatıldı")
	fmt.Println("=" + strings.Repeat("=", 60))

	// Örnek Senaryo: Bir müşterinin art arda soruları
	customerID := "customer_12345"

	// Soru 1: İade hakkında
	fmt.Println("\n📝 Müşteri Sorusu 1:")
	query1 := "Aldığım ürünü iade edebilir miyim? Ne kadar sürem var?"
	fmt.Println(query1)

	response1, err := agent.AnswerQuery(ctx, customerID, query1)
	if err != nil {
		log.Printf("Hata: %v", err)
	} else {
		fmt.Println("\n🤖 AI Yanıtı:")
		fmt.Println(response1)
	}

	// Müşteri memnun kaldı
	agent.UpdateSatisfaction(customerID, 5)

	time.Sleep(2 * time.Second)

	// Soru 2: Kargo hakkında (Memory-Based: önceki soruyu hatırlayacak)
	fmt.Println("\n" + strings.Repeat("-", 60))
	fmt.Println("\n📝 Müşteri Sorusu 2:")
	query2 := "Kargo ne kadar sürer?"
	fmt.Println(query2)

	response2, err := agent.AnswerQuery(ctx, customerID, query2)
	if err != nil {
		log.Printf("Hata: %v", err)
	} else {
		fmt.Println("\n🤖 AI Yanıtı:")
		fmt.Println(response2)
	}

	agent.UpdateSatisfaction(customerID, 4)

	time.Sleep(2 * time.Second)

	// YENİ DÖKÜMAN EKLEME (Adaptation örneği)
	fmt.Println("\n" + strings.Repeat("-", 60))
	fmt.Println("\n📄 Bilgi tabanına yeni döküman ekleniyor...")

	newDoc := Document{
		ID:       "doc006",
		Title:    "Yeni Promosyon",
		Content:  "Ocak ayı boyunca tüm ürünlerde %20 indirim! Kampanya kodu: OCAK2026. 1000 TL üzeri alışverişlerde ek %10 indirim kazanın.",
		Category: "promotion",
	}
	agent.AddDocument(newDoc)

	time.Sleep(2 * time.Second)

	// Soru 3: Yeni eklenen bilgi hakkında (Online Learning)
	fmt.Println("\n" + strings.Repeat("-", 60))
	fmt.Println("\n📝 Müşteri Sorusu 3:")
	query3 := "Şu anda bir kampanyanız var mı?"
	fmt.Println(query3)

	response3, err := agent.AnswerQuery(ctx, customerID, query3)
	if err != nil {
		log.Printf("Hata: %v", err)
	} else {
		fmt.Println("\n🤖 AI Yanıtı:")
		fmt.Println(response3)
	}

	agent.UpdateSatisfaction(customerID, 5)

	// İstatistikler
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("\n📊 AGENT İSTATİSTİKLERİ:")
	stats := agent.GetStatistics()
	statsJSON, _ := json.MarshalIndent(stats, "", "  ")
	fmt.Println(string(statsJSON))

	fmt.Println("\n✅ Demo tamamlandı!")
	fmt.Println("\n💡 LEARNING & ADAPTATION ÖZETİ:")
	fmt.Println("  • RAG: Her soruda ilgili dökümanlar aranıp kullanıldı")
	fmt.Println("  • Memory-Based Learning: Müşteri geçmişi hatırlandı")
	fmt.Println("  • Online Learning: Yeni döküman anında sisteme eklendi")
	fmt.Println("  • Feedback Loop: Müşteri memnuniyeti kaydedildi")
}
