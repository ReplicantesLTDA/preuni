package mail

import (
	"net"
	"testing"
)

// closedPortAddr binds a real TCP listener, immediately closes it, and
// returns the (host, port) it was bound to -- the OS won't reuse the port
// for the lifetime of this test process's connection attempt, so dialing
// it back genuinely fails with "connection refused". This exercises
// SMTPSender.Send's real network-dial error paths (both the implicit-TLS
// and STARTTLS branches) without needing a full fake SMTP/TLS server.
func closedPortAddr(t *testing.T) (string, int) {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := l.Addr().(*net.TCPAddr)
	if err := l.Close(); err != nil {
		t.Fatalf("close listener: %v", err)
	}
	return addr.IP.String(), addr.Port
}

func TestSMTPSender_Send_ImplicitTLSDialFails(t *testing.T) {
	host, port := closedPortAddr(t)
	s := NewSMTPSender(SMTPConfig{Host: host, Port: port, Username: "u", Password: "p", UseTLS: true})
	err := s.Send(Message{From: "noreply@preuni.test", To: "someone@preuni.test", Subject: "hi", TextBody: "hi", HTMLBody: "<p>hi</p>"})
	if err == nil {
		t.Fatal("expected a real dial error against a closed port")
	}
}

func TestSMTPSender_Send_STARTTLSDialFails(t *testing.T) {
	host, port := closedPortAddr(t)
	s := NewSMTPSender(SMTPConfig{Host: host, Port: port, Username: "u", Password: "p", UseTLS: false})
	// From left empty on purpose: exercises the "from = s.cfg.Username"
	// fallback branch on the way to the (still-failing) dial.
	err := s.Send(Message{To: "someone@preuni.test", Subject: "hi", TextBody: "hi", HTMLBody: "<p>hi</p>"})
	if err == nil {
		t.Fatal("expected a real dial error against a closed port")
	}
}

// acceptCloseListener accepts exactly one connection and immediately closes
// it (no bytes written), forcing whatever reads the connection next to see
// a real EOF/connection-reset -- genuine network behavior, not a mocked
// SMTP server.
func acceptCloseListener(t *testing.T) (string, int) {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := l.Addr().(*net.TCPAddr)
	go func() {
		conn, err := l.Accept()
		if err == nil {
			conn.Close()
		}
		l.Close()
	}()
	return addr.IP.String(), addr.Port
}

func TestSMTPSender_Send_ImplicitTLSHandshakeFailsOnPlainSocket(t *testing.T) {
	host, port := acceptCloseListener(t)
	s := NewSMTPSender(SMTPConfig{Host: host, Port: port, Username: "u", Password: "p", UseTLS: true})
	err := s.Send(Message{From: "noreply@preuni.test", To: "someone@preuni.test", Subject: "hi", TextBody: "hi", HTMLBody: "<p>hi</p>"})
	if err == nil {
		t.Fatal("expected a real TLS handshake failure against a plain socket that closes immediately")
	}
}
