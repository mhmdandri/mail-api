# Mail API Microservice (iCloud SMTP Relay)

Microservice REST API ringan berbasis **Go (Golang)** dan **Docker** untuk mengirim email konfirmasi secara andal melalui **SMTP relay iCloud Mail** (`smtp.mail.me.com`).

---

## 🌟 Fitur Utama

- **Lightweight & Fast**: Dibangun menggunakan pustaka standar Go dan `gopkg.in/gomail.v2`.
- **MIME & Multipart Ready**: Mendukung pengiriman email plaintext, HTML, atau multipart alternative (plaintext fallback + HTML).
- **STARTTLS Port 587**: Koneksi aman dengan verifikasi sertifikat TLS lengkap (`ca-certificates`).
- **Endpoint Protection**: Proteksi autentikasi via header `X-API-Key` dengan constant-time comparison untuk mencegah timing attack.
- **Structured Logging**: Menggunakan `log/slog` bawaan Go untuk log JSON terstruktur (status pengiriman, durasi, error SMTP).
- **Multi-Stage Docker Build**: Image Docker Alpine minimalis, aman (non-root user), dan siap produksi.

---

## 📁 Struktur Direktori

```text
mail-api/
├── cmd/
│   └── api/
│       └── main.go           # Entry point aplikasi & graceful shutdown
├── internal/
│   ├── config/
│   │   └── config.go         # Parsing environment variables & .env
│   ├── handler/
│   │   ├── handler.go        # HTTP handlers, validasi & response helper
│   │   └── handler_test.go   # Unit test endpoint & validasi
│   └── mailer/
│       └── mailer.go         # Konfigurasi SMTP gomail & STARTTLS
├── .env.example              # Template variabel konfigurasi
├── .gitignore
├── Dockerfile                # Multi-stage build Alpine
├── docker-compose.yml        # Orkestrasi Docker Compose
├── go.mod
├── go.sum
└── README.md
```

---

## ⚙️ Konfigurasi Environment Variables

Salin file `.env.example` menjadi `.env`:

```bash
cp .env.example .env
```

| Variabel     | Default                          | Keterangan                                |
| :----------- | :------------------------------- | :---------------------------------------- |
| `PORT`       | `8082`                           | Port server HTTP                          |
| `API_KEY`    | _(wajib diisi)_                  | Kunci rahasia untuk header `X-API-Key`    |
| `SMTP_HOST`  | `smtp.mail.me.com`               | Host SMTP relay iCloud                    |
| `SMTP_PORT`  | `587`                            | Port SMTP (STARTTLS)                      |
| `SMTP_USER`  | `mhmdandri03@icloud.com`         | Apple ID / iCloud email                   |
| `SMTP_PASS`  | _(wajib diisi)_                  | 16-digit App-Specific Password dari Apple |
| `FROM_EMAIL` | `no-reply@weddingofandricica.me` | Alamat email pengirim                     |
| `FROM_NAME`  | `Andri & Cica`                   | Nama pengirim email                       |

> [!IMPORTANT]
> **Cara Membuat App-Specific Password Apple:**
>
> 1. Buka [appleid.apple.com](https://appleid.apple.com) dan login dengan Apple ID Anda.
> 2. Masuk ke menu **Sign-In and Security** -> **App-Specific Passwords**.
> 3. Buat password baru (misal: "Mail API Relay") dan salin 16 digit password (contoh format: `abcd-efgh-ijkl-mnop`).
> 4. Tempelkan ke nilai `SMTP_PASS` di file `.env`.

---

## 🚀 Cara Menjalankan

### Opsi 1: Menggunakan Docker Compose (Direkomendasikan)

1. Pastikan file `.env` sudah terisi dengan benar.
2. Jalankan container:
   ```bash
   docker compose up -d --build
   ```
3. Cek log service:
   ```bash
   docker compose logs -f
   ```
4. Matikan service:
   ```bash
   docker compose down
   ```

### Opsi 2: Menjalankan Secara Lokal (Go Native)

1. Pastikan Go versi 1.22+ terpasang:
   ```bash
   go version
   ```
2. Jalankan unit test:
   ```bash
   go test -v ./...
   ```
3. Jalankan aplikasi:
   ```bash
   go run ./cmd/api
   ```

---

## 📡 Dokumentasi Endpoint

### 1. Health Check

- **URL**: `/health`
- **Method**: `GET`
- **Response**: `200 OK`

```json
{
  "status": "success",
  "message": "Service is healthy"
}
```

---

### 2. Kirim Email

- **URL**: `/api/send-email`
- **Method**: `POST`
- **Headers**:
  - `Content-Type: application/json`
  - `X-API-Key: <SECRET_KEY>`
- **Request Body (JSON)**:
  ```json
  {
    "to": "tamu@example.com",
    "subject": "Konfirmasi Kehadiran - Andri & Cica",
    "text": "Halo! Terima kasih telah melakukan konfirmasi kehadiran.",
    "html": "<h2>Halo!</h2><p>Terima kasih telah melakukan konfirmasi kehadiran pada pernikahan <b>Andri & Cica</b>.</p>"
  }
  ```

#### Aturan Validasi Payload:

- `to`: Wajib diisi dan harus berupa format email yang valid (RFC 5322).
- `subject`: Wajib diisi.
- `text` / `html`: Minimal salah satu harus diisi. Jika keduanya diisi, email akan dikirim dengan format multipart/alternative (fallback teks + tampilan HTML).

#### Status Respon:

- `200 OK`: Email berhasil dikirim ke server relay iCloud.
  ```json
  {
    "status": "success",
    "message": "Email berhasil dikirim"
  }
  ```
- `400 Bad Request`: Format JSON rusak atau field validasi gagal.
  ```json
  {
    "status": "error",
    "message": "Field 'to' wajib diisi"
  }
  ```
- `401 Unauthorized`: Header `X-API-Key` tidak ada atau salah.
  ```json
  {
    "status": "error",
    "message": "Unauthorized: API Key tidak valid atau tidak disertakan"
  }
  ```
- `500 Internal Server Error`: Autentikasi atau koneksi SMTP relay iCloud gagal.
  ```json
  {
    "status": "error",
    "message": "Gagal mengirim email: kegagalan pengiriman SMTP: 535 5.7.8 Error: authentication failed"
  }
  ```

---

## 🧪 Contoh Pengujian dengan `curl`

### 1. Pengiriman Sukses (HTML + Plaintext Fallback)

```bash
curl -X POST http://localhost:8082/api/send-email \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-secure-api-key-here" \
  -d '{
    "to": "recipient@example.com",
    "subject": "Konfirmasi Kehadiran - Andri & Cica",
    "text": "Terima kasih atas konfirmasi kehadiran Anda.",
    "html": "<div style=\"font-family:sans-serif;\"><h2>Konfirmasi Kehadiran</h2><p>Terima kasih atas konfirmasi kehadiran Anda pada pernikahan <strong>Andri &amp; Cica</strong>.</p></div>"
  }'
```

### 2. Pengujian Gagal Autentikasi (401 Unauthorized)

```bash
curl -i -X POST http://localhost:8082/api/send-email \
  -H "Content-Type: application/json" \
  -H "X-API-Key: key-yang-salah" \
  -d '{
    "to": "recipient@example.com",
    "subject": "Tes Auth",
    "text": "Percobaan tanpa otorisasi"
  }'
```

### 3. Pengujian Validasi Gagal (400 Bad Request)

```bash
curl -i -X POST http://localhost:8082/api/send-email \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-secure-api-key-here" \
  -d '{
    "to": "bukan-alamat-email",
    "subject": "Tes Email Tidak Valid",
    "text": "Testing"
  }'
```

### 4. Cek Health Server

```bash
curl -i http://localhost:8082/health
```
