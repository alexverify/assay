const STATS = [
  {
    figure: "140,963",
    caption:
      "security findings in an audit of 22,511 public agent skills across four registries (Mobb.ai, March 2026)",
  },
  {
    figure: "~1 in 6",
    caption: "skills containing a curl | sh remote-code-execution pattern (same audit)",
  },
  {
    figure: "341+",
    caption:
      'malicious skills found on a single registry carrying the Atomic Stealer (AMOS) infostealer payload — malware that exfiltrates crypto wallets, browser passwords, and SSH keys ("ClawHavoc," February 2026)',
  },
  {
    figure: "36%",
    caption:
      'of analyzed skills contained security flaws; 76 confirmed malicious payloads (Snyk "ToxicSkills," February 2026)',
  },
]

export function Stats() {
  return (
    <section className="border-b border-border">
      <div className="mx-auto max-w-[1100px] px-6 py-20">
        <p className="font-mono text-sm uppercase tracking-wider text-primary">
          The supply chain is already compromised.
        </p>
        <div className="mt-10 grid gap-x-8 gap-y-10 sm:grid-cols-2">
          {STATS.map((stat) => (
            <div key={stat.figure} className="border-l-2 border-border pl-5">
              <div className="font-mono text-3xl font-semibold tracking-tight text-foreground md:text-4xl">
                {stat.figure}
              </div>
              <p className="mt-3 text-sm leading-relaxed text-muted-foreground">{stat.caption}</p>
            </div>
          ))}
        </div>
        <p className="mt-12 max-w-3xl text-pretty leading-relaxed text-muted-foreground">
          Every one of these runs on a developer machine with full system permissions the moment it&apos;s installed.
          There is no signature verification, no runtime scanning, and no way to know whether what&apos;s on disk is the
          same code that was audited. Your <code className="font-mono text-foreground">.env</code> files, deploy
          credentials, SSH keys, and client codebases sit on the same machine.
        </p>
      </div>
    </section>
  )
}
