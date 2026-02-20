# Go ve Gemini ile Ajan Tasarım Kalıpları (Agentic Design Patterns)

Bu depo, **Go (Golang)** ve **Google Gemini API** kullanılarak çeşitli **Ajan Tasarım Kalıplarının** (Agentic Patterns) uygulanmasını gösterir.

Bu kalıplar, Büyük Dil Modelleri (LLM'ler) ile uygulama mantığı arasındaki etkileşimi yapılandırarak daha güvenilir, ölçeklenebilir ve akıllı yapay zeka sistemlerinin nasıl oluşturulacağını keşfeder.

## ⚠️ Uyarı ve Bilgilendirme

Bu proje tamamen **eğitim ve öğrenme amaçlı** ("Code Kata") hazırlanmıştır.

- **Üretim Ortamı İçin Değildir:** Kodlar, temel tasarım desenlerini (pattern) göstermek için basitleştirilmiştir. Hata yönetimi, güvenlik kontrolleri ve performans optimizasyonları üretim standartlarında olmayabilir.
- **Token Kullanımı:** Örnekler, LLM çağrılarını tekrarlayan döngüler veya büyük bağlam pencereleri (Context Window) kullanabilir. Bu durum, API kotanızın hızlı tükenmesine veya beklenmedik maliyetlere yol açabilir.
- **Hata Payı:** LLM'ler (Gemini dahil) deterministik değildir; yani aynı kodla farklı zamanlarda farklı sonuçlar üretebilirler. Bazen halüsinasyon görebilir veya hatalı kod/yanıt döndürebilirler.
- **Veri Güvenliği:** API anahtarlarınızı asla `git` geçmişine eklemeyin. `.env` dosyalarını `.gitignore` içinde tuttuğunuzdan emin olun.


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

### 7. Multi-Agent Collaboration (Çoklu Ajan İşbirliği)
Farklı uzmanlıklara sahip ajanların bir koordinatör yönetiminde ortak bir amacı gerçekleştirmesidir.

```mermaid
graph TD
    User[Görev] --> Coord{Koordinatör}
    
    subgraph Flow [Koordinasyon Akışı]
        Coord -->|1- Konuyu Araştır| Res[Araştırmacı Ajan]
        Res -->|Ham Veriler| Coord
        
        Coord -->|2- Verileri Analiz Et| Analyst[Analist Ajan]
        Analyst -->|İçgörüler| Coord
        
        Coord -->|3- Hepsini Düzenle| Writer[Yazar Ajan]
        Writer -->|Final Rapor| Coord
    end
    
    Coord --> Final[Nihai Çıktı]
```

---

### 8. Memory Management (Bellek Yönetimi)
Müşteri etkileşimlerini, tercihlerini ve geçmiş konuşmaları saklayarak bağlam farkındalığı sağlar. Kişiselleştirilmiş yanıtlar üretmek için geçmiş bilgileri kullanır.

```mermaid
graph TD
    Query[Kullanıcı Sorgusu] --> MemCheck{Bellek Kontrolü}
    MemCheck -->|Geçmiş Var| Context[Bağlam Oluştur]
    MemCheck -->|Geçmiş Yok| NewContext[Yeni Bağlam]
    Context --> Memory[(Müşteri Hafızası)]
    Memory -->|Tercihler| Context
    Memory -->|Geçmiş Etkileşimler| Context
    NewContext --> LLM[LLM İşleme]
    Context --> LLM
    LLM --> Response[Yanıt]
    Response --> SaveMemory[Hafızayı Güncelle]
    SaveMemory --> Memory
```

---

### 9. Advanced Memory Management (Gelişmiş Bellek Yönetimi)
Embedding tabanlı benzerlik araması, farklı bellek türleri (kısa/uzun vadeli, episodik, semantik, prosedürel) ve otomatik bellek birleştirme ile gelişmiş hafıza yönetimi sunar.

```mermaid
graph TD
    User[Kullanıcı] -->|Sorgu| QueryProcess[Sorgu İşleme & Embedding]
    
    subgraph MemoryStore [Hafıza Deposu]
        ST[(Kısa Vadeli)]
        LT[(Uzun Vadeli)]
    end
    
    QueryProcess -->|Vektör Arama| Retriever{Alakalı Olanları Getir}
    Retriever <--> ST & LT
    
    Retriever -->|En İlgili Kayıtlar + Profil| Context[Bağlam Oluştur]
    Context --> LLM{Gemini Model}
    
    LLM -->|Yanıt| Response[Yanıt Üretimi]
    Response -->|Çıktı| User
    
    %% Kayıt Akışı
    QueryProcess -.->|Sorguyu Kaydet| ST
    Response -.->|Yanıtı Kaydet| LT
    
    %% Arka Plan İşlemi
    subgraph Background [Arka Plan: Konsolidasyon]
        ST -.->|Periyodik Kontrol| Logic{Benzerlik Analizi}
        Logic -->|Grupla & Özetle| Merger[Birleştirici Ajan]
        Merger -->|Kalıcı Bilgi Olarak Sakla| LT
    end
```

---

### 10. Learning and Adaptation (Öğrenme ve Adaptasyon)
RAG (Retrieval-Augmented Generation), Memory-Based Learning ve Online Learning tekniklerini birleştirerek sürekli öğrenen ve adapte olan bir AI sistemi. Müşteri etkileşimlerinden öğrenir, bilgi tabanını dinamik olarak günceller ve kişiselleştirilmiş yanıtlar üretir.

```mermaid
graph TD
    Query[Müşteri Sorusu] --> Retrieve[Bilgi Tabanı Araması]
    Retrieve --> KB[(Bilgi Tabanı)]
    Query --> Memory{Müşteri Geçmişi}
    Memory --> History[(Etkileşim Hafızası)]
    KB --> Context[Bağlam Oluştur]
    History --> Context
    Context --> LLM[LLM: Yanıt Üret]
    LLM --> Response[Yanıt]
    Response --> Feedback{Müşteri Geri Bildirimi}
    Feedback -->|Memnuniyet Skoru| Store[Etkileşimi Kaydet]
    Store --> History
    NewDoc[Yeni Bilgi] --> Update[Bilgi Tabanını Güncelle]
    Update --> KB
    style Query fill:#f9f,stroke:#333,stroke-width:2px
    style Response fill:#9f9,stroke:#333,stroke-width:2px
    style KB fill:#bbf,stroke:#333,stroke-width:2px
    style History fill:#fbb,stroke:#333,stroke-width:2px
```

---

### 11. Model Context Protocol (MCP) (Model Bağlam Protokolü)
Yapay zeka modellerini harici veri kaynaklarına (Resources) ve araçlara (Tools) standartlaştırılmış, bağımsız bir protokol üzerinden bağlar. Bu, güvenlik sınırlarını korumanızı ve modüler bir mimari kurmanızı sağlar.

```mermaid
graph TD
    User[Görev İstendi] --> Agent{AI Agent}
    
    subgraph Protocol [MCP İstemci/Sunucu Protokolü]
        Agent -->|Kaynak Listesi Talebi| MCPServer[MCP Sunucusu]
        MCPServer -->|Bağlam/Veri| Agent
        Agent -->|Araç Çalıştırma Talebi| MCPServer
        MCPServer -->|Araç Sonucu| Agent
    end
    
    Agent -->|Cevap| User
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

