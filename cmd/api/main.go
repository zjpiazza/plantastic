// cmd/api/main.go
package main

import (
	// Keep for future use, though not directly by Clerk in this pattern
	"fmt"
	"log"
	"net/http"
	"os"

	// "time" // Not explicitly needed by Clerk in this pattern, unless for WithLeeway if we re-add it

	"github.com/clerk/clerk-sdk-go/v2"                // Main Clerk package
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http" // For WithHeaderAuthorization

	"github.com/gin-contrib/cors" // Import CORS middleware
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/zjpiazza/plantastic/cmd/api/handlers"
	"github.com/zjpiazza/plantastic/cmd/api/internal/routes"
	"github.com/zjpiazza/plantastic/cmd/api/internal/storage"
	"github.com/zjpiazza/plantastic/internal/device"
	"github.com/zjpiazza/plantastic/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const clerkSessionKey = "clerk_session_claims" // For storing claims in Gin's context

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: Error loading .env file, proceeding with environment variables if set.")
	}

	// Initialize Clerk client globally (as per your example)
	clerkKey := os.Getenv("CLERK_SECRET_KEY")
	if clerkKey == "" {
		log.Fatal("CLERK_SECRET_KEY environment variable not set.")
	}
	clerk.SetKey(clerkKey) // Set the key globally

	// Initialize database connection
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=plantastic port=5432 sslmode=disable TimeZone=UTC"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	fmt.Println("Database connection successful.")

	// AutoMigrate
	db.AutoMigrate(&models.Garden{}, &models.Bed{}, &models.Task{}, &models.Device{})
	fmt.Println("Database migration complete")

	// Create storage instances
	gardenStore := storage.NewGormGardenStore(db)
	bedStore := storage.NewGormBedStore(db)
	taskStore := storage.NewGormTaskStore(db)

	// Initialize device manager
	deviceManager := device.NewManager(db)

	// Initialize handlers
	deviceApiHandler, err := handlers.NewDeviceHandler(deviceManager)
	if err != nil {
		log.Fatal("Failed to initialize API device handler:", err)
	}

	// Initialize Gin router
	router := gin.Default()

	// Apply CORS middleware
	// This allows http://localhost:8080, specified methods, and specified headers.
	cDefault := cors.DefaultConfig()
	cDefault.AllowOrigins = []string{"http://localhost:8080"} // Your web app's origin
	cDefault.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	cDefault.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	// Allow credentials (cookies, authorization headers, etc.)
	cDefault.AllowCredentials = true
	router.Use(cors.New(cDefault))

	// Public Device Routes
	router.POST("/device/request-code", deviceApiHandler.GenerateCode)
	router.GET("/device/check-status", deviceApiHandler.CheckStatus)

	// Public routes (example)
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Welcome to Plantastic API!"})
	})

	// Protected route group
	protected := router.Group("/")
	// ClerkMiddleware might not need the client passed if SetKey is global
	protected.Use(ClerkMiddleware())

	// Initialize routes
	routes.SetupProtectedRoutes(protected, gardenStore, bedStore, taskStore, deviceApiHandler)

	// Start server
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8000"
	}
	fmt.Printf("Starting server on :%s\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// ClerkMiddleware creates a Gin middleware for Clerk authentication
func ClerkMiddleware() gin.HandlerFunc { // Removed clerkClient from params
	// This returns a function: func(next http.Handler) http.Handler
	clerkMiddlewareWrapper := clerkhttp.WithHeaderAuthorization()

	return func(c *gin.Context) {
		nextHttpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// SessionClaimsFromContext is from the main clerk package, as per your example
			claims, ok := clerk.SessionClaimsFromContext(r.Context())
			if !ok {
				log.Println("Clerk middleware: SessionClaimsFromContext found no claims or 'ok' is false.")
				// Since we are in an http.HandlerFunc, we write directly to the ResponseWriter
				// Gin's context `c` is available via closure, but c.AbortWithStatusJSON might not be suitable here.
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized) // Use the http.ResponseWriter from this scope
				fmt.Fprintln(w, `{"error":"Unauthorized: No valid session claims found in context"}`)
				return
			}

			log.Printf("Clerk middleware: Successfully authenticated session for user %s, session %s\n", claims.Subject, claims.SessionID)

			// Store claims in Gin's context for easier access in Gin handlers
			c.Set(clerkSessionKey, claims)

			// Ensure Gin uses the request that Clerk's middleware might have modified (e.g., context values)
			c.Request = r
			c.Next() // Tell Gin to proceed to the next handler in its chain
		})

		// Apply the clerkMiddlewareWrapper to our nextHttpHandler
		wrappedHandler := clerkMiddlewareWrapper(nextHttpHandler)

		// Serve the request through the wrapped handler
		wrappedHandler.ServeHTTP(c.Writer, c.Request)

		// If the response was already written by Clerk or our nextHttpHandler (due to error),
		// we should abort Gin's chain to prevent further processing (like Gin writing its own 404).
		if c.Writer.Written() {
			c.Abort()
		}
		// If c.Next() was called in nextHttpHandler, Gin will continue.
		// If an error was written and returned from nextHttpHandler, c.Next() wasn't called,
		// and c.Abort() here ensures no further Gin handlers run.
	}
}
