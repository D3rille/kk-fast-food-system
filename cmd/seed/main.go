// Package main is a one-shot CLI seeder for mock development data.
// Run via: make mock-seed
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/D3rille/kk-fast-food-system/internal/config"
	"github.com/D3rille/kk-fast-food-system/internal/database"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// Hardcoded UUIDs so re-running the seeder is fully idempotent (upserts on id).
const (
	storeID   = "aaaaaaaa-0000-0000-0000-000000000001"
	adminID   = "aaaaaaaa-0000-0000-0000-000000000002"
	cashierID = "aaaaaaaa-0000-0000-0000-000000000003"
	managerID = "aaaaaaaa-0000-0000-0000-000000000004"
	kitchenID = "aaaaaaaa-0000-0000-0000-000000000005"

	catBurgers  = "aaaaaaaa-0001-0000-0000-000000000001"
	catChicken  = "aaaaaaaa-0002-0000-0000-000000000001"
	catSides    = "aaaaaaaa-0003-0000-0000-000000000001"
	catDrinks   = "aaaaaaaa-0004-0000-0000-000000000001"
	catDesserts = "aaaaaaaa-0005-0000-0000-000000000001"

	// Products that carry modifier groups (subset of seedProducts, referenced by id below).
	prodSpicyCrispyBurger = "aaaaaaaa-0001-0004-0000-000000000001"
	prodFriedChickenMeal  = "aaaaaaaa-0002-0001-0000-000000000001"
	prodChickenStrips     = "aaaaaaaa-0002-0002-0000-000000000001"
	prodChickenWings      = "aaaaaaaa-0002-0003-0000-000000000001"
	prodNashvilleHot      = "aaaaaaaa-0002-0004-0000-000000000001"
	prodClassicSmash      = "aaaaaaaa-0001-0001-0000-000000000001"
	prodDoubleSmash       = "aaaaaaaa-0001-0002-0000-000000000001"
	prodMushroomSwiss     = "aaaaaaaa-0001-0003-0000-000000000001"
	prodFrenchFries       = "aaaaaaaa-0003-0001-0000-000000000001"
	prodOnionRings        = "aaaaaaaa-0003-0002-0000-000000000001"
	prodMozzarellaSticks  = "aaaaaaaa-0003-0003-0000-000000000001"
	prodLoadedFries       = "aaaaaaaa-0003-0004-0000-000000000001"
	prodIcedTea           = "aaaaaaaa-0004-0001-0000-000000000001"
	prodSoda              = "aaaaaaaa-0004-0002-0000-000000000001"
	prodLemonade          = "aaaaaaaa-0004-0003-0000-000000000001"
	prodMilkshake         = "aaaaaaaa-0004-0004-0000-000000000001"

	// Modifier groups
	groupSize            = "bbbbbbbb-0001-0000-0000-000000000001"
	groupSpiceLevel      = "bbbbbbbb-0002-0000-0000-000000000001"
	groupFriesFlavor     = "bbbbbbbb-0003-0000-0000-000000000001"
	groupSodaType        = "bbbbbbbb-0004-0000-0000-000000000001"
	groupSweetness       = "bbbbbbbb-0005-0000-0000-000000000001"
	groupMilkshakeFlavor = "bbbbbbbb-0006-0000-0000-000000000001"
	groupBurgerAddOns    = "bbbbbbbb-0007-0000-0000-000000000001"
)

type seedProduct struct {
	id          string
	categoryID  string
	name        string
	description string
	basePrice   int64 // centavos
}

var seedProducts = []seedProduct{
	// Burgers
	{"aaaaaaaa-0001-0001-0000-000000000001", catBurgers, "Classic Smash Burger", "Hand-smashed beef patty, American cheese, pickles, special sauce", 15900},
	{"aaaaaaaa-0001-0002-0000-000000000001", catBurgers, "Double Smash Burger", "Two smashed patties, double cheese, caramelized onions", 20900},
	{"aaaaaaaa-0001-0003-0000-000000000001", catBurgers, "Mushroom Swiss Burger", "Beef patty, sautéed mushrooms, Swiss cheese, garlic aioli", 18900},
	{"aaaaaaaa-0001-0004-0000-000000000001", catBurgers, "Spicy Crispy Burger", "Crispy fried chicken thigh, sriracha mayo, coleslaw, jalapeños", 17500},

	// Chicken
	{"aaaaaaaa-0002-0001-0000-000000000001", catChicken, "Fried Chicken Meal", "Golden fried chicken served with garlic rice and gravy", 14900},
	{"aaaaaaaa-0002-0002-0000-000000000001", catChicken, "Chicken Strips (3 pcs)", "Crispy tenders with choice of dipping sauce", 12900},
	{"aaaaaaaa-0002-0003-0000-000000000001", catChicken, "Chicken Wings (6 pcs)", "Buffalo or classic seasoning, fried to perfection", 19900},
	{"aaaaaaaa-0002-0004-0000-000000000001", catChicken, "Nashville Hot Chicken", "Extra crispy, coated in spicy Nashville sauce", 17900},

	// Sides & Snacks
	{"aaaaaaaa-0003-0001-0000-000000000001", catSides, "French Fries", "Golden shoestring fries, lightly seasoned", 7900},
	{"aaaaaaaa-0003-0002-0000-000000000001", catSides, "Onion Rings", "Beer-battered onion rings with ranch dip", 8900},
	{"aaaaaaaa-0003-0003-0000-000000000001", catSides, "Mozzarella Sticks (6 pcs)", "Crispy breaded mozzarella with marinara sauce", 10900},
	{"aaaaaaaa-0003-0004-0000-000000000001", catSides, "Loaded Fries", "Fries topped with cheese sauce, bacon bits, and jalapeños", 13900},

	// Drinks
	{"aaaaaaaa-0004-0001-0000-000000000001", catDrinks, "Iced Tea", "Freshly brewed, sweet or unsweetened", 4900},
	{"aaaaaaaa-0004-0002-0000-000000000001", catDrinks, "Soda", "Coke, Sprite, or Rootbeer — regular or diet", 3900},
	{"aaaaaaaa-0004-0003-0000-000000000001", catDrinks, "Fresh Lemonade", "House-squeezed lemonade with mint", 6900},
	{"aaaaaaaa-0004-0004-0000-000000000001", catDrinks, "Milkshake", "Chocolate, vanilla, or strawberry", 9900},

	// Desserts
	{"aaaaaaaa-0005-0001-0000-000000000001", catDesserts, "Chocolate Brownie", "Warm fudge brownie, served with vanilla ice cream", 8900},
	{"aaaaaaaa-0005-0002-0000-000000000001", catDesserts, "Ice Cream Sundae", "Three scoops with hot fudge and whipped cream", 7900},
	{"aaaaaaaa-0005-0003-0000-000000000001", catDesserts, "Apple Turnover", "Flaky pastry with cinnamon apple filling, dusted with powdered sugar", 5900},
}

type seedModifierOption struct {
	id         string
	name       string
	extraPrice int64 // centavos
	isDefault  bool
}

type seedModifierGroup struct {
	id           string
	name         string
	minSelection int
	maxSelection int
	options      []seedModifierOption
}

// Required groups (minSelection >= 1) always carry exactly one isDefault option, so the kiosk
// can pre-select a valid choice instead of forcing the customer to configure every product from
// a blank state.
var seedModifierGroups = []seedModifierGroup{
	{
		id: groupSize, name: "Size", minSelection: 1, maxSelection: 1,
		options: []seedModifierOption{
			{"bbbbbbbb-0001-0001-0000-000000000001", "Regular", 0, true},
			{"bbbbbbbb-0001-0002-0000-000000000001", "Large", 2500, false},
		},
	},
	{
		id: groupSpiceLevel, name: "Spice Level", minSelection: 1, maxSelection: 1,
		options: []seedModifierOption{
			{"bbbbbbbb-0002-0001-0000-000000000001", "Not Spicy", 0, true},
			{"bbbbbbbb-0002-0002-0000-000000000001", "Mild", 0, false},
			{"bbbbbbbb-0002-0003-0000-000000000001", "Spicy", 0, false},
			{"bbbbbbbb-0002-0004-0000-000000000001", "Extra Spicy", 0, false},
		},
	},
	{
		id: groupFriesFlavor, name: "Fries Flavor", minSelection: 1, maxSelection: 1,
		options: []seedModifierOption{
			{"bbbbbbbb-0003-0001-0000-000000000001", "Classic Salted", 0, true},
			{"bbbbbbbb-0003-0002-0000-000000000001", "Sour Cream", 1000, false},
			{"bbbbbbbb-0003-0003-0000-000000000001", "Cheese", 1500, false},
			{"bbbbbbbb-0003-0004-0000-000000000001", "BBQ", 1200, false},
		},
	},
	{
		id: groupSodaType, name: "Soda Type", minSelection: 1, maxSelection: 1,
		options: []seedModifierOption{
			{"bbbbbbbb-0004-0001-0000-000000000001", "Coke", 0, true},
			{"bbbbbbbb-0004-0002-0000-000000000001", "Sprite", 0, false},
			{"bbbbbbbb-0004-0003-0000-000000000001", "Rootbeer", 0, false},
		},
	},
	{
		id: groupSweetness, name: "Sweetness", minSelection: 1, maxSelection: 1,
		options: []seedModifierOption{
			{"bbbbbbbb-0005-0001-0000-000000000001", "Sweetened", 0, true},
			{"bbbbbbbb-0005-0002-0000-000000000001", "Unsweetened", 0, false},
		},
	},
	{
		id: groupMilkshakeFlavor, name: "Milkshake Flavor", minSelection: 1, maxSelection: 1,
		options: []seedModifierOption{
			{"bbbbbbbb-0006-0001-0000-000000000001", "Chocolate", 0, true},
			{"bbbbbbbb-0006-0002-0000-000000000001", "Vanilla", 0, false},
			{"bbbbbbbb-0006-0003-0000-000000000001", "Strawberry", 0, false},
		},
	},
	{
		// Optional add-on group — no default needed since min_selection is 0.
		id: groupBurgerAddOns, name: "Add-ons", minSelection: 0, maxSelection: 3,
		options: []seedModifierOption{
			{"bbbbbbbb-0007-0001-0000-000000000001", "Extra Cheese", 2000, false},
			{"bbbbbbbb-0007-0002-0000-000000000001", "Bacon", 3500, false},
			{"bbbbbbbb-0007-0003-0000-000000000001", "Extra Patty", 6000, false},
		},
	},
}

// seedProductModifierGroups associates each product with the modifier groups a customer must
// (or may) configure before adding it to the cart.
var seedProductModifierGroups = []struct{ productID, groupID string }{
	{prodClassicSmash, groupBurgerAddOns},
	{prodDoubleSmash, groupBurgerAddOns},
	{prodMushroomSwiss, groupBurgerAddOns},
	{prodSpicyCrispyBurger, groupSpiceLevel},
	{prodSpicyCrispyBurger, groupBurgerAddOns},

	{prodFriedChickenMeal, groupSpiceLevel},
	{prodChickenStrips, groupSpiceLevel},
	{prodChickenWings, groupSpiceLevel},
	{prodNashvilleHot, groupSpiceLevel},

	{prodFrenchFries, groupSize},
	{prodFrenchFries, groupFriesFlavor},
	{prodLoadedFries, groupSize},
	{prodLoadedFries, groupFriesFlavor},
	{prodOnionRings, groupSize},
	{prodMozzarellaSticks, groupSize},

	{prodIcedTea, groupSweetness},
	{prodIcedTea, groupSize},
	{prodSoda, groupSodaType},
	{prodSoda, groupSize},
	{prodLemonade, groupSize},
	{prodMilkshake, groupMilkshakeFlavor},
	{prodMilkshake, groupSize},
}

func hashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func seed(ctx context.Context, pool *pgxpool.Pool) error {
	// 1. Upsert store
	_, err := pool.Exec(ctx, `
		INSERT INTO stores (id, name, address, timezone, is_active, created_at, updated_at)
		VALUES ($1, 'NextGen Kitchen', 'SM Mall of Asia, Pasay City, Philippines', 'Asia/Manila', TRUE, NOW(), NOW())
		ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, address = EXCLUDED.address, updated_at = NOW()
	`, storeID)
	if err != nil {
		return fmt.Errorf("seed store: %w", err)
	}
	fmt.Println("  ✓ store")

	// 2. Upsert users
	users := []struct {
		id, username, password, role string
	}{
		{adminID, "admin", "admin123", "admin"},
		{cashierID, "kiosk", "kiosk123", "cashier"},
		{managerID, "manager", "manager123", "manager"},
		{kitchenID, "kitchen", "kitchen123", "kitchen"},
	}
	for _, u := range users {
		hash, err := hashPassword(u.password)
		if err != nil {
			return fmt.Errorf("hash password for %s: %w", u.username, err)
		}
		_, err = pool.Exec(ctx, `
			INSERT INTO users (id, store_id, username, password_hash, role, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, TRUE, NOW(), NOW())
			ON CONFLICT (username) DO UPDATE SET id = EXCLUDED.id, password_hash = EXCLUDED.password_hash, store_id = EXCLUDED.store_id, updated_at = NOW()
		`, u.id, storeID, u.username, hash, u.role)
		if err != nil {
			return fmt.Errorf("seed user %s: %w", u.username, err)
		}
	}
	fmt.Printf("  ✓ %d users\n", len(users))

	// 3. Upsert categories
	categories := []struct {
		id, name  string
		sortOrder int
	}{
		{catBurgers, "Burgers", 1},
		{catChicken, "Chicken", 2},
		{catSides, "Sides & Snacks", 3},
		{catDrinks, "Drinks", 4},
		{catDesserts, "Desserts", 5},
	}
	for _, c := range categories {
		_, err = pool.Exec(ctx, `
			INSERT INTO categories (id, store_id, name, sort_order, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, TRUE, NOW(), NOW())
			ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, sort_order = EXCLUDED.sort_order, updated_at = NOW()
		`, c.id, storeID, c.name, c.sortOrder)
		if err != nil {
			return fmt.Errorf("seed category %s: %w", c.name, err)
		}
	}
	fmt.Printf("  ✓ %d categories\n", len(categories))

	// 4. Upsert products
	for _, p := range seedProducts {
		_, err = pool.Exec(ctx, `
			INSERT INTO products (id, category_id, name, description, base_price, image_url, is_available, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, '', TRUE, NOW(), NOW())
			ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description, base_price = EXCLUDED.base_price, updated_at = NOW()
		`, p.id, p.categoryID, p.name, p.description, p.basePrice)
		if err != nil {
			return fmt.Errorf("seed product %s: %w", p.name, err)
		}
	}
	fmt.Printf("  ✓ %d products\n", len(seedProducts))

	// 5. Upsert modifier groups and their options
	optionCount := 0
	for _, g := range seedModifierGroups {
		_, err = pool.Exec(ctx, `
			INSERT INTO modifier_groups (id, name, min_selection, max_selection, created_at, updated_at)
			VALUES ($1, $2, $3, $4, NOW(), NOW())
			ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, min_selection = EXCLUDED.min_selection, max_selection = EXCLUDED.max_selection, updated_at = NOW()
		`, g.id, g.name, g.minSelection, g.maxSelection)
		if err != nil {
			return fmt.Errorf("seed modifier group %s: %w", g.name, err)
		}

		for _, o := range g.options {
			_, err = pool.Exec(ctx, `
				INSERT INTO modifier_options (id, modifier_group_id, name, extra_price, is_default, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
				ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, extra_price = EXCLUDED.extra_price, is_default = EXCLUDED.is_default, updated_at = NOW()
			`, o.id, g.id, o.name, o.extraPrice, o.isDefault)
			if err != nil {
				return fmt.Errorf("seed modifier option %s: %w", o.name, err)
			}
			optionCount++
		}
	}
	fmt.Printf("  ✓ %d modifier groups (%d options)\n", len(seedModifierGroups), optionCount)

	// 6. Attach modifier groups to products
	for _, pmg := range seedProductModifierGroups {
		_, err = pool.Exec(ctx, `
			INSERT INTO product_modifier_groups (product_id, modifier_group_id)
			VALUES ($1, $2)
			ON CONFLICT (product_id, modifier_group_id) DO NOTHING
		`, pmg.productID, pmg.groupID)
		if err != nil {
			return fmt.Errorf("attach modifier group %s to product %s: %w", pmg.groupID, pmg.productID, err)
		}
	}
	fmt.Printf("  ✓ %d product-modifier group associations\n", len(seedProductModifierGroups))

	return nil
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	if cfg.Database.URL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	pool, err := database.New(ctx, cfg.Database.URL, cfg.Database.MaxConns, cfg.Database.MinConns, logger)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer pool.Close()

	fmt.Println("Seeding mock data...")
	if err := seed(ctx, pool); err != nil {
		log.Fatalf("seed: %v", err)
	}

	fmt.Println()
	fmt.Println("Done. Credentials:")
	fmt.Println()
	fmt.Printf("  Store ID : %s\n", storeID)
	fmt.Println()
	fmt.Println("  Username   Password     Role")
	fmt.Println("  ---------  -----------  --------")
	fmt.Println("  admin      admin123     admin")
	fmt.Println("  kiosk      kiosk123     cashier")
	fmt.Println("  manager    manager123   manager")
	fmt.Println("  kitchen    kitchen123   kitchen")
	fmt.Println()
	fmt.Println("Kiosk .env (web/kiosk/.env):")
	fmt.Printf("  VITE_STORE_ID=%s\n", storeID)
	fmt.Println("  VITE_KIOSK_USERNAME=kiosk")
	fmt.Println("  VITE_KIOSK_PASSWORD=kiosk123")
}
