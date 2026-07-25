# İmtahan nəticələrinin yoxlanılması

Bun ilə işləyən SSR Next.js interfeysi, Go Fiber hot-path API-si, xarici PostgreSQL mənbə bazası və Redis cache-aside qatı.

## Lokal işə salma

~~~sh
cp .env.example .env
# .env-də DATABASE_URL-ni əlçatan PostgreSQL instansına uyğunlaşdırın.
docker compose --env-file .env -f docker-compose.yml -f docker-compose.local.yml up --build
~~~

Sistem ilk işə düşəndə sxem avtomatik tətbiq edilir. Sintetik məlumat ayrıca yüklənir; API startında heç vaxt seed edilmir:

> Diqqət: seed əmri `DATABASE_URL`-də göstərilən bazaya yazır. İstehsal bazasında yalnız sintetik məlumatı qəsdən yükləmək istədikdə işlədin.

~~~sh
# sürətli yoxlama üçün
docker compose --env-file .env -f docker-compose.yml -f docker-compose.local.yml run --rm --entrypoint /app/seed api -total 10000

# tam test həcmi üçün
docker compose --env-file .env -f docker-compose.yml -f docker-compose.local.yml run --rm --entrypoint /app/seed api -total 10000000 -batch-size 10000
~~~

İlk sintetik FIN `T000000` olur. Web: http://localhost:3000, API liveness: http://localhost:8080/healthz.

## Coolify və xarici PostgreSQL

`DATABASE_URL` Coolify-də secret kimi təyin edilməlidir; repository-yə yazılmır. Dəyişəni yalnız **Runtime Variable** edin; Build Variable və **Use Docker Build Secrets** aktiv olmasın. Coolify runtime `.env`-i Compose interpolationu ilə həm `migrate`, həm API konteynerinə ötürülür.

Ayrı Coolify PostgreSQL resursu üçün hazırkı ən etibarlı seçim həmin resursun **External connection URL**-idir: onu `DATABASE_URL` dəyəri kimi istifadə edin. Raw resource UUID hostunu (məsələn, `u...`) təkbaşına yazmayın; o yalnız düzgün shared Docker şəbəkəsində resolve olunur. Xarici URL istifadə edilirsə **Connect to Predefined Network** lazım deyil. Daxili URL seçilirsə, bu seçim aktiv olmalı və host `postgres-<database-resource-uuid>` formasında yazılmalıdır.

`migrate` servisi URL əlçatan olduqda `exam_results` cədvəlini idempotent şəkildə yaradır, API isə eyni bazaya read-only bağlantı qurur.

## Yoxlama

~~~sh
bun run build 2>&1 | tail -n 50
bun run test
~~~

## İstehsal qeydi

60.000–100.000 RPS yalnız WAF/load balancer arxasında üfüqi Fiber replika sayı, yüksək Redis hit nisbəti və PostgreSQL read-replica/topoloji ölçüləndikdə real hədəfdir. Bu Compose faylı funksional lokal topologiyadır, həmin throughput üçün sübut deyil. Real şəxsi nəticələr üçün FIN koduna əlavə identifikasiya, CAPTCHA/WAF rate-limit, Redis TLS/ACL və ayrıca read-only DB rolu tələb olunur.
