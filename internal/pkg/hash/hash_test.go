package hash

import (
	"testing"
)

func TestHashAndComparePassword(t *testing.T) {
	password := "my-secret-password"
	wrongPassword := "wrong-password"

	// 1. Uji Hashing Password
	hashed, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Gagal melakukan hashing password: %v", err)
	}

	if hashed == "" {
		t.Fatal("Hasil hash kosong")
	}

	// 2. Uji Verifikasi Sukses (Password Benar)
	match, err := ComparePassword(password, hashed)
	if err != nil {
		t.Fatalf("Gagal memverifikasi password: %v", err)
	}
	if !match {
		t.Error("Password valid terdeteksi salah (diharapkan cocok)")
	}

	// 3. Uji Verifikasi Gagal (Password Salah)
	matchWrong, err := ComparePassword(wrongPassword, hashed)
	if err != nil {
		t.Fatalf("Gagal menjalankan perbandingan untuk password salah: %v", err)
	}
	if matchWrong {
		t.Error("Password salah terdeteksi benar (diharapkan tidak cocok)")
	}

	// 4. Uji Format Hash Tidak Valid
	invalidHash := "$argon2id$v=19$m=65536, t=3, p=4$invalid-salt" // Kurang 1 bagian (hash)
	_, errInvalid := ComparePassword(password, invalidHash)
	if errInvalid == nil {
		t.Error("Diharapkan error karena format hash tidak lengkap/valid")
	}
}
