package types

import (
	"strconv"
	"strings"
	"time"
)

const DateTimeLayout = "2006-01-02 15:04:05"

// CustomTime adalah tipe kustom untuk time.Time agar format JSON-nya "2006-01-02 15:04:05"
type CustomTime time.Time

// MarshalJSON mengimplementasikan json.Marshaler
// Ini akan dipanggil otomatis saat struct di-encode ke JSON
func (ct CustomTime) MarshalJSON() ([]byte, error) {
	t := time.Time(ct)
	if t.IsZero() {
		return []byte(`""`), nil // Mengembalikan string kosong jika waktu nol (bisa diubah ke []byte("null"))
	}
	// strconv.Quote memastikan string dibungkus tanda kutip ganda dengan aman untuk JSON
	return []byte(strconv.Quote(t.Format(DateTimeLayout))), nil
}

// UnmarshalJSON mengimplementasikan json.Unmarshaler
// Ini akan dipanggil otomatis saat menerima JSON request (binding)
func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "null" || s == "" {
		*ct = CustomTime(time.Time{})
		return nil
	}
	t, err := time.Parse(DateTimeLayout, s)
	if err != nil {
		return err
	}
	*ct = CustomTime(t)
	return nil
}

// Time mengembalikan nilai time.Time asli (berguna jika perlu operasi waktu)
func (ct CustomTime) Time() time.Time {
	return time.Time(ct)
}

// ToCustomTimePtr adalah helper untuk mengubah *time.Time dari model menjadi *CustomTime
func ToCustomTimePtr(t *time.Time) *CustomTime {
	if t == nil {
		return nil
	}
	ct := CustomTime(*t)
	return &ct
}
