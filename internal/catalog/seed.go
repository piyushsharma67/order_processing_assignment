package catalog

import (
	"context"
	"log"

	"order_processing/internal/domain"
	"order_processing/internal/repository"
)

func EnsureProducts(ctx context.Context, repo repository.ProductRepository) error {
	count, err := repo.Count(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		log.Printf("product catalog already seeded (%d products)", count)
		return nil
	}

	products := SeedProducts()
	if err := repo.CreateMany(ctx, products); err != nil {
		return err
	}

	log.Printf("seeded %d products into catalog", len(products))
	return nil
}

func SeedProducts() []domain.Product {
	return []domain.Product{
		{
			ID:          "B08N5WRWNW",
			Name:        "Echo Dot (5th Gen) Smart Speaker",
			Description: "Compact smart speaker with Alexa, improved audio, and temperature sensor.",
			Category:    "Electronics",
			Price:       49.99,
			ImageURL:    "https://m.media-amazon.com/images/I/714Rq4k05UL._AC_SL1000_.jpg",
			Stock:       120,
		},
		{
			ID:          "B0BSHF7WHW",
			Name:        "Kindle Paperwhite (16 GB)",
			Description: "6.8\" display, adjustable warm light, weeks of battery life, waterproof.",
			Category:    "Electronics",
			Price:       139.99,
			ImageURL:    "https://m.media-amazon.com/images/I/61j555sNaaL._AC_SL1500_.jpg",
			Stock:       85,
		},
		{
			ID:          "B09B8V1LZ3",
			Name:        "Fire TV Stick 4K Max",
			Description: "Streaming device with Wi-Fi 6, Alexa Voice Remote, and support for 4K HDR.",
			Category:    "Electronics",
			Price:       54.99,
			ImageURL:    "https://m.media-amazon.com/images/I/51TjJOTfslL._AC_SL1000_.jpg",
			Stock:       200,
		},
		{
			ID:          "B08C1W5N87",
			Name:        "Anker PowerCore 10000 Portable Charger",
			Description: "Ultra-compact 10000mAh battery pack with high-speed PowerIQ charging.",
			Category:    "Electronics",
			Price:       21.99,
			ImageURL:    "https://m.media-amazon.com/images/I/61+v+0QKzGL._AC_SL1500_.jpg",
			Stock:       350,
		},
		{
			ID:          "B07FZ8S74R",
			Name:        "Logitech MX Master 3S Wireless Mouse",
			Description: "Ergonomic wireless mouse with 8000 DPI sensor and quiet clicks.",
			Category:    "Computers",
			Price:       99.99,
			ImageURL:    "https://m.media-amazon.com/images/I/61ni3t1krXL._AC_SL1500_.jpg",
			Stock:       60,
		},
		{
			ID:          "B08KTZ8249",
			Name:        "Samsung 970 EVO Plus 1TB NVMe SSD",
			Description: "Internal solid state drive with V-NAND technology for fast read/write speeds.",
			Category:    "Computers",
			Price:       89.99,
			ImageURL:    "https://m.media-amazon.com/images/I/71gS8V+YJEL._AC_SL1500_.jpg",
			Stock:       45,
		},
		{
			ID:          "B07G4LN56J",
			Name:        "Instant Pot Duo 7-in-1",
			Description: "7-in-1 electric pressure cooker: pressure cook, slow cook, rice, steam, and more.",
			Category:    "Home & Kitchen",
			Price:       79.95,
			ImageURL:    "https://m.media-amazon.com/images/I/71aFt4+OTpL._AC_SL1500_.jpg",
			Stock:       90,
		},
		{
			ID:          "B07DGR98VQ",
			Name:        "Ninja AF101 Air Fryer",
			Description: "4-quart air fryer with one-touch programs for fries, chicken, and veggies.",
			Category:    "Home & Kitchen",
			Price:       89.99,
			ImageURL:    "https://m.media-amazon.com/images/I/71bLT8j+8VL._AC_SL1500_.jpg",
			Stock:       75,
		},
		{
			ID:          "B00TTD9BRC",
			Name:        "Keurig K-Classic Coffee Maker",
			Description: "Single serve K-Cup pod coffee brewer with 48 oz water reservoir.",
			Category:    "Home & Kitchen",
			Price:       89.99,
			ImageURL:    "https://m.media-amazon.com/images/I/71k3V2X-TyL._AC_SL1500_.jpg",
			Stock:       55,
		},
		{
			ID:          "B07D6C9M6R",
			Name:        "Dyson V8 Cordless Vacuum",
			Description: "Lightweight cordless stick vacuum with up to 40 minutes of fade-free suction.",
			Category:    "Home & Kitchen",
			Price:       349.99,
			ImageURL:    "https://m.media-amazon.com/images/I/61fF2X+0QGL._AC_SL1500_.jpg",
			Stock:       30,
		},
		{
			ID:          "B07H65KP63",
			Name:        "Apple AirPods (2nd Generation)",
			Description: "Wireless earbuds with charging case and hands-free Siri access.",
			Category:    "Electronics",
			Price:       129.00,
			ImageURL:    "https://m.media-amazon.com/images/I/61SUj2aKoEL._AC_SL1500_.jpg",
			Stock:       150,
		},
		{
			ID:          "B08G9PRS1K",
			Name:        "Sony WH-1000XM4 Headphones",
			Description: "Industry-leading noise canceling over-ear headphones with 30-hour battery.",
			Category:    "Electronics",
			Price:       278.00,
			ImageURL:    "https://m.media-amazon.com/images/I/71o8QHeX-yL._AC_SL1500_.jpg",
			Stock:       40,
		},
		{
			ID:          "B07PXGQC1Q",
			Name:        "Apple Watch Series 9 (GPS, 41mm)",
			Description: "Advanced health features, bright Always-On Retina display, S9 chip.",
			Category:    "Electronics",
			Price:       399.00,
			ImageURL:    "https://m.media-amazon.com/images/I/61cTfzS2s1L._AC_SL1500_.jpg",
			Stock:       25,
		},
		{
			ID:          "B08N36XNTT",
			Name:        "Atomic Habits by James Clear",
			Description: "Bestselling guide to building good habits and breaking bad ones.",
			Category:    "Books",
			Price:       13.49,
			ImageURL:    "https://m.media-amazon.com/images/I/81wgvm5S9DL._AC_SL1500_.jpg",
			Stock:       500,
		},
		{
			ID:          "B0030LY7VG",
			Name:        "The Lean Startup by Eric Ries",
			Description: "How today's entrepreneurs use continuous innovation to build successful businesses.",
			Category:    "Books",
			Price:       14.99,
			ImageURL:    "https://m.media-amazon.com/images/I/81-QB7IUNzL._AC_SL1500_.jpg",
			Stock:       320,
		},
		{
			ID:          "B00I8BICB2",
			Name:        "CeraVe Moisturizing Cream",
			Description: "Daily face and body moisturizer with hyaluronic acid and ceramides.",
			Category:    "Beauty",
			Price:       16.08,
			ImageURL:    "https://m.media-amazon.com/images/I/81U5Rj+5KGL._AC_SL1500_.jpg",
			Stock:       280,
		},
		{
			ID:          "B00U2VQZOS",
			Name:        "Neutrogena Hydro Boost Water Gel",
			Description: "Lightweight gel moisturizer with hyaluronic acid for dry skin.",
			Category:    "Beauty",
			Price:       18.97,
			ImageURL:    "https://m.media-amazon.com/images/I/71X3F8B5iVL._AC_SL1500_.jpg",
			Stock:       210,
		},
		{
			ID:          "B07PFFMP9P",
			Name:        "LEGO Classic Creative Bricks Set",
			Description: "1500-piece LEGO set with bricks, wheels, windows, and eyes for open-ended play.",
			Category:    "Toys",
			Price:       59.99,
			ImageURL:    "https://m.media-amazon.com/images/I/81O2+6X+4dL._AC_SL1500_.jpg",
			Stock:       95,
		},
		{
			ID:          "B07H8QMZPV",
			Name:        "Nintendo Switch with Neon Joy-Con",
			Description: "Hybrid gaming console for TV and handheld play with two Joy-Con controllers.",
			Category:    "Video Games",
			Price:       299.99,
			ImageURL:    "https://m.media-amazon.com/images/I/61-PblYntsL._AC_SL1500_.jpg",
			Stock:       35,
		},
		{
			ID:          "B08F7PTF53",
			Name:        "Amazon Basics Microfiber Sheet Set",
			Description: "Queen-size 4-piece bed sheet set, soft and wrinkle-resistant microfiber.",
			Category:    "Home & Kitchen",
			Price:       24.99,
			ImageURL:    "https://m.media-amazon.com/images/I/81s0rA0f3XL._AC_SL1500_.jpg",
			Stock:       400,
		},
	}
}
