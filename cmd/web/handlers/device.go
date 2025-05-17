package handlers

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/zjpiazza/plantastic/internal/device"
)

type DeviceHandler struct {
	manager   *device.Manager
	templates *template.Template
}

type PageData struct {
	ClerkPublishableKey string
	ClerkFrontendAPI    string
	Code                string
}

func NewDeviceHandler(manager *device.Manager, templates *template.Template) *DeviceHandler {
	return &DeviceHandler{
		manager:   manager,
		templates: templates,
	}
}

func (h *DeviceHandler) HandleHome(w http.ResponseWriter, r *http.Request) {
	h.templates.ExecuteTemplate(w, "home.html", nil)
}

func (h *DeviceHandler) HandleDeviceLink(w http.ResponseWriter, r *http.Request) {
	h.templates.ExecuteTemplate(w, "link.html", nil)
}

func (h *DeviceHandler) HandleDeviceLinkCode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	code := vars["code"]

	_, err := h.manager.ValidateUserCode(code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	clerkPublishableKey := os.Getenv("CLERK_PUBLISHABLE_KEY")
	if clerkPublishableKey == "" {
		log.Println("Warning: CLERK_PUBLISHABLE_KEY environment variable not set for web handler.")
		// Optionally, handle this more gracefully, e.g., render an error page or don't init ClerkJS
	}

	log.Println("Clerk Publishable Key:", clerkPublishableKey)

	data := PageData{
		ClerkPublishableKey: clerkPublishableKey,
		ClerkFrontendAPI:    "relative-seasnail-33.clerk.accounts.dev",
		Code:                code,
	}
	h.templates.ExecuteTemplate(w, "link_code.html", data)
}
