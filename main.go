package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"golang.org/x/net/proxy"
	"gopkg.in/yaml.v3"
)

type site struct {
	Isim string `yaml:"isim"`
	Url  string `yaml:"url"`
}

type Config struct {
	Site []site `yaml:"sites"`
}

// Aktif Tor portunu bul (9050 veya 9150)
func findActiveTorPort() (string, error) {
	ports := []string{"9050", "9150"}

	for _, port := range ports {
		fmt.Printf("[TEST] Port %s kontrol ediliyor...\n", port)
		timeout := 2 * time.Second
		conn, err := net.DialTimeout("tcp", "127.0.0.1:"+port, timeout)
		if err == nil {
			conn.Close()
			fmt.Printf("[BAŞARI] Port %s AÇIK.\n", port)
			return port, nil
		}
		fmt.Printf("[BİLGİ] Port %s kapalı.\n", port)
	}
	return "", fmt.Errorf("hiçbir Tor portu (9050, 9150) açık değil. Tor Browser açık mı?")
}

// Tor üzerinden çalışan HTTP client oluştur (Port parametresi alıyor)
func createTorClient(torPort string) (*http.Client, error) {
	// Bulunan port ile SOCKS5 dialer oluştur
	torAddr := "127.0.0.1:" + torPort
	dialer, err := proxy.SOCKS5("tcp", torAddr, nil, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("SOCKS5 proxy hatası: %v", err)
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		},
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   60 * time.Second,
	}

	return client, nil
}

// ChromeDP ile Tor üzerinden screenshot ve HTML al
func captureWithTor(siteUrl, siteName string, torPort string) ([]byte, string, error) {
	// ChromeDP ayarları - Tor proxy ile
	proxyURL := fmt.Sprintf("socks5://127.0.0.1:%s", torPort)
	fmt.Printf("[ChromeDP] Proxy kullanılıyor: %s\n", proxyURL)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ProxyServer(proxyURL),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-features", "IsolateOrigins,site-per-process"),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("ignore-certificate-errors", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Timeout ekle (onion siteleri yavaş olabilir)
	ctx, cancel = context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	var screenshot []byte
	var htmlContent string

	// Site yükle, HTML al ve screenshot çek
	err := chromedp.Run(ctx,
		chromedp.Navigate(siteUrl),
		chromedp.Sleep(15*time.Second), // Sayfanın yüklenmesini bekle (biraz artırdım)
		chromedp.OuterHTML("html", &htmlContent),
		chromedp.FullScreenshot(&screenshot, 100),
	)

	if err != nil {
		return nil, "", err
	}

	return screenshot, htmlContent, nil
}

// HTML dosyasını kaydet
func saveHTML(siteName, url, content string) error {
	os.MkdirAll("results", 0755)

	timestamp := time.Now().Format("20060102_150405")
	safeName := strings.ReplaceAll(siteName, " ", "_")
	filename := fmt.Sprintf("results/%s_%s.html", safeName, timestamp)

	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return err
	}

	fmt.Printf("[KAYIT] HTML: %s\n", filename)
	return nil
}

// Screenshot kaydet
func saveScreenshot(siteName string, screenshot []byte) error {
	os.MkdirAll("results", 0755)

	timestamp := time.Now().Format("20060102_150405")
	safeName := strings.ReplaceAll(siteName, " ", "_")
	filename := fmt.Sprintf("results/%s_%s.png", safeName, timestamp)

	err := os.WriteFile(filename, screenshot, 0644)
	if err != nil {
		return err
	}

	fmt.Printf("[KAYIT] Screenshot: %s\n", filename)
	return nil
}

// Log dosyasına yaz
func writeLog(message string) {
	f, err := os.OpenFile("scan_report.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("[UYARI] Log yazılamadı: %v\n", err)
		return
	}
	defer f.Close()
	logEntry := fmt.Sprintf("[%s] %s\n", time.Now().Format(time.RFC3339), message)
	f.WriteString(logEntry)
}

func main() {
	// 1. Önce aktif Tor portunu bul
	activePort, err := findActiveTorPort()
	if err != nil {
		log.Fatal("[KRİTİK HATA] Tor bağlantısı sağlanamadı: ", err)
	}

	fmt.Printf("[BİLGİ] İşlemler %s portu üzerinden yapılacak.\n", activePort)

	// 2. Tor client oluştur (Bulunan port ile)
	fmt.Println("[İŞLEM] Tor HTTP client oluşturuluyor...")
	torClient, err := createTorClient(activePort)
	if err != nil {
		log.Fatal("[HATA] Client oluşturulamadı: ", err)
	}

	// 3. YAML dosyasını oku
	yamlFile, err := os.ReadFile("sites.yaml")
	if err != nil {
		log.Fatal("Dosya okunurken hata oluştu: ", err)
	}

	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatal("YAML işleme hatası: ", err)
	}

	// Sayaçlar
	basarisiz := 0

	// 4. Her bir siteyi sırayla tara
	for i, s := range config.Site {
		fmt.Println("------------------------------------------------")
		fmt.Printf("[%d/%d] Hedef: %s | URL: %s\n", i+1, len(config.Site), s.Isim, s.Url)

		// Önce HTTP status kontrolü yap
		resp, err := torClient.Get(s.Url)
		if err != nil {
			fmt.Printf("[HATA] Site erişilemez: %v\n", err)
			fmt.Println("[LOG] Dead/Offline - Sonraki URL'e geçiliyor")
			writeLog(fmt.Sprintf("OFFLINE | %s | %s | Hata: %v", s.Isim, s.Url, err))
			continue
		}

		// Body kapatılmalı ama okumadığımız için hata vermesin diye drain ediyoruz
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		fmt.Printf("[DURUM] HTTP Status: %d %s\n", resp.StatusCode, resp.Status)

		if resp.StatusCode != 200 {
			fmt.Println("[UYARI] Sayfa 200 döndürmedi, yine de ChromeDP deneniyor...")
		}

		// ChromeDP ile screenshot ve HTML al (Port bilgisini gönderiyoruz)
		fmt.Println("[İŞLEM] ChromeDP başlatılıyor...")
		screenshot, htmlContent, err := captureWithTor(s.Url, s.Isim, activePort)
		if err != nil {
			fmt.Printf("[HATA] ChromeDP hatası: %v\n", err)
			writeLog(fmt.Sprintf("BAŞARISIZ | %s | %s | ChromeDP Hatası: %v", s.Isim, s.Url, err))
			basarisiz++
			continue
		}

		// HTML kaydet
		err = saveHTML(s.Isim, s.Url, htmlContent)
		if err != nil {
			fmt.Printf("[UYARI] HTML kaydedilemedi: %v\n", err)
		}

		// Screenshot kaydet
		err = saveScreenshot(s.Isim, screenshot)
		if err != nil {
			fmt.Printf("[UYARI] Screenshot kaydedilemedi: %v\n", err)
		}

		writeLog(fmt.Sprintf("BAŞARILI | %s | %s", s.Isim, s.Url))
	}

	fmt.Println("------------------------------------------------")
	fmt.Println("Tarama tamamlandı.")
}
