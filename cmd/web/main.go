package main

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/zjpiazza/plantastic/cmd/web/handlers"
	"github.com/zjpiazza/plantastic/internal/device"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize Clerk client
	clerkKey := os.Getenv("CLERK_SECRET_KEY")
	if clerkKey == "" {
		log.Fatal("CLERK_SECRET_KEY environment variable not set.")
	}
	clerk.SetKey(clerkKey)

	// Initialize database connection
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=plantastic port=5432 sslmode=disable TimeZone=UTC"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn), // Or logger.Silent
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("Web server successfully connected to database")

	// Initialize device manager
	deviceManager := device.NewManager(db)

	// Load templates
	templates := template.Must(template.ParseGlob("cmd/web/templates/*.html"))

	// Initialize handlers
	deviceHandler := handlers.NewDeviceHandler(deviceManager, templates)

	// Set up router
	r := mux.NewRouter()

	// Static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Public routes
	r.HandleFunc("/", deviceHandler.HandleHome)

	// Protected routes (require Clerk authentication)
	protected := r.NewRoute().Subrouter()
	protected.Use(func(next http.Handler) http.Handler {
		return clerkhttp.WithHeaderAuthorization()(next)
	})
	protected.HandleFunc("/link", deviceHandler.HandleDeviceLink)
	protected.HandleFunc("/link/{code}", deviceHandler.HandleDeviceLinkCode)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting web server on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
