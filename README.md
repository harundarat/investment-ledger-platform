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

`POST /users` wajib menyertakan header `Idempotency-Key`. Retry dengan key dan payload yang sama akan mengembalikan user yang sama tanpa membuat wallet tambahan.

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
