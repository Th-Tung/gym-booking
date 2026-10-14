package models_test

import (
	"fmt"
	"testing"
	"time"

	"gym-booking/internal/config"
	"gym-booking/internal/models"
)

func TestAuthenticationAndHashing(t *testing.T) {
	plain := "TestSecret@123"
	hash, err := models.HashPassword(plain)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	u := &models.User{PasswordHash: hash}
	if !u.CheckPassword(plain) {
		t.Errorf("Password check failed for valid password")
	}
	if u.CheckPassword("WrongPassword") {
		t.Errorf("Password check succeeded for invalid password")
	}
}

func TestBookingCollisionDetection(t *testing.T) {
	cfg := config.GetConfig()
	db, err := config.InitDB(cfg)
	if err != nil {
		t.Skipf("Skipping DB test, cannot connect to MySQL: %v", err)
	}
	defer db.Close()

	equipmentID := 1
	testDate := time.Now().AddDate(0, 0, 5).Format("2006-01-02")
	start1 := fmt.Sprintf("%s 14:00:00", testDate)
	end1 := fmt.Sprintf("%s 15:00:00", testDate)

	// Clean up any test booking
	_, _ = db.Exec("DELETE FROM bookings WHERE equipment_id = ? AND start_time = ?", equipmentID, start1)

	// Create test booking
	b, err := models.CreateBooking(db, 2, equipmentID, start1, end1, "Go Unit Test")
	if err != nil {
		t.Fatalf("Failed to create initial booking: %v", err)
	}
	defer func() {
		_, _ = db.Exec("DELETE FROM bookings WHERE id = ?", b.ID)
	}()

	// 1. Overlapping slot 14:30 - 15:30 MUST fail
	overlapStart := fmt.Sprintf("%s 14:30:00", testDate)
	overlapEnd := fmt.Sprintf("%s 15:30:00", testDate)

	hasOverlap, err := models.CheckOverlap(db, equipmentID, overlapStart, overlapEnd)
	if err != nil {
		t.Fatalf("CheckOverlap error: %v", err)
	}
	if !hasOverlap {
		t.Errorf("Expected overlap detected for %s - %s, but got none", overlapStart, overlapEnd)
	}

	// Attempt to create overlapping booking via CreateBooking
	_, err = models.CreateBooking(db, 2, equipmentID, overlapStart, overlapEnd, "Should Fail")
	if err == nil {
		t.Errorf("Expected CreateBooking to reject overlapping time, but succeeded")
	}

	// 2. Non-overlapping slot 15:00 - 16:00 MUST pass
	freeStart := fmt.Sprintf("%s 15:00:00", testDate)
	freeEnd := fmt.Sprintf("%s 16:00:00", testDate)

	hasOverlapFree, err := models.CheckOverlap(db, equipmentID, freeStart, freeEnd)
	if err != nil {
		t.Fatalf("CheckOverlap error: %v", err)
	}
	if hasOverlapFree {
		t.Errorf("Expected NO overlap for %s - %s, but got overlap", freeStart, freeEnd)
	}

	// 3. Test cancellation frees up slot
	err = models.CancelBooking(db, b.ID, 2, false, "Test cancel")
	if err != nil {
		t.Fatalf("Failed to cancel booking: %v", err)
	}

	hasOverlapAfterCancel, err := models.CheckOverlap(db, equipmentID, overlapStart, overlapEnd)
	if err != nil {
		t.Fatalf("CheckOverlap error: %v", err)
	}
	if hasOverlapAfterCancel {
		t.Errorf("Expected slot to be freed after cancellation, but still reported overlap")
	}
}
