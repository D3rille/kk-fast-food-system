-- Marks a modifier option as the pre-selected choice for its group.
-- Required groups (min_selection >= 1) must have at least one default option
-- so kiosk customers always start with a valid selection.
ALTER TABLE modifier_options ADD COLUMN IF NOT EXISTS is_default BOOLEAN NOT NULL DEFAULT FALSE;
