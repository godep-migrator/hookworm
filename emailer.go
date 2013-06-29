package hookworm

import (
	"crypto/tls"
	"log"
	"net"
	"net/smtp"
	"net/url"
)

// Emailer encapsulates emailing functionality which mostly means it wraps
// goodies from net/smtp
type Emailer struct {
	serverUri *url.URL
	auth      smtp.Auth
}

// NewEmailer creates an Emailer from an smtp server uri
func NewEmailer(serverUri string) *Emailer {
	parsedUri, err := url.Parse(serverUri)
	if err != nil {
		log.Panicln("Failed to parse smtp url:", err)
	}

	var auth smtp.Auth

	if parsedUri.User != nil {
		username := parsedUri.User.Username()
		password, _ := parsedUri.User.Password()
		auth = smtp.PlainAuth("",
			username, password, parsedUri.Host)
	}

	return &Emailer{serverUri: parsedUri, auth: auth}
}

// Send wraps sending via InsecureSendMail, using this Emailer's server uri
// host and auth
func (me *Emailer) Send(from string, to []string, msg []byte) error {
	return InsecureSendMail(me.serverUri.Host, me.auth, from, to, msg)
}

// InsecureSendMail is a patched version of net/smtp.SendMail that sets
// InsecureSkipVerify on the TLS config so that self-signed mail certs don't
// cause x509 errors.
func InsecureSendMail(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	config := &tls.Config{}
	c, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	if ok, _ := c.Extension("STARTTLS"); ok {
		host, _, _ := net.SplitHostPort(addr)
		config.InsecureSkipVerify = true
		config.ServerName = host
		if err = c.StartTLS(config); err != nil {
			return err
		}
	}
	if a != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err = c.Auth(a); err != nil {
				return err
			}
		}
	}
	if err = c.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}
