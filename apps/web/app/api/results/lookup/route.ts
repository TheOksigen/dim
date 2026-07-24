import { NextResponse } from "next/server";

export const runtime = "nodejs";
export const dynamic = "force-dynamic";

const upstreamTimeoutMs = 2_500;
const maxRequestBytes = 1_024;

export async function POST(request: Request) {
  const apiUrl = process.env.RESULTS_API_URL;

  const contentLength = Number(request.headers.get("content-length") ?? "0");
  if (Number.isFinite(contentLength) && contentLength > maxRequestBytes) {
    return NextResponse.json(
      { error: { code: "payload_too_large", message: "Sorğu qəbul edilən həddən böyükdür." } },
      { status: 413 },
    );
  }

  if (!apiUrl) {
    return NextResponse.json(
      { error: { code: "service_unavailable", message: "Xidmət hazırda əlçatan deyil." } },
      { status: 503 },
    );
  }

  let body: unknown;
  try {
    body = await request.json();
  } catch {
    return NextResponse.json(
      { error: { code: "invalid_request", message: "Göndərilən məlumat düzgün deyil." } },
      { status: 400 },
    );
  }

  try {
    const response = await fetch(`${apiUrl.replace(/\/$/, "")}/api/v1/results/lookup`, {
      method: "POST",
      headers: {
        "content-type": "application/json",
        "x-request-id": crypto.randomUUID(),
      },
      body: JSON.stringify(body),
      cache: "no-store",
      signal: AbortSignal.timeout(upstreamTimeoutMs),
    });

    const payload = await response.text();
    return new NextResponse(payload, {
      status: response.status,
      headers: {
        "cache-control": "no-store",
        "content-type": response.headers.get("content-type") ?? "application/json; charset=utf-8",
      },
    });
  } catch {
    return NextResponse.json(
      { error: { code: "service_unavailable", message: "Xidmətə qoşulmaq mümkün olmadı. Yenidən cəhd edin." } },
      { status: 503, headers: { "cache-control": "no-store" } },
    );
  }
}
