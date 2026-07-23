# GoTask — REST API Task Management

A production-ready task management REST API built with Go, featuring workflow-based task tracking similar to mini-Jira.

---

## 📑 Daftar Isi

1. [✨ Fitur](#features)
2. [🔧 Technology Stack](#technology-stack)
3. [📁 Struktur Folder](#project-structure)
4. [🧠 Panduan untuk Pemula (Bahasa Indonesia)](#-panduan-struktur-folder--urutan-coding-untuk-pemula)
   - [📁 Gambaran Besar](#-gambaran-besar-ibarat-gedung-perkantoran)
   - [🧱 Arsitektur Modul (Rumus 6M)](#-arsitektur-dalam-setiap-modul-pola-6-file)
   - [🐣 Urutan Membuat dari Nol](#-urutan-membuat-project-dari-nol-step-by-step)
   - [❌ 5 Kesalahan Umum](#-5-kesalahan-umum-pemula-jangan-ditiru)
   - [✅ Ceklis Sebelum Lanjut](#-ceklis-sebelum-pindah-ke-modul-berikutnya)
   - [📐 Diagram Alur](#-diagram-alur-coding-dari-nol-sampai-selesai)
   - [🙋 FAQ](#-faq--pertanyaan-yang-sering-muncul)
   - [💡 Tips Pengembangan](#-tips-pengembangan)
5. [⚙️ Setup & Instalasi](#setup)
   - [Prasyarat](#prerequisites)
   - [Environment Variables](#environment-variables)
   - [Docker Compose](#running-with-docker-compose-recommended)
   - [Running Locally](#running-locally)
   - [☁️ Opsi Supabase (Cloud)](#-opsi-pakai-supabase-postgresql-cloud--tanpa-install-lokal)
6. [▶️ Menjalankan & Menghentikan Server](#-menjalankan--menghentikan-server)
7. [🌐 Deployment (Koyeb)](#-deployment-ke-koyeb-gratis)
8. [📡 API Endpoints](#api-endpoints)
9. [📮 Contoh Request/Response](#api-examples)
10. [🔄 Task Workflow](#task-workflow)
11. [📋 Business Rules](#business-rules)
12. [🗺️ ERD (Diagram Database)](#erd-entity-relationship-diagram)
13. [🏗️ Arsitektur](#architecture)
14. [🔒 Keamanan](#security)
15. [⚡ Makefile Commands](#makefile-commands)
16. [🛠️ Troubleshooting](#-troubleshooting)
17. [🧪 Testing](#testing)
18. [📦 Postman Collection](#-postman-collection)
19. [📖 Kamus Istilah](#-kamus-istilah-glossary)
20. [🤝 Berkontribusi](#-berkontribusi-contribution-guide)
21. [📊 Status Project](#-status-project)

---

## Features

- 🔐 **JWT Authentication** — Register, login, refresh token rotation, logout
- 📋 **Task Lists** — Create and manage multiple task boards
- ✅ **Task Management** — Full CRUD with filtering, pagination, search
- 🔄 **Workflow Engine** — Backlog → Todo → In Progress → Review → Done
- 📊 **Progress Tracking** — Track task completion percentage with notes
- 🔍 **Review System** — Submit for review, approve, or request changes
- 💬 **Comments** — Add and manage comments on tasks
- 📜 **Activity Log** — Automatic audit trail for all task changes
- 📈 **Dashboard** — Summary, progress analytics, deadlines, priority distribution

## Technology Stack

| Component        | Technology              |
| ---------------- | ----------------------- |
| Language         | Go 1.24                 |
| HTTP Router      | go-chi/chi v5           |
| Database         | PostgreSQL 16           |
| Migrations       | golang-migrate          |
| Query Generator  | sqlc                    |
| Validation       | go-playground/validator |
| Authentication   | JWT (golang-jwt)        |
| Password Hashing | bcrypt (x/crypto)       |
| Logging          | log/slog (structured)   |
| UUID             | google/uuid             |
| Containerization | Docker + Docker Compose |

## Project Structure

```
gotask/
├── cmd/api/main.go              # Application entry point
├── internal/
│   ├── activity/                # Activity log module
│   │   ├── handler.go, service.go, repository.go, repository_impl.go, model.go
│   ├── auth/                    # Authentication module
│   │   ├── handler.go, service.go, repository.go, repository_impl.go
│   │   ├── model.go, dto.go, token.go, password.go
│   ├── comment/                 # Comment module
│   │   ├── handler.go, repository_impl.go (includes Service), model.go
│   ├── dashboard/               # Dashboard analytics module
│   │   ├── handler.go, service.go
│   ├── middleware/               # HTTP middleware
│   │   ├── auth.go, logging.go, recovery.go, request_id.go
│   │   ├── cors.go, content_type.go, body_limit.go, rate_limiter.go
│   ├── platform/                # Platform utilities
│   │   ├── config/config.go
│   │   ├── database/postgres.go
│   │   ├── logger/logger.go
│   │   ├── response/response.go
│   │   └── validator/validator.go
│   ├── progress/                # Progress tracking module
│   │   ├── handler.go, service.go, repository.go, repository_impl.go
│   │   ├── model.go, dto.go
│   ├── review/                  # Review module
│   │   ├── handler.go, service.go, repository.go, repository_impl.go, model.go
│   ├── task/                    # Task module
│   │   ├── handler.go, service.go, repository.go, repository_impl.go
│   │   ├── model.go, dto.go
│   └── tasklist/                # Task List module
│       ├── handler.go, service.go, repository.go, repository_impl.go
│       ├── model.go, dto.go
├── db/
│   ├── migrations/              # PostgreSQL migrations
│   │   └── 000001_init_schema.{up,down}.sql
│   ├── queries/                 # SQL queries for sqlc
│   │   ├── activity.sql, auth.sql, comment.sql
│   │   ├── dashboard.sql, health.sql, progress.sql
│   │   ├── review.sql, task.sql, task_list.sql
│   └── sqlc.yaml                # sqlc configuration
├── Dockerfile                   # Multi-stage Docker build
├── docker-compose.yml           # Docker Compose setup
├── Makefile                     # Development commands
├── .env.example                 # Environment variables template
└── README.md
```

---

## 🧠 Panduan Struktur Folder & Urutan Coding (Untuk Pemula)

> **Bahasa Indonesia** — Penjelasan sederhana tentang cara membuat project GoTask dari nol, langkah demi langkah.

### 📁 Gambaran Besar: Ibarat Gedung Perkantoran

Bayangkan GoTask seperti sebuah **gedung perkantoran**:

| Folder                 | Ibarat                       | Fungsi                                                |
| ---------------------- | ---------------------------- | ----------------------------------------------------- |
| `cmd/api/main.go`      | 🚪 **Pintu masuk**           | Tombol START untuk menjalankan aplikasi               |
| `internal/platform/`   | ⚙️ **Ruang utilitas**        | Listrik, air, AC (config, logger, database)           |
| `internal/middleware/` | 🛂 **Pos satpam**            | Pengecekan sebelum masuk (auth, rate limit, CORS)     |
| `internal/auth/`       | 🔑 **Ruang login**           | Register, login, lupa password                        |
| `internal/tasklist/`   | 📋 **Papan kerja**           | Tempat menempel daftar tugas                          |
| `internal/task/`       | ✅ **Kartu tugas**           | Satu per satu pekerjaan yang harus diselesaikan       |
| `internal/progress/`   | 📊 **Laporan harian**        | Catatan "sudah selesai berapa persen"                 |
| `internal/review/`     | 🔍 **Ruang pemeriksa**       | Atasan periksa hasil kerja, approve atau minta revisi |
| `internal/comment/`    | 💬 **Post-it note**          | Tempelan diskusi di setiap kartu tugas                |
| `internal/activity/`   | 📜 **Buku sejarah**          | Catatan otomatis semua perubahan yang terjadi         |
| `internal/dashboard/`  | 📈 **Layar ringkasan**       | Grafik dan angka rekapitulasi                         |
| `db/migrations/`       | 📝 **Cetak biru database**   | Skema tabel yang bisa di-_upgrade_                    |
| `db/queries/`          | 📄 **Kumpulan perintah SQL** | Mentahan query yang akan di-generate oleh sqlc        |

---

### 🧱 Arsitektur Dalam Setiap Modul: Pola 6 File

Setiap modul (seperti `auth/`, `task/`, dll.) memiliki **6 file standar** yang bekerja dalam rantai:

```
┌──────────────────────────────────────────────────────────────┐
│                     ALUR KERJA STANDAR                         │
│                                                               │
│   handler.go   →   service.go   →   repository.go             │
│      ↑                 ↑                  ↑                   │
│  (terima HTTP)    (logika bisnis)   (simpan/ambil data)        │
│                                                               │
│   dto.go          model.go           repository_impl.go       │
│  (formulir)      (cetakan data)      (kode SQL asli)           │
└──────────────────────────────────────────────────────────────┘
```

| File                 | Analogi Sederhana    | Isinya                                                      |
| -------------------- | -------------------- | ----------------------------------------------------------- |
| `model.go`           | 🧁 **Cetakan kue**   | Struktur data asli (ID, nama, email, password, dll.)        |
| `dto.go`             | 📝 **Formulir**      | Data yang dikirim/diterima user (Request & Response)        |
| `repository.go`      | 📜 **Kontrak kerja** | Interface — janji method apa saja yang harus ada            |
| `repository_impl.go` | 👷 **Pelaksana**     | Kode SQL sungguhan (INSERT, SELECT, UPDATE, DELETE)         |
| `service.go`         | 👔 **Manager**       | Aturan bisnis: validasi, kalkulasi, pengambilan keputusan   |
| `handler.go`         | 📞 **Resepsionis**   | Terima HTTP request → panggil service → kirim JSON response |

> **💡 Kenapa dipisah?** Supaya kalau ada perubahan (misal ganti database dari PostgreSQL ke MySQL), kamu cukup ganti `repository_impl.go` saja. File lain tidak perlu disentuh!

---

### 🐣 Urutan Membuat Project dari Nol (Step by Step)

#### 🔰 FASE 0: Fondasi Project

```
Langkah 1: Buat folder & inisialisasi Go module
─────────────────────────────────────────────────
$ mkdir gotask && cd gotask
$ go mod init github.com/namamu/gotask
```

#### 🔰 FASE 1: Platform Layer (Paling Dasar — HARUS PERTAMA!)

**Jangan dibalik urutannya!** Karena setiap langkah bergantung pada langkah sebelumnya:

| No  | File                                 | Kenapa Harus Duluan?                                         |
| --- | ------------------------------------ | ------------------------------------------------------------ |
| 1   | `config.go`                          | Semua file butuh pengaturan (port, secret key, database URL) |
| 2   | `logger.go`                          | Supaya bisa mencatat error sejak awal development            |
| 3   | `postgres.go`                        | Semua data disimpan di database, wajib connect dulu          |
| 4   | `response.go`                        | Semua jawaban ke user harus seragam formatnya                |
| 5   | `validator.go`                       | Semua input user harus divalidasi                            |
| 6   | Middleware (8 file)                  | "Gerbang" pengecekan sebelum request masuk ke handler        |
| 7   | File migrasi `.sql`                  | Tabel harus sudah dibuat sebelum bisa menyimpan data         |
| 8   | File query `.sql`                    | Repository butuh query SQL yang sudah siap                   |
| 9   | `main.go`                            | "Otak" yang merangkai semua komponen menjadi satu            |
| 10  | Dockerfile, docker-compose, Makefile | Untuk deployment dan kemudahan development                   |

#### 🔰 FASE 2: Modul Authentication (Login Dulu!)

Ini modul **paling kritis** — karena SEMUA modul lain butuh user yang sudah login.

**Urutan dalam 1 modul (Rumus 6M):**

| Urutan | Pertanyaan Kunci                | File                        |
| ------ | ------------------------------- | --------------------------- |
| **M1** | "Datanya bentuknya gimana?"     | `model.go`                  |
| **M2** | "User ngirim/terima data apa?"  | `dto.go`                    |
| **M3** | "Fitur apa saja yang tersedia?" | `repository.go` (interface) |
| **M4** | "SQL-nya gimana?"               | `repository_impl.go`        |
| **M5** | "Aturan mainnya apa?"           | `service.go`                |
| **M6** | "Gimana terima request HTTP?"   | `handler.go`                |
| ⚠️     | **DAFTARKAN KE `main.go`!**     | Edit `main.go`              |
| ✅     | Tes unit                        | `service_test.go`           |

#### 🔰 FASE 3–6: Modul Berikutnya (Pola Sama)

Setelah auth beres, **ulangi Rumus 6M** untuk setiap modul:

```
Phase 3: tasklist → task         (tasklist dulu, karena task ada DI DALAM list)
Phase 4: progress → activity     (activity dibuat otomatis oleh backend)
Phase 5: review                  (bergantung pada task & activity)
Phase 6: comment → dashboard     (dashboard baca agregat dari task)
```

---

### ❌ 5 Kesalahan Umum Pemula (Jangan Ditiru!)

| #   | Kesalahan                                             | Akibatnya                                                     |
| --- | ----------------------------------------------------- | ------------------------------------------------------------- |
| 1   | **Bikin `handler.go` duluan** sebelum `repository.go` | Bingung sendiri: "datanya dari mana?!"                        |
| 2   | **Skip bikin interface** (`repository.go`)            | Susah ganti database & susah bikin unit test                  |
| 3   | **Lupa daftarin ke `main.go`**                        | Fitur sudah jadi 100% tapi tidak jalan 😭                     |
| 4   | **Nulis SQL langsung di `handler.go`**                | Kode berantakan, rawan SQL injection, susah di-_debug_        |
| 5   | **Tidak pakai database transaction**                  | Data bisa "setengah jadi" kalau terjadi error di tengah jalan |

---

### ✅ Ceklis Sebelum Pindah ke Modul Berikutnya

Sebelum lanjut mengerjakan modul baru, pastikan **5 hal ini**:

```
☐ go build ./...        → Kompilasi SUKSES, tidak ada error
☐ go fmt ./...          → Kode sudah rapi formatnya
☐ go vet ./...          → Tidak ada warning dari Go
☐ go test ./...         → Semua unit test LULUS
☐ Handler terdaftar     → Endpoint baru sudah ada di main.go
```

---

### 📐 Diagram Alur Coding (Dari Nol Sampai Selesai)

```mermaid
graph TD
    A[Mulai] --> B[Buat go mod init]
    B --> C[FASE 1: Platform Layer]
    C --> C1[config.go]
    C1 --> C2[logger.go]
    C2 --> C3[postgres.go]
    C3 --> C4[response.go]
    C4 --> C5[validator.go]
    C5 --> C6[middleware/*]
    C6 --> C7[migrations/*.sql]
    C7 --> C8[queries/*.sql]
    C8 --> C9[main.go]
    C9 --> D{Fase 1 ceklis}
    D -->|Lulus| E[FASE 2: Auth Module]
    D -->|Error| C3

    E --> E1[model.go]
    E1 --> E2[dto.go]
    E2 --> E3[repository.go]
    E3 --> E4[repository_impl.go]
    E4 --> E5[service.go]
    E5 --> E6[handler.go]
    E6 --> E7[Daftar ke main.go]
    E7 --> E8[service_test.go]
    E8 --> F{Fase 2 ceklis}

    F -->|Lulus| G[FASE 3-6: Modul Lainnya]
    F -->|Error| E4

    G --> G1[tasklist]
    G1 --> G2[task]
    G2 --> G3[progress]
    G3 --> G4[activity]
    G4 --> G5[review]
    G5 --> G6[comment]
    G6 --> G7[dashboard]

    G7 --> H[SELESAI 🎉]
```

---

### 🙋 FAQ — Pertanyaan yang Sering Muncul

<details>
<summary><b>Q: Kenapa harus pakai interface di repository?</b></summary>

**A:** Supaya kode mudah diuji (unit test). Saat testing, kita bisa pakai repository "palsu" (mock) yang tidak benar-benar menyimpan ke database. Juga memudahkan kalau suatu saat ingin ganti database (dari PostgreSQL ke MySQL, misalnya) — cukup ganti `repository_impl.go`, file lain tidak perlu diubah.

</details>

<details>
<summary><b>Q: Kenapa password harus di-hash? Kenapa tidak disimpan apa adanya?</b></summary>

**A:** Kalau database bocor, penyerang tidak bisa langsung tahu password asli user. Hashing itu satu arah — dari password jadi hash gampang, tapi dari hash kembali ke password hampir mustahil. bcrypt bahkan sengaja dibuat lambat supaya penyerang tidak bisa "menebak" password secara brutal (brute force).

</details>

<details>
<summary><b>Q: Bedanya access token dan refresh token apa?</b></summary>

**A:** Ibarat kunci hotel. **Access token** = kunci kamar, berlaku sebentar (15 menit). Kalau hilang, pencuri hanya bisa akses sebentar. **Refresh token** = kartu identitas di resepsionis, berlaku lama (30 hari) untuk minta kunci baru tanpa perlu login ulang. Refresh token juga di-_rotate_ (diganti setiap kali dipakai) untuk mencegah pencurian.

</details>

<details>
<summary><b>Q: Soft delete itu apa? Kenapa tidak hard delete saja?</b></summary>

**A:** Soft delete = data tidak benar-benar dihapus, hanya diberi tanda `deleted_at = sekarang`. Ini penting supaya: (1) data bisa dikembalikan kalau tidak sengaja terhapus, (2) activity log tetap punya referensi ke task yang "dihapus", (3) memenuhi kebutuhan audit trail.

</details>

<details>
<summary><b>Q: Kenapa setiap query harus menyertakan `user_id`?</b></summary>

**A:** Ini disebut **ownership authorization**. Meskipun seorang user tahu UUID task milik orang lain, dia tetap tidak bisa mengaksesnya karena query selalu mengecek: "apakah user ini pemilik task list yang menaungi task tersebut?" Contoh:

```sql
SELECT * FROM tasks t
JOIN task_lists tl ON tl.id = t.list_id
WHERE t.id = $1 AND tl.user_id = $2  -- ← BARIS INI KRUSIAL!
```

Tanpa pengecekan ini, user A bisa membaca data user B hanya dengan menebak UUID.

</details>

<details>
<summary><b>Q: Gimana cara nambah fitur baru?</b></summary>

**A:** Ikuti Rumus 6M:

1. Tentukan dulu bentuk datanya → `model.go`
2. Buat formulir request/response → `dto.go`
3. Tentukan method di interface → `repository.go`
4. Tulis SQL-nya → `repository_impl.go`
5. Tulis aturan bisnis → `service.go`
6. Buat HTTP handler → `handler.go`
7. **Jangan lupa daftarin ke `main.go`!**
8. Buat unit test → `service_test.go`
</details>

---

### 💡 Tips Pengembangan

1. **Gunakan `make` commands** — `make run`, `make test`, `make fmt` akan menghemat waktumu.
2. **Jalankan `go vet` sebelum commit** — ini seperti spell-checker untuk kode Go.
3. **Baca pesan error dengan teliti** — Go memberikan pesan error yang sangat deskriptif.
4. **Gunakan Postman atau curl** untuk mengetes endpoint satu per satu.
5. **Commit sering-sering** — lebih baik commit 10x kecil daripada 1x raksasa.
6. **Jangan menonaktifkan unit test hanya supaya build berhasil** — perbaiki kodenya, bukan testnya.
7. **Pakai Docker Compose kalau malas install PostgreSQL manual** — satu command: `docker compose up -d`.

---

## Setup

### Prerequisites

- Go 1.24+
- PostgreSQL 16+
- Docker & Docker Compose (optional)
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI
- [sqlc](https://sqlc.dev/) CLI

### Environment Variables

Copy `.env.example` to `.env`:

```bash
cp .env.example .env
```

Configure:

| Variable               | Description                          | Default / Contoh                                                                                                                                       |
| ---------------------- | ------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `APP_ENV`              | Environment (development/production) | `development`                                                                                                                                          |
| `APP_PORT`             | HTTP port                            | `8080`                                                                                                                                                 |
| `DATABASE_URL`         | PostgreSQL connection URL            | Local: `postgres://gotask:gotask@localhost:5432/gotask?sslmode=disable`<br>Supabase: `postgresql://postgres:[PASS]@db.xxxxx.supabase.co:5432/postgres` |
| `JWT_ACCESS_SECRET`    | JWT access token secret              | `change-me`                                                                                                                                            |
| `JWT_REFRESH_SECRET`   | JWT refresh token secret             | `change-me`                                                                                                                                            |
| `JWT_ACCESS_TTL`       | Access token TTL                     | `15m`                                                                                                                                                  |
| `JWT_REFRESH_TTL`      | Refresh token TTL                    | `720h`                                                                                                                                                 |
| `CORS_ALLOWED_ORIGINS` | Allowed CORS origins                 | `http://localhost:3000`                                                                                                                                |

### Running with Docker Compose (Recommended)

```bash
# Start PostgreSQL and API
make docker-up

# Check health
curl http://localhost:8080/health

# Stop
make docker-down
```

### Running Locally

```bash
# Start PostgreSQL (via Docker or locally)
docker compose up -d postgres

# Run migrations
make migrate-up

# Run the application
make run
```

### ☁️ Opsi: Pakai Supabase (PostgreSQL Cloud — Tanpa Install Lokal)

Malas install PostgreSQL? Pakai [Supabase](https://supabase.com) (gratis 500MB). Cukup ganti `DATABASE_URL` di `.env`:

```bash
# .env
DATABASE_URL=postgresql://postgres:[PASSWORD-KAMU]@db.xxxxx.supabase.co:5432/postgres
```

**Cara dapat URL-nya:**

1. Buka [supabase.com/dashboard](https://supabase.com/dashboard) → bikin project (gratis)
2. Setelah project siap, buka **Settings → Database**
3. Scroll ke **Connection string** → copy yang format `postgresql://`
4. Ganti `[YOUR-PASSWORD]` dengan password yang kamu set saat bikin project

**Jalankan migrasi ke Supabase:**

```bash
export DATABASE_URL="postgresql://postgres:[PASSWORD]@db.xxxxx.supabase.co:5432/postgres"
migrate -path db/migrations -database "$DATABASE_URL" up
```

**Keuntungan pakai Supabase:**

- ✅ Tidak perlu install PostgreSQL di laptop
- ✅ Database online, bisa diakses dari mana saja
- ✅ Gratis sampai 500MB
- ✅ Ada dashboard GUI untuk lihat isi tabel
- ✅ Tetap pakai PostgreSQL standar — semua query GoTask langsung kompatibel

> 💡 **Driver `lib/pq` yang dipakai GoTask sudah kompatibel dengan Supabase.** Tidak perlu install `pgx` atau ganti kode apapun.

### Generate sqlc

```bash
make sqlc
```

---

## ▶️ Menjalankan & Menghentikan Server

> **Untuk sehari-hari setelah instalasi.** Simpan perintah ini, kamu akan sering memakainya.

### 1️⃣ Install Tools & Database (sekali saja)

```bash
# Install PostgreSQL 16 dan golang-migrate
brew install postgresql@16 golang-migrate

# Start PostgreSQL (jalan di background)
brew services start postgresql@16

# Buat user & database
/opt/homebrew/opt/postgresql@16/bin/psql -d postgres <<SQL
CREATE USER gotask WITH PASSWORD 'gotask';
CREATE DATABASE gotask OWNER gotask;
GRANT ALL PRIVILEGES ON DATABASE gotask TO gotask;
SQL

# Copy .env (sekali saja)
cp .env.example .env

# Jalankan migrasi
export $(grep -v '^#' .env | xargs)
migrate -path db/migrations -database "$DATABASE_URL" up
```

### 2️⃣ Menjalankan Server (setiap kali coding)

```bash
# Buka terminal di folder project
cd /Volumes/SSD/LEARN/GO/GoTask

# Jalankan server
GOPATH=$HOME/go GOMODCACHE=$HOME/go/pkg/mod \
  go run ./cmd/api
```

**Kalau berhasil, akan muncul:**

```
time=... level=INFO msg="starting GoTask API" env=development
time=... level=INFO msg="connected to PostgreSQL"
time=... level=INFO msg="server started" port=8080
```

> ⚠️ **Server tidak otomatis baca file `.env`** — environment variable harus di-export dulu. Lihat langkah 3 di bawah.

### 3️⃣ Cara Paling Simpel (pakai .env otomatis)

Buat file `run.sh` di root project:

```bash
cat > run.sh << 'EOF'
#!/bin/bash
# Matikan server lama (kalau ada)
pkill -f gotask 2>/dev/null
sleep 1

# Export env dari .env dan jalankan
cd "$(dirname "$0")"
set -a; source .env; set +a
export GOPATH=$HOME/go
export GOMODCACHE=$HOME/go/pkg/mod

echo "▶️  Menjalankan GoTask di http://localhost:${APP_PORT:-8080}"
echo "   Tekan Ctrl+C untuk berhenti"
echo ""
go run ./cmd/api
EOF

chmod +x run.sh
```

Sekarang tinggal:

```bash
./run.sh
```

### 4️⃣ Menghentikan Server

| Cara                   | Perintah                                           |
| ---------------------- | -------------------------------------------------- |
| **Ctrl+C**             | Tekan `Ctrl+C` di terminal yang menjalankan server |
| **Dari terminal lain** | `pkill -f gotask`                                  |
| **Cek masih jalan?**   | `lsof -i :8080`                                    |

### 5️⃣ Cek Apakah Server Sudah Berjalan

```bash
# Cek port 8080
lsof -i :8080

# Kalau ada output → server masih hidup
# Kalau kosong     → aman, bisa start baru
```

### 6️⃣ Restart Cepat (stop + start)

```bash
pkill -f gotask 2>/dev/null
sleep 1
go run ./cmd/api
```

### 7️⃣ Tes Server Berjalan

```bash
# Health check
curl http://localhost:8080/health

# Response yang diharapkan:
# {"success":true,"message":"OK","data":{"status":"ok"}}
```

---

## 🌐 Deployment ke Koyeb (Gratis)

Biar API bisa diakses 24 jam dari mana saja — tanpa laptop menyala, tanpa `go run`.

### Kenapa Koyeb?

| Kelebihan | Detail |
|-----------|--------|
| ✅ Gratis selamanya | Gak perlu kartu kredit |
| ✅ Server gak tidur | Berbeda dengan Render |
| ✅ Deploy dari GitHub | Tinggal connect repo |
| ✅ Auto-deploy | Push ke GitHub → otomatis update |

### Cara Deploy (5 menit)

**1. Buka [koyeb.com](https://www.koyeb.com) → Sign Up with GitHub**

**2. Klik Create App → pilih GitHub → pilih repo `IqromAbadi/GoTask`**

**3. Setting:**

| Setting | Isi |
|---------|-----|
| Type | Dockerfile |
| Port | `8080` |
| Health Check | `GET /health` |

**4. Environment Variables (copy dari `.env`):**

```
APP_ENV=production
APP_PORT=8080
DATABASE_URL=postgresql://postgres.fkuzsqncaacvpbbjiwgd:PASSWORD@aws-1-ap-northeast-2.pooler.supabase.com:5432/postgres?sslmode=require
JWT_ACCESS_SECRET=your-secret-here
JWT_REFRESH_SECRET=your-secret-here
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=720h
CORS_ALLOWED_ORIGINS=*
```

**5. Klik Deploy** 🚀

Tunggu 2-3 menit, dapat URL: `https://gotask-xxxxx.koyeb.app`

**6. Tes:**

```bash
curl https://gotask-xxxxx.koyeb.app/health
# → {"success":true,"message":"OK","data":{"status":"ok"}}
```

> 💡 **Setelah deploy, frontend kamu bisa pakai `https://gotask-xxxxx.koyeb.app/api/v1` sebagai base URL.**

---

## API Endpoints

### Health

| Method | Path      | Auth   | Description  |
| ------ | --------- | ------ | ------------ |
| GET    | `/health` | Public | Health check |

### Authentication

| Method | Path                    | Auth   | Description          |
| ------ | ----------------------- | ------ | -------------------- |
| POST   | `/api/v1/auth/register` | Public | Register new user    |
| POST   | `/api/v1/auth/login`    | Public | Login                |
| POST   | `/api/v1/auth/refresh`  | Public | Refresh access token |
| POST   | `/api/v1/auth/logout`   | Public | Logout               |

### User

| Method | Path                        | Auth     | Description     |
| ------ | --------------------------- | -------- | --------------- |
| GET    | `/api/v1/users/me`          | Required | Get profile     |
| PATCH  | `/api/v1/users/me`          | Required | Update profile  |
| PATCH  | `/api/v1/users/me/password` | Required | Change password |

### Task Lists

| Method | Path                             | Auth     | Description       |
| ------ | -------------------------------- | -------- | ----------------- |
| POST   | `/api/v1/lists`                  | Required | Create task list  |
| GET    | `/api/v1/lists`                  | Required | List task lists   |
| GET    | `/api/v1/lists/{listId}`         | Required | Get task list     |
| PATCH  | `/api/v1/lists/{listId}`         | Required | Update task list  |
| DELETE | `/api/v1/lists/{listId}`         | Required | Delete task list  |
| PATCH  | `/api/v1/lists/{listId}/archive` | Required | Archive task list |
| PATCH  | `/api/v1/lists/{listId}/restore` | Required | Restore task list |
| GET    | `/api/v1/lists/{listId}/board`   | Required | Get board view    |

### Tasks

| Method | Path                              | Auth     | Description               |
| ------ | --------------------------------- | -------- | ------------------------- |
| POST   | `/api/v1/lists/{listId}/tasks`    | Required | Create task               |
| GET    | `/api/v1/lists/{listId}/tasks`    | Required | List tasks (with filters) |
| GET    | `/api/v1/tasks/{taskId}`          | Required | Get task                  |
| PATCH  | `/api/v1/tasks/{taskId}`          | Required | Update task               |
| DELETE | `/api/v1/tasks/{taskId}`          | Required | Soft delete task          |
| PATCH  | `/api/v1/tasks/{taskId}/status`   | Required | Change task status        |
| PATCH  | `/api/v1/tasks/{taskId}/priority` | Required | Change task priority      |
| POST   | `/api/v1/tasks/{taskId}/reopen`   | Required | Reopen completed task     |

**Task Filters**: `?status=in_progress&priority=high&search=login&due_date_from=2026-07-01&due_date_to=2026-07-31&is_overdue=true&sort_by=due_date&sort_order=asc&page=1&limit=20`

### Progress

| Method | Path                                           | Auth     | Description            |
| ------ | ---------------------------------------------- | -------- | ---------------------- |
| POST   | `/api/v1/tasks/{taskId}/progress`              | Required | Add progress update    |
| GET    | `/api/v1/tasks/{taskId}/progress`              | Required | List progress updates  |
| GET    | `/api/v1/tasks/{taskId}/progress/{progressId}` | Required | Get progress update    |
| PATCH  | `/api/v1/tasks/{taskId}/progress/{progressId}` | Required | Update progress note   |
| DELETE | `/api/v1/tasks/{taskId}/progress/{progressId}` | Required | Delete progress update |

### Review

| Method | Path                                                        | Auth     | Description       |
| ------ | ----------------------------------------------------------- | -------- | ----------------- |
| POST   | `/api/v1/tasks/{taskId}/submit-review`                      | Required | Submit for review |
| GET    | `/api/v1/tasks/{taskId}/reviews`                            | Required | List reviews      |
| GET    | `/api/v1/tasks/{taskId}/reviews/{reviewId}`                 | Required | Get review        |
| POST   | `/api/v1/tasks/{taskId}/reviews/{reviewId}/approve`         | Required | Approve review    |
| POST   | `/api/v1/tasks/{taskId}/reviews/{reviewId}/request-changes` | Required | Request changes   |

### Comments

| Method | Path                                          | Auth     | Description    |
| ------ | --------------------------------------------- | -------- | -------------- |
| POST   | `/api/v1/tasks/{taskId}/comments`             | Required | Add comment    |
| GET    | `/api/v1/tasks/{taskId}/comments`             | Required | List comments  |
| PATCH  | `/api/v1/tasks/{taskId}/comments/{commentId}` | Required | Update comment |
| DELETE | `/api/v1/tasks/{taskId}/comments/{commentId}` | Required | Delete comment |

### Activity

| Method | Path                                | Auth     | Description     |
| ------ | ----------------------------------- | -------- | --------------- |
| GET    | `/api/v1/tasks/{taskId}/activities` | Required | Task activities |
| GET    | `/api/v1/activities`                | Required | User activities |

### Dashboard

| Method | Path                                      | Auth     | Description           |
| ------ | ----------------------------------------- | -------- | --------------------- |
| GET    | `/api/v1/dashboard/summary`               | Required | Task summary          |
| GET    | `/api/v1/dashboard/progress`              | Required | Progress analytics    |
| GET    | `/api/v1/dashboard/upcoming-deadlines`    | Required | Upcoming deadlines    |
| GET    | `/api/v1/dashboard/overdue-tasks`         | Required | Overdue tasks         |
| GET    | `/api/v1/dashboard/priority-distribution` | Required | Priority distribution |

## API Examples

### Register

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Agung","email":"agung@example.com","password":"Password123!"}'
```

### Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"agung@example.com","password":"Password123!"}'
```

Response:

```json
{
  "success": true,
  "message": "Login berhasil",
  "data": {
    "access_token": "eyJhbGci...",
    "refresh_token": "dGhpcyBp...",
    "expires_in": 900,
    "token_type": "Bearer"
  }
}
```

### Create Task List

```bash
curl -X POST http://localhost:8080/api/v1/lists \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -d '{"name":"Sprint 1","description":"First sprint tasks"}'
```

### Create Task

```bash
curl -X POST http://localhost:8080/api/v1/lists/{listId}/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -d '{"title":"Implement login","priority":"high","status":"todo","due_date":"2026-07-30","estimated_minutes":240}'
```

### Add Progress

```bash
curl -X POST http://localhost:8080/api/v1/tasks/{taskId}/progress \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -d '{"progress":60,"note":"Login and register endpoints done"}'
```

### Submit for Review

```bash
curl -X POST http://localhost:8080/api/v1/tasks/{taskId}/submit-review \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -d '{"submission_note":"All auth endpoints completed and tested"}'
```

## Task Workflow

```mermaid
stateDiagram-v2
    [*] --> backlog
    backlog --> todo
    todo --> backlog
    todo --> in_progress
    in_progress --> todo
    in_progress --> review: progress = 100%
    review --> in_progress
    review --> done: approved
    done --> in_progress: reopen
```

### Valid Status Transitions

| From        | To                   |
| ----------- | -------------------- |
| backlog     | todo                 |
| todo        | backlog, in_progress |
| in_progress | todo, review         |
| review      | in_progress, done    |
| done        | in_progress          |

## Business Rules

1. Task can only enter `review` when progress is 100%
2. Task can only reach `done` through review approval
3. Status `done` cannot be set directly via status endpoint
4. Progress can only be added when task is `in_progress`
5. New progress must be ≥ previous progress (unless `allow_rollback: true` with a note)
6. Rollback requires a note explaining the reason
7. All status changes are logged to activity
8. Deleted tasks and comments use soft delete
9. Comments can only be edited/deleted by their owner
10. Each user can only access their own data

## ERD (Entity Relationship Diagram)

```mermaid
erDiagram
    users ||--o{ refresh_tokens : has
    users ||--o{ task_lists : owns
    task_lists ||--o{ tasks : contains
    users ||--o{ tasks : creates
    tasks ||--o{ task_progress_updates : has
    tasks ||--o{ task_reviews : has
    tasks ||--o{ task_comments : has
    tasks ||--o{ task_activities : logs
    users ||--o{ task_progress_updates : records
    users ||--o{ task_reviews : reviews
    users ||--o{ task_comments : writes
    users ||--o{ task_activities : performs

    users {
        uuid id PK
        string name
        string email UK
        string password_hash
        string avatar_url
        timestamp created_at
        timestamp updated_at
    }

    refresh_tokens {
        uuid id PK
        uuid user_id FK
        string token_hash
        timestamp expires_at
        timestamp revoked_at
        timestamp created_at
    }

    task_lists {
        uuid id PK
        uuid user_id FK
        string name
        text description
        boolean is_archived
        timestamp created_at
        timestamp updated_at
    }

    tasks {
        uuid id PK
        uuid list_id FK
        uuid created_by FK
        string title
        text description
        string status
        string priority
        int progress
        date due_date
        int estimated_minutes
        timestamp started_at
        timestamp completed_at
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    task_progress_updates {
        uuid id PK
        uuid task_id FK
        uuid user_id FK
        int progress
        text note
        timestamp created_at
        timestamp updated_at
    }

    task_reviews {
        uuid id PK
        uuid task_id FK
        uuid reviewer_id FK
        string status
        text submission_note
        text review_note
        timestamp reviewed_at
        timestamp created_at
        timestamp updated_at
    }

    task_comments {
        uuid id PK
        uuid task_id FK
        uuid user_id FK
        text content
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    task_activities {
        uuid id PK
        uuid task_id FK
        uuid user_id FK
        string action
        text old_value
        text new_value
        jsonb metadata
        timestamp created_at
    }
```

## Architecture

The project follows a modular layered architecture:

```mermaid
graph TD
    HTTP[HTTP Request] --> MW[Middleware Stack]
    MW --> R[Router chi]
    R --> H[Handler]
    H --> S[Service]
    S --> REP[Repository]
    REP --> DB[(PostgreSQL)]

    subgraph Middleware
        MW1[RequestID]
        MW2[Logging]
        MW3[Recovery]
        MW4[CORS]
        MW5[BodyLimit]
        MW6[ContentType]
        MW7[RateLimit]
        MW8[Authenticate]
    end

    subgraph Modules
        AUTH[Auth]
        USER[User]
        TL[TaskList]
        TASK[Task]
        PROG[Progress]
        REV[Review]
        COM[Comment]
        ACT[Activity]
        DASH[Dashboard]
    end
```

**Pattern**: Handler → Service → Repository → Database

Each module is self-contained with:

- `model.go` — Domain entity
- `dto.go` — Request/Response DTOs
- `repository.go` — Repository interface
- `repository_impl.go` — PostgreSQL implementation
- `service.go` — Business logic
- `handler.go` — HTTP handlers

## Security

- ✅ Password hashing with bcrypt
- ✅ JWT signature validation
- ✅ Access token expiration (15 min)
- ✅ Refresh token rotation
- ✅ Refresh token storage as hash (not plain text)
- ✅ Rate limiting on auth endpoints
- ✅ Generic login error (no email enumeration)
- ✅ Parameterized queries (SQL injection prevention)
- ✅ Input validation
- ✅ Ownership authorization (user-scoped queries)
- ✅ Soft delete
- ✅ Request body size limit (1MB)
- ✅ Panic recovery
- ✅ CORS configuration
- ✅ Request ID tracking
- ✅ Structured logging (no passwords/tokens)

## Makefile Commands

```bash
make run              # Run the application
make build            # Build binary
make test             # Run tests
make test-cover       # Run tests with coverage
make migrate-up       # Run database migrations up
make migrate-down     # Run database migrations down
make migrate-create name=xxx  # Create a new migration
make sqlc             # Generate sqlc code
make docker-up        # Start Docker Compose
make docker-down      # Stop Docker Compose
make lint             # Run linter
make fmt              # Format code
```

---

## 🛠️ Troubleshooting

### ❌ Port 8080 sudah digunakan (`bind: address already in use`)

**Penyebab:** Server sebelumnya masih berjalan di background.

**Cek apakah port sedang dipakai:**

```bash
lsof -i :8080
```

**Solusi — matikan server lama dulu:**

```bash
pkill -f gotask
```

Lalu jalankan ulang:

```bash
go run ./cmd/api
```

**Atau gunakan satu command untuk restart otomatis:**

```bash
pkill -f gotask 2>/dev/null; sleep 1; \
  export GOPATH=$HOME/go GOMODCACHE=$HOME/go/pkg/mod && \
  export $(grep -v '^#' .env | xargs) && \
  go run ./cmd/api
```

---

### ❌ Tabel database tidak ditemukan (`relation "users" does not exist`)

**Penyebab:** Migrasi belum dijalankan atau gagal.

**Cek status migrasi:**

```bash
psql -U gotask -d gotask -c "SELECT * FROM schema_migrations;"
```

**Cek daftar tabel:**

```bash
psql -U gotask -d gotask -c "\dt"
```

**Solusi — jalankan migrasi:**

```bash
export $(grep -v '^#' .env | xargs)
migrate -path db/migrations -database "$DATABASE_URL" up
```

---

### ❌ PostgreSQL tidak berjalan (`connection refused`)

**Cek status PostgreSQL:**

```bash
pg_isready
```

**Solusi — start PostgreSQL (Homebrew):**

```bash
brew services start postgresql@16
```

**Cek apakah user & database sudah dibuat:**

```bash
psql -U gotask -d gotask -c "\dt"
```

Jika belum, buat database:

```bash
psql -d postgres -c "CREATE USER gotask WITH PASSWORD 'gotask';"
psql -d postgres -c "CREATE DATABASE gotask OWNER gotask;"
```

---

### ❌ `go: could not create module cache: mkdir /Volumes/...`

**Penyebab:** `GOPATH` atau `GOMODCACHE` mengarah ke volume yang tidak bisa ditulis.

**Solusi — arahkan ke home directory:**

```bash
export GOPATH=$HOME/go
export GOMODCACHE=$HOME/go/pkg/mod
```

Tambahkan ke `~/.zshrc` agar permanen:

```bash
echo 'export GOPATH=$HOME/go' >> ~/.zshrc
echo 'export GOMODCACHE=$HOME/go/pkg/mod' >> ~/.zshrc
source ~/.zshrc
```

---

### ❌ Format tanggal tidak bisa diparse (`Format request tidak valid`)

**Penyebab:** Format tanggal di JSON body salah.

**Format yang benar:** `"YYYY-MM-DD"` (contoh: `"2026-07-30"`)

**Contoh benar:**

```json
{
  "title": "Task baru",
  "due_date": "2026-07-30"
}
```

**Contoh salah:**

```json
{
  "due_date": "2026-07-30T00:00:00Z"   ← JANGAN pakai format ini
}
```

---

## Testing

```bash
# Run all tests
make test

# Run with coverage
make test-cover
```

Tests cover:

- Authentication (register, login, refresh, logout)
- Task list CRUD and ownership
- Task CRUD, filtering, and status transitions
- Password validation
- Status workflow rules
- Progress business rules
- User ownership enforcement

---

## � Postman Collection

Tinggal import, langsung bisa tes semua 39 endpoint tanpa nulis kode.

### File

| File | Fungsi |
|------|--------|
| `docs/postman-collection.json` | Kumpulan 39 endpoint siap pakai |
| `docs/postman-environment.json` | Variable (base_url, token, dll.) |

### Cara Import

1. Buka Postman → **Import** → pilih kedua file di atas
2. Pilih Environment **"GoTask Local"** di dropdown kanan atas
3. Jalankan **Login** → token otomatis tersimpan ✨

### Struktur Folder

```
📁 GoTask API
├── 🏥 Health              (1 endpoint)
├── 🔐 Auth                (4 endpoint) — Register, Login, Refresh, Logout
├── 👤 User                (3 endpoint) — Profile, Update, Password
├── 📋 Task Lists          (8 endpoint) — CRUD + Archive + Board
├── ✅ Tasks               (9 endpoint) — CRUD + Status + Priority + Reopen
├── 📊 Progress            (6 endpoint) — CRUD + Rollback
├── 🔍 Review              (5 endpoint) — Submit, Approve, Request Changes
├── 💬 Comments            (4 endpoint) — CRUD
├── 📜 Activity            (2 endpoint) — Task & User
└── 📈 Dashboard           (6 endpoint) — Summary, Analytics, Deadlines
```

> 💡 **Body support `form-data` dan `raw JSON`** — backend otomatis detect formatnya.

---

## �📖 Kamus Istilah (Glossary)

Biar tidak bingung dengan istilah-istilah teknis, berikut daftar kata yang sering muncul:

| Istilah            | Arti Sederhana                                                             |
| ------------------ | -------------------------------------------------------------------------- |
| **API**            | Application Programming Interface — jembatan antara frontend dan backend   |
| **REST**           | Gaya arsitektur API yang pakai method HTTP (GET, POST, PUT, DELETE)        |
| **JSON**           | Format data ringan yang mudah dibaca manusia dan mesin                     |
| **JWT**            | JSON Web Token — "kartu identitas" digital untuk user yang sudah login     |
| **Access Token**   | Token jangka pendek (15 menit) untuk akses API                             |
| **Refresh Token**  | Token jangka panjang (30 hari) untuk minta access token baru               |
| **Middleware**     | "Pos pengecekan" yang dilewati setiap request sebelum sampai ke handler    |
| **Handler**        | Fungsi yang menangani HTTP request di endpoint tertentu                    |
| **Service**        | Lapisan yang berisi aturan bisnis (business logic)                         |
| **Repository**     | Lapisan yang berhubungan langsung dengan database                          |
| **DTO**            | Data Transfer Object — struct untuk data yang masuk/keluar API             |
| **Model**          | Struct yang merepresentasikan tabel di database                            |
| **Interface**      | "Kontrak" di Go — mendefinisikan method apa saja yang harus ada            |
| **Migration**      | Cara mengubah struktur database secara terstruktur dan bisa di-_rollback_  |
| **sqlc**           | Tool yang mengubah file SQL biasa menjadi kode Go yang siap pakai          |
| **UUID**           | Universally Unique Identifier — ID unik sepanjang masa, tidak akan bentrok |
| **Soft Delete**    | Menghapus data dengan menandai, bukan menghapus permanen                   |
| **Transaction**    | Operasi database yang "all or nothing" — semua sukses atau semua gagal     |
| **Hash**           | Mengubah data (password) menjadi string acak yang tidak bisa dikembalikan  |
| **bcrypt**         | Algoritma hashing khusus password yang sengaja dibuat lambat               |
| **CORS**           | Cross-Origin Resource Sharing — aturan domain mana yang boleh akses API    |
| **Rate Limiting**  | Membatasi jumlah request per detik untuk mencegah penyalahgunaan           |
| **Docker**         | Wadah untuk menjalankan aplikasi secara terisolasi dan konsisten           |
| **Docker Compose** | Cara menjalankan banyak container Docker sekaligus (API + PostgreSQL)      |

---

## 🤝 Berkontribusi (Contribution Guide)

### Untuk Pemula yang Ingin Belajar

1. **Clone repository ini**

   ```bash
   git clone https://github.com/IqromAbadi/GoTask.git
   cd GoTask
   ```

2. **Jalankan dengan Docker Compose** (cara termudah)

   ```bash
   cp .env.example .env
   docker compose up -d
   ```

3. **Coba-coba endpoint** pakai Postman atau curl

4. **Baca kode dengan urutan yang benar:**
   - Mulai dari `cmd/api/main.go` (pintu masuk)
   - Lalu `internal/auth/` (modul paling sederhana untuk dipahami)
   - Ikuti Rumus 6M: `model.go` → `dto.go` → `repository.go` → `repository_impl.go` → `service.go` → `handler.go`

### Ide untuk Latihan (Coba Implementasikan Sendiri!)

| Tingkat     | Ide Fitur                             | Modul yang Disentuh                |
| ----------- | ------------------------------------- | ---------------------------------- |
| 🟢 Mudah    | Tambah field `color` di task list     | `tasklist/`                        |
| 🟢 Mudah    | Ganti pesan error ke Bahasa Inggris   | `response.go` + semua `handler.go` |
| 🟡 Menengah | Tambah fitur tag/label di task        | `task/` + migration baru           |
| 🟡 Menengah | Assign task ke user lain (multi-user) | `task/` + `auth/`                  |
| 🔴 Sulit    | Notifikasi real-time via WebSocket    | Module baru `notification/`        |
| 🔴 Sulit    | Upload file attachment di task        | Module baru `attachment/`          |

### Aturan Kontribusi

1. Buat branch baru: `git checkout -b fitur-anda`
2. Ikuti struktur folder dan pola kode yang sudah ada
3. Tulis unit test untuk fitur baru
4. Jalankan `go fmt ./... && go vet ./... && go test ./...`
5. Commit dengan pesan yang jelas
6. Buat Pull Request

---

## 📊 Status Project

| Kategori               | Status              |
| ---------------------- | ------------------- |
| Authentication         | ✅ Selesai          |
| User Profile           | ✅ Selesai          |
| Task List CRUD         | ✅ Selesai          |
| Task CRUD              | ✅ Selesai          |
| Task Workflow          | ✅ Selesai          |
| Progress Tracking      | ✅ Selesai          |
| Review System          | ✅ Selesai          |
| Comments               | ✅ Selesai          |
| Activity Log           | ✅ Selesai          |
| Dashboard              | ✅ Selesai          |
| Unit Tests             | ✅ 40+ test case    |
| Docker Support         | ✅ Selesai          |
| API Documentation      | ✅ README + OpenAPI |
| Multi-user Support     | 🔜 Coming Soon      |
| File Attachments       | 🔜 Coming Soon      |
| Real-time Notification | 🔜 Coming Soon      |

---

## License

MIT
