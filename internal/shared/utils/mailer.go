package utils

import (
	"fmt"

	"gopkg.in/gomail.v2"
)

// ─── Mailer ────────────────────────────────────────────────────────────────────

// MailConfig konfigurasi SMTP
type MailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
}

// Mailer mengirim email via SMTP
type Mailer struct {
	cfg MailConfig
}

func NewMailer(cfg MailConfig) *Mailer {
	return &Mailer{cfg: cfg}
}

// SendResetPasswordEmail mengirim email reset password
func (m *Mailer) SendResetPasswordEmail(toEmail, toName, resetURL string) error {
	subject := "Reset Password - Neosim"
	body := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2>Reset Password</h2>
			<p>Halo %s,</p>
			<p>Kami menerima permintaan untuk mereset password akun Anda.</p>
			<p>Klik tombol di bawah ini untuk mereset password Anda:</p>
			<p style="text-align: center; margin: 30px 0;">
				<a href="%s" 
				   style="background-color: #4F46E5; color: white; padding: 12px 24px; 
				          text-decoration: none; border-radius: 6px; display: inline-block;">
					Reset Password
				</a>
			</p>
			<p>Atau salin link berikut ke browser Anda:</p>
			<p style="word-break: break-all; color: #6B7280;">%s</p>
			<p><strong>Link ini akan kedaluwarsa dalam beberapa menit.</strong></p>
			<hr style="border: none; border-top: 1px solid #E5E7EB; margin: 20px 0;">
			<p style="color: #9CA3AF; font-size: 12px;">
				Jika Anda tidak merasa meminta reset password, abaikan email ini.
			</p>
		</body>
		</html>
	`, toName, resetURL, resetURL)

	return m.send(toEmail, subject, body)
}

// SendVerificationEmail mengirim email verifikasi akun
func (m *Mailer) SendVerificationEmail(toEmail, toName, verifyURL string) error {
	subject := "Verifikasi Akun - Neosim"
	body := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2>Verifikasi Akun Anda</h2>
			<p>Halo %s,</p>
			<p>Terima kasih telah mendaftar. Klik tombol di bawah untuk memverifikasi akun Anda:</p>
			<p style="text-align: center; margin: 30px 0;">
				<a href="%s"
				   style="background-color: #10B981; color: white; padding: 12px 24px;
				          text-decoration: none; border-radius: 6px; display: inline-block;">
					Verifikasi Akun
				</a>
			</p>
			<p>Atau salin link berikut:</p>
			<p style="word-break: break-all; color: #6B7280;">%s</p>
			<hr style="border: none; border-top: 1px solid #E5E7EB; margin: 20px 0;">
			<p style="color: #9CA3AF; font-size: 12px;">
				Jika Anda tidak mendaftar, abaikan email ini.
			</p>
		</body>
		</html>
	`, toName, verifyURL, verifyURL)

	return m.send(toEmail, subject, body)
}

// send mengirim email via SMTP
func (m *Mailer) send(to, subject, htmlBody string) error {
	msg := gomail.NewMessage()
	msg.SetAddressHeader("From", m.cfg.From, m.cfg.FromName)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", htmlBody)

	dialer := gomail.NewDialer(m.cfg.Host, m.cfg.Port, m.cfg.Username, m.cfg.Password)

	return dialer.DialAndSend(msg)
}
