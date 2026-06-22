package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"order_processing/internal/auth"
	"order_processing/internal/catalog"
	"order_processing/internal/handler"
	"order_processing/internal/repository"
	"order_processing/internal/service"
	"order_processing/internal/worker"
)

func main() {
	addr := envOrDefault("PORT", "8080")
	interval := envDurationOrDefault("STATUS_UPDATE_INTERVAL", 5*time.Second)
	pendingDelay := envDurationOrDefault("PENDING_PROCESS_DELAY", 10*time.Second)

	ctx := context.Background()
	repo, err := newOrderRepository(ctx)
	if err != nil {
		log.Fatalf("repository init failed: %v", err)
	}
	svc := service.NewOrderService(repo, pendingDelay)
	orderHandler := handler.NewOrderHandler(svc)

	productRepo, err := newProductRepository(ctx)
	if err != nil {
		log.Fatalf("product repository init failed: %v", err)
	}
	if err := catalog.EnsureProducts(ctx, productRepo); err != nil {
		log.Fatalf("product seed failed: %v", err)
	}
	productSvc := service.NewProductService(productRepo)
	productHandler := handler.NewProductHandler(productSvc)

	authSvc, err := newAuthService(ctx)
	if err != nil {
		log.Fatalf("auth init failed: %v", err)
	}
	authHandler := handler.NewAuthHandler(authSvc)

	mux := http.NewServeMux()
	authHandler.RegisterRoutes(mux)
	orderHandler.RegisterRoutes(mux)
	productHandler.RegisterRoutes(mux)
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/index.html")
	})
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	statusUpdater := worker.NewStatusUpdater(svc, interval)
	go statusUpdater.Start(ctx)
	log.Printf("pending orders move to PROCESSING after %s (poll interval=%s)", pendingDelay, interval)

	server := &http.Server{
		Addr:              ":" + addr,
		Handler:           loggingMiddleware(auth.Middleware(authSvc, mux)),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("server listening on http://localhost:%s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func newAuthService(ctx context.Context) (*auth.Service, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-jwt-secret-change-me"
		log.Println("JWT_SECRET not set, using development default")
	}

	expiry := envDurationOrDefault("JWT_EXPIRY", 24*time.Hour)

	userRepo, err := newUserRepository(ctx)
	if err != nil {
		return nil, err
	}

	svc, err := auth.NewService(secret, userRepo, expiry)
	if err != nil {
		return nil, err
	}

	username := envOrDefault("AUTH_USERNAME", "admin")
	password := envOrDefault("AUTH_PASSWORD", "admin")
	if err := svc.EnsureUser(ctx, username, password); err != nil {
		return nil, err
	}

	return svc, nil
}

func newUserRepository(ctx context.Context) (repository.UserRepository, error) {
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Println("MONGODB_URI not set, using in-memory user repository")
		return repository.NewInMemoryUserRepository(), nil
	}

	database := envOrDefault("MONGODB_DATABASE", "order_processing")
	log.Printf("connecting user store to MongoDB at %s (database: %s)", mongoURI, database)
	return repository.NewMongoUserRepository(ctx, mongoURI, database)
}

func newProductRepository(ctx context.Context) (repository.ProductRepository, error) {
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Println("MONGODB_URI not set, using in-memory product repository")
		return repository.NewInMemoryProductRepository(), nil
	}

	database := envOrDefault("MONGODB_DATABASE", "order_processing")
	log.Printf("connecting product store to MongoDB at %s (database: %s)", mongoURI, database)
	return repository.NewMongoProductRepository(ctx, mongoURI, database)
}

func newOrderRepository(ctx context.Context) (repository.OrderRepository, error) {
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Println("MONGODB_URI not set, using in-memory repository")
		return repository.NewInMemoryOrderRepository(), nil
	}

	database := envOrDefault("MONGODB_DATABASE", "order_processing")
	log.Printf("connecting to MongoDB at %s (database: %s)", mongoURI, database)
	return repository.NewMongoOrderRepository(ctx, mongoURI, database)
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envDurationOrDefault(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		log.Printf("invalid %s=%q, using default %s", key, value, fallback)
		return fallback
	}
	return parsed
}
