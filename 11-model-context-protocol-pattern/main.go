// Model Context Protocol (MCP) Pattern - Bağlam ve Araç Sağlama Protokolü (Go)
// Standartlaştırılmış bir protokol aracılığıyla LLM'i harici veri kaynaklarına (Resources) ve araçlara (Tools) bağlayan sistem.

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

// MCPResource MCP üzerinden sunulan bir veri kaynağını temsil eder
type MCPResource struct {
	URI         string
	Name        string
	Description string
	TextContent string
}

// MCPTool MCP üzerinden sunulan çalıştırılabilir bir aracı temsil eder
type MCPTool struct {
	Name        string
	Description string
	Schema      *genai.Schema
	Execute     func(args map[string]any) (string, error)
}

// MCPServer standartlaştırılmış bağlam protokolü sunucusunu temsil eder
type MCPServer interface {
	ListResources() []MCPResource
	ReadResource(uri string) (*MCPResource, error)
	ListTools() []MCPTool
	CallTool(name string, args map[string]any) (string, error)
}

// MockDatabaseMCP veritabanı işlemlerini simüle eden bir MCP Sunucusu
type MockDatabaseMCP struct {
	resources []MCPResource
	tools     []MCPTool
}

// NewMockDatabaseMCP yeni bir MockDatabaseMCP oluşturur ve başlatır
func NewMockDatabaseMCP() *MockDatabaseMCP {
	mcp := &MockDatabaseMCP{
		resources: []MCPResource{
			{
				URI:         "db://users/schema",
				Name:        "Kullanıcı Tablosu Şeması",
				Description: "Kullanıcı veritabanı tablosunun yapısı",
				TextContent: "CREATE TABLE users (id INT, username VARCHAR, email VARCHAR, role VARCHAR);",
			},
			{
				URI:         "db://kpi/monthly",
				Name:        "Aylık KPI Verileri",
				Description: "Şirketin aylık performans göstergeleri",
				TextContent: "Ocak: 100K TL ciro, Şubat: 120K TL ciro, Mart: 150K TL ciro",
			},
		},
	}

	mcp.tools = []MCPTool{
		{
			Name:        "query_user_db",
			Description: "Kullanıcı veritabanında sorgu çalıştırır ve sonuçları döndürür. (Sadece email ile arama yapar)",
			Schema: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"email": {Type: genai.TypeString, Description: "Aranacak kullanıcının e-posta adresi"},
				},
				Required: []string{"email"},
			},
			Execute: func(args map[string]any) (string, error) {
				email, ok := args["email"].(string)
				if !ok {
					return "", fmt.Errorf("email parametresi eksik veya geçersiz")
				}
				// Mock DB sorgusu
				if email == "admin@sirket.com" {
					return `{"id": 1, "username": "admin", "role": "superadmin"}`, nil
				}
				if email == "test@sirket.com" {
					return `{"id": 2, "username": "testuser", "role": "user"}`, nil
				}
				return "Kullanıcı bulunamadı.", nil
			},
		},
	}

	return mcp
}

// ListResources tüm kaynakları listeler
func (m *MockDatabaseMCP) ListResources() []MCPResource {
	return m.resources
}

// ReadResource verilen URI'ye sahip kaynağı okur
func (m *MockDatabaseMCP) ReadResource(uri string) (*MCPResource, error) {
	for _, r := range m.resources {
		if r.URI == uri {
			return &r, nil
		}
	}
	return nil, fmt.Errorf("kaynak bulunamadı: %s", uri)
}

// ListTools tüm araçları listeler
func (m *MockDatabaseMCP) ListTools() []MCPTool {
	return m.tools
}

// CallTool ismi verilen aracı argümanlarla çalıştırır
func (m *MockDatabaseMCP) CallTool(name string, args map[string]any) (string, error) {
	for _, t := range m.tools {
		if t.Name == name {
			fmt.Printf("🔧 MCP Sunucusu Araç Çalıştırıyor: %s (%v)\n", name, args)
			return t.Execute(args)
		}
	}
	return "", fmt.Errorf("araç bulunamadı: %s", name)
}

// MCPAgent MCP sunucularıyla haberleşen ana Ajan
type MCPAgent struct {
	client  *genai.GenerativeModel
	session *genai.ChatSession
	mcp     MCPServer
}

// NewMCPAgent yeni bir ajan oluşturur
func NewMCPAgent(apiKey string, mcp MCPServer) (*MCPAgent, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("client oluşturulamadı: %v", err)
	}

	model := client.GenerativeModel(os.Getenv("GEMINI_MODEL"))
	model.SetTemperature(0.1) // Daha tahmin edilebilir sonuçlar için düşük tutuldu

	// MCP üzerinden sağlanan araçları modele tanıtma
	var genaiTools []*genai.Tool
	var functionDeclarations []*genai.FunctionDeclaration

	for _, tool := range mcp.ListTools() {
		fmt.Printf("🔌 MCP'den Model İçin Araç Yükleniyor: %s\n", tool.Name)
		funcDecl := &genai.FunctionDeclaration{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  tool.Schema,
		}
		functionDeclarations = append(functionDeclarations, funcDecl)
	}

	if len(functionDeclarations) > 0 {
		genaiTools = append(genaiTools, &genai.Tool{
			FunctionDeclarations: functionDeclarations,
		})
		model.Tools = genaiTools
	}

	// MCP Kaynaklarını (Resources) Sistem Promt'una (Context) Enjekte Etme
	instructionText := "Sen MCP (Model Context Protocol) destekli bir yapay zeka asistanısın. Aşağıdaki bağlam kaynaklarına erişimin var:\n\n"

	for _, res := range mcp.ListResources() {
		instructionText += fmt.Sprintf("URI: %s\nAdı: %s\nAçıklama: %s\nİçerik: %s\n\n", res.URI, res.Name, res.Description, res.TextContent)
	}

	instructionText += "\nGerektiğinde bu bağlamı okuyarak veya sağlanan fonksiyonları(tools) çalıştırarak kullanıcıya cevap ver. Cevaplarını Türkçe ver."

	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(instructionText)},
	}

	session := model.StartChat()

	return &MCPAgent{
		client:  model,
		session: session,
		mcp:     mcp,
	}, nil
}

// Ask ajana bir soru sorar, araç kullanımı gerektiğinde devralarak aracı çalıştırır ve sonucu modele tekrar iletir
func (agent *MCPAgent) Ask(ctx context.Context, query string) (string, error) {
	fmt.Printf("\n👤 Kullanıcı Sorgusu: %s\n", query)

	// Gemini modeline soruyu gönder
	resp, err := agent.session.SendMessage(ctx, genai.Text(query))
	if err != nil {
		return "", fmt.Errorf("mesaj gönderilemedi: %v", err)
	}

	// Modelin araç çağırması gerekiyorsa süreç bitene kadar dön
	for {
		requiresAction := false

		if len(resp.Candidates) == 0 {
			return "", fmt.Errorf("model yanıt üretemedi")
		}

		for _, part := range resp.Candidates[0].Content.Parts {
			if funcCall, ok := part.(genai.FunctionCall); ok {
				requiresAction = true
				fmt.Printf("🤖 Model Araç Kullanımı İstedi: %s\n", funcCall.Name)

				// Argümanları bir map'e çıkar
				args := make(map[string]any)
				for k, v := range funcCall.Args {
					args[k] = v
				}

				// MCP Sunucusu üzerinden aracı çalıştır
				result, err := agent.mcp.CallTool(funcCall.Name, args)
				if err != nil {
					result = fmt.Sprintf("Araç çalıştırma hatası: %v", err)
				}

				fmt.Printf("✅ Araç Sonucu: %s\n", result)

				// Sonucu modele geri gönder (FunctionResponse olarak)
				resp, err = agent.session.SendMessage(ctx, genai.FunctionResponse{
					Name: funcCall.Name,
					Response: map[string]any{
						"result": result,
					},
				})
				if err != nil {
					return "", fmt.Errorf("fonksiyon yanıtı gönderilemedi: %v", err)
				}

				break // Döngüyü kırıp modelden gelen Response'ta tekrar FunctionCall var mı diye bakacağız
			}
		}

		// Araç gereksinimi kalmadıysa veya salt metin gelmişse döngüden çık
		if !requiresAction {
			break
		}
	}

	// Nihai metin (Text) yanıtını birleştir
	var finalResponse string
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			finalResponse += string(text)
		}
	}

	return finalResponse, nil
}

func main() {
	if err := godotenv.Load("../.env"); err != nil {
		// Go çalışma dizini uyuşmazlığında güncel dizine de bakarız
		if err := godotenv.Load(); err != nil {
			log.Println("Uyarı: .env dosyası bulunamadı, sistem environment variable'ları kullanılacak")
		}
	}

	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		log.Fatal("GOOGLE_API_KEY environment variable tanımlanmamış")
	}

	fmt.Println("🤖 MCP (Model Context Protocol) Deseni Başlatılıyor...")
	fmt.Println("=" + strings.Repeat("=", 60))

	// 1. MCP Sunucusunu başlat (Örneğin: Şirket veritabanına bağlı bir sunucu simülasyonu)
	mcpServer := NewMockDatabaseMCP()

	// 2. Ajanı MCP sunucusuna bağlayarak oluştur
	agent, err := NewMCPAgent(apiKey, mcpServer)
	if err != nil {
		log.Fatalf("Ajan oluşturulamadı: %v", err)
	}

	ctx := context.Background()

	// SENARYO 1: MCP Resoures (Bağlam) Kullanımı
	// Ajan, Context'ine eklenen bilgilere (Resources) bakarak cevabı bilecektir.
	query1 := "Şirketin Mart ayındaki cirosu nedir?"
	response1, err := agent.Ask(ctx, query1)
	if err != nil {
		log.Printf("Hata: %v\n", err)
	} else {
		fmt.Printf("\n🤖 AI Yanıtı: %s\n", response1)
	}
	fmt.Println("\n" + strings.Repeat("-", 60))

	// SENARYO 2: MCP Tools (Araç) Kullanımı
	// Ajan, Tools listesinde olan ve veritabanı simülasyonunu çalıştıran aracı kullanacaktır.
	query2 := "Veritabanında 'admin@sirket.com' e-posta adresine sahip olan kullanıcının rolü nedir?"
	response2, err := agent.Ask(ctx, query2)
	if err != nil {
		log.Printf("Hata: %v\n", err)
	} else {
		fmt.Printf("\n🤖 AI Yanıtı: %s\n", response2)
	}

	fmt.Println("\n✅ MCP Demo tamamlandı!")
}
