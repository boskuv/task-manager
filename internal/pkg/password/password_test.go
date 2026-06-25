package password

import "testing"

func TestHashAndCompare(t *testing.T) {
	t.Parallel()

	hash, err := Hash("secret-password")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	if hash == "secret-password" {
		t.Fatal("hash must not equal plain password")
	}
	if err := Compare(hash, "secret-password"); err != nil {
		t.Fatalf("Compare valid password: %v", err)
	}
	if err := Compare(hash, "wrong-password"); err == nil {
		t.Fatal("expected error for wrong password")
	}
}
