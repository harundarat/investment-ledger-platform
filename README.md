# Investment Ledger Platform

## Menjalankan API secara lokal

Aplikasi membutuhkan PostgreSQL dan Goose untuk menjalankan migration.

1. Salin `.env.example` menjadi `.env`, lalu isi setidaknya nilai berikut:

   ```env
   DATABASE_URL=postgres://ilp:ilp@localhost:5432/investment_ledger?sslmode=disable
   IDEMPOTENCY_HASH_SECRET=ganti-dengan-secret-lokal-yang-panjang
   PORT=3000
   ```

2. Jalankan PostgreSQL:

   ```sh
   docker compose up -d postgres
   ```

3. Jalankan seluruh migration pada database development yang masih kosong:

   ```sh
   goose -dir migrations postgres "$DATABASE_URL" up
   ```

4. Jalankan API:

   ```sh
   go run ./cmd/api
   ```

## Endpoint user

### Buat user

`POST /users` wajib menyertakan header `Idempotency-Key`. Retry dengan key
dan payload yang sama akan mengembalikan user yang sama tanpa membuat wallet
tambahan.

```sh
curl -X POST http://localhost:3000/users \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: register-harun-001' \
  -d '{
    "name": "Harun",
    "email": "harun@example.com",
    "password": "password-yang-kuat"
  }'
```

Request pertama mengembalikan `201 Created` dan menandai bahwa respons bukan
hasil replay:

```json
{
  "message": "user created",
  "data": {
    "id": "019f7d55-6e6d-7852-9c1f-f4c07422d534",
    "name": "Harun",
    "email": "harun@example.com"
  },
  "meta": {
    "idempotency_replayed": false
  }
}
```

Retry dengan key dan payload yang sama mengembalikan `200 OK`, header
`Idempotency-Replayed: true`, dan metadata berikut:

```json
{
  "message": "idempotent request replayed",
  "data": {
    "id": "019f7d55-6e6d-7852-9c1f-f4c07422d534",
    "name": "Harun",
    "email": "harun@example.com"
  },
  "meta": {
    "idempotency_replayed": true
  }
}
```

Key yang sama dengan payload berbeda ditolak sebagai konflik idempotensi.
Key baru dengan email yang telah terdaftar ditolak sebagai konflik email.

Jika body bukan JSON yang valid, API mengembalikan `400 Bad Request`:

```json
{
  "error": {
    "code": "MALFORMED_JSON",
    "message": "request body must contain valid JSON"
  }
}
```

Jika JSON valid tetapi field tidak memenuhi aturan input, API mengembalikan
`422 Unprocessable Entity` dengan seluruh detail validasi:

```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "request validation failed",
    "details": [
      {
        "field": "email",
        "rule": "email",
        "message": "email must be a valid email address"
      },
      {
        "field": "password",
        "rule": "min",
        "message": "password must be at least 8 characters"
      }
    ]
  }
}
```

Client sebaiknya menggunakan `error.code` untuk menentukan alur program dan
`error.message` atau `error.details` untuk ditampilkan kepada pengguna.

### Ambil profile

```sh
curl http://localhost:3000/users/<user-id>
```

Endpoint profile ini belum memiliki autentikasi. Jangan gunakan `GET /users/:id` sebagai profile privat sampai modul autentikasi dan authorization tersedia.

## Pengembangan

```sh
go test ./...
go vet ./...
```

> `migrations/00001_create_user_module.sql` telah diubah untuk menambahkan tabel idempotensi. Jika migration versi lama sudah pernah diterapkan pada database lokal, reset database development sebelum menjalankan Goose lagi. Jangan mengubah migration yang sudah diterapkan pada environment dengan data yang harus dipertahankan.
