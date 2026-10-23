package ai

import "testing"

func stub(id string, pol Policy) *Stub {
	return &Stub{IDValue: id, Pol: pol, Reply: "ok"}
}

func ids(ps []Provider) []string {
	out := make([]string, 0, len(ps))
	for _, p := range ps {
		out = append(out, p.ID())
	}
	return out
}

func eq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func productionPolicy(max DataClass) Policy {
	return Policy{
		MaxClass:            max,
		ProductionPermitted: true,
		Retention:           RetentionNone,
		Source:              "https://example.test/policy (2026-10-16)",
	}
}

func TestNewRegistryRejectsBadChains(t *testing.T) {
	t.Run("nil provider", func(t *testing.T) {
		if _, err := NewRegistry([]Provider{nil}, Options{}); err == nil {
			t.Fatal("want error for nil provider")
		}
	})
	t.Run("empty id", func(t *testing.T) {
		if _, err := NewRegistry([]Provider{stub("  ", productionPolicy(DataPublic))}, Options{}); err == nil {
			t.Fatal("want error for empty id")
		}
	})
	t.Run("duplicate id", func(t *testing.T) {
		_, err := NewRegistry([]Provider{
			stub("groq", productionPolicy(DataPublic)),
			stub("groq", productionPolicy(DataPublic)),
		}, Options{})
		if err == nil {
			t.Fatal("want error for duplicate id")
		}
	})
	t.Run("unsourced sensitive policy", func(t *testing.T) {
		_, err := NewRegistry([]Provider{
			stub("groq", Policy{MaxClass: DataIdentified, ProductionPermitted: true}),
		}, Options{})
		if err == nil {
			t.Fatal("want error for sensitive policy with no source")
		}
	})
	t.Run("dev and production permitted", func(t *testing.T) {
		_, err := NewRegistry([]Provider{
			stub("trial", Policy{MaxClass: DataPublic, DevOnly: true, ProductionPermitted: true}),
		}, Options{})
		if err == nil {
			t.Fatal("want error for DevOnly+ProductionPermitted")
		}
	})
	t.Run("empty chain is valid", func(t *testing.T) {
		r, err := NewRegistry(nil, Options{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if r.Enabled() {
			t.Fatal("empty registry must report disabled")
		}
	})
}

func TestSelectFiltersByDataClass(t *testing.T) {
	pub := stub("pub", productionPolicy(DataPublic))
	psu := stub("psu", productionPolicy(DataPseudonymous))
	idn := stub("idn", productionPolicy(DataIdentified))
	r, err := NewRegistry([]Provider{pub, psu, idn}, Options{})
	if err != nil {
		t.Fatal(err)
	}

	if got := ids(r.Select(DataPublic)); !eq(got, []string{"pub", "psu", "idn"}) {
		t.Errorf("public select = %v", got)
	}
	if got := ids(r.Select(DataPseudonymous)); !eq(got, []string{"psu", "idn"}) {
		t.Errorf("pseudonymous select = %v", got)
	}
	if got := ids(r.Select(DataIdentified)); !eq(got, []string{"idn"}) {
		t.Errorf("identified select = %v", got)
	}
}

func TestSelectExcludesDevOnlyInProduction(t *testing.T) {
	prod := stub("groq", productionPolicy(DataPseudonymous))
	trial := stub("trial", Policy{MaxClass: DataIdentified, DevOnly: true, Retention: RetentionUnknown, Source: "dev-only fixture"})

	// Production: the trial endpoint must never be returned, even though it
	// permits the class, so a head failure cannot leak learner data to it.
	rp, err := NewRegistry([]Provider{prod, trial}, Options{Production: true, AllowDev: true, ZDRConfirmed: true})
	if err != nil {
		t.Fatal(err)
	}
	if got := ids(rp.Select(DataIdentified)); !eq(got, []string{}) {
		t.Errorf("production select leaked dev-only provider: %v", got)
	}
	if got := ids(rp.Select(DataPublic)); !eq(got, []string{"groq"}) {
		t.Errorf("production public select = %v, want [groq]", got)
	}

	// Development with the dev flag: the trial provider is usable, after the
	// vetted one, but only for data its policy and retention both permit. Its
	// retention is unverified, so learner content still does not reach it.
	rd, err := NewRegistry([]Provider{prod, trial}, Options{AllowDev: true})
	if err != nil {
		t.Fatal(err)
	}
	if got := ids(rd.Select(DataPseudonymous)); !eq(got, []string{"groq"}) {
		t.Errorf("dev pseudonymous select = %v, want [groq] — unverified retention must not take learner data", got)
	}
	if got := ids(rd.Select(DataPublic)); !eq(got, []string{"groq", "trial"}) {
		t.Errorf("dev public select = %v, want [groq trial]", got)
	}

	// Development without the dev flag: inert.
	rn, err := NewRegistry([]Provider{prod, trial}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got := ids(rn.Select(DataPublic)); !eq(got, []string{"groq"}) {
		t.Errorf("dev-no-flag select = %v, want [groq]", got)
	}
}

func TestSelectExcludesNonProductionPolicyInProduction(t *testing.T) {
	// A provider whose terms do not permit production learner traffic is not
	// selectable in production even for public data.
	preview := stub("preview", Policy{MaxClass: DataIdentified, ProductionPermitted: false, Source: "preview tier"})
	r, err := NewRegistry([]Provider{preview}, Options{Production: true})
	if err != nil {
		t.Fatal(err)
	}
	if got := ids(r.Select(DataPublic)); !eq(got, []string{}) {
		t.Errorf("production select included non-production provider: %v", got)
	}
}

// TestProductionRequiresZeroRetentionConfirmation pins the gate that keeps a
// declared-but-unenabled zero-retention policy from being trusted: in
// production a provider claiming RetentionNone is a configuration error until
// the operator confirms the account setting, because the policy text says the
// switch exists, not that it is on.
func TestProductionRequiresZeroRetentionConfirmation(t *testing.T) {
	prod := stub("groq", productionPolicy(DataPseudonymous))

	if _, err := NewRegistry([]Provider{prod}, Options{Production: true}); err == nil {
		t.Error("production must refuse an unconfirmed zero-retention claim")
	}
	if _, err := NewRegistry([]Provider{prod}, Options{Production: true, ZDRConfirmed: true}); err != nil {
		t.Errorf("a confirmed deployment must build: %v", err)
	}
	// Outside production the gate does not apply: there is no learner traffic to
	// protect, and failing here would make the test suite unrunnable.
	if _, err := NewRegistry([]Provider{prod}, Options{}); err != nil {
		t.Errorf("non-production must not require the confirmation: %v", err)
	}
}

// TestSelectEnforcesRetention pins the rule that retention is a control rather
// than a description: an unverified or training provider is denied anything
// that identifies a learner, and only zero retention may receive identifiable
// data. "Nobody checked" is treated as the worst case, not the default.
func TestSelectEnforcesRetention(t *testing.T) {
	want := map[Retention]map[DataClass]bool{
		RetentionNone:      {DataPublic: true, DataPseudonymous: true, DataIdentified: true},
		RetentionTemporary: {DataPublic: true, DataPseudonymous: true, DataIdentified: false},
		RetentionUnknown:   {DataPublic: true, DataPseudonymous: false, DataIdentified: false},
		RetentionTrained:   {DataPublic: true, DataPseudonymous: false, DataIdentified: false},
	}
	for ret, byClass := range want {
		r, err := NewRegistry([]Provider{stub("p", Policy{
			MaxClass: DataIdentified, ProductionPermitted: true,
			Retention: ret, Source: "retention fixture",
		})}, Options{})
		if err != nil {
			t.Fatalf("retention %v: %v", ret, err)
		}
		for _, class := range []DataClass{DataPublic, DataPseudonymous, DataIdentified} {
			got := len(r.Select(class)) == 1
			if got != byClass[class] {
				t.Errorf("retention %v, class %v: selectable = %v, want %v", ret, class, got, byClass[class])
			}
		}
	}
}

func TestNilRegistryIsSafe(t *testing.T) {
	var r *Registry
	if r.Enabled() {
		t.Error("nil registry must be disabled")
	}
	if r.Select(DataPublic) != nil {
		t.Error("nil registry Select must be nil")
	}
	if r.IDs() != nil {
		t.Error("nil registry IDs must be nil")
	}
}

func TestRegistry_IDs_WithProviders(t *testing.T) {
	s1 := &Stub{IDValue: "p1", Pol: Policy{ProductionPermitted: true}}
	s2 := &Stub{IDValue: "p2", Pol: Policy{ProductionPermitted: true}}
	r, err := NewRegistry([]Provider{s1, s2}, Options{})
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	ids := r.IDs()
	if len(ids) != 2 || ids[0] != "p1" || ids[1] != "p2" {
		t.Errorf("IDs = %v, want [p1 p2]", ids)
	}
}
