# İmtahan nəticələrinin yoxlanılması

Bun ilə işləyən SSR Next.js interfeysi, Go Fiber hot-path API-si, PostgreSQL mənbə bazası və Redis cache-aside qatı.

## Lokal işə salma

~~~sh
cp .env.example .env
docker compose up --build
~~~

Sistem ilk işə düşəndə sxem avtomatik tətbiq edilir. Sintetik məlumat ayrıca yüklənir; API startında heç vaxt seed edilmir:

~~~sh
# sürətli yoxlama üçün
docker compose run --rm --entrypoint /app/seed api -total 10000

# tam test həcmi üçün
docker compose run --rm --entrypoint /app/seed api -total 10000000 -batch-size 10000
~~~

İlk sintetik FIN `T000000` olur. Web: http://localhost:3000, API liveness: http://localhost:8080/healthz.

## Yoxlama

~~~sh
bun run build 2>&1 | tail -n 50
bun run test
~~~

## İstehsal qeydi

60.000–100.000 RPS yalnız WAF/load balancer arxasında üfüqi Fiber replika sayı, yüksək Redis hit nisbəti və PostgreSQL read-replica/topoloji ölçüləndikdə real hədəfdir. Bu Compose faylı funksional lokal topologiyadır, həmin throughput üçün sübut deyil. Real şəxsi nəticələr üçün FIN koduna əlavə identifikasiya, CAPTCHA/WAF rate-limit, Redis TLS/ACL və ayrıca read-only DB rolu tələb olunur.
