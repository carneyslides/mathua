package storage

import (
	"fmt"
	"testing"
	"time"
)

func newAdminV2Store(t *testing.T) *SQLiteStore {
	t.Helper()
	s, err := NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestAdminV2_StaffAndOwnerCountsAndList(t *testing.T) {
	s := newAdminV2Store(t)

	// Create 4 users with distinct roles
	u1, err := s.CreateStudent("Learner")
	if err != nil {
		t.Fatal(err)
	}
	_ = u1
	u2, err := s.CreateStudent("Mod")
	if err != nil {
		t.Fatal(err)
	}
	u3, err := s.CreateStudent("Admin")
	if err != nil {
		t.Fatal(err)
	}
	u4, err := s.CreateStudent("Owner")
	if err != nil {
		t.Fatal(err)
	}

	if err := s.SetStudentRole(u2.ID, RoleModerator); err != nil {
		t.Fatal(err)
	}
	if err := s.SetStudentRole(u3.ID, RoleAdmin); err != nil {
		t.Fatal(err)
	}
	if err := s.SetStudentRole(u4.ID, RoleOwner); err != nil {
		t.Fatal(err)
	}

	// CountStaff counts owner, admin, moderator (not learner u1)
	staffCount, err := s.CountStaff()
	if err != nil {
		t.Fatalf("CountStaff: %v", err)
	}
	if staffCount != 3 {
		t.Errorf("CountStaff = %d, want 3", staffCount)
	}

	// CountOwners counts only owner
	ownerCount, err := s.CountOwners()
	if err != nil {
		t.Fatalf("CountOwners: %v", err)
	}
	if ownerCount != 1 {
		t.Errorf("CountOwners = %d, want 1", ownerCount)
	}

	// ListStaff returns the 3 staff members
	staffList, err := s.ListStaff(10, 0)
	if err != nil {
		t.Fatalf("ListStaff: %v", err)
	}
	if len(staffList) != 3 {
		t.Fatalf("ListStaff len = %d, want 3", len(staffList))
	}
	roleMap := make(map[string]Role)
	for _, u := range staffList {
		roleMap[u.ID] = u.Role
	}
	if roleMap[u2.ID] != RoleModerator {
		t.Errorf("u2 role = %v, want %v", roleMap[u2.ID], RoleModerator)
	}
	if roleMap[u3.ID] != RoleAdmin {
		t.Errorf("u3 role = %v, want %v", roleMap[u3.ID], RoleAdmin)
	}
	if roleMap[u4.ID] != RoleOwner {
		t.Errorf("u4 role = %v, want %v", roleMap[u4.ID], RoleOwner)
	}
}

func TestAdminV2_ReportsModerationAndAudit(t *testing.T) {
	s := newAdminV2Store(t)
	now := time.Now().UTC()

	// Create 3 reports on topo.dimension (2 open, 1 resolved)
	r1, err := s.CreateReport(QuestionReport{
		ConceptID: "topo.dimension",
		Question:  "What is dimension of R^2?",
		Expected:  "2",
		Status:    "open",
		CreatedAt: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	r2, err := s.CreateReport(QuestionReport{
		ConceptID: "topo.dimension",
		Question:  "What is dimension of S^1?",
		Expected:  "1",
		Status:    "open",
		CreatedAt: now.Add(time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	r3, err := s.CreateReport(QuestionReport{
		ConceptID: "topo.dimension",
		Question:  "What is dimension of Cantor set?",
		Expected:  "0",
		Status:    "resolved",
		CreatedAt: now.Add(2 * time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	_ = r3

	// Import a report with origin id
	imported, err := s.ImportReport("prod-report-777", QuestionReport{
		ConceptID: "ra.order.supremum",
		Question:  "Find sup of (0, 1)",
		Expected:  "1",
		Status:    "open",
		CreatedAt: now.Add(3 * time.Minute),
	})
	if err != nil || !imported {
		t.Fatalf("ImportReport failed: imported=%v, err=%v", imported, err)
	}

	// CountReportsByStatus
	counts, err := s.CountReportsByStatus()
	if err != nil {
		t.Fatalf("CountReportsByStatus: %v", err)
	}
	if counts["open"] != 3 || counts["resolved"] != 1 {
		t.Errorf("CountReportsByStatus = %+v, want open:3 resolved:1", counts)
	}

	// GetReport
	rep, err := s.GetReport(r1)
	if err != nil || rep == nil {
		t.Fatalf("GetReport(r1): %v", err)
	}
	if rep.ConceptID != "topo.dimension" || rep.Status != "open" {
		t.Errorf("GetReport(r1) = %+v", rep)
	}

	// GetReport missing
	missingRep, err := s.GetReport(999999)
	if err != nil || missingRep != nil {
		t.Errorf("GetReport(missing) expected nil, got %+v, err=%v", missingRep, err)
	}

	// GetReportByOrigin
	repByOrig, err := s.GetReportByOrigin("prod-report-777")
	if err != nil || repByOrig == nil {
		t.Fatalf("GetReportByOrigin(prod-report-777): %v", err)
	}
	if repByOrig.OriginReportID != "prod-report-777" {
		t.Errorf("GetReportByOrigin OriginReportID = %q, want prod-report-777", repByOrig.OriginReportID)
	}
	if nonOrig, _ := s.GetReportByOrigin("nonexistent"); nonOrig != nil {
		t.Errorf("GetReportByOrigin nonexistent got %+v", nonOrig)
	}
	if emptyOrig, _ := s.GetReportByOrigin(""); emptyOrig != nil {
		t.Errorf("GetReportByOrigin empty got %+v", emptyOrig)
	}

	// ListUnsettledReportIDsForConcept: should return r1 and r2, but not r3
	unsettled, err := s.ListUnsettledReportIDsForConcept("topo.dimension")
	if err != nil {
		t.Fatalf("ListUnsettledReportIDsForConcept: %v", err)
	}
	if len(unsettled) != 2 || unsettled[0] != r1 || unsettled[1] != r2 {
		t.Errorf("ListUnsettledReportIDsForConcept = %v, want [%d, %d]", unsettled, r1, r2)
	}

	// UpdateReportStatusAudited: resolve r1
	resolvedRep, err := s.UpdateReportStatusAudited("mod_1", r1, "resolved", "fixed in commit 123")
	if err != nil {
		t.Fatalf("UpdateReportStatusAudited: %v", err)
	}
	if resolvedRep.Status != "resolved" || resolvedRep.Resolution != "fixed in commit 123" || resolvedRep.ResolvedBy != "mod_1" || resolvedRep.ResolvedAt == "" {
		t.Errorf("resolved report fields not updated: %+v", resolvedRep)
	}

	// Audit trail recorded
	audits, err := s.ListAdminAuditFor("report", fmt.Sprintf("%d", r1), 10)
	if err != nil {
		t.Fatalf("ListAdminAuditFor: %v", err)
	}
	if len(audits) != 1 || audits[0].Action != ActionReportStatusChange || audits[0].ActorID != "mod_1" {
		t.Errorf("ListAdminAuditFor = %+v", audits)
	}

	// Reopen r1: should clear resolution and resolved_by/resolved_at
	reopenedRep, err := s.UpdateReportStatusAudited("mod_1", r1, "open", "reopened")
	if err != nil {
		t.Fatalf("reopen UpdateReportStatusAudited: %v", err)
	}
	if reopenedRep.Status != "open" || reopenedRep.ResolvedBy != "" || reopenedRep.ResolvedAt != "" {
		t.Errorf("reopened report stamps not cleared: %+v", reopenedRep)
	}

	auditsAfter, err := s.ListAdminAuditFor("report", fmt.Sprintf("%d", r1), 10)
	if err != nil || len(auditsAfter) != 2 {
		t.Errorf("expected 2 audit entries after reopen, got %+v", auditsAfter)
	}
}

func TestAdminV2_InvitationLifecycle(t *testing.T) {
	s := newAdminV2Store(t)
	now := time.Now().UTC().Format(time.RFC3339)
	future := time.Now().UTC().Add(48 * time.Hour).Format(time.RFC3339)

	tokenHash1 := "token_hash_alpha_123"
	inv1 := AdminInvitation{
		Email:     "Alice@Mathua.org",
		Role:      RoleModerator,
		InvitedBy: "owner_1",
		CreatedAt: now,
		ExpiresAt: future,
	}

	id1, err := s.CreateAdminInvitation(inv1, tokenHash1)
	if err != nil || id1 <= 0 {
		t.Fatalf("CreateAdminInvitation: %v (id=%d)", err, id1)
	}

	// FindAdminInvitationByTokenHash
	found, err := s.FindAdminInvitationByTokenHash(tokenHash1)
	if err != nil || found == nil {
		t.Fatalf("FindAdminInvitationByTokenHash: %v", err)
	}
	if found.Email != "alice@mathua.org" || found.Role != RoleModerator || found.InvitedBy != "owner_1" {
		t.Errorf("invitation fields mismatch: %+v", found)
	}

	// Missing token hash
	missing, err := s.FindAdminInvitationByTokenHash("nonexistent_hash")
	if err != nil || missing != nil {
		t.Errorf("expected nil for missing token hash, got %+v, err=%v", missing, err)
	}

	// ListAdminInvitations
	invList, err := s.ListAdminInvitations(10, 0)
	if err != nil || len(invList) != 1 || invList[0].ID != id1 {
		t.Errorf("ListAdminInvitations = %+v", invList)
	}

	// ConsumeAdminInvitation
	stu, err := s.CreateStudent("Alice")
	if err != nil {
		t.Fatal(err)
	}
	consumed, ok, err := s.ConsumeAdminInvitation(tokenHash1, stu.ID)
	if err != nil || !ok || consumed == nil {
		t.Fatalf("ConsumeAdminInvitation: ok=%v, err=%v, consumed=%+v", ok, err, consumed)
	}
	if consumed.AcceptedBy != stu.ID || consumed.AcceptedAt == "" {
		t.Errorf("consumed invitation accepted fields not set: %+v", consumed)
	}

	// Verify student role updated to RoleModerator
	roleGot, err := s.GetStudentRole(stu.ID)
	if err != nil {
		t.Fatal(err)
	}
	if roleGot != RoleModerator {
		t.Errorf("student role = %v, want %v", roleGot, RoleModerator)
	}

	// Re-consuming same token must fail
	reconsume, ok2, err := s.ConsumeAdminInvitation(tokenHash1, stu.ID)
	if err != nil || ok2 || reconsume != nil {
		t.Errorf("re-consuming token should fail: ok=%v, err=%v", ok2, err)
	}

	// Revocation lifecycle with second invitation
	tokenHash2 := "token_hash_beta_456"
	inv2 := AdminInvitation{
		Email:     "bob@mathua.org",
		Role:      RoleAdmin,
		InvitedBy: "owner_1",
		CreatedAt: now,
		ExpiresAt: future,
	}
	id2, err := s.CreateAdminInvitation(inv2, tokenHash2)
	if err != nil {
		t.Fatal(err)
	}

	revoked, err := s.RevokeAdminInvitation("owner_1", id2)
	if err != nil {
		t.Fatalf("RevokeAdminInvitation: %v", err)
	}
	if revoked.RevokedBy != "owner_1" || revoked.RevokedAt == "" {
		t.Errorf("revoked invitation stamps missing: %+v", revoked)
	}

	// Consuming revoked token fails
	stuBob, err := s.CreateStudent("Bob")
	if err != nil {
		t.Fatal(err)
	}
	bobConsume, bobOk, err := s.ConsumeAdminInvitation(tokenHash2, stuBob.ID)
	if err != nil || bobOk || bobConsume != nil {
		t.Errorf("consuming revoked token should fail: ok=%v, err=%v", bobOk, err)
	}

	// Re-revoking already revoked invitation fails
	if _, err := s.RevokeAdminInvitation("owner_1", id2); err == nil {
		t.Error("expected error re-revoking already revoked invitation")
	}
}
