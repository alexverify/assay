import { cn } from "@/lib/utils"

interface StatCardProps {
  label: string
  value: string | number
  hint?: string
  accent?: "critical" | "high" | "medium" | "ok" | "neutral"
}

const ACCENTS: Record<NonNullable<StatCardProps["accent"]>, string> = {
  critical: "text-sev-critical",
  high: "text-sev-high",
  medium: "text-sev-medium",
  ok: "text-ok",
  neutral: "text-foreground",
}

export function StatCard({ label, value, hint, accent = "neutral" }: StatCardProps) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <p className="font-mono text-[11px] uppercase tracking-wide text-muted-foreground">{label}</p>
      <p className={cn("mt-2 font-mono text-3xl font-semibold tabular-nums", ACCENTS[accent])}>
        {value}
      </p>
      {hint ? <p className="mt-1 text-xs text-muted-foreground">{hint}</p> : null}
    </div>
  )
}
