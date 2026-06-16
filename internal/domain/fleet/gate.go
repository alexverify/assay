package fleet

// GateResult is the fleet CI gate's decision: a single OK plus the specific
// machines and exposures that caused a failure, so the CLI can print exactly
// what to fix.
type GateResult struct {
	OK            bool               `json:"ok"`
	NonCompliant  []OwnerConformance `json:"nonCompliant,omitempty"`  // machines out of policy
	BlastBreaches []Exposure         `json:"blastBreaches,omitempty"` // drifted/quarantined artifacts whose reach exceeds the threshold
}

// Gate turns the fleet rollup into a CI pass/fail. It fails when either:
//
//   - any machine is out of policy (from CheckConformance), or
//   - a drifted or quarantined artifact reaches more than maxBlast machines —
//     a wide blast radius the team chose to block.
//
// maxBlast <= 0 disables the reach check (conformance alone gates). Pure: the
// caller supplies the already-computed report and conformance, so the gate
// reuses the exact same semantics the dashboard shows.
func Gate(rep Report, con Conformance, maxBlast int) GateResult {
	res := GateResult{OK: true}

	for _, m := range con.Machines {
		if !m.Compliant {
			res.NonCompliant = append(res.NonCompliant, m)
			res.OK = false
		}
	}

	if maxBlast > 0 {
		for _, e := range rep.Exposures {
			if e.Drifted > maxBlast || e.Quarantine > maxBlast {
				res.BlastBreaches = append(res.BlastBreaches, e)
				res.OK = false
			}
		}
	}

	return res
}
