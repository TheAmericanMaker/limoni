<p align="center">
  <img src="assets/logo.png" alt="Limoni Logo" width="180" />
</p>

<h1 align="center">🍋 Limoni</h1>

<p align="center">
  <strong>Go Dili İçin Ultra Hızlı, Sıfır Bellek Tahsisatlı (Zero-Alloc), İş Parçacığı Güvenli Modern TUI Motoru.</strong>
</p>

<p align="center">
  <a href="https://github.com/thebanri/limoni/actions"><img src="https://img.shields.io/github/actions/workflow/status/thebanri/limoni/ci.yml?branch=main&style=flat-square&logo=github" alt="Derleme Durumu"></a>
  <a href="https://pkg.go.dev/github.com/thebanri/limoni"><img src="https://img.shields.io/badge/go.dev-referans-007d9c?style=flat-square&logo=go&logoColor=white" alt="Go.Dev Referans"></a>
  <a href="https://golang.org"><img src="https://img.shields.io/badge/go-%3E%3D%201.25-blue?style=flat-square&logo=go" alt="Go Sürümü"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/lisans-Apache_2.0-blue?style=flat-square" alt="Lisans"></a>
  <a href="#-performans-ve-kıyaslamalar"><img src="https://img.shields.io/badge/tahsisat-0_B%2Fop-brightgreen?style=flat-square" alt="Sıfır Bellek Tahsisatı"></a>
</p>

<p align="center">
  <strong>Dil / Language:</strong>
  <a href="README.md">English</a> •
  <a href="README_TR.md">Türkçe</a>
</p>

<p align="center">
  <a href="#-neden-limoni">Neden Limoni?</a> •
  <a href="#-vitrin--canlı-demolar">Vitrin</a> •
  <a href="#-temel-özellikler">Özellikler</a> •
  <a href="#-hızlı-başlangıç">Hızlı Başlangıç</a> •
  <a href="#-dokümantasyon">Dokümantasyon</a> •
  <a href="#-zengin-widget-ekosistemi">Widget'lar</a> •
  <a href="#-örnek-uygulamalar">Örnekler</a>
</p>

---

<p align="center">
  <strong><a href="https://thebanri.github.io/limoni/">▶ Limoni'yi tarayıcında dene</a></strong><br>
  <sub>Aynı motor, WebAssembly'ye derlenmiş ve xterm.js üzerinde çalışıyor — kurulum yok.</sub>
</p>

---

## ⚡ Genel Bakış

**Limoni**, Go dili için sıfırdan tasarlanmış kurumsal düzeyde, yüksek performanslı bir Terminal Kullanıcı Arayüzü (TUI) motorudur. Veri yoğun izleme panelleri, DevOps araçları ve modern CLI uygulamaları için Go'nun geliştirici ergonomisini Rust benzeri ham render hızıyla buluşturur.

**1D düz hücre matrisi**, **sıfır bellek tahsisatlı sıcak yollar (zero-alloc hot-paths)** ve **yüksek performanslı diferansiyel ANSI motoru (~50 µs tam ekran diff, 0 B/op)** sayesinde Go Garbage Collector'ını tetiklemeden yüksek FPS'te pürüzsüz çizim sağlar.

---

## 💡 Neden Limoni?

| Özellik / Hedef | 🍋 Limoni (Go) | 🫧 Bubble Tea **v1** + Lip Gloss v1 (Go) | 🌈 Bubble Tea **v2** + Ultraviolet (Go) | 🐀 Ratatui **0.30** (Rust) |
| :--- | :--- | :--- | :--- | :--- |
| **Dil ve Araçlar** | **Go (Yerel)** | Go (Yerel) | Go (Yerel) | Rust (Yerel) |
| **Render Mimarisi** | **1D Düz Matris + Adaptif ANSI Diff** | String birleştirme / TEA | Hücre tamponu + ncurses tarzı diff | Çift Tamponlu Immediate Mode |
| **Kritik Yol Tahsisatı**| **`0 B/op` (Sıfır Alloc)** | Yüksek heap tahsisatı | Azaltılmış; sıfır-alloc bir tasarım hedefi değil — Ultraviolet her glif için bir `Cell` tahsis ediyor | Stack / RAII |
| **Düzen Paradigması** | **Bildirimsel Flexbox & Yığın Çözücü** | String dilimleme (`JoinHorizontal/Vertical`) | Cassowary kısıt çözücü | Kısıt çözücü (Constraint solver) |
| **Fare Etkileşimi** | **Hücresel Koordinat & Z-Index Yönlendirme** | Yok (manuel koordinat hesabı) | SGR fare olayları; dahili hit-testing yok | Manuel koordinat |
| **Çift Tampon & Diff** | **Mikrosaniye altı diff + Adaptif tam akış** | Yok (tüm string stdout'a dökülür) | Hücre diff + `ECH`/`REP`/`ICH`/`DCH` + kaydırma optimizasyonu | Çift tamponlu diff |
| **Grapheme Cluster** | **UAX #29 cluster'ları, Unicode 17.0, resmî 766 kırılım testinin tamamı geçiyor** + Mod 2027 isteği; desteklemeyen terminaller için her cluster'dan sonra imleç yeniden konumlanır | `uniseg` | `uniseg` + Mod 2027 müzakeresi | `unicode-width` |
| **Yetenek Tespiti** | Yalnızca ortam değişkenleri | Ortam / terminfo | Çalışma anında sorgulama (terminfo'suz) | terminfo / crossterm |
| **Büyük Veri / Tablolar**| **1M satır sanallaştırma (sürekli kaydırma altında ~2,6 ms/kare)** | Yüksek GC yükü | v1'e göre iyileştirilmiş | Her karede tüm satırları yeniden kurar — `Table` satır iterator'ının sahibidir |
| **3D & Vektör Grafikleri**| **Dahili 3D (OBJ/STL/PLY/GLB) & Shaders** | Harici eklenti gerekir | Harici eklenti gerekir | Eklenti gerekir |
| **Erişilebilirlik (A11y)** | **Dahili Semantik Ağaç ve Ekran Okuyucu** | Kısıtlı / Manuel | Kısıtlı / Manuel | Deneysel |
| **Harici Bağımlılık** | **2 (`golang.org/x/sys`, `golang.org/x/crypto`)** | ~15 dolaylı modül | ~15 dolaylı modül | crates.io grafiği |
| **Eşzamanlılık (Concurrency)** | **Kilit-Serbest Kanallar / İş Parçacığı Güvenli** | Tek iş parçacıklı TEA | Tek iş parçacıklı TEA | Manuel iş parçacığı yönetimi |

> **Bubble Tea v2 sütunu hakkında:** bu satırlar Limoni'nin kendi ölçümlerinden değil, üst akış dokümantasyonundan alınmıştır. Charm, render motorunu hücre tabanlı diff yapan [Ultraviolet](https://github.com/charmbracelet/ultraviolet) üzerine yeniden inşa etti; dolayısıyla Limoni'nin **v1**'e karşı açtığı mimari fark **v2** için olduğu gibi geçerli değildir.
>
> Ultraviolet artık burada **ölçülüyor** ve karşılaştırılabilir render iş yüklerinde Limoni 1,9–20 kat daha hızlı, üstelik kare başına belirgin biçimde daha az bayt yayıyor ([§2.4](docs/benchmark-methodology.md#24-ultraviolet)). Bu bir **Bubble Tea v2 sonucu değildir**: bir v2 programı ayrıca kendi çalışma zamanını, mesaj dağıtımını ve view kurulumunu da öder; bunların hiçbiri burada ölçülmüyor. Bu depodaki hiçbir koşucu Bubble Tea v2'yi link etmiyor, dolayısıyla v2'nin kendisine karşı her performans iddiasını kanıtlanmamış sayın.

### 🍋 Limoni Composable (Lego UI) vs. 🎀 Charm Lip Gloss **v1**

**Lip Gloss v1** Go ekosisteminde bildirimsel stili popülerleştirmiş olsa da, string birleştirmeye dayalı mimarisi yüksek frekanslı ve etkileşimli modern TUI uygulamalarında yapısal kısıtlamalar getiriyordu. Aşağıdaki karşılaştırma **özellikle v1**'e karşıdır:

> ⚠️ **Lip Gloss v2 bu tabloyu değiştiriyor.** v2, ham string birleştirme yerine [Ultraviolet](https://github.com/charmbracelet/ultraviolet) hücre tamponu üzerine kuruludur; bu yüzden aşağıdaki "Veri İlkesi", "Render Hattı" ve "Ekran Kırpma" satırları güncel Charm yığınını tarif etmez. Limoni'nin v2'ye karşı kalan yapısal üstünlükleri hit-testing, sanallaştırma, dahili 3D ve bağımlılık ayak izidir — string-hücre mimarisi değil.

| Yetenek | 🍋 Limoni Composable (`component`) | 🎀 Charm Lip Gloss **v1** |
| :--- | :--- | :--- |
| **Veri İlkesi** | **16 baytlık önbellek uyumlu `Cell` yapısı** | Ham ANSI kaçışlı metin (`string`) |
| **Kritik Yol Bellek Tahsisi** | **`0 B/op` (0 allocs/op)** layout ve render | Yüksek tahsisat oranı (~Yüzlerce KB - MB/sn) |
| **Düzen Modeli** | **Gerçek Flexbox & Grid kısıt çözücü** | String dilimleme (`JoinHorizontal`, `JoinVertical`) |
| **Boyut Kısıtları** | **Orantısal `Flex`, `Ratio`, `Min`, `Max`** | Sadece sabit manuel karakter genişlikleri |
| **Fare Hit-Testing** | **Otomatik uzamsal sınırlar & z-index yönlendirme** | Yok (manuel koordinat ve karakter hesabı gerekir) |
| **Ekran Kırpma (Clipping)** | **Hücre seviyesinde dikdörtgensel uzamsal kırpma** | String kesme (bozuk ANSI kaçış dizilerine yol açar) |
| **Z-Index & Katmanlar** | **Donanım benzeri katman yığını & modal izole** | Satır satır string yamama (`PlaceOverlay`) |
| **Render Hattı** | **Çift tamponlu ANSI diffing (`~14 µs` seyrek, `~50 µs` tam ekran)** | Tüm terminale string dökme (ekranda titreme yapar) |
| **Geçiş Köprüsü** | **`compat/bubbletea` akıcı stil oluşturucu** | Charm ekosistemi yerel standardı |

#### Neden Sıfır Bellek Tahsisatlı Mimari Önemlidir?
1. **Garbage Collector Donmalarını (GC Stutter) Yok Eder**: Lipgloss her kenarlık, boşluk ve yatay birleştirme için bellekte yeni string nesneleri tahsis eder. 60 FPS çalışan hareketli bir ekranda bu durum saniyede yüz binlerce nesne üreterek Go GC'sini devreye sokar ve arayüzde mikro donmalara (stutter) yol açar. Limoni bileşenleri çağrı yığınında (call stack) çalışır ve doğrudan yeniden kullanılan 1D tampona yazar; **sıfır bellek tahsisatı (`0 B/op`)** garantilenir.
2. **Kutudan Çıkan Fare ve Tıklama Desteği**: Lipgloss sadece düz bir metin ürettiğinden kullanıcının nereye tıkladığını bilemez. Limoni bileşenleri çizildikleri ekran alanını (`cell.Rect`) otomatik kaydeder; tıklama, üzerine gelme (hover), sürükleme ve tekerlek olayları doğrudan ilgili bileşenin callback'ine yönlendirilir.

---

## 🎬 Vitrin & Canlı Demolar

### 🎮 3D Vektör ve Model İşleme Motoru
Terminal hücrelerinde 60+ FPS hızında gerçek zamanlı 3D yazılımsal rasterizasyon. `.obj`, `.stl` ve `.ply` model desteği, derinlik tamponlu Gouraud gölgelendirme, Lambertian aydınlatma ve etkileşimli fare/klavye kamera yörünge kontrolleri.

<p align="center">
  <img src="assets/3d.gif" alt="Limoni 3D Model İşleme" width="100%" />
</p>

```bash
# Yerel çalıştırma (-fps bayrağı veya [F] tuşu ile 240 FPS destekler):
go run ./examples/3d_viewer -fps 240

# Veya repoyu indirmeden doğrudan terminalinizden çalıştırın:
go run github.com/thebanri/limoni/examples/3d_viewer@latest -fps 240
```

---

### 📁 Süper Dosya Gezgini & Görsel Önizleme
Açılır/kapanır klasörler, hiyerarşik kılavuz çizgileri, dosya meta verileri ve yerleşik TrueColor yarım-blok (half-block) görsel önizleme desteğine sahip gelişmiş dosya ağacı bileşeni (`widgets.TreeView`).

<p align="center">
  <img src="assets/treeview.gif" alt="Limoni Dosya Gezgini ve Görsel Önizleme" width="100%" />
</p>

```bash
go run ./examples/treeview
```

---

### 📊 Yüksek Çözünürlüklü Grafikler & Veri Görselleştirme
Alt-piksel Braille eğrileri (`widgets.LineChart`), dikey gradyan spektrum çubukları (`widgets.BarChart`) ve pasta/halka dağılımları (`widgets.PieChart`) ile sıfır bellek tahsisatlı yüksek frekanslı telemetri görselleştirme.

<p align="center">
  <img src="assets/chart.gif" alt="Limoni Grafikler ve Veri Görselleştirme" width="100%" />
</p>

```bash
go run ./examples/charts
```

---

## ✨ Temel Özellikler

* 🚀 **Ultra Hızlı ANSI Diffing**: Ekrandaki değişiklikleri tespit edip tam ekran yenilemede ~50 µs (~19.800 FPS) sürede sıfır bellek tahsisatıyla minimum ANSI kaçış dizilerini terminale gönderir; ekran değişmediğinde ~2 ns içinde anında döner.
* 📉 **Koşu Sıkıştırmalı Çıktı**: Boş diziler `ECH`/`EL`, tekrarlayan glifler `REP` olur; bu, tam ekran yenilemeyi 4.897 bayttan **377 bayta** indiriyor — aynı karede Ratatui 0.30.2'nin 15 kat altında. SSH üzerinde hissettiğiniz şey CPU süresi değil, yayılan bayttır.
* 📃 **Inline Render**: `limoni.WithInline(height)` normal ekran tamponunda bir bant içinde çizer — alternatif ekran yok, scrollback korunur, çıkıştan sonra çıktı ekranda kalır. `gum` ve CI ilerleme göstergelerinin kurulu olduğu mod.
* 📦 **1D Düz Tampon (Flat Buffer)**: Bellek parçalanmasını önler ve CPU L1/L2 önbellek erişimini maksimize eder.
* 🎨 **24-Bit TrueColor & Otomatik Geri Dönüş**: TrueColor desteği olmayan terminallerde otomatik 256 ve 16 renk dönüşümü.
* 📐 **Esnek Flexbox & Grid Düzeni**: Proportional, Fixed, Min/Max, GridArea ve boyut pazarlığı (negotiation) desteği.
* 🎬 **60 FPS Animasyon & Fizik Motoru**: Yay (spring) fizikleri, renk enterpolasyonu ve akıcı easing eğrileri.
* 🕶️ **Dahili 3D & Vektör Grafik Motoru**: `.obj`, `.stl`, `.ply` 3D modelleri Gouraud/Lambertian gölgelendirme ile doğrudan terminalde işleme.
* ♿ **Dahili Erişilebilirlik (A11y)**: Ekran okuyucular için semantik gezinme ağacı ve satır satır denetim modu.
* 🤖 **Semantik Otomasyon**: Çalışan bir uygulamayı ekran koordinatı yerine seçiciyle sürün — aşağıya bakın.
* 🔲 **Otomatik Kenarlık Birleştirme**: Komşu `Block`'lar `MergeBorders` ile tek bir kenarı paylaşır ve `┬ ┼ ├ ┤ ┴` ile birleşir; hücrede zaten duran box-drawing parçalarının birleşimi alınarak — string birleştiren bir render motorunun çoktan üzerine yazdığı bir bilgi.

---

## 🤖 Semantik Otomasyon

Bugün bir terminal uygulamasını otomatikleştiren her araç — [termwright](https://github.com/fcoury/termwright), [mcp-tui-test](https://github.com/GeorgePearse/mcp-tui-test) — süreci bir sözde terminale sarıp çizilen karakter ızgarasını parse eder. Başka çareleri yoktur: alttaki uygulamanın sunacak bir semantiği yoktur. Dolayısıyla test, bir metnin bir koordinatta olduğunu doğrular ve düzen bir sütun kaydığı anda kırılır.

Limoni zaten ekran okuyucular için her karede bir semantik ağaç kuruyor. `WithAutomation` aynı ağacı bir Unix soketinde sunar; böylece

```
"Submit" metni 42,7'de mi?  →  42,7'ye tıkla
```

yerine

```go
client.Click(automation.Selector{Role: "button", Label: "Submit"})
```

yazarsınız. Seçici yeniden düzenlemeye, boyutlandırmaya ve stil değişimine dayanır, çünkü hiçbir şeyin nerede çizildiğinden söz etmez.

```go
// Uygulama açık bir politikayla ister. -tags limoni_debug ile derleyin.
limoni.Run(draw, limoni.WithAutomation("/run/user/1000/myapp.sock", limoni.AutomationPolicy{
	AllowInput:   true, // varsayılan kapalı: olmadan istemci yalnızca gözlemleyebilir
	ExposeScreen: true, // varsayılan kapalı: ızgara ekrandaki her karakteri içerir
}))
```

```go
// Test ya da ajan sürer.
client, _ := automation.Dial("/run/user/1000/myapp.sock")
defer client.Close()

list, _ := client.WaitFor(automation.Selector{Role: "list"}, 3*time.Second)
// list.Value == "beta", list.Position == 2, list.SetSize == 3

client.Key("down")
client.Type("hello")
client.Click(automation.Selector{Role: "button", Label: "Submit"})

screen, _ := client.Screen() // ağacın ifade edemediği doğrulamalar için ham ızgara
```

Protokol satır ayrımlı JSON, yani hata ayıklarken `socat` kullanılabilir bir istemcidir. Belirsiz bir seçici **hatadır**, yazı tura değil — iki eşleşen düğmeden sessizce ilkini seçen bir test, ikincisi eklendiği anda yanlış sebeple geçer; hangisini kastettiğinizi `Nth` ile söyleyin.

### Bir yapay zekâ ajanıyla sürmek (MCP)

`cmd/limoni-mcp` aynı ağacı [Model Context Protocol](https://modelcontextprotocol.io) konuşan her ajanın önüne koyar — Claude Code, Claude Desktop, Cursor ve diğerleri. Standart kütüphane dışında bağımlılığı olmayan bir köprüdür: bir yanda stdio üzerinden MCP, öbür yanda uygulamanın soketi.

```bash
go install github.com/thebanri/limoni/cmd/limoni-mcp@latest

# Bunun için yazılmış örnekte deneyin:
# Bir terminalde — uygulama $XDG_RUNTIME_DIR/limoni-checklist.sock üzerinde dinler:
go run -tags limoni_debug ./examples/agent_checklist
# Bir diğerinde:
claude mcp add limoni -- limoni-mcp -socket "$XDG_RUNTIME_DIR/limoni-checklist.sock"
```

Ajan sekiz araç alır: `tree`, `find`, `click`, `press_key`, `type_text`, `wait_for`, `screen` ve `status`. Pratikte iki ayrıntı önemli:

- **Girdi araçları, uygulama yeniden çizdikten sonraki ağacı döndürür.** Girdi asenkrondur — soket bir tuşu, yol açtığı kareden önce onaylar — bu yüzden `click` ağacın değişip durulmasını bekler ve onu döndürür. Ajan hiçbir zaman kendi tıklamasından *önceki* ekran üzerinden akıl yürütmez. Hiçbir şey değişmediyse sonuç bunu söyler, anlayabildiğinde de nedenini: alan gizlidir ya da politika girdi değerlerini saklıyordur.
- **Hatalar açıklamadır.** Belirsiz bir seçici, yanlış yazılmış bir argüman ya da politika reddi, modelin üzerine iş yapabileceği bir metin olarak döner — `label="Remove" matches 2 nodes; set nth to choose one` — protokol hatası olarak değil.

Köprü aşağıdaki sınırların hiçbirini değiştirmez: yalnızca uygulamanın politikasının izin verdiğini yapabilir ve port açmaz. Girdi araçları yıkıcı (destructive) olarak işaretlidir; yan etkiden önce soran bir istemci soracaktır.

Bir denemede Claude Code 2.1.270, yalnızca hedef söylenerek, bir görev ekledi, üçünü işaretledi, bir deploy token'ı yazdı ve 17 araç çağrısında deploy etti. Denemenin kaydındaki hiçbir araç sonucunda token geçmiyor. Bu bir denemedir, benchmark değil.

### Playwright gibi test etmek (`uitest`)

Aynı ağaç, Playwright'ın web sayfaları için yazdığı testlere benzeyen testler yazmayı sağlar. `uitest` widget'ları rol, etiket ve ID ile bulur, onlar üzerinde işlem yapar ve **bekleyen** kontrollerle doğrular: bir işlem, konumlandırıcısı tam olarak bir widget'la eşleşene kadar bekler; bir doğrulama tutana kadar yeniden dener. Böylece test hiçbir zaman `sleep` kullanmaz. Bir kontrol başarısız olduğunda mesaj neyin beklendiğini, neyin görüldüğünü ve son karenin tüm semantik ağacını gösterir.

```go
func TestReleaseFlow(t *testing.T) {
	app := newChecklist()
	page := uitest.Run(t, 80, 24, app.draw) // süreç içinde: terminal yok, build tag yok

	page.GetByRole("input", "New task").Type("Tag v1.0")
	page.GetByRole("button", "Add task").Click()

	rows := page.GetByRole("list-item", "").Within(page.GetByRole("list", "Tasks"))
	page.Expect(rows).ToHaveCount(3)
	rows.Nth(-1).Click()
	page.Press("space")

	page.Expect(page.GetByID("status")).ToContainLabel("Added")
	page.Expect(page.GetByRole("dialog", "")).Not().ToBeVisible()
}
```

```
uitest: expected id="status" to have label "Deployed with 3 tasks complete.": got label "Blocked: the deploy token is empty." after 5s
last frame:
  input#new-task "New task" bounds=2,2 36x1
  button#add "Add task" bounds=40,2 14x1
  …
```

Her aksiyon loglanır; bir hata, ona götüren adımlarla birlikte gelir. `uitest.WithSlowMo` ise testi izleyen biri için yavaşlatır. `uitest.Connect` ile çalışan bir uygulamaya yöneltilen aynı test onu ekranda canlı sürer — [`examples/agent_checklist`](examples/agent_checklist) içindeki `TestLiveDemo` tam olarak bunu yapar.

Tek API, üç hedef: immediate-mode çizim fonksiyonu için `uitest.Run`, gerçek mesaj döngüsünden (komutlar dahil) geçen declarative model için `uitest.Program`, otomasyon soketi üzerinden çalışan bir binary için `uitest.Connect`. [`examples/agent_checklist`](examples/agent_checklist) bununla test ediliyor ve bir ajanın `limoni-mcp` üzerinden sürdüğü uygulamanın ta kendisi.

Listeler görünür satırlarını `list-item` alt düğümleri olarak sunar; bir satıra tuş basışı sayarak değil, metniyle ulaşılır.

> [!WARNING]
> **Bu, çalışan bir sürece kontrol kanalı açar.** Her biri kendi başına kapalı başarısız olan katmanlar hâlinde kuruldu.
>
> - **Release derlemelerinde yok.** Geçit yalnızca `-tags limoni_debug` ile derlenen binary'lerde bulunur. Etiket olmadan `WithAutomation`, `Run`'ın `ErrAutomationNotCompiled` döndürmesine yol açar ve soket sunucusu binary'de hiç yer almaz — CI bir release binary derleyip sembol tablosunda otomasyon kodu arıyor. Hiçbir yapılandırma hatası, orada olmayan kodu açamaz.
> - **Varsayılan kapalı.** Sıfır değerli `AutomationPolicy` yalnızca yapıyı gösterir: roller, etiketler, konumlar, sınırlar. Girdi değerleri, ekran görüntüsü ve girdi sentezi her biri kendi alanının açılmasını ister.
> - **Sırlar politika ne derse desin çıkmaz.** `TextInput{Secret: true}` her karakter yerine maske glifi çizer, yani sır hücre tamponuna hiç girmez; düğümü değer taşımaz ve hassas olarak işaretlenir. Geçit hassas değerleri yine de temizler, böylece bunu unutan bir widget sızdırmaz. Seçiciler redakte edilmiş ağaçta çözülür, yani bir istemci `value="…"` eşleşip eşleşmediğine bakarak parola tahmin edemez.
> - **Yalnızca sizin kullanıcınız bağlanabilir.** Linux, macOS ve FreeBSD'de sunucu, soketin 0600 izinlerine ek olarak bağlanan sürecin sahibini çekirdeğe sorar ve başka her kullanıcıyı reddeder. Çekirdeğin bunu söyleyemediği yerlerde — Windows dahil — dosya izinleri tek koruma kalacağı için `AllowUnverifiedPeers` açılmadıkça her bağlantı reddedilir.
> - **Yalnızca Unix soketi, TCP seçeneği yok.** Bilinçli ve yapılandırılamaz: bir port, uygulama kontrolünü makineye erişebilen her şeye açardı.
>
> **Kalan riskler, açıkça:** *aynı kullanıcı olarak* çalışan başka bir süreç yine bağlanabilir — işletim sistemi yerel bir sokette kullanıcıdan güçlü bir kimlik sunmuyor. Ve siz işaretlemedikçe geçit, çizdiğiniz bir paragrafın sır olduğunu bilemez; `ExposeScreen` ekranda ne varsa gönderir. Bir `limoni_debug` binary'sine hata ayıklama konsolu gibi davranın.

---

## ⏺️ Oturum Kaydı ve Yeniden Oynatma

*"Birkaç tuşa bastım, çöktü"* diyen bir hata raporuyla bir şey yapmak zordur. Oturumun kaydıyla değil. `session` paketi declarative bir uygulamayı kaydeder — modelin aldığı her mesajı, `Update`'in gördüğü sırayla, artı her karenin semantik ağacını — ve modele karşı yeniden oynatıp her kareyi doğrular.

```go
rec, _ := session.Create("hata.limoni", model, genislik, yukseklik, session.Policy{})
defer rec.Close()
limoni.RunProgram(ctx, model, limoni.WithProgramObserver(rec))
```

```go
// Sonra, bir testte: kayıt bir regresyon testine dönüşür.
report, err := session.Replay("hata.limoni", func() limoni.Model { return yeniModel() }, session.ReplayOptions{})
if err != nil { t.Fatal(err) }                       // kayda güvenilemedi
if !report.Verified() { t.Fatal(report.Divergence) } // "replay diverged at step 3 ..."
```

Oynatma **karesi farklılaşan ilk adımı** ve kaydedilmiş bir **çökmenin yeniden üretilip üretilmediğini** raporlar — yani aynı dosya önce hatanın hâlâ orada olduğunu, sonra düzeldiğini söyler.

Terminal yerine `Update` noktasında kaydeder; oynatılabilir olmasını sağlayan şey bu: canlı çalışmada zamanlayıcılar ve komutlar yarışır, `Update`'in çağrıldığı yerde kaydetmek bu yarışların nasıl sonuçlandığını dondurur. Bir komutun etkisi kayda ürettiği mesaj olarak girer, yani oynatma bir isteği yeniden göndermez ya da saati okumaz.

> [!WARNING]
> **Sınırlar, açıkça.**
>
> - **Yalnızca declarative mod.** Immediate mode `Run`'da uygulamayla yan etkileri arasında bir sınır yoktur, kaydedilecek bir nokta da yoktur.
> - **`Update` ve `View` deterministik olmalı.** İçlerinde `time.Now` ya da `math/rand` çağıran bir model oynatmada farklılaşır — oynatma bunu tespit edip adımı gösterir ama düzeltemez. Saati `limoni.NowCmd` ile mesaj olarak alın. [`tools/limonivet`](tools/limonivet) bu çağrıları raporlar; CI onu bu depoda çalıştırıyor.
> - **Yalnızca kayıtlı mesajlar.** Bir uygulama mesaj tipi `session.Register` ile kaydedilmedikçe yalnızca adıyla kaydedilir; oynatma ona ulaştığında atlamak yerine yüksek sesle başarısız olur.
>
> **Gizlilik varsayılan olarak kapalı.** Yazılan metin ve yapıştırmalar `RecordText` açılmadıkça `x` olarak yazılır; girdi alanı değerleri `ExposeInputValues` açılmadıkça kayıtlı ağaçlardan düşürülür; hassas alanlar her zaman düşürülür. `RecordText` açıkken bile, gizli bir alan odakta olabileceği her durumda karakterler redakte edilir — *herhangi bir* mesajdan sonraki, bir sonraki kare odağın nereye gittiğini gösterene kadarki pencere dahil; çünkü bir komut sonucu odağı parola alanına Tab kadar kolay taşıyabilir. Dosyalar `0600`, asla üzerine yazılmaz, sağlama toplamlıdır ve değiştirilmişse oynatmada reddedilir.
>
> **Kalan risk:** kaydettiğiniz bir mesaj tipi eksiksiz kaydedilir — sır taşıyan biri için `session.RegisterRedacted` kullanın. Değerler kaydedilmese bile etiketler kaydedilir; *etiketinde* sır gösteren bir widget onu dosyaya koyar. Ve kayıt diskte bir dosyadır: kullanıcılarınızın gördüğünü içerebilecek bir log gibi davranın.

---

## 🚀 Hızlı Başlangıç

### Kurulum

```bash
go get github.com/thebanri/limoni
```

### Örnek Uygulama:

```go
package main

import (
	"os"
	"github.com/thebanri/limoni/core/cell"
	"github.com/thebanri/limoni/core/driver"
	"github.com/thebanri/limoni/core/terminal"
	"github.com/thebanri/limoni/widgets"
)

func main() {
	d := driver.NewDriver(os.Stdin, os.Stdout)
	d.Setup()
	defer d.Close()

	t, err := terminal.New(d)
	if err != nil {
		panic(err)
	}

	d.StartEventLoop()

	t.Draw(func(f *terminal.Frame) {
		f.RenderWidget(widgets.Block{
			Title:         " 🍋 LIMONI TUI ",
			Borders:       widgets.BorderAll,
			BorderSymbols: widgets.SymbolsRounded,
			BorderStyle:   cell.Style{Fg: cell.NewColorRGB(0, 210, 255)},
		}, f.Buffer.Area)
	})

	for ev := range d.Events() {
		if ev.Type == driver.EventKey && ev.Key.Type == driver.KeyEsc {
			return
		}
	}
}
```

---

## 📊 Performans ve Kıyaslamalar (Benchmarks)

> 📐 **Önce [`docs/benchmark-methodology.md`](docs/benchmark-methodology.md) dosyasını okuyun.** Hangi sürümlerin ölçüldüğünü, harness'ın neyi yakalayıp neyi yakalamadığını ve hangi iddiaların henüz kanıtlanmadığını açıklar. Çapraz-framework koşucuları **Ratatui 0.30.2**, **Ultraviolet** (Bubble Tea v2 ve Lip Gloss v2'nin altındaki hücre render motoru) ve **Bubble Tea v1.3.10**'u hedefliyor. **Bubble Tea v2 koşucusu yoktur**: burada hiçbir şey onu link etmiyor, dolayısıyla bu depodaki hiçbir iddia v2'nin kendisi hakkında değildir.

Limoni, standart sanal terminal ortamında (120×40 hücre = 4.800 hücre) gerçek dirty diffing, kısmi güncellemeler, sanal kaydırma ve bellek tahsisatlarını ölçen kapsamlı bir kıyaslama paketine sahiptir.

Testleri yerel ortamınızda çalıştırmak için:
```bash
# Buffer Diff kıyaslamaları (kirli ve temiz kare testleri)
go test ./core/buffer -run '^$' -bench . -benchmem

# Widget ve Düzen kıyaslamaları
go test ./benchmarks -run '^$' -bench . -benchmem

# Çapraz-framework karşılaştırması: her koşucuyu derler, üçünü de üçer kez
# çalıştırır, medyan gecikmeyi koşular arası yayılımla raporlar ve koşucuların
# karşılaştırılamaz işaretlediği her oranı gizler. Ratatui için Rust gerekir.
./benchmarks/compare.sh
```

### Ölçülen sonuçlar

**AMD Ryzen 5 5600 (6Ç/12İ), Linux 6.17, Go 1.27.1, `-count=3`, medyan.** Mutlak
değerler donanıma bağlıdır; anlamlı olan, tek bir makinede commit'ler arasındaki
orandır. Yukarıdaki komutla yeniden üretilebilir.

| Kıyaslama İşlemi | Ölçülen Gecikme | Kare / İşlem Hızı | Bellek Tahsisatı | Açıklama |
| :--- | :--- | :--- | :--- | :--- |
| **`BenchmarkDiff_FullChanges`** | **`~50.5 µs`** | **~19.800 FPS** | **`0 B/op (0 allocs)`** | %100 tam ekran hücre değişimi (4.800 hücre) çift tampon diff işlemi ve ANSI akışı üretimi |
| **`BenchmarkDiff_PartialChanges`** | **`~23.7 µs`** | **~42.200 FPS** | **`0 B/op (0 allocs)`** | %10 ekran alanı değişimi (480 hücre) çift tampon diff işlemi |
| **`BenchmarkDiff_NoChanges`** | **`~1.94 ns`** | **~516.000.000 FPS** | **`0 B/op (0 allocs)`** | Tamponda hiçbir değişiklik olmadığında fast-path ile anında dönüş |
| **`BenchmarkTextHeavyFrame`** | **`~31.7 µs`** | **~31.500 FPS** | **`2 B/op (0 allocs)`** | 120 sütuna yayılan 40 satırlık unicode sembollü ve kelime kaydırmalı metin çizimi |
| **`BenchmarkHundredLayers`** | **`~68.2 µs`** | **~14.700 FPS** | **`1 B/op (0 allocs)`** | 100 katmanlı Block widget çizimi ve değerlendirmesi (Ratatui hundred-layers denklik testi) |
| **`BenchmarkTenThousandRowTable`** | **`~68.5 µs`** | **~14.600 FPS** | **`611 B/op (4 allocs)`** | 10.000 satırlık tabloda aktif imleç kaydırma (scrolling) ve görünür satır çizimi |
| **`BenchmarkOneMillionRowVirtualScroll`**| **`~2.70 ms`** | **~370 FPS** | **`4.9 KB/op (6 allocs)`** | 1.000.000 satırlık sanal veri kaynağında aktif kaydırma ve görünür alan yönetimi |
| **`BenchmarkMouseHitTest`** | **`~63.4 ns`** | **~15.800.000 op/s** | **`0 B/op (0 allocs)`** | 100 tıklama bölgesi üzerinde hiyerarşik uzamsal fare tıklama tespiti |
| **`BenchmarkAsyncUpdateBurst`** | **`~232 ns`** | **~4.310.000 msg/s** | **`7 B/op (0 allocs)`** | Elm çalışma mimarisinde yüksek verimli asenkron mesaj kuyruğu iletimi |

> [!NOTE]
> **Bu değerler iki kez, iki yönde de değişti.** Tablonun önceki hâlinde adı
> geçen donanım sınıfında yeniden üretilemeyen gecikmeler vardı;
> `BenchmarkHundredLayers` 47 µs iddia ediliyor, 159 µs ölçülüyordu. Bunu
> düzeltmek zamanın asıl nerede gittiğini ortaya çıkardı: katmanlı bir karenin
> %39'u `cell.RuneWidth`'te geçiyordu, yazılan her hücre için yirmi aralık
> karşılaştırması dolaşarak. Artık cevabı bir arama tablosundan veriyor ve bu,
> çizim yolunu genel olarak 2–3 kat aşağı çekti. Yani yukarıdaki sayılar hem
> düzeltilmiş hâlden hem de orijinal şişirilmiş iddialardan daha düşük — bu kez
> arkalarında bir profil var. Sıfır-tahsisat garantileri baştan sona korundu.

> [!NOTE]
> **Tablo grapheme cluster desteğinden önceye ait.** Metni bölütlemenin ASCII
> dışı karakterlerde bir maliyeti var; düz ASCII bunu atlayan hızlı yoldan
> geçiyor. Yukarıdaki makinede, değişiklikten önceki commit'e karşı art arda
> ölçüldü (`-count=3`, medyan): `BenchmarkTextHeavyFrame` +%5 (30,7 → 32,3 µs —
> metninde satır başına üç sembol var), `BenchmarkDiff_FullChanges` +%2,
> `BenchmarkHundredLayers` ve `BenchmarkDiff_PartialChanges` değişmedi; hepsi
> hâlâ sıfır tahsisatta. Tablodaki mutlak değerler yeniden ölçülmedi; satırları
> değil oranları karşılaştırın.

> [!NOTE]
> **Şeffaflık ve Mühendislik Dürüstlüğü Garantisi**:
> Sentetik kısayollar, yapay tampon temizlemeleri veya statik sıfır-offset döngüleri kullanılmaz.
> - **Diff Kıyaslamaları**: Hücrelerin her karede bizzat değiştiği kalıcı çift tampon üzerinde çalışır; diff motorunu ve ANSI kodlayıcısını uçtan uca çalıştırır.
> - **Kaydırma Kıyaslamaları**: `Select((i * 7) % N)` ile satırlar arasında aktif olarak kaydırma yapar ve sürekli kaydırma altında bellek tüketimini test eder.
> - **100 Katman Testi**: Tıklama kestirmesi yerine 100 adet `Block` widget'ını ekrana bizzat çizer.

---

## 🖥️ Görsel İşleme İpuçları ve SSS (Rendering Quirks & FAQ)

### 1. Çizgiler veya 3D modeller bazı terminallerde neden ince çizgili (hairline gap) veya delikli görünür?
Standart terminal emülatörlerinde varsayılan monospace yazı tipi satır yüksekliği (`line-height` / hücre dolgusu) genellikle bitişik karakter satırları arasına 1–2 piksellik boşluk ekler. Bitişik Braille alt-piksel matrisleri veya yarım-bloklar (`▄`, `▀`) çizilirken bu boşluk yüzeylerin delikli veya ızgara gibi görünmesine yol açabilir.

#### Limoni Bu Sorunu Nasıl Çözer: Alt Yarım-Blok (`▄`, U+2584) Taban Standardı
Geleneksel TUI kütüphaneleri genellikle Üst Yarım-Blok (`▀`, `U+2580`) kullanır. Yazı tipi motorları karakter gliflerini hücrenin **taban çizgisine** (baseline) kilitlediğinden, satır yüksekliği eklendiğinde boşluk hücrenin *üst kısmında* oluşur ve `▀` karakterini üst satırdan ayırır.

Limoni tüm piksel ve 3D çizimlerini **Alt Yarım-Blok (`▄`, `U+2584`)** standardına geçirmiştir:
- **Üst Piksel Rengi:** Hücre arkaplanına (`Cell.Bg`) yazılır.
- **Alt Piksel Rengi:** Hücre önplanına (`Cell.Fg`) yazılır.
- **Karakter:** `▄` (Alt Yarım Blok).

Arka plan rengi karakter hücresinin tamamını kapladığı ve `▄` glifi tam taban çizgisine oturduğu için, gevşek satır yüksekliğine sahip terminallerde dahi pikseller sıfır aralıkla pürüzsüzce birleşir.

### 2. Kusursuz Görsel Deneyim İçin Önerilen Terminal Ayarları
Limoni'nin 3D rasterizasyonunu, grafiklerini ve Braille vektör çizimlerini en yüksek netlikte deneyimlemek için:

* **Satır Yüksekliğini 1.0 Yapın:** Terminalinizin yapılandırma dosyasında `line-height` / `cell-height` değerini `1.0` (veya `%100` / 0 piksel dikey boşluk) olarak ayarlayın.
* **Önerilen Modern Terminaller:**
  - **[Ghostty](https://ghostty.org):** Kutu çizimlerini, Braille ve blok gliflerini sıfır hücre boşluğuyla kusursuz çizen modern GPU terminali.
  - **[Kitty](https://sw.kovidgoyal.net/kitty/):** Yüksek performanslı OpenGL motoru, yerleşik grafik protokolleri ve bitişik glif desteği.
  - **[WezTerm](https://wezfurlong.org/wezterm/):** Mükemmel font fallback ve bitişik kutu glifi desteği.
  - **[Alacritty](https://alacritty.org):** `alacritty.toml` içinde `font.offset.y: 0` ve satır yüksekliği 1.0 kullanın.
* **Önerilen Yazı Tipleri:** [JetBrains Mono](https://www.jetbrains.com/lp/mono/), [Fira Code](https://github.com/tonsky/FiraCode) veya yamalanmış [Nerd Font](https://www.nerdfonts.com/) monospace fontları.

### 3. Limoni Hızlı Animasyonlarda 60+ FPS Performansı Nasıl Korur?
Limoni eşik tabanlı bir **Adaptif Flush Motoruna (Adaptive Flush Engine)** sahiptir:
* **Seyrek Diffing (`dirtyRatio < 0.45`):** Yazma, imleç yanıp sönmesi veya sayaç güncellemeleri gibi seyrek durumlarda sadece değişen hücreleri hesaplayıp hassas imleç sıçramaları (`CUP`) gönderir (**`~14.2 µs`**, 0 B/op).
* **Tam Akış Yenileme (`dirtyRatio >= 0.45`):** 3D model dönüşü veya hızlı kaydırma gibi ekranın %45'inden fazlasının değiştiği durumlarda imleç sıçramaları terk edilir; DEC senkronize güncelleme modu (`\x1b[?2026h`) ve ana konuma dönüş (`\x1b[H`) ile ardışık tam akış gönderilir. Böylece yırtılma ve titreme olmadan **`0 B/op`** hız korunur.

### 4. Emoji, bayraklar ve aksanlı harfler
Limoni her hücrede bir **grapheme cluster** tutar — kaç kod noktasından oluşursa oluşsun, okuyucunun tek karakter olarak gördüğü şey. `🇹🇷` (iki bölgesel gösterge), `👨‍👩‍👧` (ZWJ ile birleşmiş beş kod noktası), `👍🏽` (emoji + ten rengi) ve `e` + U+0301 olarak yazılmış `é` birer hücre kaplar; emojiler iki sütun genişliğindedir. Rune rune yürümek bayrağı iki harf olarak çiziyor, aileyi altı sütun ölçüyor ve birleşik aksanı düşürüyordu.

* **Kurallar:** bölütleme Unicode 17.0 için UAX #29'u izler ve resmî `GraphemeBreakTest.txt`'nin 766 durumunun tamamını geçer. Bir cluster'ın genişliği en geniş kod noktasınınkidir (East Asian Width, emoji sunumu); VS16 iki sütuna, VS15 bire zorlar.
* **Saklama:** hücre hâlâ tek bir `rune` tutar. Çok kod noktalı bir cluster paylaşılan bir tabloya bir kez kaydedilir ve hücre ona bir tutamaç saklar; böylece `Cell` 16 bayt kalır ve tek kod noktaları — metnin neredeyse tamamı — tabloya hiç dokunmaz. Tablo yaklaşık bir milyon farklı cluster ile sınırlıdır; sonrasında yeni cluster'lar belleği büyütmek yerine ilk kod noktalarına düşer.
* **Terminaller:** Limoni mod 2027'yi (`CSI ? 2027 h`) ister; Ghostty, WezTerm, foot ve Contour bunu uygular. Desteklemeyen terminaller imleci kod noktası başına ilerletir ve bir aile emojisini altı sütun çizebilir. Bunun satırın geri kalanını kaydırmasını önlemek için diff her cluster'dan hemen sonra imleci yeniden konumlandırır. Böyle bir terminalde cluster'ın kendisi yine yanlış görünebilir, ama ondan sonraki hiçbir şey yerinden oynamaz.
* **Kapatmak:** `LIMONI_GRAPHEME=0` (ya da `cell.SetGraphemeClusters(false)`) hücre başına bir kod noktasına döner ve mod 2027 isteğini göndermez.
* **Henüz dönüştürülmedi:** `Buffer.SetString` ile çizilen ve `cell.StringWidth` ile ölçülen metin cluster'ları tanır. Metni rune sayısına göre kesen ya da yerleştiren widget'lar — `TextInput`, `TextArea`, tablo ve toast kırpması ve diğerleri — kestikleri yerde bir cluster'ı hâlâ bölebilir.

---

## 🧩 Zengin Widget Ekosistemi

Limoni kutudan çıkan, üretime hazır geniş bir widget takımıyla gelir:

| Kategori | Mevcut Widget'lar |
| :--- | :--- |
| **Yapı & Düzen** | `Block`, `Dialog / Modal`, `Popup`, `ResponsiveGrid`, `Flexbox`, `Viewport` |
| **Veri Gösterimi** | `Table (Sanal/Sayfalı)`, `List (Sanal)`, `TreeView`, `Sparkline`, `ProgressBar`, `RichText` |
| **Grafikler** | `LineChart (Braille)`, `BarChart`, `PieChart`, `Sparkline` |
| **Girdi Kontrolleri** | `TextInput`, `TextArea`, `Checkbox`, `RadioGroup`, `Select / Dropdown`, `Slider`, `ColorPicker` |
| **Gezinme & Arama** | `Tabs`, `Scrollbar`, `CommandPalette`, `FuzzySearch (FZF tarzı)`, `KeybindingManager` |
| **Geri Bildirim** | `Spinner`, `Toast`, `ProgressBar` |
| **Grafik & 3D** | `Canvas (Braille / Blok)`, `Vector3D Mesh (OBJ/STL/PLY/GLB)`, `Lambert & Gouraud Shader`, `Image (Kitty/Sixel/iTerm2/HalfBlock)` |
| **Metin & Doküman** | `Markdown (Tam GFM)`, `RichText Vurgulama`, `Label`, `Paragraph` |
| **Erişilebilirlik & Araçlar** | `AccessibleTree`, `DevTools`, `Theme`, `Validation` |

**Kaydırma hakkında:** `List` ve `Table` kendi içinde sanallaştırma yapar ve yalnızca görünür satırları çizer; bu yüzden veri seti ne kadar büyürse büyüsün maliyetleri sabit kalır. `Viewport` ise geri kalan her şey için genel amaçlı kaydırma kapsayıcısıdır — alanından uzun herhangi bir widget'ı sarar, istenirse `Scrollbar` ile birlikte. `Viewport` ve `Scrollbar` kararlı durumda bellek tahsisatı yapmaz; maliyet modeli için paket dokümantasyonuna bakın.

---

## 📂 Örnek Uygulamalar

| Dizin | Başlık | Çalıştırma |
| :--- | :--- | :--- |
| **[`examples/composable`](examples/composable)** | **Lego Mimarisi Composable UI (VStack, HStack, Border, Sıfır Allokasyon)** | `go run ./examples/composable` |
| **[`examples/3d_viewer`](examples/3d_viewer)** | **3D Model & Shader Viewer** | `go run ./examples/3d_viewer` |
| **[`examples/paint`](examples/paint)** | **Noktasal Paint & Çizim Stüdyosu** | `go run ./examples/paint` |
| **[`examples/dashboard`](examples/dashboard)** | **Sistem Telemetri Paneli** | `go run ./examples/dashboard` |
| **[`examples/table_virtual`](examples/table_virtual)** | **1M Satırlı Sanal Tablo** | `go run ./examples/table_virtual` |
| **[`examples/todo`](examples/todo)** | **TEA Todo Uygulaması** | `go run ./examples/todo` |
| **[`examples/demo`](examples/demo)** | **3D Limon Modeli (GLB/ASCII/Braille/Half-Block) & Tanıtım Vitrini** | `go run ./examples/demo` |
| **[`examples/showcase`](examples/showcase)** | **Gelişmiş Vitrin Demosu (Matrix, DevTools F12, Formlar, Komut Paleti)** | `go run ./examples/showcase` |
| **[`examples/forms`](examples/forms)** | **Form & Girdi Kontrolleri** | `go run ./examples/forms` |
| **[`examples/layer_demo`](examples/layer_demo)** | **Katman & Modal Demosu** | `go run ./examples/layer_demo` |

---

## 📚 Türkçe Dokümantasyon

Detaylı Türkçe rehberler için [docs/tr/ dizinine](docs/tr/README.md) göz atabilirsiniz:
- [Hızlı Başlangıç Rehberi](docs/tr/getting-started.md)
- [Çekirdek Motor API Referansı](docs/tr/core-api.md)
- [Widget Kataloğu & Kullanım Kılavuzu](docs/tr/widgets-reference.md)
- [Örnek Uygulamalar Rehberi](docs/tr/examples.md)

---

## 💡 Mühendislik Felsefesi & Teşekkür

Limoni, Go ekosisteminde terminal performansının sınırlarını zorlamak ve Rust seviyesinde gecikme ve bellek determinizmini Go'ya kazandırmak amacıyla geliştirilmiştir.

> [!NOTE]
> AI tools were used for generating initial boilerplates, documentation drafts, and test cases, while the core architecture, memory layout, and debugging were directed and implemented by the author.

---

## 📄 Lisans

Bu proje **Apache License 2.0** altında lisanslanmıştır.
