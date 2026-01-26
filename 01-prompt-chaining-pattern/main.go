// Prompt Chaining Pattern - Blog Yazısı Oluşturma (Go)
// Gemini API kullanarak karmaşık bir görevi adımlara böler

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

// ChainResult her adımın sonucunu tutar
type ChainResult struct {
	Research string
	Outline  string
	Draft    string
	Final    string
}

// BlogWriterChain blog yazısı oluşturmak için prompt chaining uygular
type BlogWriterChain struct {
	client  *genai.Client
	model   *genai.GenerativeModel
	results ChainResult
}

// NewBlogWriterChain yeni bir chain instance oluşturur
func NewBlogWriterChain(ctx context.Context, apiKey string) (*BlogWriterChain, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("client oluşturulamadı: %w", err)
	}

	model := client.GenerativeModel(os.Getenv("GEMINI_MODEL"))

	// Güvenlik ayarları (opsiyonel)
	model.SafetySettings = []*genai.SafetySetting{
		{
			Category:  genai.HarmCategoryHarassment,
			Threshold: genai.HarmBlockMediumAndAbove,
		},
	}

	return &BlogWriterChain{
		client: client,
		model:  model,
	}, nil
}

// Close client'ı kapatır
func (b *BlogWriterChain) Close() error {
	return b.client.Close()
}

// Step1Research konu hakkında araştırma yapar
func (b *BlogWriterChain) Step1Research(ctx context.Context, topic string) (string, error) {
	prompt := fmt.Sprintf(`
Konu: %s

Bu konu hakkında bir blog yazısı için araştırma yap.
Aşağıdakileri liste halinde ver:
1. Ana tema ve odak noktası
2. Hedef kitle
3. 5 önemli alt başlık önerisi
4. Her alt başlık için kısa açıklama

Sadece bu bilgileri yapılandırılmış şekilde ver.
`, topic)

	resp, err := b.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("araştırma adımı başarısız: %w", err)
	}

	result := extractText(resp)
	b.results.Research = result
	fmt.Println("✓ Adım 1 tamamlandı: Araştırma")
	return result, nil
}

// Step2Outline araştırmaya dayalı taslak oluşturur
func (b *BlogWriterChain) Step2Outline(ctx context.Context, research string) (string, error) {
	prompt := fmt.Sprintf(`
Aşağıdaki araştırma notlarına dayanarak detaylı bir blog yazısı taslağı oluştur:

%s

Taslak şunları içermeli:
- Çekici bir başlık
- Giriş paragrafı için anahtar noktalar
- Her alt başlık için 2-3 cümlelik açıklama
- Sonuç bölümü için ana mesaj

Taslağı yapılandırılmış şekilde hazırla.
`, research)

	resp, err := b.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("taslak oluşturma adımı başarısız: %w", err)
	}

	result := extractText(resp)
	b.results.Outline = result
	fmt.Println("✓ Adım 2 tamamlandı: Taslak oluşturma")
	return result, nil
}

// Step3Write taslağı tam bir blog yazısına dönüştürür
func (b *BlogWriterChain) Step3Write(ctx context.Context, outline, tone string) (string, error) {
	prompt := fmt.Sprintf(`
Aşağıdaki taslağı kullanarak tam bir blog yazısı yaz:

%s

Ton: %s

Yazı şunları içermeli:
- İlgi çekici giriş
- Her alt başlık için detaylı paragraflar (200-300 kelime)
- Örnekler ve açıklamalar
- Güçlü bir sonuç

800-1000 kelime uzunluğunda, akıcı ve ilgi çekici bir yazı oluştur.
`, outline, tone)

	resp, err := b.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("yazı yazma adımı başarısız: %w", err)
	}

	result := extractText(resp)
	b.results.Draft = result
	fmt.Println("✓ Adım 3 tamamlandı: Yazı yazma")
	return result, nil
}

// Step4Edit yazıyı düzenler ve iyileştirir
func (b *BlogWriterChain) Step4Edit(ctx context.Context, draft string) (string, error) {
	prompt := fmt.Sprintf(`
Aşağıdaki blog yazısını düzenle ve iyileştir:

%s

Şunlara dikkat et:
- Gramer ve yazım hatalarını düzelt
- Cümle yapısını iyileştir
- Gereksiz tekrarları kaldır
- Akışı düzgünleştir
- SEO için anahtar kelimeleri vurgula

Düzenlenmiş, yayına hazır versiyonu ver.
`, draft)

	resp, err := b.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("düzenleme adımı başarısız: %w", err)
	}

	result := extractText(resp)
	b.results.Final = result
	fmt.Println("✓ Adım 4 tamamlandı: Düzenleme")
	return result, nil
}

// GenerateBlog tüm chain'i çalıştırarak blog yazısı oluşturur
func (b *BlogWriterChain) GenerateBlog(ctx context.Context, topic, tone string) (*ChainResult, error) {
	fmt.Printf("\n🚀 Blog yazısı oluşturma başlıyor: '%s'\n\n", topic)

	// Adım 1: Araştırma
	research, err := b.Step1Research(ctx, topic)
	if err != nil {
		return nil, err
	}
	printSeparator()

	// Adım 2: Taslak
	outline, err := b.Step2Outline(ctx, research)
	if err != nil {
		return nil, err
	}
	printSeparator()

	// Adım 3: Yazma
	draft, err := b.Step3Write(ctx, outline, tone)
	if err != nil {
		return nil, err
	}
	printSeparator()

	// Adım 4: Düzenleme
	_, err = b.Step4Edit(ctx, draft)
	if err != nil {
		return nil, err
	}
	printSeparator()

	fmt.Println("✅ Blog yazısı başarıyla oluşturuldu!")
	return &b.results, nil
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

// printSeparator görsel ayırıcı yazdırır
func printSeparator() {
	fmt.Println("\n" + string(make([]byte, 60)) + "\n")
}

func main() {
	// .env dosyasını yükle
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env dosyası bulunamadı, sistem environment variable'ları kullanılacak")
	}

	// API anahtarını al
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		log.Fatal("GOOGLE_API_KEY environment variable tanımlanmamış")
	}

	ctx := context.Background()

	// Blog writer chain oluştur
	blogWriter, err := NewBlogWriterChain(ctx, apiKey)
	if err != nil {
		log.Fatalf("BlogWriterChain oluşturulamadı: %v", err)
	}
	defer blogWriter.Close()

	// Blog yazısı oluştur
	topic := "Yapay Zeka ve Etik: Gelecekte Karşılaşacağımız Zorluklar"
	tone := "bilgilendirici ve düşündürücü"

	results, err := blogWriter.GenerateBlog(ctx, topic, tone)
	if err != nil {
		log.Fatalf("Blog oluşturma hatası: %v", err)
	}

	// Final sonucu göster
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("FİNAL BLOG YAZISI:")
	fmt.Println(strings.Repeat("=", 60) + "\n")
	fmt.Println(results.Final)

	// İsterseniz ara adımları da görebilirsiniz
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("ARAŞTIRMA NOTLARI:")
	fmt.Println(strings.Repeat("=", 60) + "\n")
	fmt.Println(results.Research)
}
