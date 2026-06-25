package lib

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/smtp"
	"net/url"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
)

type loginAuth struct {
	username, password string
}

type Mail struct {
	Sender  string
	To      []string
	Subject string
	Body    string
}

func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}
func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}
func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, errors.New("Unkown fromServer")
		}
	}
	return nil, nil
}
func BuildMail(data []byte, mail Mail, attachName string) []byte {

	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("From: %s\r\n", mail.Sender))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(mail.To, ";")))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", mail.Subject))

	if len(data) > 0 {
		boundary := "my-boundary-779"
		buf.WriteString("MIME-Version: 1.0\r\n")
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\n",
			boundary))

		buf.WriteString(fmt.Sprintf("\r\n--%s\r\n", boundary))
		buf.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
		buf.WriteString(fmt.Sprintf("\r\n%s", mail.Body))

		buf.WriteString(fmt.Sprintf("\r\n--%s\r\n", boundary))
		buf.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
		buf.WriteString("Content-Transfer-Encoding: base64\r\n")
		buf.WriteString("Content-Disposition: attachment; filename=" + attachName + "\r\n")
		buf.WriteString("Content-ID: <" + attachName + ">\r\n\r\n")

		b := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
		base64.StdEncoding.Encode(b, data)
		buf.Write(b)
		buf.WriteString(fmt.Sprintf("\r\n--%s", boundary))

		buf.WriteString("--")
	} else {
		buf.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
		buf.WriteString(fmt.Sprintf("\r\n%s", mail.Body))
	}

	return buf.Bytes()
}
func SendMail(mailer Mailer) error {

	from_email := mailer.Sender
	host := mailer.SmtpHost + ":" + mailer.Port

	request := Mail{
		Sender:  from_email,
		To:      mailer.Receivers,
		Subject: mailer.Subject,
		Body:    mailer.Msg,
	}

	data := BuildMail(mailer.Attach, request, mailer.AttachName)

	// auth := smtp.PlainAuth("", from_email, password, smtp)
	auth := LoginAuth(from_email, mailer.Passwd)

	ctx := context.Background()

	err := smtp.SendMail(host, auth, from_email, mailer.Receivers, data)
	if err == nil {
		Logga(ctx, os.Getenv("JsonLog"), "Email Sent Successfully")
		return nil
	} else {
		Logga(ctx, os.Getenv("JsonLog"), "ERROR: "+err.Error(), "error")
		return err
	}
}
func TelegramSendMessage(botToken, cftoolDevopsChatID, text string) error {
	type telegramResStruct struct {
		Ok     bool `json:"ok"`
		Result struct {
			MessageID  int `json:"message_id"`
			SenderChat struct {
				ID       int64  `json:"id"`
				Title    string `json:"title"`
				Username string `json:"username"`
				Type     string `json:"type"`
			} `json:"sender_chat"`
			Chat struct {
				ID       int64  `json:"id"`
				Title    string `json:"title"`
				Username string `json:"username"`
				Type     string `json:"type"`
			} `json:"chat"`
			Date int    `json:"date"`
			Text string `json:"text"`
		} `json:"result"`
	}

	var erro error

	// Guard: senza token o chat id la URL Telegram e' malformata (path /bot/sendMessage)
	// e l'API ritorna una pagina HTML 404, non JSON -> unmarshal fallisce con
	// "invalid character '<'". Notify non configurato = skip silenzioso.
	if strings.TrimSpace(botToken) == "" || strings.TrimSpace(cftoolDevopsChatID) == "" {
		return nil
	}

	clientTelegram := resty.New()
	clientTelegram.Debug = false
	resTelegram, errTelegram := clientTelegram.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParam("chat_id", cftoolDevopsChatID).
		SetQueryParam("text", text).
		Post("https://api.telegram.org/bot" + url.PathEscape(botToken) + "/sendMessage")

	var telegramRes telegramResStruct
	if errTelegram != nil {
		return errTelegram
	}
	err1 := json.Unmarshal(resTelegram.Body(), &telegramRes)
	if err1 != nil {
		// Risposta non-JSON (es. HTML 404 da token errato): ritorna corpo grezzo troncato
		body := strings.TrimSpace(string(resTelegram.Body()))
		if len(body) > 200 {
			body = body[:200]
		}
		return fmt.Errorf("telegram: risposta non-JSON (status %d): %s", resTelegram.StatusCode(), body)
	}

	if !telegramRes.Ok {
		erro = errors.New(telegramRes.Result.Text)
		return erro
	}
	return nil
}
