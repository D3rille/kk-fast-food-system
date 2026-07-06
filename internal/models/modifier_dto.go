package models

// CreateModifierGroupRequest represents the payload to add a new modifier group.
type CreateModifierGroupRequest struct {
	Name         string `json:"name"`
	MinSelection int    `json:"min_selection"`
	MaxSelection int    `json:"max_selection"`
}

// UpdateModifierGroupRequest represents the payload to modify an existing modifier group.
type UpdateModifierGroupRequest struct {
	Name         string `json:"name"`
	MinSelection int    `json:"min_selection"`
	MaxSelection int    `json:"max_selection"`
}

// CreateModifierOptionRequest represents the payload to add an option to a modifier group.
type CreateModifierOptionRequest struct {
	Name       string `json:"name"`
	ExtraPrice int64  `json:"extra_price"`
	IsDefault  bool   `json:"is_default"`
}

// UpdateModifierOptionRequest represents the payload to modify an existing modifier option.
type UpdateModifierOptionRequest struct {
	Name       string `json:"name"`
	ExtraPrice int64  `json:"extra_price"`
	IsDefault  bool   `json:"is_default"`
}

// AttachModifierGroupRequest represents the payload to associate a modifier group with a product.
type AttachModifierGroupRequest struct {
	ModifierGroupID string `json:"modifier_group_id"`
}

// ModifierOptionResponse represents the serialized modifier option output.
type ModifierOptionResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	ExtraPrice int64  `json:"extra_price"`
	IsDefault  bool   `json:"is_default"`
}

// ModifierGroupResponse represents the serialized modifier group output, including its options.
// IsRequired is derived from MinSelection >= 1, so the kiosk always knows whether it must
// force a selection before the item can be added to the cart.
type ModifierGroupResponse struct {
	ID           string                   `json:"id"`
	Name         string                   `json:"name"`
	MinSelection int                      `json:"min_selection"`
	MaxSelection int                      `json:"max_selection"`
	IsRequired   bool                     `json:"is_required"`
	Options      []ModifierOptionResponse `json:"options"`
}
