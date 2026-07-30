package programs

import (
	"testing"
)

func TestProfilesStoreCRUDAndMembership(t *testing.T) {
	root := t.TempDir()
	store, err := NewProfilesStore(root)
	if err != nil {
		t.Fatalf("NewProfilesStore() error = %v", err)
	}
	profile := CustomProfile{
		ID:       "portable",
		Name:     "Portable",
		Programs: []string{"7-Zip"},
		Tweaks:   []string{"show-extensions"},
	}
	if err := store.Save(profile); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if err := store.AddProgramToProfile(profile.ID, "Rufus"); err != nil {
		t.Fatalf("AddProgramToProfile() error = %v", err)
	}
	if err := store.AddProgramToProfile(profile.ID, "Rufus"); err != nil {
		t.Fatalf("duplicate AddProgramToProfile() error = %v", err)
	}

	reloaded, err := NewProfilesStore(root)
	if err != nil {
		t.Fatalf("reload error = %v", err)
	}
	got := reloaded.GetAll()
	if len(got) != 1 || len(got[0].Programs) != 2 {
		t.Fatalf("unexpected profiles after reload: %#v", got)
	}

	if err := reloaded.RemoveProgramFromProfile(profile.ID, "7-Zip"); err != nil {
		t.Fatalf("RemoveProgramFromProfile() error = %v", err)
	}
	if err := reloaded.Delete(profile.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if got := reloaded.GetAll(); len(got) != 0 {
		t.Fatalf("profiles remain after delete: %#v", got)
	}
}

func TestProfilesStoreRejectsMissingIDAndUnknownProfile(t *testing.T) {
	store, err := NewProfilesStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Save(CustomProfile{Name: "Missing ID"}); err == nil {
		t.Fatal("Save() should reject an empty ID")
	}
	if err := store.Delete("missing"); err == nil {
		t.Fatal("Delete() should reject an unknown profile")
	}
	if err := store.AddProgramToProfile("missing", "Rufus"); err == nil {
		t.Fatal("AddProgramToProfile() should reject an unknown profile")
	}
}

func TestProfilesStoreGetAllReturnsIndependentCopy(t *testing.T) {
	store, err := NewProfilesStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Save(CustomProfile{
		ID:       "safe",
		Name:     "Safe",
		Programs: []string{"Rufus"},
		Tweaks:   []string{"one"},
	}); err != nil {
		t.Fatal(err)
	}

	first := store.GetAll()
	first[0].Name = "mutated"
	first[0].Programs[0] = "mutated"
	first[0].Tweaks[0] = "mutated"

	second := store.GetAll()
	if second[0].Name != "Safe" ||
		second[0].Programs[0] != "Rufus" ||
		second[0].Tweaks[0] != "one" {
		t.Fatalf("GetAll() leaked internal state: %#v", second[0])
	}
}
