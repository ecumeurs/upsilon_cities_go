package region

import "testing"

func TestGenerateRegions(t *testing.T) {
	Load()
}

func TestGenerateOneRegion(t *testing.T) {
	Load()
	reg, err := Generate("Elvenwood")
	if err != nil {
		t.Errorf("Failed to generate Elvenwood: %s", err)
		return
	}

	_, err = reg.Generate()
	if err != nil {
		t.Errorf("Failed to generate a grid based on Elvenwood region: %s", err)
		return
	}
}
