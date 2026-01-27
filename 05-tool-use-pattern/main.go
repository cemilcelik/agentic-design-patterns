// Tool Use Pattern - Hava Durumu ve Para Birimi Dönüştürücü (Go)
// Gemini API'nin araç kullanım (Function Calling) yeteneklerini gösterir

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

// Hava durumu verisi için struct
type WeatherData struct {
	Temperature float64 `json:"temperature"`
	Condition   string  `json:"condition"`
	Humidity    int     `json:"humidity"`
}

// Para birimi dönüşümü için struct
type CurrencyConversion struct {
	Amount   float64 `json:"amount"`
	FromRate float64 `json:"from_rate"`
	ToRate   float64 `json:"to_rate"`
	Result   float64 `json:"result"`
}

// Hava durumu sorgulama fonksiyonu (simülasyon)
func getWeather(city string) (WeatherData, error) {
	// Gerçek uygulamada bir API çağrısı yapılır
	weatherMap := map[string]WeatherData{
		"istanbul": {Temperature: 15.5, Condition: "Parçalı Bulutlu", Humidity: 65},
		"ankara":   {Temperature: 8.2, Condition: "Güneşli", Humidity: 45},
		"izmir":    {Temperature: 18.0, Condition: "Açık", Humidity: 70},
	}

	weather, exists := weatherMap[city]
	if !exists {
		return WeatherData{Temperature: 12.0, Condition: "Bilinmiyor", Humidity: 50}, nil
	}
	return weather, nil
}

// Para birimi dönüştürme fonksiyonu (simülasyon)
func convertCurrency(amount float64, from, to string) (CurrencyConversion, error) {
	// Gerçek uygulamada bir döviz API'si kullanılır
	rates := map[string]float64{
		"USD": 1.0,
		"EUR": 0.92,
		"TRY": 32.50,
		"GBP": 0.79,
	}

	fromRate, fromExists := rates[from]
	toRate, toExists := rates[to]

	if !fromExists || !toExists {
		return CurrencyConversion{}, fmt.Errorf("geçersiz para birimi")
	}

	// USD üzerinden dönüşüm
	inUSD := amount / fromRate
	result := inUSD * toRate

	return CurrencyConversion{
		Amount:   amount,
		FromRate: fromRate,
		ToRate:   toRate,
		Result:   result,
	}, nil
}

// Tool çağrısını işleme fonksiyonu
func handleToolCall(funcCall *genai.FunctionCall) (interface{}, error) {
	switch funcCall.Name {
	case "get_weather":
		city := funcCall.Args["city"].(string)
		return getWeather(city)

	case "convert_currency":
		amount := funcCall.Args["amount"].(float64)
		from := funcCall.Args["from"].(string)
		to := funcCall.Args["to"].(string)
		return convertCurrency(amount, from, to)

	default:
		return nil, fmt.Errorf("bilinmeyen fonksiyon: %s", funcCall.Name)
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env dosyası bulunamadı.")
	}

	// Gemini API anahtarını çevre değişkeninden al
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		log.Fatal("GOOGLE_API_KEY çevre değişkeni ayarlanmalı")
	}

	ctx := context.Background()

	// Gemini client oluştur
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Client oluşturulamadı: %v", err)
	}
	defer client.Close()

	// Model oluştur
	model := client.GenerativeModel(os.Getenv("GEMINI_MODEL"))

	// Tool (fonksiyon) tanımlamaları
	weatherTool := &genai.Tool{
		FunctionDeclarations: []*genai.FunctionDeclaration{
			{
				Name:        "get_weather",
				Description: "Belirtilen şehir için hava durumu bilgilerini getirir",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"city": {
							Type:        genai.TypeString,
							Description: "Hava durumu sorgulanacak şehir adı (örn: istanbul, ankara, izmir)",
						},
					},
					Required: []string{"city"},
				},
			},
			{
				Name:        "convert_currency",
				Description: "Para birimi dönüşümü yapar",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"amount": {
							Type:        genai.TypeNumber,
							Description: "Dönüştürülecek miktar",
						},
						"from": {
							Type:        genai.TypeString,
							Description: "Kaynak para birimi kodu (USD, EUR, TRY, GBP)",
						},
						"to": {
							Type:        genai.TypeString,
							Description: "Hedef para birimi kodu (USD, EUR, TRY, GBP)",
						},
					},
					Required: []string{"amount", "from", "to"},
				},
			},
		},
	}

	model.Tools = []*genai.Tool{weatherTool}

	// Chat oturumu başlat
	session := model.StartChat()

	// Örnek sorgular
	queries := []string{
		"İstanbul'un hava durumu nasıl?",
		"100 USD'yi Türk Lirası'na çevir",
		"Ankara'da hava nasıl ve 50 Euro kaç TRY eder?",
	}

	for i, query := range queries {
		fmt.Printf("\n=== Sorgu %d ===\n", i+1)
		fmt.Printf("Kullanıcı: %s\n", query)

		// İlk yanıtı al
		resp, err := session.SendMessage(ctx, genai.Text(query))
		if err != nil {
			log.Printf("Hata: %v", err)
			continue
		}

		// Tool çağrılarını işle
		for {
			hasToolCall := false

			for _, part := range resp.Candidates[0].Content.Parts {
				if funcCall, ok := part.(genai.FunctionCall); ok {
					hasToolCall = true
					fmt.Printf("🔧 Araç Çağrısı: %s\n", funcCall.Name)

					// Tool'u çalıştır
					result, err := handleToolCall(&funcCall)
					if err != nil {
						log.Printf("Tool çağrısı hatası: %v", err)
						continue
					}

					// Sonucu JSON'a çevir
					resultJSON, _ := json.Marshal(result)
					fmt.Printf("📊 Sonuç: %s\n", string(resultJSON))

					// Sonucu modele geri gönder
					resp, err = session.SendMessage(ctx,
						genai.FunctionResponse{
							Name:     funcCall.Name,
							Response: map[string]interface{}{"result": result},
						})
					if err != nil {
						log.Printf("Yanıt gönderme hatası: %v", err)
						continue
					}
				}
			}

			// Tool çağrısı yoksa döngüden çık
			if !hasToolCall {
				break
			}
		}

		// Nihai yanıtı yazdır
		for _, part := range resp.Candidates[0].Content.Parts {
			if textPart, ok := part.(genai.Text); ok {
				fmt.Printf("🤖 Asistan: %s\n", textPart)
			}
		}
	}
}
