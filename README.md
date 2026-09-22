# شبایار | Card Number to IBAN

سرویس full-stack تبدیل شماره کارت به شبا با Go، PostgreSQL و Nuxt/Vue.

## ساختار

- `backend/`: API، منطق دامنه، repository، کلاینت ZarinHub و worker queue
- `frontend/`: رابط RTL واکنش‌گرا با Nuxt، Vue و Tailwind
- `backend/migrations/`: migrationهای قابل تکرار دیتابیس

## اجرای backend

پیش‌نیازها: Go 1.27+، PostgreSQL و ابزار `migrate`.

```bash
cd backend
cp .env.example .env
# مقدارهای DATABASE_*، ZARINHUB_* و CARD_FINGERPRINT_KEY را در .env تنظیم کن
migrate -path migrations -database "$DATABASE_URL" up
go run ./cmd/api
```

اگر از تنظیمات فعلی پروژه استفاده می‌کنی، migration سوم جدول DLQ را نیز اعمال
کن:

```bash
migrate -path migrations -database "postgres://USER:PASSWORD@localhost:5432/zarrinpal?sslmode=disable" up
```

سرور روی `http://localhost:8080` اجرا می‌شود. health check:

```bash
curl http://localhost:8080/healthz
```

## API

```bash
curl --request POST http://localhost:8080/api/v1/card-to-iban \
  --header 'Content-Type: application/json' \
  --data '{"card":"6037997512345670"}'
```

پاسخ موفق در `data` و خطا در `error` برمی‌گردد. شماره کارت قبل از ثبت
درخواست normalize و validate می‌شود و در audit فقط به صورت mask و fingerprint
ذخیره می‌شود.

## Queue، rate limit و DLQ

فراخوانی ZarinHub از یک صف bounded داخل process عبور می‌کند. تعداد workerها،
ظرفیت صف، نرخ درخواست و retry از `.env` قابل تنظیم هستند:

```dotenv
CARD_QUEUE_CAPACITY=100
CARD_QUEUE_WORKERS=4
ZARINHUB_RATE_LIMIT_RPS=5
ZARINHUB_RATE_LIMIT_BURST=5
ZARINHUB_MAX_ATTEMPTS=3
ZARINHUB_RETRY_BASE_DELAY=250ms
```

فقط خطاهای timeout، unavailable و transport retry می‌شوند. هر retry یک attempt
و `trackId` جدید دارد. خطاهای authentication یا retryهای تمام‌شده در جدول
`card_to_iban_dead_letters` ثبت می‌شوند. DLQ عمداً شماره کارت کامل را نگه
نمی‌دارد؛ replay باید با ارسال دوبارهٔ درخواست از سمت کاربر انجام شود.

## اجرای frontend

پیش‌نیاز: Node.js 20+ و pnpm.

```bash
cd frontend
cp .env.example .env
pnpm install
pnpm dev
```

Nuxt در حالت development مسیر `/api/*` را به `http://localhost:8080` proxy
می‌کند. آدرس proxy با `NUXT_API_PROXY_TARGET` قابل تغییر است.

## تست و بررسی

```bash
cd backend
GOCACHE=/tmp/card-to-iban-go-build go test ./...
GOCACHE=/tmp/card-to-iban-go-build go test -race ./...
GOCACHE=/tmp/card-to-iban-go-build go vet ./...

cd ../frontend
pnpm typecheck
pnpm build
```

credentialهای واقعی فقط در `.env` محلی قرار می‌گیرند و نباید commit شوند.
