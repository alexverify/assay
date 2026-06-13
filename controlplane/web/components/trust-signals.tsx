const SIGNALS = [
  "Open-source core",
  "Reproducible builds",
  "Local-first — your code never leaves your machine",
  "Built and dogfooded by a CTO running an engineering team on Claude Code daily.",
]

export function TrustSignals() {
  return (
    <section className="border-b border-border">
      <div className="mx-auto max-w-[1100px] px-6 py-8">
        <ul className="flex flex-col flex-wrap items-start gap-x-8 gap-y-3 font-mono text-xs text-muted-foreground sm:flex-row sm:items-center">
          {SIGNALS.map((signal, idx) => (
            <li key={signal} className="flex items-center gap-3">
              {idx > 0 && (
                <span className="hidden text-border sm:inline" aria-hidden="true">
                  ·
                </span>
              )}
              {signal}
            </li>
          ))}
        </ul>
      </div>
    </section>
  )
}
