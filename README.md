# İmtahan nəticələri API

Yalnız Go Fiber backend: PostgreSQL-dən FIN kodu ilə nəticə oxuyur, Redis ilə cache edir və Swagger UI təqdim edir. Frontend yoxdur.

## API və Swagger

- Swagger UI: `https://api.dim.davidjs.dev/docs/index.html`
- OpenAPI JSON: `https://api.dim.davidjs.dev/openapi.json`
- Canlılıq: `GET /healthz`
- Hazırlıq: `GET /readyz`
- Nəticə axtarışı: `POST /api/v1/results/lookup`

~~~json
{ "finCode": "T000000" }
~~~

## Lokal seed və işə salma

`apps/api/.env` yaradın və yalnız lokal/external PostgreSQL URL-ni ora yazın. Bu fayl git-ə daxil edilmir.

~~~sh
cp apps/api/.env.example apps/api/.env
cd apps/api

# Cədvəl ilk dəfə yaradılırsa, bir dəfə tətbiq edin.
psql "$DATABASE_URL" -X -v ON_ERROR_STOP=1 -f migrations/0001_exam_results.sql

# 10 milyon sintetik nəticəni əlavə edin.
go run ./cmd/seed -total=10000000 -batch-size=10000

# API-ni başladın.
go run ./cmd/api
~~~

Seed heç vaxt API startında işləmir. Eyni əmri yenidən başladanda mövcud sintetik sıra yoxlanılır və qaldığı yerdən davam edir. `--truncate` bütün `exam_results` cədvəlini silir; adi seed üçün istifadə etməyin.

İlk test FIN-i `T000000`-dır. 10 milyon sıra üçün PostgreSQL-də bir neçə GB boş disk sahəsi olmalıdır.

## Coolify

Coolify-də yalnız `DATABASE_URL`-ni **Runtime Variable** kimi əlavə edin; build secret kimi verməyin. Compose artıq `web` və `migrate` service-lərini işə salmır, buna görə `SERVICE_FQDN_WEB` və `SERVICE_FQDN_API` dəyişənlərinə ehtiyac qalmır. Xarici PostgreSQL üçün Compose `API_DB_TIMEOUT=1s` və `API_REQUEST_TIMEOUT=3s` verir; köhnə `180ms` dəyəri lookup-ları vaxtından əvvəl dayandırır. API domenini Coolify-dən `api.dim.davidjs.dev` olaraq təyin etdikdən sonra Swagger yuxarıdakı ünvanla açılacaq.

## Yoxlama

~~~sh
bun run build 2>&1 | tail -n 50
bun run test
~~~
