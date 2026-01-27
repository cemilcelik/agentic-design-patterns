# Go ve Gemini ile Ajan Tasarım Kalıpları (Agentic Design Patterns)

Bu depo, **Go (Golang)** ve **Google Gemini API** kullanılarak çeşitli **Ajan Tasarım Kalıplarının** (Agentic Patterns) uygulanmasını gösterir.

Bu kalıplar, Büyük Dil Modelleri (LLM'ler) ile uygulama mantığı arasındaki etkileşimi yapılandırarak daha güvenilir, ölçeklenebilir ve akıllı yapay zeka sistemlerinin nasıl oluşturulacağını keşfeder.

## 📂 Proje İçeriği ve Mimari

Aşağıda, bu projede yer alan tasarım kalıpları ve işleyiş şemaları bulunmaktadır.

### 1. Prompt Chaining (Zincirleme İstemi)
Karmaşık bir görevi (örneğin blog yazısı yazmak) doğrusal, yönetilebilir adımlara böler. Bir adımın çıktısı, diğerinin girdisi olur.

```mermaid
graph LR
    A[Girdi: Konu] -->|Adım 1| B(Araştırma Ajanı)
    B -->|Bulgular| C(Taslak Ajanı)
    C -->|Taslak| D(Yazar Ajanı)
    D -->|Ham Metin| E(Editör Ajanı)
    E --> F[Final Çıktı]
    style A fill:#f9f,stroke:#333,stroke-width:2px
    style F fill:#9f9,stroke:#333,stroke-width:2px
```

---

### 2. Routing (Yönlendirme)
Gelen kullanıcı isteğini analiz eder ve en uygun işleyiciye (veya modele) yönlendirir.

```mermaid
graph TD
    User[Kullanıcı Sorgusu] --> Router{Router LLM}
    Router -->|Teknik Sorun| Tech[Teknik Destek]
    Router -->|Fatura/Ödeme| Bill[Fatura Departmanı]
    Router -->|Satış| Sales[Satış Danışmanı]
    Router -->|Diğer| Gen[Genel Asistan]
    Tech --> Response[Yanıt]
    Bill --> Response
    Sales --> Response
    Gen --> Response
```

---

### 3. Parallelization (Paralelleştirme)
Bir görevi bağımsız parçalara ayırır, bunları aynı anda (concurrent) işler ve sonuçları birleştirir.

```mermaid
graph TD
    Input[Girdi] --> Split((Dağıtıcı))
    Split --> TaskA[Görev A: Güvenlik Kontrolü]
    Split --> TaskB[Görev B: Performans Analizi]
    Split --> TaskC[Görev C: Maliyet Hesabı]
    TaskA --> Join((Birleştirici))
    TaskB --> Join
    TaskC --> Join
    Join --> Output[Birleştirilmiş Rapor]
```

### 4. Reflection (Düşünme/İyileştirme)
Modelin kendi ürettiği çıktıyı eleştirmesini ve hataları düzelterek yeniden üretmesini sağlar (Self-Correction).

```mermaid
graph TD
    Start[Girdi] --> Draft[Taslak Oluştur]
    Draft --> Critique{Eleştir/Kontrol Et}
    Critique -->|Hatalı/Eksik| Feedback[Geri Bildirim Oluştur]
    Feedback --> Draft
    Critique -->|Yeterli| Final[Final Çıktı]
```

---

### 5. Tool Use (Araç Kullanımı)
Modelin kendi bilgisi dışındaki verilere erişmesi veya eylem gerçekleştirmesi için harici fonksiyonları çağırmasıdır.

```mermaid
graph LR
    User[Sorgu] --> Model{Gemini Model}
    Model -->|Fonksiyon Çağrısı| Tool[Hava Durumu API]
    Tool -->|Veri| Model
    Model -->|Nihai Yanıt| User
```

---

### 6. Planning (Planlama)
Model, görevi doğrudan yapmak yerine önce adım adım bir plan hazırlar, sonra bu planı sırasıyla uygular.

```mermaid
graph TD
    Goal[Girdi: Hedef/Görev] -->|Plan İsteği| Planner{Planlayıcı Ajan}
    Planner -->|JSON Plan| Plan[Adım Adım İş Planı]
    
    subgraph Execution [Yürütme Döngüsü]
        Plan -->|Adım 1| Step1[Adım 1'i Yürüt]
        Step1 -->|Sonuç + Context| Step2[Adım 2'yi Yürüt]
        Step2 -->|Sonuç + Context| StepN[Adım N'i Yürüt]
    end
    
    StepN -->|Tüm Sonuçlar| Combiner{Sonuç Birleştirici}
    Combiner --> Output[Final Çıktı: Blog Yazısı]
```

---

## 🚀 Başlarken

### Gereksinimler

- **Go 1.25** veya üzeri
- Bir **Google Gemini API Anahtarı** ([aistudio.google.com](https://aistudio.google.com/) adresinden alabilirsiniz)

### Kurulum

1.  **Depoyu klonlayın:**
    ```bash
    git clone https://github.com/kullaniciadiniz/agentic-design-patterns.git
    cd agentic-design-patterns
    ```

2.  **Bağımlılıkları yükleyin:**
    ```bash
    go mod download
    ```

3.  **Ortam Değişkenlerini Ayarlayın:**
    Kök dizinde `.env` dosyası oluşturun:
    ```bash
    cp .env.example .env
    ```
    Ve içine API anahtarınızı yapıştırın:
    ```env
    GOOGLE_API_KEY=AIzaSy...
    GEMINI_MODEL=gemini-2.0-flash-exp
    ```

## 🏃‍♂️ Kullanım (Makefile ile)

Projeyi kolayca çalıştırmak için `make` komutlarını kullanabilirsiniz:

```bash
make help                # Tüm komutları listeler
make run-chain           # Prompt Chaining desenini çalıştırır
make run-routing         # Routing desenini çalıştırır
make run-parallel        # Parallelization desenini çalıştırır
```

Veya doğrudan Go komutu ile:

```bash
go run 01-prompt-chaining-pattern/main.go
```

## 🛠️ Kullanılan Teknolojiler

- **Go** (Golang)
- **Google Generative AI Go SDK**
- **Mermaid.js** (Diyagramlar için)
