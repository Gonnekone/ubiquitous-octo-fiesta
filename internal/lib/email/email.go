package email

import (
	"crypto/tls"
	"fmt"
	"gopkg.in/gomail.v2"
)

func SendEmailWarning(email string) error {
	m := gomail.NewMessage()
	m.SetHeader("Subject", "Auth warning")
	m.SetHeader("From", "gtafakon@gmail.com")
	m.SetAddressHeader("To", email, email)
	m.SetBody("text/html", "You ip was changed!")

	d := gomail.NewDialer("smtp.gmail.com", 587, "gtafakon@gmail.com", "zsmq jisv xreq kkpr")
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	return nil
}
