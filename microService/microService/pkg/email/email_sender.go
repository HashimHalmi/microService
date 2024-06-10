package email

import (
	"gopkg.in/gomail.v2"
	"io"
	"net/smtp"
)

const (
	SMTPServer   = "smtp.gmail.com"
	SMTPPort     = "587"
	AuthEmail    = "bdauren06@gmail.com"
	AuthPassword = "lrhcvmwjvdkbjrkc"
)

func SendEmail(to, subject, body string) error {
	auth := smtp.PlainAuth("", AuthEmail, AuthPassword, SMTPServer)
	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")
	return smtp.SendMail(SMTPServer+":"+SMTPPort, auth, AuthEmail, []string{to}, msg)
}

func SendReceiptEmail(to, subject, body string, attachment []byte) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "bdauren06@gmail.com")
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	m.Attach("receipt.pdf", gomail.SetCopyFunc(func(w io.Writer) error {
		_, err := w.Write(attachment)
		return err
	}))

	d := gomail.NewDialer("smtp.gmail.com", 587, "bdauren06@gmail.com", "lrhcvmwjvdkbjrkc")

	return d.DialAndSend(m)
}
