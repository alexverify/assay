export function Tiers() {
  return (
    <section className="border-b border-border bg-card">
      <div className="mx-auto max-w-[1100px] px-6 py-20">
        <h2 className="text-balance font-mono text-2xl font-semibold tracking-tight md:text-3xl">
          From solo to team in 15 minutes
        </h2>

        <div className="mt-12 grid gap-10 md:grid-cols-2">
          <div className="rounded-lg border border-border bg-background p-6">
            <h3 className="font-mono text-sm uppercase tracking-wider text-muted-foreground">Solo (free, forever)</h3>
            <p className="mt-4 leading-relaxed text-foreground">
              Run the audit. Commit the lockfile. Done — no account, no config, no agent rewiring.
            </p>
          </div>

          <div className="rounded-lg border border-primary/40 bg-background p-6">
            <h3 className="font-mono text-sm uppercase tracking-wider text-primary">Team ($10–20/dev/month)</h3>
            <p className="mt-4 leading-relaxed text-foreground">
              One shared lockfile + the GitHub Action = &quot;only approved, unmodified skills run here&quot; across the
              whole team. Centralized policy, drift alerts, and audit logs for the person who&apos;s actually liable when
              a dev machine reaches production.
            </p>
            <a
              href="#team"
              className="mt-6 inline-flex items-center rounded-md bg-primary px-4 py-2.5 font-mono text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background"
            >
              Set up team policy
            </a>
          </div>
        </div>
      </div>
    </section>
  )
}
