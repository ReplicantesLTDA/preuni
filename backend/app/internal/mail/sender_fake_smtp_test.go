package mail

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"testing"
	"time"
)

// genSelfSignedCert generates a real self-signed TLS certificate for
// 127.0.0.1, valid for the duration of the test. Returns the cert (to
// serve) and a CertPool trusting it (for the client's RootCAs knob) --
// real crypto/tls verification against a real cert chain, not disabled
// verification.
func genSelfSignedCert(t *testing.T) (tls.Certificate, *x509.CertPool) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "127.0.0.1"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("create cert: %v", err)
	}
	cert := tls.Certificate{Certificate: [][]byte{der}, PrivateKey: priv}
	leaf, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parse cert: %v", err)
	}
	pool := x509.NewCertPool()
	pool.AddCert(leaf)
	return cert, pool
}

// fakeSMTPServer speaks real SMTP protocol bytes over a real TLS listener:
// 220 greeting, 250 on EHLO/AUTH/MAIL/RCPT, 354 on DATA, 250 after the
// terminating "\r\n.\r\n". If rcptFails is true it returns 550 for RCPT TO
// instead, exercising a genuine protocol-level failure.
func fakeSMTPServer(t *testing.T, cert tls.Certificate, rcptFails bool) (host string, port int) {
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
				write("235 Authentication successful")
			case len(line) >= 4 && line[:4] == "MAIL":
				write("250 OK")
			case len(line) >= 4 && line[:4] == "RCPT":
				if rcptFails {
					write("550 No such user")
					continue
				}
				write("250 OK")
			case len(line) >= 4 && line[:4] == "DATA":
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
				write("250 OK: queued")
			case len(line) >= 4 && line[:4] == "QUIT":
				write("221 Bye")
				return
			default:
				write("500 unrecognized command")
			}
		}
	}()

	return addr.IP.String(), addr.Port
}

func TestSMTPSender_Send_ImplicitTLSSucceeds(t *testing.T) {
	cert, pool := genSelfSignedCert(t)
	host, port := fakeSMTPServer(t, cert, false)
	s := NewSMTPSender(SMTPConfig{Host: host, Port: port, Username: "u", Password: "p", UseTLS: true, RootCAs: pool})
	err := s.Send(Message{From: "noreply@preuni.test", To: "someone@preuni.test", Subject: "hi", TextBody: "hi", HTMLBody: "<p>hi</p>"})
	if err != nil {
		t.Fatalf("expected a successful send against a real fake SMTP server, got: %v", err)
	}
}

func TestSMTPSender_Send_ImplicitTLSRcptRejected(t *testing.T) {
	cert, pool := genSelfSignedCert(t)
	host, port := fakeSMTPServer(t, cert, true)
	s := NewSMTPSender(SMTPConfig{Host: host, Port: port, Username: "u", Password: "p", UseTLS: true, RootCAs: pool})
	err := s.Send(Message{From: "noreply@preuni.test", To: "someone@preuni.test", Subject: "hi", TextBody: "hi", HTMLBody: "<p>hi</p>"})
	if err == nil {
		t.Fatal("expected the server's real 550 RCPT rejection to surface as an error")
	}
}
