🧅 Tor Web Scraper (Go)
Bu proje, Go dili kullanılarak geliştirilmiş, Tor ağı üzerinden .onion sitelerine erişebilen, sayfaların tam boy ekran görüntülerini (screenshot) alan ve HTML içeriklerini kaydeden bir otomasyon aracıdır.

✨ Özellikler
Otomatik Port Tespiti: Tor Browser'ın kullandığı varsayılan portları (9050 veya 9150) otomatik olarak tarar ve aktif olanı kullanır.

Tor Proxy Desteği: Tüm trafik SOCKS5 proxy üzerinden güvenli bir şekilde Tor ağına yönlendirilir.

Headless Tarayıcı: chromedp kütüphanesi sayesinde sayfalar gerçek bir tarayıcı gibi yüklenir (JavaScript desteği mevcuttur).

Otomatik Çıktı Yönetimi: Her tarama için results/ klasörü altında zaman damgalı HTML ve PNG dosyaları oluşturur.

Detaylı Loglama: İşlem durumlarını ve hataları scan_report.log dosyasına kaydeder.

YAML Yapılandırması: Hedef siteler kolayca düzenlenebilir bir sites.yaml dosyasından okunur.

🛠️ Kurulum
1. Ön Gereksinimler
Go: Bilgisayarınızda Go yüklü olmalıdır.

Tor: Tor Browser'ın açık olması veya arka planda bir Tor servisinin çalışıyor olması gerekir.

Chrome/Chromium: chromedp için sistemde yüklü bir tarayıcı bulunmalıdır.

2. Bağımlılıkları Yükle
Proje klasöründe terminali açın ve gerekli modülleri indirin:

Bash

go mod tidy

⚙️ Yapılandırma
sites.yaml dosyasını kullanarak taranacak siteleri şu formatta belirleyebilirsiniz:

🚀 Kullanım
Programı çalıştırmadan önce Tor Browser'ın açık olduğundan emin olun.

Bash
go run main.go

📂 Dosya Yapısı
main.go: Ana uygulama mantığı ve proxy ayarları.

sites.yaml: Hedef sitelerin listesi.

results/: Tarama sonuçlarının (HTML ve PNG) saklandığı klasör.

scan_report.log: Tarama geçmişi ve hata kayıtları.

go.mod & go.sum: Bağımlılık yönetim dosyaları.

⚠️ Önemli Notlar
Timeout: Onion siteleri normal sitelere göre çok daha yavaş yüklenebilir. Kod içerisinde varsayılan olarak 120 saniyelik bir zaman aşımı süresi tanımlanmıştır.

Güvenlik: Bu araç eğitim ve siber güvenlik araştırmaları amacıyla geliştirilmiştir. Kullanım sırasında yasal sorumluluk kullanıcıya aittir.
