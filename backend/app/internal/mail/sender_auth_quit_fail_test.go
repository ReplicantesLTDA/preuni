package mail

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"testing"
)

// fakeSMTPServerBehavior speaks the same real SMTP protocol as
// fakeSMTPServer in sender_fake_smtp_test.go but with configurable failure
// points, to exercise SMTPSender.Send's remaining error branches (Auth,
// wc.Close, c.Quit) with a genuine protocol-level failure or connection
// drop -- never a mock of the Sender interface.
type fakeSMTPServerBehavior struct {
	badGreeting   bool // send a non-220 greeting and close, failing smtp.NewClient
	authFails     bool // respond 535 to AUTH instead of 235
	mailFails     bool // respond 451 to MAIL FROM instead of 250
	dataRejected  bool // respond 550 to DATA instead of 354
	dropOnDataAck bool // close the connection right after the 354 DATA prompt, before the client can write
	dropAfterData bool // close the connection right after the DATA terminator, before responding
	dropOnQuit    bool // close the connection on QUIT instead of responding 221
}

func fakeSMTPServerWithBehavior(t *testing.T, cert tls.Certificate, b fakeSMTPServerBehavior) (host string, port int) {
	t.Helper()
	l, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{Certificates: []tls.Certificate{cert}})
	if err != nil {
		t.Fatalf("tls listen: %v", err)
	}
	addr := l.Addr().(*net.TCPAddr)
	t.Cleanup(func() { l.Close() })

	go func() {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		r := bufio.NewReader(conn)
		write := func(s string) { fmt.Fprintf(conn, "%s\r\n", s) }

		if b.badGreeting {
			write("421 fake.smtp.test service not available")
			return
		}
		write("220 fake.smtp.test ESMTP")
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				return
			}
			switch {
			case len(line) >= 4 && line[:4] == "EHLO":
				write("250-fake.smtp.test greets you")
				write("250 AUTH PLAIN")
			case len(line) >= 4 && line[:4] == "AUTH":
				if b.authFails {
					write("535 Authentication failed")
					continue
				}
				write("235 Authentication successful")
			case len(line) >= 4 && line[:4] == "MAIL":
				if b.mailFails {
					write("451 requested action aborted")
					continue
				}
				write("250 OK")
			case len(line) >= 4 && line[:4] == "RCPT":
				write("250 OK")
			case len(line) >= 4 && line[:4] == "DATA":
				if b.dataRejected {
					write("550 no mail service")
					continue
				}
				if b.dropOnDataAck {
					write("354 Start mail input")
					return
				}
				write("354 Start mail input")
				for {
					dl, err := r.ReadString('\n')
					if err != nil {
						return
					}
					if dl == ".\r\n" {
						break
					}
				}
				if b.dropAfterData {
					// Close without responding: wc.Close()'s read of the
					// final response genuinely fails (real EOF).
					return
				}
				write("250 OK: queued")
			case len(line) >= 4 && line[:4] == "QUIT":
				if b.dropOnQuit {
					// Close without responding: c.Quit()'s read genuinely fails.
					return
				}
				write("221 Bye")
				return
			default:
				write("500 unrecognized command")
			}
		}
	}()

	return addr.IP.String(), addr.Port
}

func TestSMTPSender_Send_AuthRejected(t *testing.T) {
	cert, pool := genSelfSignedCert(t)
	host, port := fakeSMTPServerWithBehavior(t, cert, fakeSMTPServerBehavior{authFails: true})
	s := NewSMTPSender(SMTPConfig{Host: host, Port: port, Username: "u", Password: "p", UseTLS: true, RootCAs: pool})
	err := s.Send(Message{From: "noreply@preuni.test", To: "someone@preuni.test", Subject: "hi", TextBody: "hi", HTMLBody: "<p>hi</p>"})
	if err == nil {
		t.Fatal("expected the server's real 535 AUTH rejection to surface as an error")
	}
}

func TestSMTPSender_Send_CloseFailsWhenServerDropsAfterData(t *testing.T) {
	cert, pool := genSelfSignedCert(t)
	host, port := fakeSMTPServerWithBehavior(t, cert, fakeSMTPServerBehavior{dropAfterData: true})
	s := NewSMTPSender(SMTPConfig{Host: host, Port: port, Username: "u", Password: "p", UseTLS: true, RootCAs: pool})
	err := s.Send(Message{From: "noreply@preuni.test", To: "someone@preuni.test", Subject: "hi", TextBody: "hi", HTMLBody: "<p>hi</p>"})
	if err == nil {
		t.Fatal("expected wc.Close() to fail when the server drops the connection instead of responding")
	}
}

func TestSMTPSender_Send_NewClientFailsOnBadGreeting(t *testing.T) {
	cert, pool := genSelfSignedCert(t)
	host, port := fakeSMTPServerWithBehavior(t, cert, fakeSMTPServerBehavior{badGreeting: true})
	s := NewSMTPSender(SMTPConfig{Host: host, Port: port, Username: "u", Password: "p", UseTLS: true, RootCAs: pool})
	err := s.Send(Message{From: "noreply@preuni.test", To: "someone@preuni.test", Subject: "hi", TextBody: "hi", HTMLBody: "<p>hi</p>"})
	if err == nil {
		t.Fatal("expected smtp.NewClient to fail on a real non-220 greeting")
	}
}

func TestSMTPSender_Send_MailFromRejected(t *testing.T) {
	cert, pool := genSelfSignedCert(t)
	host, port := fakeSMTPServerWithBehavior(t, cert, fakeSMTPServerBehavior{mailFails: true})
	s := NewSMTPSender(SMTPConfig{Host: host, Port: port, Username: "u", Password: "p", UseTLS: true, RootCAs: pool})
	err := s.Send(Message{From: "noreply@preuni.test", To: "someone@preuni.test", Subject: "hi", TextBody: "hi", HTMLBody: "<p>hi</p>"})
	if err == nil {
		t.Fatal("expected the server's real 451 MAIL FROM rejection to surface as an error")
	}
}

func TestSMTPSender_Send_DataRejected(t *testing.T) {
	cert, pool := genSelfSignedCert(t)
	host, port := fakeSMTPServerWithBehavior(t, cert, fakeSMTPServerBehavior{dataRejected: true})
	s := NewSMTPSender(SMTPConfig{Host: host, Port: port, Username: "u", Password: "p", UseTLS: true, RootCAs: pool})
	err := s.Send(Message{From: "noreply@preuni.test", To: "someone@preuni.test", Subject: "hi", TextBody: "hi", HTMLBody: "<p>hi</p>"})
	if err == nil {
		t.Fatal("expected the server's real 550 DATA rejection to surface as an error")
	}
}

func TestSMTPSender_Send_WriteFailsWhenServerDropsAfterDataPrompt(t *testing.T) {
	cert, pool := genSelfSignedCert(t)
	host, port := fakeSMTPServerWithBehavior(t, cert, fakeSMTPServerBehavior{dropOnDataAck: true})
	s := NewSMTPSender(SMTPConfig{Host: host, Port: port, Username: "u", Password: "p", UseTLS: true, RootCAs: pool})
	err := s.Send(Message{From: "noreply@preuni.test", To: "someone@preuni.test", Subject: "hi", TextBody: "hi", HTMLBody: "<p>hi</p>"})
	if err == nil {
		t.Fatal("expected wc.Write to fail when the server drops the connection right after the 354 prompt")
	}
}

func TestSMTPSender_Send_QuitFailsWhenServerDropsOnQuit(t *testing.T) {
	cert, pool := genSelfSignedCert(t)
	host, port := fakeSMTPServerWithBehavior(t, cert, fakeSMTPServerBehavior{dropOnQuit: true})
	s := NewSMTPSender(SMTPConfig{Host: host, Port: port, Username: "u", Password: "p", UseTLS: true, RootCAs: pool})
	err := s.Send(Message{From: "noreply@preuni.test", To: "someone@preuni.test", Subject: "hi", TextBody: "hi", HTMLBody: "<p>hi</p>"})
	if err == nil {
		t.Fatal("expected c.Quit() to fail when the server drops the connection instead of responding 221")
	}
}
