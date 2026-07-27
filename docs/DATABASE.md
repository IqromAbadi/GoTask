# GoTask — Dokumentasi Database

Dokumentasi lengkap pembuatan database GoTask dari nol (0) hingga struktur saat ini. Ditulis dengan gaya yang mudah dipahami oleh pemula maupun developer baru.

---

## Daftar Isi

1. [Prasyarat](#1-prasyarat)
2. [Arsitektur Database](#2-arsitektur-database)
3. [Konsep Dasar](#3-konsep-dasar)
4. [Langkah-Langkah Pembuatan](#4-langkah-langkah-pembuatan)
   - [4.1 Install PostgreSQL](#41-install-postgresql)
   - [4.2 Buat User & Database](#42-buat-user--database)
   - [4.3 Aktifkan UUID Extension](#43-aktifkan-uuid-extension)
   - [4.4 Migrasi: Apa & Kenapa](#44-migrasi-apa--kenapa)
   - [4.5 Menjalankan Migrasi](#45-menjalankan-migrasi)
5. [Penjelasan Setiap Tabel](#5-penjelasan-setiap-tabel)
   - [5.1 users](#51-users)
   - [5.2 refresh_tokens](#52-refresh_tokens)
   - [5.3 task_lists](#53-task_lists)
   - [5.4 tasks](#54-tasks)
   - [5.5 task_progress_updates](#55-task_progress_updates)
   - [5.6 task_reviews](#56-task_reviews)
   - [5.7 task_comments](#57-task_comments)
   - [5.8 task_activities](#58-task_activities)
   - [5.9 schema_migrations](#59-schema_migrations)
6. [Urutan Pembuatan Tabel](#6-urutan-pembuatan-tabel)
7. [Relasi Antar Tabel](#7-relasi-antar-tabel)
8. [Entity Relationship Diagram (ERD)](#8-entity-relationship-diagram-erd)
9. [Flowchart Alur Pembuatan Database](#9-flowchart-alur-pembuatan-database)
10. [Contoh Data (Sample Data)](#10-contoh-data-sample-data)
11. [Index & Performance](#11-index--performance)
12. [Best Practice & Catatan Penting](#12-best-practice--catatan-penting)
13. [Cara Menjalankan Ulang dari Nol](#13-cara-menjalankan-ulang-dari-nol)
14. [Troubleshooting Database](#14-troubleshooting-database)

---

## 1. Prasyarat

Sebelum memulai, pastikan sudah terinstall:

| Tool | Versi | Cara Install (macOS) |
|------|-------|---------------------|
| PostgreSQL | 16+ | `brew install postgresql@16` |
| golang-migrate | 4+ | `brew install golang-migrate` |
| psql (PostgreSQL client) | 16+ | Otomatis terinstall bersama PostgreSQL |

---

## 2. Arsitektur Database

GoTask menggunakan **PostgreSQL 16** sebagai database. Arsitekturnya dirancang dengan prinsip:

```
┌─────────────────────────────────────────────────────┐
│                  ARSITEKTUR DATABASE                 │
│                                                      │
│  users (1) ──────< tasks (N)                        │
│    │                 │                               │
│    │                 ├── task_progress_updates (N)   │
│    │                 ├── task_reviews (N)            │
│    │                 ├── task_comments (N)           │
│    │                 └── task_activities (N)         │
│    │                                                 │
│    ├── task_lists (N) ──< tasks (N)                 │
│    ├── refresh_tokens (N)                            │
│    └── (semua aksi tercatat di activities)           │
└─────────────────────────────────────────────────────┘
```

**Pola relasi:**
- **1 user** punya **banyak task list**
- **1 task list** punya **banyak task**
- **1 task** punya **banyak** progress updates, reviews, comments, activities

---

## 3. Konsep Dasar

### UUID sebagai Primary Key

Semua tabel menggunakan **UUID** (Universally Unique Identifier) sebagai primary key, bukan auto-increment integer.

**Kenapa UUID?**
- Tidak bisa ditebak — user tidak bisa "menebak" ID milik user lain
- Unik secara global — tidak akan bentrok meskipun data digabung dari banyak server
- Aman untuk URL publik — `GET /tasks/550e8400-e29b-41d4-a716-446655440000` tidak membocorkan jumlah data

**Contoh UUID:** `550e8400-e29b-41d4-a716-446655440000`

### Timestamp UTC

Semua kolom waktu (`created_at`, `updated_at`, dll.) menggunakan **TIMESTAMP WITH TIME ZONE** dan disimpan dalam **UTC**. Ini menghindari masalah zona waktu saat aplikasi diakses dari berbagai negara.

### Soft Delete

Tabel `tasks` dan `task_comments` menggunakan **soft delete** — data tidak benar-benar dihapus, hanya diberi tanda `deleted_at`. Ini memungkinkan:
- Data bisa dikembalikan jika tidak sengaja terhapus
- Activity log tetap memiliki referensi ke data yang "dihapus"
- Memenuhi kebutuhan audit trail

### Check Constraints

Beberapa tabel memiliki **check constraint** untuk memastikan data selalu valid di level database:

| Tabel | Constraint | Aturan |
|-------|-----------|--------|
| tasks | `chk_tasks_status` | Status harus: backlog, todo, in_progress, review, done |
| tasks | `chk_tasks_priority` | Priority harus: low, medium, high, urgent |
| tasks | `chk_tasks_progress` | Progress harus 0-100 |
| task_progress_updates | `chk_progress_updates_progress` | Progress harus 0-100 |
| task_reviews | `chk_reviews_status` | Status harus: pending, approved, changes_requested |

---

## 4. Langkah-Langkah Pembuatan

### 4.1 Install PostgreSQL

```bash
# macOS
brew install postgresql@16

# Start service
brew services start postgresql@16

# Verifikasi
/opt/homebrew/opt/postgresql@16/bin/psql --version
# Output: psql (PostgreSQL) 16.x
```

### 4.2 Buat User & Database

```sql
-- Masuk sebagai superuser
psql -d postgres

-- Buat user aplikasi
CREATE USER gotask WITH PASSWORD 'gotask';

-- Buat database
CREATE DATABASE gotask OWNER gotask;

-- Beri hak akses penuh
GRANT ALL PRIVILEGES ON DATABASE gotask TO gotask;

-- Keluar
\q
```

### 4.3 Aktifkan UUID Extension

PostgreSQL perlu extension `uuid-ossp` untuk generate UUID otomatis:

```sql
-- Extension ini diaktifkan di awal migration
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
```

Setelah aktif, kita bisa pakai `uuid_generate_v4()` untuk generate UUID random.

### 4.4 Migrasi: Apa & Kenapa

**Migrasi** adalah cara untuk mengelola perubahan struktur database secara terstruktur dan bisa di-rollback.

**Kenapa pakai migrasi?**
- **Version control** — setiap perubahan database tercatat, seperti git untuk kode
- **Rollback** — kalau ada masalah, bisa kembali ke versi sebelumnya
- **Kolaborasi** — tim bisa menjalankan migrasi yang sama di laptop masing-masing
- **Deployment** — production database bisa di-upgrade dengan aman

**File migrasi GoTask:**
```
db/migrations/
└── 000001_init_schema.up.sql    # Membuat semua tabel
└── 000001_init_schema.down.sql  # Menghapus semua tabel (rollback)
```

Format penamaan: `{nomor}_{deskripsi}.{arah}.sql`
- `000001` = nomor urut
- `init_schema` = deskripsi
- `up` = migrasi naik (create)
- `down` = migrasi turun (drop)

### 4.5 Menjalankan Migrasi

```bash
# Setup database URL
export DATABASE_URL="postgres://gotask:gotask@localhost:5432/gotask?sslmode=disable"

# Jalankan migrasi (UP)
migrate -path db/migrations -database "$DATABASE_URL" up

# Output:
# 1/u init_schema (452ms)

# Cek status
psql "$DATABASE_URL" -c "SELECT * FROM schema_migrations;"
# Output: version=1, dirty=false

# Untuk rollback (DOWN)
migrate -path db/migrations -database "$DATABASE_URL" down
```

---

## 5. Penjelasan Setiap Tabel

### 5.1 users

**Tujuan:** Menyimpan data pengguna yang terdaftar di aplikasi.

```sql
CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name          VARCHAR(255) NOT NULL,
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    avatar_url    VARCHAR(500) NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

| Kolom | Tipe | Keterangan |
|-------|------|------------|
| `id` | UUID | Primary key, auto-generate |
| `name` | VARCHAR(255) | Nama lengkap pengguna |
| `email` | VARCHAR(255) | Email (UNIQUE, lowercase) |
| `password_hash` | VARCHAR(255) | Hash password (bcrypt) — **bukan** password asli! |
| `avatar_url` | VARCHAR(500) | URL foto profil (nullable) |
| `created_at` | TIMESTAMPTZ | Waktu pendaftaran (UTC) |
| `updated_at` | TIMESTAMPTZ | Waktu terakhir diupdate (UTC) |

**Index:**
- `idx_users_email` — mempercepat pencarian berdasarkan email (login)

**Catatan penting:**
- `password_hash` menyimpan hasil bcrypt, **bukan password mentah**
- Email selalu disimpan dalam **lowercase** (dinormalisasi di backend)
- ON DELETE CASCADE: saat user dihapus, semua data terkait ikut terhapus

**Relasi:**
- 1 user → banyak task_lists
- 1 user → banyak refresh_tokens
- 1 user → banyak tasks (sebagai created_by)
- 1 user → banyak progress_updates, reviews, comments, activities

---

### 5.2 refresh_tokens

**Tujuan:** Menyimpan token untuk fitur "remember me" dan auto-refresh JWT.

```sql
CREATE TABLE refresh_tokens (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  VARCHAR(255) NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked_at  TIMESTAMPTZ NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

| Kolom | Tipe | Keterangan |
|-------|------|------------|
| `id` | UUID | Primary key |
| `user_id` | UUID | Foreign key ke users |
| `token_hash` | VARCHAR(255) | **Hash** dari refresh token — bukan token asli |
| `expires_at` | TIMESTAMPTZ | Waktu kadaluarsa (30 hari) |
| `revoked_at` | TIMESTAMPTZ | Waktu pencabutan (NULL = masih aktif) |
| `created_at` | TIMESTAMPTZ | Waktu pembuatan |

**Kenapa simpan hash, bukan token asli?**
- Kalau database bocor, penyerang tidak bisa pakai token hasil curian
- Token asli hanya dikirim ke user sekali (saat login/refresh)

**Index:**
- `idx_refresh_tokens_user_id` — cepat mencari token milik user tertentu
- `idx_refresh_tokens_token_hash` — cepat mencocokkan token saat refresh

---

### 5.3 task_lists

**Tujuan:** Menyimpan daftar tugas (board) milik user.

```sql
CREATE TABLE task_lists (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    description TEXT NULL,
    is_archived BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

| Kolom | Tipe | Keterangan |
|-------|------|------------|
| `id` | UUID | Primary key |
| `user_id` | UUID | Foreign key ke users (pemilik) |
| `name` | VARCHAR(255) | Nama task list, misal "Sprint 1" |
| `description` | TEXT | Deskripsi (nullable) |
| `is_archived` | BOOLEAN | Status arsip (default: FALSE) |
| `created_at` | TIMESTAMPTZ | Waktu pembuatan |
| `updated_at` | TIMESTAMPTZ | Waktu update terakhir |

**Catatan:**
- `is_archived` digunakan untuk fitur arsip, bukan delete
- User hanya bisa lihat task list miliknya sendiri (authorization di query)

---

### 5.4 tasks

**Tujuan:** Tabel inti — menyimpan setiap task/pekerjaan.

```sql
CREATE TABLE tasks (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    list_id           UUID NOT NULL REFERENCES task_lists(id) ON DELETE CASCADE,
    created_by        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title             VARCHAR(500) NOT NULL,
    description       TEXT NULL,
    status            VARCHAR(50) NOT NULL DEFAULT 'backlog',
    priority          VARCHAR(50) NOT NULL DEFAULT 'medium',
    progress          INTEGER NOT NULL DEFAULT 0,
    due_date          DATE NULL,
    estimated_minutes INTEGER NULL,
    started_at        TIMESTAMPTZ NULL,
    completed_at      TIMESTAMPTZ NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ NULL,

    CONSTRAINT chk_tasks_status   CHECK (status IN ('backlog','todo','in_progress','review','done')),
    CONSTRAINT chk_tasks_priority CHECK (priority IN ('low','medium','high','urgent')),
    CONSTRAINT chk_tasks_progress CHECK (progress >= 0 AND progress <= 100)
);
```

| Kolom | Tipe | Keterangan |
|-------|------|------------|
| `id` | UUID | Primary key |
| `list_id` | UUID | Foreign key ke task_lists |
| `created_by` | UUID | Foreign key ke users (pembuat) |
| `title` | VARCHAR(500) | Judul task |
| `description` | TEXT | Deskripsi detail (nullable) |
| `status` | VARCHAR(50) | Status workflow (default: backlog) |
| `priority` | VARCHAR(50) | Prioritas (default: medium) |
| `progress` | INTEGER | Persentase 0-100 (default: 0) |
| `due_date` | DATE | Tenggat waktu (nullable) |
| `estimated_minutes` | INTEGER | Estimasi pengerjaan dalam menit |
| `started_at` | TIMESTAMPTZ | Kapan mulai dikerjakan (auto-set saat in_progress) |
| `completed_at` | TIMESTAMPTZ | Kapan selesai (auto-set saat review approved) |
| `created_at` | TIMESTAMPTZ | Waktu pembuatan |
| `updated_at` | TIMESTAMPTZ | Waktu update terakhir |
| `deleted_at` | TIMESTAMPTZ | Soft delete marker |

**Status workflow:**
```
backlog → todo → in_progress → review → done
```

**Prioritas:**
```
low < medium < high < urgent
```

**Check constraints:**
- Status hanya boleh 5 nilai di atas
- Priority hanya boleh 4 nilai di atas
- Progress harus antara 0 dan 100

**Index:**
- `idx_tasks_list_id` — mempercepat Board View (group by list)
- `idx_tasks_status` — mempercepat filter & dashboard
- `idx_tasks_priority` — mempercepat filter prioritas
- `idx_tasks_due_date` — mempercepat deadline & overdue
- `idx_tasks_deleted_at` — mempercepat exclude soft-deleted tasks

---

### 5.5 task_progress_updates

**Tujuan:** Mencatat riwayat perubahan progress setiap task.

```sql
CREATE TABLE task_progress_updates (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id    UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    progress   INTEGER NOT NULL,
    note       TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_progress_updates_progress CHECK (progress >= 0 AND progress <= 100)
);
```

| Kolom | Tipe | Keterangan |
|-------|------|------------|
| `id` | UUID | Primary key |
| `task_id` | UUID | Foreign key ke tasks |
| `user_id` | UUID | Foreign key ke users (siapa yang update) |
| `progress` | INTEGER | Nilai progress (0-100) |
| `note` | TEXT | Catatan progress (nullable) |
| `created_at` | TIMESTAMPTZ | Waktu update |
| `updated_at` | TIMESTAMPTZ | Waktu edit terakhir |

**Aturan bisnis:**
- Progress hanya bisa 0-100
- Progress baru tidak boleh < progress lama (kecuali allow_rollback)
- Hanya bisa update saat task status = in_progress

---

### 5.6 task_reviews

**Tujuan:** Mencatat proses review task sebelum dinyatakan selesai.

```sql
CREATE TABLE task_reviews (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id         UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    reviewer_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status          VARCHAR(50) NOT NULL DEFAULT 'pending',
    submission_note TEXT NULL,
    review_note     TEXT NULL,
    reviewed_at     TIMESTAMPTZ NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_reviews_status CHECK (status IN ('pending','approved','changes_requested'))
);
```

| Kolom | Tipe | Keterangan |
|-------|------|------------|
| `id` | UUID | Primary key |
| `task_id` | UUID | Foreign key ke tasks |
| `reviewer_id` | UUID | Foreign key ke users (siapa yang review) |
| `status` | VARCHAR(50) | Status review (pending/approved/changes_requested) |
| `submission_note` | TEXT | Catatan saat submit (nullable) |
| `review_note` | TEXT | Catatan hasil review (nullable) |
| `reviewed_at` | TIMESTAMPTZ | Kapan direview (NULL sebelum direview) |
| `created_at` | TIMESTAMPTZ | Waktu submit |
| `updated_at` | TIMESTAMPTZ | Waktu update |

**Flow review:**
```
pending → approved        (task jadi done, completed_at terisi)
pending → changes_requested (task kembali ke in_progress)
```

---

### 5.7 task_comments

**Tujuan:** Menyimpan komentar/diskusi pada task.

```sql
CREATE TABLE task_comments (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id    UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content    TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);
```

| Kolom | Tipe | Keterangan |
|-------|------|------------|
| `id` | UUID | Primary key |
| `task_id` | UUID | Foreign key ke tasks |
| `user_id` | UUID | Foreign key ke users (pemberi komentar) |
| `content` | TEXT | Isi komentar |
| `created_at` | TIMESTAMPTZ | Waktu pembuatan |
| `updated_at` | TIMESTAMPTZ | Waktu edit terakhir |
| `deleted_at` | TIMESTAMPTZ | Soft delete marker |

**Catatan:**
- Soft delete — komentar yang "dihapus" masih ada di database
- User hanya bisa edit/hapus komentar miliknya sendiri
- Query selalu filter `WHERE deleted_at IS NULL`

---

### 5.8 task_activities

**Tujuan:** Log otomatis semua aktivitas yang terjadi pada task.

```sql
CREATE TABLE task_activities (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id    UUID NULL REFERENCES tasks(id) ON DELETE SET NULL,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action     VARCHAR(100) NOT NULL,
    old_value  TEXT NULL,
    new_value  TEXT NULL,
    metadata   JSONB NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

| Kolom | Tipe | Keterangan |
|-------|------|------------|
| `id` | UUID | Primary key |
| `task_id` | UUID | Foreign key ke tasks (NULL jika task dihapus) |
| `user_id` | UUID | Foreign key ke users (siapa yang melakukan) |
| `action` | VARCHAR(100) | Jenis aksi (task_created, status_changed, dll.) |
| `old_value` | TEXT | Nilai sebelum perubahan |
| `new_value` | TEXT | Nilai setelah perubahan |
| `metadata` | JSONB | Data tambahan (misal: task_title) |
| `created_at` | TIMESTAMPTZ | Waktu kejadian |

**Jenis aksi yang dicatat:**
```
task_created, task_updated, task_deleted,
task_status_changed, task_priority_changed,
task_progress_updated, task_progress_rolled_back,
task_submitted_for_review, review_approved,
review_changes_requested, comment_created,
comment_updated, comment_deleted, task_reopened
```

**Catatan:**
- Activity dibuat **otomatis oleh backend** — tidak ada endpoint create/update/delete
- `ON DELETE SET NULL` — jika task dihapus, activity tetap ada (task_id jadi NULL)
- `metadata` pakai JSONB — fleksibel untuk data tambahan

---

### 5.9 schema_migrations

**Tujuan:** Tabel otomatis dari `golang-migrate` untuk melacak versi migrasi.

| Kolom | Tipe | Keterangan |
|-------|------|------------|
| `version` | BIGINT | Nomor versi migrasi |
| `dirty` | BOOLEAN | FALSE = migrasi sukses, TRUE = migrasi gagal di tengah jalan |

Tabel ini **otomatis dibuat dan dikelola** oleh golang-migrate. Tidak perlu diurus manual.

---

## 6. Urutan Pembuatan Tabel

Urutan pembuatan tabel **tidak boleh sembarangan** karena ada ketergantungan Foreign Key:

```
1. users              (tidak ada FK — bisa dibuat pertama)
2. refresh_tokens     (FK ke users)
3. task_lists         (FK ke users)
4. tasks              (FK ke task_lists + users)
5. task_progress_updates (FK ke tasks + users)
6. task_reviews       (FK ke tasks + users)
7. task_comments      (FK ke tasks + users)
8. task_activities    (FK ke tasks + users)
```

**Aturan:** Tabel yang direferensi (parent) HARUS dibuat lebih dulu dari tabel yang mereferensi (child).

Contoh: `tasks` mereferensi `task_lists` → `task_lists` harus dibuat dulu.

---

## 7. Relasi Antar Tabel

```
users (1) ──────< task_lists (N)
  │                  │
  │                  └──< tasks (N)
  │                        │
  │                        ├──< task_progress_updates (N)
  │                        ├──< task_reviews (N)
  │                        ├──< task_comments (N)
  │                        └──< task_activities (N)
  │
  ├──< refresh_tokens (N)
  ├──< tasks.created_by (N)
  └──< (semua tabel dengan user_id)
```

**Jenis relasi:**
- **One-to-Many:** users → task_lists, task_lists → tasks
- **One-to-Many:** tasks → progress_updates, reviews, comments, activities
- **Many-to-One:** task_activities → tasks (ON DELETE SET NULL)

---

## 8. Entity Relationship Diagram (ERD)

```mermaid
erDiagram
    users ||--o{ refresh_tokens : has
    users ||--o{ task_lists : owns
    users ||--o{ tasks : creates
    task_lists ||--o{ tasks : contains
    tasks ||--o{ task_progress_updates : has
    tasks ||--o{ task_reviews : reviewed_in
    tasks ||--o{ task_comments : discussed_in
    tasks ||--o{ task_activities : logged_in
    users ||--o{ task_progress_updates : records
    users ||--o{ task_reviews : performs
    users ||--o{ task_comments : writes
    users ||--o{ task_activities : triggers

    users {
        uuid id PK "uuid_generate_v4()"
        varchar name "NOT NULL"
        varchar email "UNIQUE NOT NULL"
        varchar password_hash "NOT NULL (bcrypt)"
        varchar avatar_url "NULL"
        timestamptz created_at "DEFAULT NOW()"
        timestamptz updated_at "DEFAULT NOW()"
    }

    refresh_tokens {
        uuid id PK
        uuid user_id FK "REFERENCES users"
        varchar token_hash "NOT NULL (SHA-256)"
        timestamptz expires_at "NOT NULL"
        timestamptz revoked_at "NULL = masih aktif"
        timestamptz created_at
    }

    task_lists {
        uuid id PK
        uuid user_id FK "REFERENCES users"
        varchar name "NOT NULL"
        text description "NULL"
        boolean is_archived "DEFAULT FALSE"
        timestamptz created_at
        timestamptz updated_at
    }

    tasks {
        uuid id PK
        uuid list_id FK "REFERENCES task_lists"
        uuid created_by FK "REFERENCES users"
        varchar title "NOT NULL"
        text description "NULL"
        varchar status "CHECK (5 nilai)"
        varchar priority "CHECK (4 nilai)"
        int progress "CHECK (0-100)"
        date due_date "NULL"
        int estimated_minutes "NULL"
        timestamptz started_at "auto-set"
        timestamptz completed_at "auto-set"
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at "soft delete"
    }

    task_progress_updates {
        uuid id PK
        uuid task_id FK "REFERENCES tasks"
        uuid user_id FK "REFERENCES users"
        int progress "CHECK (0-100)"
        text note "NULL"
        timestamptz created_at
        timestamptz updated_at
    }

    task_reviews {
        uuid id PK
        uuid task_id FK "REFERENCES tasks"
        uuid reviewer_id FK "REFERENCES users"
        varchar status "CHECK (pending/approved/changes_requested)"
        text submission_note "NULL"
        text review_note "NULL"
        timestamptz reviewed_at "NULL"
        timestamptz created_at
        timestamptz updated_at
    }

    task_comments {
        uuid id PK
        uuid task_id FK "REFERENCES tasks"
        uuid user_id FK "REFERENCES users"
        text content "NOT NULL"
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at "soft delete"
    }

    task_activities {
        uuid id PK
        uuid task_id FK "REFERENCES tasks ON DELETE SET NULL"
        uuid user_id FK "REFERENCES users"
        varchar action "NOT NULL"
        text old_value "NULL"
        text new_value "NULL"
        jsonb metadata "NULL"
        timestamptz created_at
    }
```

---

## 9. Flowchart Alur Pembuatan Database

```mermaid
flowchart TD
    A[Mulai: Database Kosong] --> B[Install PostgreSQL 16]
    B --> C[Start PostgreSQL Service]
    C --> D[Buat User gotask]
    D --> E[Buat Database gotask]
    E --> F[Grant Privileges]
    F --> G[Aktifkan Extension uuid-ossp]
    G --> H{Jalankan Migrasi?}
    
    H -->|Ya| I[Migrate UP]
    H -->|Tidak| J[Jalankan SQL Manual]
    
    I --> K[Check schema_migrations]
    K --> L{version=1, dirty=false?}
    L -->|Ya| M[SEMUA TABEL SIAP]
    L -->|Tidak| N[Force reset & ulangi]
    
    J --> M

    M --> O[Tabel 1: users]
    O --> P[Tabel 2: refresh_tokens]
    P --> Q[Tabel 3: task_lists]
    Q --> R[Tabel 4: tasks]
    R --> S[Tabel 5: task_progress_updates]
    S --> T[Tabel 6: task_reviews]
    T --> U[Tabel 7: task_comments]
    U --> V[Tabel 8: task_activities]
    
    V --> W[Indexes terpasang]
    W --> X[Constraints aktif]
    X --> Y[Database Siap Digunakan]
```

---

## 10. Contoh Data (Sample Data)

### 10.1 User

```sql
INSERT INTO users (id, name, email, password_hash) VALUES
('a1000000-0000-0000-0000-000000000001', 'Agung', 'agung@example.com', '$2a$10$...hash...');
```

### 10.2 Task List

```sql
INSERT INTO task_lists (id, user_id, name, description) VALUES
('b1000000-0000-0000-0000-000000000001', 'a1000000-0000-0000-0000-000000000001', 'Sprint 1', 'Sprint pertama — modul authentication');
```

### 10.3 Tasks

```sql
-- Task 1: Backlog
INSERT INTO tasks (id, list_id, created_by, title, status, priority, due_date) VALUES
('c1000000-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', 'a1000000-0000-0000-0000-000000000001', 'Desain halaman login', 'backlog', 'medium', '2026-08-01');

-- Task 2: Todo
INSERT INTO tasks (id, list_id, created_by, title, status, priority, estimated_minutes) VALUES
('c1000000-0000-0000-0000-000000000002', 'b1000000-0000-0000-0000-000000000001', 'a1000000-0000-0000-0000-000000000001', 'Implementasi JWT', 'todo', 'high', 240);

-- Task 3: In Progress
INSERT INTO tasks (id, list_id, created_by, title, status, priority, progress, started_at) VALUES
('c1000000-0000-0000-0000-000000000003', 'b1000000-0000-0000-0000-000000000001', 'a1000000-0000-0000-0000-000000000001', 'Buat endpoint register', 'in_progress', 'urgent', 60, '2026-07-22T09:00:00Z');

-- Task 4: Done
INSERT INTO tasks (id, list_id, created_by, title, status, priority, progress, started_at, completed_at) VALUES
('c1000000-0000-0000-0000-000000000004', 'b1000000-0000-0000-0000-000000000001', 'a1000000-0000-0000-0000-000000000001', 'Setup project structure', 'done', 'high', 100, '2026-07-20T08:00:00Z', '2026-07-22T17:00:00Z');
```

### 10.4 Progress Updates

```sql
INSERT INTO task_progress_updates (task_id, user_id, progress, note) VALUES
('c1000000-0000-0000-0000-000000000003', 'a1000000-0000-0000-0000-000000000001', 30, 'Sudah buat model & DTO'),
('c1000000-0000-0000-0000-000000000003', 'a1000000-0000-0000-0000-000000000001', 60, 'Handler & service selesai');
```

### 10.5 Review

```sql
INSERT INTO task_reviews (task_id, reviewer_id, status, submission_note) VALUES
('c1000000-0000-0000-0000-000000000003', 'a1000000-0000-0000-0000-000000000001', 'pending', 'Semua endpoint auth selesai, tolong di-review');
```

### 10.6 Comment

```sql
INSERT INTO task_comments (task_id, user_id, content) VALUES
('c1000000-0000-0000-0000-000000000003', 'a1000000-0000-0000-0000-000000000001', 'Jangan lupa tambah validasi password strength');
```

### 10.7 Activity

```sql
INSERT INTO task_activities (task_id, user_id, action, old_value, new_value) VALUES
('c1000000-0000-0000-0000-000000000003', 'a1000000-0000-0000-0000-000000000001', 'task_status_changed', 'todo', 'in_progress'),
('c1000000-0000-0000-0000-000000000003', 'a1000000-0000-0000-0000-000000000001', 'task_progress_updated', '0%', '60%');
```

---

## 11. Index & Performance

### Daftar Index

| Index | Tabel | Kolom | Tujuan |
|-------|-------|-------|--------|
| `idx_users_email` | users | email | Cepat mencari user saat login |
| `idx_refresh_tokens_user_id` | refresh_tokens | user_id | Cepat revoke semua token user |
| `idx_refresh_tokens_token_hash` | refresh_tokens | token_hash | Cepat validasi token saat refresh |
| `idx_task_lists_user_id` | task_lists | user_id | Cepat list task list milik user |
| `idx_tasks_list_id` | tasks | list_id | Cepat Board View (group by list) |
| `idx_tasks_status` | tasks | status | Cepat filter & dashboard aggregation |
| `idx_tasks_priority` | tasks | priority | Cepat filter prioritas |
| `idx_tasks_due_date` | tasks | due_date | Cepat deadline & overdue queries |
| `idx_tasks_deleted_at` | tasks | deleted_at | Cepat exclude soft-deleted tasks |
| `idx_task_progress_updates_task_id` | progress_updates | task_id | Cepat list progress per task |
| `idx_task_reviews_task_id` | reviews | task_id | Cepat list review per task |
| `idx_task_comments_task_id` | comments | task_id | Cepat list komentar per task |
| `idx_task_activities_task_id` | activities | task_id | Cepat list aktivitas per task |
| `idx_task_activities_user_id` | activities | user_id | Cepat list semua aktivitas user |

**Kenapa index penting?**
- Tanpa index, PostgreSQL scan seluruh tabel (lambat untuk data besar)
- Dengan index, pencarian seperti mencari di buku telepon — langsung ke halaman yang tepat

---

## 12. Best Practice & Catatan Penting

### Security

| Praktik | Penerapan |
|---------|-----------|
| Password tidak disimpan mentah | `password_hash` pakai bcrypt (hash satu arah) |
| Refresh token tidak plain text | `token_hash` pakai SHA-256 |
| ID tidak bisa ditebak | UUID, bukan auto-increment integer |
| SQL injection dicegah | Semua query pakai parameterized query (`$1`, `$2`) |

### Data Integrity

| Praktik | Penerapan |
|---------|-----------|
| Status terbatas | CHECK constraint di tasks.status, tasks.priority |
| Progress valid | CHECK constraint (0-100) |
| FK dengan CASCADE | Hapus user → hapus semua data terkait |
| Soft delete | tasks & comments tidak benar-benar dihapus |
| Timestamp UTC | Semua TIMESTAMPTZ, aplikasi menggunakan UTC |

### Performance

| Praktik | Penerapan |
|---------|-----------|
| Index di kolom pencarian | email, status, priority, due_date, list_id |
| Connection pool | Max 25 open connections, 10 idle |
| Query JOIN untuk ownership | Setiap query task JOIN task_lists untuk cek user_id |

### Migration

| Praktik | Penerapan |
|---------|-----------|
| Version control | Setiap perubahan database ada di file migrasi |
| Rollback siap | Setiap `up.sql` punya `down.sql` |
| Idempotent | `CREATE EXTENSION IF NOT EXISTS` — aman dijalankan berkali-kali |

---

## 13. Cara Menjalankan Ulang dari Nol

### Reset total (HATI-HATI: hapus semua data!)

```bash
# 1. Rollback semua migrasi
migrate -path db/migrations -database "$DATABASE_URL" down

# 2. Jalankan ulang
migrate -path db/migrations -database "$DATABASE_URL" up

# 3. Atau hapus & buat ulang database
psql -d postgres -c "DROP DATABASE gotask;"
psql -d postgres -c "CREATE DATABASE gotask OWNER gotask;"
migrate -path db/migrations -database "$DATABASE_URL" up
```

### Setup dari awal (database kosong)

```bash
# 1. Buat user & database
psql -d postgres <<SQL
CREATE USER gotask WITH PASSWORD 'gotask';
CREATE DATABASE gotask OWNER gotask;
GRANT ALL PRIVILEGES ON DATABASE gotask TO gotask;
SQL

# 2. Jalankan migrasi
export DATABASE_URL="postgres://gotask:gotask@localhost:5432/gotask?sslmode=disable"
migrate -path db/migrations -database "$DATABASE_URL" up

# 3. Cek hasil
psql "$DATABASE_URL" -c "\dt"
# Harusnya muncul 9 tabel
```

---

## 14. Troubleshooting Database

### "relation does not exist"

**Penyebab:** Migrasi belum dijalankan.

**Solusi:**
```bash
migrate -path db/migrations -database "$DATABASE_URL" up
```

### "dirty=true" di schema_migrations

**Penyebab:** Migrasi sebelumnya gagal di tengah jalan.

**Solusi:**
```bash
# Reset versi paksa
migrate -path db/migrations -database "$DATABASE_URL" force 1
```

### "extension uuid-ossp not available"

**Penyebab:** Extension belum diinstall.

**Solusi:**
```sql
psql "$DATABASE_URL" -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";"
```

### UUID tidak auto-generate

**Penyebab:** Extension uuid-ossp tidak aktif.

**Solusi:** Jalankan perintah CREATE EXTENSION di atas.

---

> **Dokumentasi ini dibuat pada 2026-07-24. Versi database: 1 (init_schema).**
