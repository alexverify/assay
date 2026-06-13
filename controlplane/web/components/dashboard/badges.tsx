import { cn } from "@/lib/utils"
import { SEVERITY_STYLES, DRIFT_STYLES } from "@/lib/scan-utils"
import type { Severity, DriftStatus } from "@/lib/scan-data"

export function SeverityBadge({ severity }: { severity: Severity }) {
  const s = SEVERITY_STYLES[severity]
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-md border px-2 py-0.5 font-mono text-[11px] font-medium uppercase tracking-wide",
        s.bg,
        s.border,
        s.text,
      )}
    >
      <span className={cn("h-1.5 w-1.5 rounded-full", s.dot)} aria-hidden />
      {s.label}
    </span>
  )
}

export function DriftBadge({ status }: { status: DriftStatus }) {
  const d = DRIFT_STYLES[status]
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-md border px-2 py-0.5 font-mono text-[11px] font-medium",
        d.bg,
        d.border,
        d.text,
      )}
    >
      {d.label}
    </span>
  )
}
