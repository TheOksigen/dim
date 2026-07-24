"use client";

import { FormEvent, useId, useRef, useState, useTransition } from "react";
import type { ApiError, ExamResult, LookupResponse } from "./types";

const finPattern = /^[A-Z0-9]{7}$/;

function normalizeFin(value: string) {
  return value.toUpperCase().replace(/[^A-Z0-9]/g, "").slice(0, 7);
}

function formatDate(value: string) {
  const date = new Date(`${value}T00:00:00`);
  if (Number.isNaN(date.getTime())) return value;

  return new Intl.DateTimeFormat("az-AZ", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
  }).format(date);
}

function InstitutionMark() {
  return (
    <svg aria-hidden="true" className="institution-mark" viewBox="0 0 48 48">
      <path d="M24 3.5 30.1 10l8.7-.3-.3 8.7 6.5 6.1-6.5 6.1.3 8.7-8.7-.3L24 44.5 17.9 39l-8.7.3.3-8.7L3.5 24l6-6.1-.3-8.7 8.7.3L24 3.5Z" />
      <path d="m24 12.2 3.2 6.6 7.3 1.1-5.2 5.1 1.2 7.2-6.5-3.5-6.5 3.5 1.2-7.2-5.2-5.1 7.3-1.1 3.2-6.6Z" />
      <circle cx="24" cy="24" r="3.4" />
    </svg>
  );
}

function SearchIcon() {
  return (
    <svg aria-hidden="true" viewBox="0 0 24 24">
      <circle cx="10.8" cy="10.8" r="6.4" />
      <path d="m16 16 4.2 4.2" />
    </svg>
  );
}

function ShieldIcon() {
  return (
    <svg aria-hidden="true" viewBox="0 0 24 24">
      <path d="M12 3.4 19 6v5.3c0 4.3-2.8 7.7-7 9.3-4.2-1.6-7-5-7-9.3V6l7-2.6Z" />
      <path d="m8.7 12.2 2.1 2.1 4.6-4.8" />
    </svg>
  );
}

function LockIcon() {
  return (
    <svg aria-hidden="true" viewBox="0 0 24 24">
      <rect x="5" y="10" width="14" height="10" rx="2" />
      <path d="M8 10V7.8a4 4 0 0 1 8 0V10M12 14v2.4" />
    </svg>
  );
}

function BoltIcon() {
  return (
    <svg aria-hidden="true" viewBox="0 0 24 24">
      <path d="m13.8 2.9-8 11h5.6l-1.2 7.2 8-11h-5.6l1.2-7.2Z" />
    </svg>
  );
}

function ResultIcon() {
  return (
    <svg aria-hidden="true" viewBox="0 0 24 24">
      <path d="M7 3.5h8.2L19 7.3v12.2a1 1 0 0 1-1 1H7a2 2 0 0 1-2-2v-13a2 2 0 0 1 2-2Z" />
      <path d="M15 3.5v4h4M8.5 12h7M8.5 16h7" />
    </svg>
  );
}

function ChevronIcon({ open }: { open: boolean }) {
  return (
    <svg aria-hidden="true" className={open ? "chevron is-open" : "chevron"} viewBox="0 0 24 24">
      <path d="m6 9 6 6 6-6" />
    </svg>
  );
}

function CandidateIcon() {
  return (
    <svg aria-hidden="true" className="candidate-icon" viewBox="0 0 48 48">
      <circle cx="24" cy="16" r="8" />
      <path d="M8.5 41c1.6-8 7.4-12 15.5-12s13.9 4 15.5 12" />
      <circle cx="24" cy="24" r="20" />
    </svg>
  );
}

function PassIndicator({ passed }: { passed: boolean }) {
  return (
    <div className={passed ? "pass-indicator is-passed" : "pass-indicator is-failed"}>
      <span className="pass-icon" aria-hidden="true">
        {passed ? "✓" : "!"}
      </span>
      <span>
        <strong>{passed ? "Müsabiqədən keçdiniz" : "Müsabiqədən keçmədiniz"}</strong>
        <small>
          {passed
            ? "Növbəti mərhələ üzrə məlumatlar ayrıca təqdim ediləcək."
            : "Nəticənizlə bağlı məlumat üçün dəstək bölməsinə müraciət edə bilərsiniz."}
        </small>
      </span>
    </div>
  );
}

function ResultDetails({ result }: { result: ExamResult }) {
  const [answersOpen, setAnswersOpen] = useState(false);
  const answersId = useId();
  const candidateName = `${result.lastName} ${result.firstName} ${result.fatherName}`;

  return (
    <section aria-labelledby="result-heading" className="result-section" id="netice">
      <div className="section-heading">
        <div>
          <p className="section-eyebrow">Nəticə</p>
          <h2 id="result-heading">İmtahan nəticəsi</h2>
        </div>
        <span className="version-label">FIN: {result.finCode}</span>
      </div>

      <div className="candidate-summary">
        <CandidateIcon />
        <div>
          <h3>{candidateName}</h3>
          <p>İmtahan tarixi: <time dateTime={result.examDate}>{formatDate(result.examDate)}</time></p>
        </div>
      </div>

      <div className="score-layout">
        <div className="score-table-wrap">
          <table className="score-table">
            <thead>
              <tr>
                <th scope="col">Fənn</th>
                <th scope="col">Bal</th>
              </tr>
            </thead>
            <tbody>
              {result.subjects.map((subject) => (
                <tr key={subject.name}>
                  <th scope="row">{subject.name}</th>
                  <td>{subject.score}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <aside aria-label="Yekun nəticə" className="total-panel">
          <span>Ümumi bal</span>
          <strong>{result.totalScore}</strong>
          <PassIndicator passed={result.passed} />
        </aside>
      </div>

      <div className="answers-block">
        <button
          aria-controls={answersId}
          aria-expanded={answersOpen}
          aria-label={answersOpen ? "Cavabları gizlət" : "Cavabları göstər"}
          className="answers-toggle"
          onClick={() => setAnswersOpen((current) => !current)}
          type="button"
        >
          <span className="answers-toggle-title">
            <ResultIcon />
            Cavablar
          </span>
          <ChevronIcon open={answersOpen} />
        </button>
        {answersOpen ? (
          <div className="answers-list" id={answersId}>
            {result.subjects.map((subject) => (
              <div className="answer-row" key={subject.name}>
                <span>{subject.name}</span>
                <code>{subject.answers}</code>
              </div>
            ))}
          </div>
        ) : null}
      </div>
    </section>
  );
}

export function ExamResultLookup() {
  const [finCode, setFinCode] = useState("");
  const [result, setResult] = useState<ExamResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isPending, startTransition] = useTransition();
  const resultRef = useRef<HTMLDivElement>(null);
  const canSubmit = finPattern.test(finCode);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setResult(null);

    if (!canSubmit) {
      setError("FIN kodu 7 böyük hərf və ya rəqəmdən ibarət olmalıdır.");
      return;
    }

    startTransition(async () => {
      try {
        const response = await fetch("/api/results/lookup", {
          method: "POST",
          headers: { "content-type": "application/json" },
          body: JSON.stringify({ finCode }),
        });
        const body = (await response.json()) as LookupResponse & ApiError;

        if (!response.ok || !body.result) {
          setError(body.error?.message ?? "Nəticə tapılmadı. FIN kodunu yenidən yoxlayın.");
          return;
        }

        setResult(body.result);
        window.setTimeout(() => resultRef.current?.scrollIntoView({ behavior: "smooth", block: "start" }), 0);
      } catch {
        setError("Sorğunu tamamlamaq mümkün olmadı. Yenidən cəhd edin.");
      }
    });
  }

  return (
    <main>
      <header className="site-header">
        <a aria-label="İmtahan nəticələri ana səhifə" className="brand" href="#ana-sehife">
          <InstitutionMark />
          <span>
            <strong>Elm və Təhsil Nazirliyi</strong>
            <small>İmtahan nəticələri</small>
          </span>
        </a>
        <nav aria-label="Əsas naviqasiya" className="desktop-nav">
          <a href="#ana-sehife">Ana səhifə</a>
          <a href="#netice">Nəticənin yoxlanılması</a>
          <a href="#melumat">Məlumat</a>
          <a href="#elaqe">Əlaqə</a>
        </nav>
      </header>

      <section className="hero" id="ana-sehife">
        <div className="hero-copy">
          <h1>İmtahan nəticənizi yoxlayın</h1>
          <div className="title-rule" aria-hidden="true"><span /><i /></div>
          <p>
            FIN kodunuzu daxil edin və iştirak etdiyiniz imtahan üzrə nəticənizi təhlükəsiz,
            sürətli şəkildə əldə edin.
          </p>
          <ul className="benefit-list">
            <li><ShieldIcon /> Rəsmi və təhlükəsiz xidmət</li>
            <li><LockIcon /> Məlumatlarınız qorunur</li>
            <li><BoltIcon /> Nəticələr anında təqdim olunur</li>
          </ul>
        </div>

        <div className="lookup-panel">
          <h2>Nəticənizi FIN kodu ilə yoxlayın</h2>
          <form noValidate onSubmit={handleSubmit}>
            <label htmlFor="fin-code">
              FIN kodunuz
              <span>7 simvol</span>
            </label>
            <input
              aria-describedby={error ? "lookup-error" : "fin-hint"}
              autoCapitalize="characters"
              autoComplete="off"
              id="fin-code"
              inputMode="text"
              maxLength={7}
              name="finCode"
              onChange={(event) => setFinCode(normalizeFin(event.target.value))}
              placeholder="Məsələn, 7A1B2C3"
              spellCheck={false}
              value={finCode}
            />
            <p className="input-hint" id="fin-hint">Yalnız böyük hərflər və rəqəmlərdən istifadə edin.</p>
            <button className="lookup-button" disabled={isPending} type="submit">
              <SearchIcon />
              {isPending ? "Axtarılır…" : "Nəticələri göstər"}
            </button>
            {error ? <p aria-live="polite" className="form-error" id="lookup-error" role="alert">{error}</p> : null}
          </form>
        </div>
      </section>

      <section aria-label="Xidmətin xüsusiyyətləri" className="assurance-strip" id="melumat">
        <div>
          <BoltIcon />
          <span><strong>Tez və rahat</strong><small>FIN kodunu daxil etməyiniz kifayətdir.</small></span>
        </div>
        <div>
          <ShieldIcon />
          <span><strong>Təhlükəsiz</strong><small>Sorğular qorunan əlaqə ilə emal olunur.</small></span>
        </div>
        <div>
          <ResultIcon />
          <span><strong>Tam nəticə</strong><small>Fənn balları və cavablar təqdim olunur.</small></span>
        </div>
      </section>

      {result ? <div ref={resultRef}><ResultDetails result={result} /></div> : null}

      <footer className="site-footer" id="elaqe">
        <div className="footer-brand">
          <InstitutionMark />
          <span><strong>İmtahan nəticələri</strong><small>© {new Date().getFullYear()} Bütün hüquqlar qorunur.</small></span>
        </div>
        <div className="footer-links">
          <a href="#melumat">Məlumatların qorunması</a>
          <a href="#ana-sehife">İstifadə şərtləri</a>
          <a href="mailto:info@example.gov.az">info@example.gov.az</a>
        </div>
      </footer>
    </main>
  );
}
