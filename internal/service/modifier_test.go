package service_test

import (
	"errors"
	"testing"

	"github.com/D3rille/kk-fast-food-system/internal/models"
	"github.com/D3rille/kk-fast-food-system/internal/service"
)

func sizeGroup() models.ModifierGroupWithOptions {
	return models.ModifierGroupWithOptions{
		ModifierGroup: models.ModifierGroup{ID: "group-size", Name: "Size", MinSelection: 1, MaxSelection: 1},
		Options: []models.ModifierOption{
			{ID: "opt-regular", ModifierGroupID: "group-size", Name: "Regular", ExtraPrice: 0, IsDefault: true},
			{ID: "opt-large", ModifierGroupID: "group-size", Name: "Large", ExtraPrice: 2500},
		},
	}
}

func addOnsGroup() models.ModifierGroupWithOptions {
	return models.ModifierGroupWithOptions{
		ModifierGroup: models.ModifierGroup{ID: "group-addons", Name: "Add-ons", MinSelection: 0, MaxSelection: 2},
		Options: []models.ModifierOption{
			{ID: "opt-cheese", ModifierGroupID: "group-addons", Name: "Extra Cheese", ExtraPrice: 2000},
			{ID: "opt-bacon", ModifierGroupID: "group-addons", Name: "Bacon", ExtraPrice: 3500},
		},
	}
}

func TestValidateAndPriceSelections_ValidRequiredAndOptional(t *testing.T) {
	groups := []models.ModifierGroupWithOptions{sizeGroup(), addOnsGroup()}

	extra, err := service.ValidateAndPriceSelections(groups, []string{"opt-large", "opt-cheese"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if extra != 2500+2000 {
		t.Errorf("expected extra price %d, got %d", 2500+2000, extra)
	}
}

func TestValidateAndPriceSelections_OptionalGroupCanBeSkipped(t *testing.T) {
	groups := []models.ModifierGroupWithOptions{sizeGroup(), addOnsGroup()}

	extra, err := service.ValidateAndPriceSelections(groups, []string{"opt-regular"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if extra != 0 {
		t.Errorf("expected extra price 0, got %d", extra)
	}
}

func TestValidateAndPriceSelections_MissingRequiredGroup(t *testing.T) {
	groups := []models.ModifierGroupWithOptions{sizeGroup()}

	_, err := service.ValidateAndPriceSelections(groups, nil)
	if !errors.Is(err, service.ErrInvalidModifierSelection) {
		t.Errorf("expected ErrInvalidModifierSelection, got %v", err)
	}
}

func TestValidateAndPriceSelections_ExceedsMaxSelection(t *testing.T) {
	groups := []models.ModifierGroupWithOptions{sizeGroup()}

	_, err := service.ValidateAndPriceSelections(groups, []string{"opt-regular", "opt-large"})
	if !errors.Is(err, service.ErrInvalidModifierSelection) {
		t.Errorf("expected ErrInvalidModifierSelection, got %v", err)
	}
}

func TestValidateAndPriceSelections_UnknownOption(t *testing.T) {
	groups := []models.ModifierGroupWithOptions{sizeGroup()}

	_, err := service.ValidateAndPriceSelections(groups, []string{"opt-regular", "opt-does-not-exist"})
	if !errors.Is(err, service.ErrInvalidModifierSelection) {
		t.Errorf("expected ErrInvalidModifierSelection, got %v", err)
	}
}
