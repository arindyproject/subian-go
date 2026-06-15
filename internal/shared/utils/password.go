package utils

import (
	"errors"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// ─── Password Policy ───────────────────────────────────────────────────────────

// PasswordPolicy konfigurasi kebijakan password
type PasswordPolicy struct {
	MinLength        int
	RequireUppercase bool
	RequireNumber    bool
	RequireSymbol    bool
}

// Validate memvalidasi password sesuai policy
// Mengembalikan list pesan error (kosong = valid)
func (p *PasswordPolicy) Validate(password string) []string {
	var errs []string

	if len(password) < p.MinLength {
		errs = append(errs, "Password minimal "+itoa(p.MinLength)+" karakter")
	}

	if p.RequireUppercase {
		hasUpper := false
		for _, r := range password {
			if unicode.IsUpper(r) {
				hasUpper = true
				break
			}
		}
		if !hasUpper {
			errs = append(errs, "Password harus mengandung minimal satu huruf kapital")
		}
	}

	if p.RequireNumber {
		hasNumber := false
		for _, r := range password {
			if unicode.IsDigit(r) {
				hasNumber = true
				break
			}
		}
		if !hasNumber {
			errs = append(errs, "Password harus mengandung minimal satu angka")
		}
	}

	if p.RequireSymbol {
		hasSymbol := false
		for _, r := range password {
			if unicode.IsPunct(r) || unicode.IsSymbol(r) {
				hasSymbol = true
				break
			}
		}
		if !hasSymbol {
			errs = append(errs, "Password harus mengandung minimal satu simbol")
		}
	}

	return errs
}

// ─── Bcrypt Helpers ────────────────────────────────────────────────────────────

// HashPassword membuat hash bcrypt dari password
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword memverifikasi password dengan hash
func VerifyPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// CheckPasswordHistory mengecek apakah password sama dengan history
func CheckPasswordHistory(newPassword string, historyHashes []string) error {
	for _, hash := range historyHashes {
		if VerifyPassword(newPassword, hash) {
			return errors.New("password tidak boleh sama dengan password sebelumnya")
		}
	}
	return nil
}

// ─── Helper ────────────────────────────────────────────────────────────────────

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}
