// Advanced Memory Management - Gelişmiş Bellek Sistemi (Go)
// Embedding, vektör benzerliği ve kısa/uzun dönem bellek optimizasyonunu içerir

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

// MemoryType bellek türlerini tanımlar
type MemoryType string

const (
	ShortTerm  MemoryType = "short_term"
	LongTerm   MemoryType = "long_term"
	Episodic   MemoryType = "episodic"
	Semantic   MemoryType = "semantic"
	Procedural MemoryType = "procedural"
)

// MemoryItem bellek öğesi
type MemoryItem struct {
	ID          string                 `json:"id"`
	Type        MemoryType             `json:"type"`
	Content     string                 `json:"content"`
	Embedding   []float64              `json:"embedding"`
	Metadata    map[string]interface{} `json:"metadata"`
	Importance  float64                `json:"importance"`
	AccessCount int                    `json:"access_count"`
	CreatedAt   time.Time              `json:"created_at"`
	LastAccess  time.Time              `json:"last_access"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
	Tags        []string               `json:"tags"`
}

// ConversationContext konuşma bağlamı
type ConversationContext struct {
	SessionID        string       `json:"session_id"`
	Messages         []Message    `json:"messages"`
	ActiveTopics     []string     `json:"active_topics"`
	UserProfile      *UserProfile `json:"user_profile"`
	ConversationGoal string       `json:"conversation_goal"`
	EmotionalTone    string       `json:"emotional_tone"`
	StartTime        time.Time    `json:"start_time"`
}

// Message mesaj yapısı
type Message struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Intent    string    `json:"intent"`
	Entities  []string  `json:"entities"`
}

// UserProfile kullanıcı profili
type UserProfile struct {
	UserID             string                 `json:"user_id"`
	Name               string                 `json:"name"`
	Preferences        map[string]interface{} `json:"preferences"`
	Interests          []string               `json:"interests"`
	ExpertiseLevel     string                 `json:"expertise_level"`
	CommunicationStyle string                 `json:"communication_style"`
	Goals              []string               `json:"goals"`
	LearningStyle      string                 `json:"learning_style"`
}

// AdvancedMemoryAgent gelişmiş bellek yönetimi yapan ajan
type AdvancedMemoryAgent struct {
	client           *genai.Client
	model            *genai.GenerativeModel
	embeddingModel   *genai.EmbeddingModel
	memories         map[string]*MemoryItem
	contexts         map[string]*ConversationContext
	memoryIndex      map[MemoryType][]*MemoryItem
	consolidationLog []ConsolidationEvent
}

// ConsolidationEvent bellek birleştirme olayı
type ConsolidationEvent struct {
	Timestamp      time.Time `json:"timestamp"`
	SourceMemories []string  `json:"source_memories"`
	ConsolidatedID string    `json:"consolidated_id"`
	Reason         string    `json:"reason"`
}

// NewAdvancedMemoryAgent yeni gelişmiş bellek ajanı oluşturur
func NewAdvancedMemoryAgent(apiKey string) (*AdvancedMemoryAgent, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("client oluşturulamadı: %v", err)
	}

	model := client.GenerativeModel(os.Getenv("GEMINI_MODEL"))
	model.SetTemperature(0.8)
	model.SetTopP(0.95)
	model.SetMaxOutputTokens(2048)

	embeddingModel := client.EmbeddingModel("text-embedding-004")

	agent := &AdvancedMemoryAgent{
		client:           client,
		model:            model,
		embeddingModel:   embeddingModel,
		memories:         make(map[string]*MemoryItem),
		contexts:         make(map[string]*ConversationContext),
		memoryIndex:      make(map[MemoryType][]*MemoryItem),
		consolidationLog: []ConsolidationEvent{},
	}

	return agent, nil
}

// CreateEmbedding metin için embedding oluşturur
func (a *AdvancedMemoryAgent) CreateEmbedding(ctx context.Context, text string) ([]float64, error) {
	result, err := a.embeddingModel.EmbedContent(ctx, genai.Text(text))
	if err != nil {
		return nil, fmt.Errorf("embedding oluşturulamadı: %v", err)
	}

	if result == nil || result.Embedding == nil || len(result.Embedding.Values) == 0 {
		return nil, fmt.Errorf("geçersiz embedding sonucu")
	}

	embedding := make([]float64, len(result.Embedding.Values))
	for i, v := range result.Embedding.Values {
		embedding[i] = float64(v)
	}

	return embedding, nil
}

// CosineSimilarity iki vektör arasındaki kosinüs benzerliğini hesaplar
func CosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0.0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// StoreMemory bellek öğesi saklar
func (a *AdvancedMemoryAgent) StoreMemory(ctx context.Context, memType MemoryType, content string, metadata map[string]interface{}, importance float64, tags []string) (*MemoryItem, error) {
	embedding, err := a.CreateEmbedding(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("embedding oluşturulamadı: %v", err)
	}

	memoryID := fmt.Sprintf("mem_%d", time.Now().UnixNano())

	var expiresAt *time.Time
	if memType == ShortTerm {
		expiry := time.Now().Add(1 * time.Hour)
		expiresAt = &expiry
	}

	memory := &MemoryItem{
		ID:          memoryID,
		Type:        memType,
		Content:     content,
		Embedding:   embedding,
		Metadata:    metadata,
		Importance:  importance,
		AccessCount: 0,
		CreatedAt:   time.Now(),
		LastAccess:  time.Now(),
		ExpiresAt:   expiresAt,
		Tags:        tags,
	}

	a.memories[memoryID] = memory
	a.memoryIndex[memType] = append(a.memoryIndex[memType], memory)

	return memory, nil
}

// RetrieveRelevantMemories ilgili bellekleri getirir
func (a *AdvancedMemoryAgent) RetrieveRelevantMemories(ctx context.Context, query string, topK int, minSimilarity float64) ([]*MemoryItem, error) {
	queryEmbedding, err := a.CreateEmbedding(ctx, query)
	if err != nil {
		return nil, err
	}

	type scoredMemory struct {
		memory     *MemoryItem
		similarity float64
		recency    float64
		finalScore float64
	}

	scored := []scoredMemory{}

	for _, memory := range a.memories {
		// Süresi dolmuş bellekleri atla
		if memory.ExpiresAt != nil && time.Now().After(*memory.ExpiresAt) {
			continue
		}

		similarity := CosineSimilarity(queryEmbedding, memory.Embedding)
		if similarity < minSimilarity {
			continue
		}

		// Recency skoru (0-1 arası normalize)
		hoursSinceCreation := time.Since(memory.CreatedAt).Hours()
		recency := 1.0 / (1.0 + hoursSinceCreation/24.0)

		// Karma skor: similarity * 0.6 + importance * 0.2 + recency * 0.2
		finalScore := similarity*0.6 + memory.Importance*0.2 + recency*0.2

		scored = append(scored, scoredMemory{
			memory:     memory,
			similarity: similarity,
			recency:    recency,
			finalScore: finalScore,
		})

		// Erişim sayısını artır
		memory.AccessCount++
		memory.LastAccess = time.Now()
	}

	// Skora göre sırala
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].finalScore > scored[j].finalScore
	})

	// Top-K seç
	if len(scored) > topK {
		scored = scored[:topK]
	}

	result := make([]*MemoryItem, len(scored))
	for i, s := range scored {
		result[i] = s.memory
	}

	return result, nil
}

// ConsolidateMemories bellekleri birleştirir
func (a *AdvancedMemoryAgent) ConsolidateMemories(ctx context.Context) error {
	// Kısa vadeli bellekleri uzun vadeli belleğe dönüştür
	shortTermMemories := a.memoryIndex[ShortTerm]

	if len(shortTermMemories) < 3 {
		return nil
	}

	// Benzer bellekleri grupla
	groups := a.groupSimilarMemories(shortTermMemories, 0.75)

	for _, group := range groups {
		if len(group) < 2 {
			continue
		}

		// Grup içeriğini birleştir
		var contents []string
		var sources []string
		var totalImportance float64

		for _, mem := range group {
			contents = append(contents, mem.Content)
			sources = append(sources, mem.ID)
			totalImportance += mem.Importance
		}

		consolidatedContent := strings.Join(contents, " | ")
		avgImportance := totalImportance / float64(len(group))

		// Yeni uzun vadeli bellek oluştur
		consolidated, err := a.StoreMemory(
			ctx,
			LongTerm,
			consolidatedContent,
			map[string]interface{}{
				"consolidated_from":  sources,
				"consolidation_date": time.Now(),
			},
			avgImportance,
			[]string{"consolidated"},
		)

		if err != nil {
			continue
		}

		// Birleştirme olayını kaydet
		a.consolidationLog = append(a.consolidationLog, ConsolidationEvent{
			Timestamp:      time.Now(),
			SourceMemories: sources,
			ConsolidatedID: consolidated.ID,
			Reason:         "Similar short-term memories consolidated",
		})

		// Eski bellekleri sil
		for _, mem := range group {
			delete(a.memories, mem.ID)
		}
	}

	return nil
}

// groupSimilarMemories benzer bellekleri gruplar
func (a *AdvancedMemoryAgent) groupSimilarMemories(memories []*MemoryItem, threshold float64) [][]*MemoryItem {
	if len(memories) == 0 {
		return nil
	}

	visited := make(map[string]bool)
	groups := [][]*MemoryItem{}

	for i, mem1 := range memories {
		if visited[mem1.ID] {
			continue
		}

		group := []*MemoryItem{mem1}
		visited[mem1.ID] = true

		for j := i + 1; j < len(memories); j++ {
			mem2 := memories[j]
			if visited[mem2.ID] {
				continue
			}

			similarity := CosineSimilarity(mem1.Embedding, mem2.Embedding)
			if similarity >= threshold {
				group = append(group, mem2)
				visited[mem2.ID] = true
			}
		}

		groups = append(groups, group)
	}

	return groups
}

// ProcessQuery gelişmiş bellek ile sorgu işler
func (a *AdvancedMemoryAgent) ProcessQuery(ctx context.Context, sessionID, userID, query string) (string, error) {
	// Konuşma bağlamını al veya oluştur
	context := a.getOrCreateContext(sessionID, userID)

	// Sorgu için ilgili bellekleri getir
	relevantMemories, err := a.RetrieveRelevantMemories(ctx, query, 5, 0.5)
	if err != nil {
		return "", err
	}

	// Intent analizi ve entity çıkarma
	intent, entities := a.analyzeIntent(query)

	// Mesajı kaydet
	message := Message{
		Role:      "user",
		Content:   query,
		Timestamp: time.Now(),
		Intent:    intent,
		Entities:  entities,
	}
	context.Messages = append(context.Messages, message)

	// Sorguyu kısa vadeli belleğe kaydet
	_, err = a.StoreMemory(ctx, ShortTerm, query, map[string]interface{}{
		"session_id": sessionID,
		"user_id":    userID,
		"intent":     intent,
	}, 0.5, []string{"user_query"})
	if err != nil {
		log.Printf("Bellek kaydedilemedi: %v", err)
	}

	// Gelişmiş prompt oluştur
	prompt := a.buildAdvancedPrompt(context, relevantMemories, query)

	// Yanıt üret
	resp, err := a.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("yanıt üretilemedi: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("yanıt alınamadı")
	}

	response := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

	// Yanıtı uzun vadeli belleğe kaydet
	_, err = a.StoreMemory(ctx, LongTerm, response, map[string]interface{}{
		"session_id": sessionID,
		"query":      query,
		"intent":     intent,
	}, 0.7, []string{"agent_response"})

	// Asistan mesajını kaydet
	assistantMessage := Message{
		Role:      "assistant",
		Content:   response,
		Timestamp: time.Now(),
	}
	context.Messages = append(context.Messages, assistantMessage)

	// Periyodik bellek birleştirme
	if len(context.Messages)%10 == 0 {
		go a.ConsolidateMemories(ctx)
	}

	return response, nil
}

// getOrCreateContext konuşma bağlamını getirir veya oluşturur
func (a *AdvancedMemoryAgent) getOrCreateContext(sessionID, userID string) *ConversationContext {
	if ctx, exists := a.contexts[sessionID]; exists {
		return ctx
	}

	ctx := &ConversationContext{
		SessionID:    sessionID,
		Messages:     []Message{},
		ActiveTopics: []string{},
		UserProfile: &UserProfile{
			UserID:      userID,
			Preferences: make(map[string]interface{}),
			Interests:   []string{},
		},
		StartTime: time.Now(),
	}

	a.contexts[sessionID] = ctx
	return ctx
}

// analyzeIntent basit intent analizi
func (a *AdvancedMemoryAgent) analyzeIntent(query string) (string, []string) {
	queryLower := strings.ToLower(query)

	intents := map[string][]string{
		"question":    {"ne", "nedir", "nasıl", "neden", "kim", "hangi"},
		"request":     {"lütfen", "yapar mısın", "yardım", "istiyorum"},
		"information": {"bilgi", "açıkla", "anlat", "göster"},
		"command":     {"yap", "oluştur", "çalıştır", "başlat"},
	}

	for intent, keywords := range intents {
		for _, keyword := range keywords {
			if strings.Contains(queryLower, keyword) {
				return intent, extractEntities(query)
			}
		}
	}

	return "general", extractEntities(query)
}

// extractEntities basit entity çıkarma
func extractEntities(text string) []string {
	words := strings.Fields(text)
	entities := []string{}

	for _, word := range words {
		if len(word) > 0 && strings.ToUpper(word[:1]) == word[:1] && len(word) > 3 {
			entities = append(entities, word)
		}
	}

	return entities
}

// buildAdvancedPrompt gelişmiş prompt oluşturur
func (a *AdvancedMemoryAgent) buildAdvancedPrompt(ctx *ConversationContext, memories []*MemoryItem, query string) string {
	var sb strings.Builder

	sb.WriteString("Sen gelişmiş bellek sistemine sahip bir AI asistanısın.\n\n")

	// Kullanıcı profili
	if ctx.UserProfile != nil {
		sb.WriteString(fmt.Sprintf("Kullanıcı: %s (ID: %s)\n", ctx.UserProfile.Name, ctx.UserProfile.UserID))
		if len(ctx.UserProfile.Interests) > 0 {
			sb.WriteString(fmt.Sprintf("İlgi Alanları: %s\n", strings.Join(ctx.UserProfile.Interests, ", ")))
		}
	}

	// İlgili bellekler
	if len(memories) > 0 {
		sb.WriteString("\n=== İlgili Geçmiş Bilgiler ===\n")
		for i, mem := range memories {
			sb.WriteString(fmt.Sprintf("%d. [%s | %s] %s\n",
				i+1,
				mem.Type,
				mem.CreatedAt.Format("2006-01-02 15:04"),
				mem.Content))
		}
	}

	// Son konuşma geçmişi
	if len(ctx.Messages) > 0 {
		sb.WriteString("\n=== Son Konuşma Geçmişi ===\n")
		start := len(ctx.Messages) - 6
		if start < 0 {
			start = 0
		}
		for _, msg := range ctx.Messages[start:] {
			sb.WriteString(fmt.Sprintf("[%s] %s: %s\n",
				msg.Timestamp.Format("15:04"),
				msg.Role,
				msg.Content))
		}
	}

	sb.WriteString(fmt.Sprintf("\n=== Güncel Soru ===\n%s\n\n", query))
	sb.WriteString("Lütfen geçmiş bilgileri ve konuşma bağlamını dikkate alarak detaylı ve kişiselleştirilmiş bir yanıt ver.")

	return sb.String()
}

// GetMemoryStatistics bellek istatistiklerini döndürür
func (a *AdvancedMemoryAgent) GetMemoryStatistics() map[string]interface{} {
	stats := make(map[string]interface{})

	typeCounts := make(map[MemoryType]int)
	totalImportance := 0.0
	totalAccess := 0

	for _, mem := range a.memories {
		typeCounts[mem.Type]++
		totalImportance += mem.Importance
		totalAccess += mem.AccessCount
	}

	stats["total_memories"] = len(a.memories)
	stats["by_type"] = typeCounts
	stats["avg_importance"] = totalImportance / float64(len(a.memories))
	stats["total_access_count"] = totalAccess
	stats["consolidation_events"] = len(a.consolidationLog)
	stats["active_sessions"] = len(a.contexts)

	return stats
}

// Close kaynakları temizler
func (a *AdvancedMemoryAgent) Close() {
	if a.client != nil {
		a.client.Close()
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

	agent, err := NewAdvancedMemoryAgent(apiKey)
	if err != nil {
		log.Fatalf("Agent oluşturulamadı: %v", err)
	}
	defer agent.Close()

	ctx := context.Background()
	sessionID := "session_001"
	userID := "user_mehmet"

	// Kullanıcı profili oluştur
	profile := &UserProfile{
		UserID:             userID,
		Name:               "Mehmet Demir",
		Interests:          []string{"Yapay Zeka", "Makine Öğrenmesi", "Go Programlama"},
		ExpertiseLevel:     "Intermediate",
		CommunicationStyle: "Teknik ve Detaylı",
		LearningStyle:      "Pratik Örneklerle",
	}

	context := agent.getOrCreateContext(sessionID, userID)
	context.UserProfile = profile

	// Bazı uzun vadeli bilgiler ekle
	agent.StoreMemory(ctx, LongTerm, "Mehmet, Go dilinde 3 yıldır çalışıyor ve mikroservis mimarileri konusunda deneyimli",
		map[string]interface{}{"category": "expertise"}, 0.9, []string{"background", "skills"})

	agent.StoreMemory(ctx, LongTerm, "Mehmet'in şu anki projesi, dağıtık sistemlerde memory management optimizasyonu",
		map[string]interface{}{"category": "current_project"}, 0.8, []string{"project"})

	fmt.Println("=== Gelişmiş Bellek Yönetimli AI Agent ===")

	// Sorgu dizisi
	queries := []string{
		"Merhaba, Go'da memory leak tespiti için hangi araçları önerirsin?",
		"Peki bu araçları mikroservis mimarilerinde nasıl kullanabilirim?",
		"Projemde kullanabileceğim bir monitoring stratejisi önerir misin?",
		"Son konuştuklarımızı özetleyebilir misin?",
	}

	for i, query := range queries {
		fmt.Printf("\n--- Soru %d ---\n", i+1)
		fmt.Printf("Kullanıcı: %s\n\n", query)

		response, err := agent.ProcessQuery(ctx, sessionID, userID, query)
		if err != nil {
			log.Printf("Hata: %v", err)
			continue
		}

		fmt.Printf("Agent: %s\n", response)
		time.Sleep(1 * time.Second)
	}

	// Bellek istatistikleri
	fmt.Println("\n\n=== Bellek İstatistikleri ===")
	stats := agent.GetMemoryStatistics()
	statsJSON, _ := json.MarshalIndent(stats, "", "  ")
	fmt.Println(string(statsJSON))

	// Bellek birleştirme
	fmt.Println("\n=== Bellek Birleştirme İşlemi ===")
	err = agent.ConsolidateMemories(ctx)
	if err != nil {
		log.Printf("Birleştirme hatası: %v", err)
	} else {
		fmt.Printf("Birleştirme tamamlandı. Toplam olay: %d\n", len(agent.consolidationLog))
	}

	// Güncellenmiş istatistikler
	fmt.Println("\n=== Güncellenmiş İstatistikler ===")
	updatedStats := agent.GetMemoryStatistics()
	updatedJSON, _ := json.MarshalIndent(updatedStats, "", "  ")
	fmt.Println(string(updatedJSON))
}
