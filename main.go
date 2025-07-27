package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"

	"github.com/joho/godotenv"
)

type ContactForm struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "https://mehmetemreok.com")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func contactHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Desteklenmeyen metot", http.StatusMethodNotAllowed)
		return
	}

	var form ContactForm
	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		http.Error(w, "JSON verisi okunamadı", http.StatusBadRequest)
		return
	}

	if form.Name == "" || form.Email == "" || form.Message == "" {
		http.Error(w, "Tüm alanlar zorunludur", http.StatusBadRequest)
		return
	}

	log.Printf("Yeni mesaj alındı: İsim: %s, Email: %s", form.Name, form.Email)

	err := sendEmail(form)
	if err != nil {
		log.Printf("E-posta gönderilirken hata oluştu: %v", err)
		http.Error(w, "Mesaj gönderilirken bir hata oluştu.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Mesajınız başarıyla gönderildi!"})
}

func sendEmail(form ContactForm) error {

	to := os.Getenv("SMTP_TO_EMAIL")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")

	subject := "Yeni İletişim Formu Mesajı - " + form.Name
	body := fmt.Sprintf("İsim: %s\nE-posta: %s\n\nMesaj:\n%s", form.Name, form.Email, form.Message)

	msg := "From: " + user + "\n" +
		"To: " + to + "\n" +
		"Reply-To: " + form.Email + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	auth := smtp.PlainAuth("", user, pass, host)
	err := smtp.SendMail(host+":"+port, auth, user, []string{to}, []byte(msg))

	return err
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Println("Uyarı: .env dosyası bulunamadı.")
	}

	handler := corsMiddleware(http.HandlerFunc(contactHandler))

	http.Handle("/api/contact", handler)

	port := "8080"
	log.Printf("Sunucu http://localhost:%s adresinde başlatılıyor...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Sunucu başlatılamadı: %v", err)
	}
}
