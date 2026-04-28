package main

import (
	"encoding/base64"
	"fmt"
	"mime"
	"net/smtp"
	"strings"
)

type attachment struct {
	Name        string
	ContentType string
	Data        []byte
}

type submission struct {
	FirstName   string
	LastName    string
	Email       string
	Answers     map[string]string
	Attachments []attachment
}

// Question labels in order for email formatting.
var questionLabels = []struct {
	Key   string
	Label string
}{
	{"q1", "Jak trafiłeś na tą ankietę?"},
	{"q1_other", "Jak trafiłeś — inne"},
	{"q2", "Metraż powierzchni projektowej"},
	{"q3", "Powierzchnia — całość czy poszczególne pomieszczenia"},
	{"q4", "Lokalizacja inwestycji"},
	{"q5", "Stan deweloperski czy rynek wtórny"},
	{"q6", "Planowane oddanie inwestycji"},
	{"q7_info", "Rzut inwestycji"},
	{"q8", "Piętro"},
	{"q9", "Skosy"},
	{"q10", "Styl wnętrzarski"},
	{"q10_other", "Styl — inne"},
	{"q11", "Zakres projektu"},
	{"q12", "Nadzór autorski"},
	{"q13", "Kiedy rozpocząć prace"},
	{"first_name", "Imię"},
	{"last_name", "Nazwisko"},
	{"phone", "Telefon"},
	{"email", "Email"},
}

// Preference form question labels in order for email formatting.
var preferenceLabels = []struct {
	Key   string
	Label string
}{
	{"first_name", "Imię"},
	{"last_name", "Nazwisko"},
	{"phone", "Telefon"},
	{"email", "Email"},
	{"p2", "Ile osób użytkuje mieszkanie"},
	{"p3", "Alergicy"},
	{"p4", "Praca w domu / biuro / warsztat"},
	{"p5", "Meble i dekoracje z poprzedniego mieszkania"},
	{"p6", "Zmiana położenia przyłączy hydraulicznych / elektrycznych"},
	{"p7", "Ulubione kolory"},
	{"p8", "Kolory i materiały do unikania"},
	{"p9", "Wykończenie podłogi w pokojach"},
	{"p10", "Szerokość zmywarki"},
	{"p11", "Lodówka w kuchni"},
	{"p12", "Dodatkowy sprzęt w kuchni"},
	{"p13", "Wyspa w kuchni"},
	{"p14", "Kuchenka"},
	{"p15", "Wymiary łóżka w sypialni głównej"},
	{"p16", "Sanitariaty w łazience"},
	{"p17", "Miejsce na ładowanie sprzętów elektrycznych"},
	{"p18", "Rodzaj ogrzewania"},
	{"p19", "Ingerencja w konstrukcję"},
	{"p20", "Inne sprzęty codziennego użytku"},
	{"p21", "Preferowany styl"},
	{"p21_other", "Preferowany styl — inne"},
	{"p22", "Zwierzęta"},
	{"p23", "Sport i sprzęt sportowy"},
	{"p24", "Instrumenty muzyczne"},
	{"p25", "Goście / sofa rozkładana"},
	{"p26", "Odkurzacz centralny"},
	{"p27", "System inteligentnego domu"},
	{"p29", "Sugestie i specjalne potrzeby"},
}

func sendPreferencesEmail(cfg config, sub submission) error {
	subject := fmt.Sprintf("Ankieta preferencji — %s %s", sub.FirstName, sub.LastName)

	var body strings.Builder
	body.WriteString("Nowe zgłoszenie z ankiety preferencji klienta\n")
	body.WriteString("==============================================\n\n")

	for _, ql := range preferenceLabels {
		if v, ok := sub.Answers[ql.Key]; ok && v != "" {
			body.WriteString(fmt.Sprintf("%s:\n%s\n\n", ql.Label, v))
		}
	}

	if len(sub.Attachments) > 0 {
		body.WriteString("Przesłane pliki:\n")
		for _, att := range sub.Attachments {
			body.WriteString(fmt.Sprintf("  - %s\n", att.Name))
		}
		body.WriteString("\n")
	}

	boundary := "----SeccoFormBoundary"
	var msg strings.Builder

	msg.WriteString(fmt.Sprintf("From: %s\r\n", cfg.smtpUser))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", cfg.recipientEmail))
	msg.WriteString(fmt.Sprintf("Reply-To: %s\r\n", sub.Email))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", mime.QEncoding.Encode("utf-8", subject)))
	msg.WriteString("MIME-Version: 1.0\r\n")

	if len(sub.Attachments) == 0 {
		msg.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
		msg.WriteString("\r\n")
		msg.WriteString(body.String())
	} else {
		msg.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n", boundary))
		msg.WriteString("\r\n")

		msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		msg.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
		msg.WriteString("\r\n")
		msg.WriteString(body.String())
		msg.WriteString("\r\n")

		for _, att := range sub.Attachments {
			msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			ct := att.ContentType
			if ct == "" {
				ct = "application/octet-stream"
			}
			msg.WriteString(fmt.Sprintf("Content-Type: %s\r\n", ct))
			msg.WriteString("Content-Transfer-Encoding: base64\r\n")
			msg.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=%q\r\n", att.Name))
			msg.WriteString("\r\n")
			msg.WriteString(base64.StdEncoding.EncodeToString(att.Data))
			msg.WriteString("\r\n")
		}

		msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	}

	addr := fmt.Sprintf("%s:%s", cfg.smtpHost, cfg.smtpPort)
	auth := smtp.PlainAuth("", cfg.smtpUser, cfg.smtpPass, cfg.smtpHost)

	return smtp.SendMail(addr, auth, cfg.smtpUser, []string{cfg.recipientEmail}, []byte(msg.String()))
}

func sendEmail(cfg config, sub submission) error {
	subject := fmt.Sprintf("Nowa wycena — %s %s", sub.FirstName, sub.LastName)

	var body strings.Builder
	body.WriteString("Nowe zgłoszenie z formularza wyceny\n")
	body.WriteString("=====================================\n\n")

	for _, ql := range questionLabels {
		if v, ok := sub.Answers[ql.Key]; ok && v != "" {
			body.WriteString(fmt.Sprintf("%s:\n%s\n\n", ql.Label, v))
		}
	}

	if len(sub.Attachments) > 0 {
		body.WriteString("Przesłane pliki:\n")
		for _, att := range sub.Attachments {
			body.WriteString(fmt.Sprintf("  - %s\n", att.Name))
		}
		body.WriteString("\n")
	}

	boundary := "----SeccoFormBoundary"
	var msg strings.Builder

	msg.WriteString(fmt.Sprintf("From: %s\r\n", cfg.smtpUser))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", cfg.recipientEmail))
	msg.WriteString(fmt.Sprintf("Reply-To: %s\r\n", sub.Email))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", mime.QEncoding.Encode("utf-8", subject)))
	msg.WriteString("MIME-Version: 1.0\r\n")

	if len(sub.Attachments) == 0 {
		msg.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
		msg.WriteString("\r\n")
		msg.WriteString(body.String())
	} else {
		msg.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n", boundary))
		msg.WriteString("\r\n")

		// Text part
		msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		msg.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
		msg.WriteString("\r\n")
		msg.WriteString(body.String())
		msg.WriteString("\r\n")

		// Attachments
		for _, att := range sub.Attachments {
			msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			ct := att.ContentType
			if ct == "" {
				ct = "application/octet-stream"
			}
			msg.WriteString(fmt.Sprintf("Content-Type: %s\r\n", ct))
			msg.WriteString("Content-Transfer-Encoding: base64\r\n")
			msg.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=%q\r\n", att.Name))
			msg.WriteString("\r\n")
			msg.WriteString(base64.StdEncoding.EncodeToString(att.Data))
			msg.WriteString("\r\n")
		}

		msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	}

	addr := fmt.Sprintf("%s:%s", cfg.smtpHost, cfg.smtpPort)
	auth := smtp.PlainAuth("", cfg.smtpUser, cfg.smtpPass, cfg.smtpHost)

	return smtp.SendMail(addr, auth, cfg.smtpUser, []string{cfg.recipientEmail}, []byte(msg.String()))
}
