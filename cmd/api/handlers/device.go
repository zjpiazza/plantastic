package handlers

import (
	// "encoding/json" // No longer needed for AuthenticateDevice if using gin.BindJSON
	"net/http"
	// Keep for Clerk initialization if still needed here, though it's often global
	// "github.com/clerk/clerk-sdk-go/v2" // Clerk client instance not directly used in this version
	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/gin-gonic/gin"
	"github.com/zjpiazza/plantastic/internal/device"

	// "github.com/zjpiazza/plantastic/internal/models" // Not directly used in handlers
	"log"
)

// DeviceHandler handles device code authentication
type DeviceHandler struct {
	deviceManager *device.Manager
	// clerk         *clerk.Client // Clerk client not directly used now in handlers for these flows
}

// NewDeviceHandler creates a new device handler
func NewDeviceHandler(deviceManager *device.Manager) (*DeviceHandler, error) {
	// Clerk key is set globally in main.go, so explicit client init might not be needed here
	// unless specific client features are used. For jwt.Verify, global SetKey is sufficient.
	// clerkClient, err := clerk.NewClient(os.Getenv("CLERK_SECRET_KEY"))
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to create clerk client: %w", err)
	// }
	return &DeviceHandler{
		deviceManager: deviceManager,
		// clerk: clerkClient,
	}, nil
}

type GenerateCodeRequest struct {
	DeviceID string `json:"device_id" binding:"required"`
}

// GenerateCode generates a new user code for a given DeviceID
func (h *DeviceHandler) GenerateCode(c *gin.Context) {
	var req GenerateCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	userCode, err := h.deviceManager.NewCode(req.DeviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate code: " + err.Error()})
		return
	}
	// Consider also returning the verification URI base path if it's fixed or configurable
	c.JSON(http.StatusOK, gin.H{
		"user_code":        userCode,
		"verification_uri": "/verify-device", // Example, TUI would append this to base URL
		"expires_in":       int(device.UserCodeDuration.Seconds()),
		"interval":         5, // Suggested polling interval in seconds
	})
}

// CheckStatus checks the status of a device code (user_code)
func (h *DeviceHandler) CheckStatus(c *gin.Context) {
	userCode := c.Query("code") // Changed to query parameter for GET
	if userCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_code query parameter is required"})
		return
	}

	log.Printf("[CheckStatus] Called for user_code: %s", userCode) // Log entry

	token, err := h.deviceManager.GetTokenForUserCode(userCode)
	if err != nil {
		log.Printf("[CheckStatus] Error from GetTokenForUserCode for %s: %v", userCode, err) // Log error
		if err.Error() == "token not yet available for this user code" {
			c.JSON(http.StatusOK, gin.H{"status": "pending_activation"})
			return
		}
		if err.Error() == "invalid user code" || err.Error() == "user code expired and was not activated" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error(), "status": "failed"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking status: " + err.Error(), "status": "failed"})
		return
	}

	log.Printf("[CheckStatus] Token retrieved for %s: %s", userCode, token) // Log retrieved token

	// Token found, now verify it with Clerk (as before)
	claims, err := jwt.Verify(c.Request.Context(), &jwt.VerifyParams{Token: token})
	if err != nil {
		log.Printf("[CheckStatus] jwt.Verify failed for token of user_code %s. Token: %s, Error: %v", userCode, token, err) // Log verification error
		// This case implies the token we stored from Clerk is somehow invalid, which is concerning.
		// Or the Clerk keys changed, or token is malformed.
		h.deviceManager.SetTokenForUserCode(userCode, "") // Clear the invalid token? Or handle differently.
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Stored token is invalid: " + err.Error(), "status": "failed"})
		return
	}

	log.Printf("[CheckStatus] jwt.Verify successful for user_code %s. Claims Subject: %s", userCode, claims.Subject) // Log success

	c.JSON(http.StatusOK, gin.H{
		"status":     "activated",
		"token":      token,
		"token_type": "Bearer", // Standard OAuth2 practice
	})
}

type AuthenticateDeviceRequest struct {
	UserCode string `json:"user_code" binding:"required"`
	Token    string `json:"clerk_session_token" binding:"required"` // This is the JWT from Clerk after web auth
}

// AuthenticateDevice is called by the web flow after user authenticates with Clerk.
// It links the Clerk session token with the UserCode.
func (h *DeviceHandler) AuthenticateDevice(c *gin.Context) {
	var req AuthenticateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// First, verify the provided Clerk session token to ensure it's valid
	// This step is crucial: the web frontend is claiming "this user is authenticated by Clerk,
	// and here's their token, and they want to activate this user_code".
	// We must verify the token before trusting it.
	_, err := jwt.Verify(c.Request.Context(), &jwt.VerifyParams{Token: req.Token})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Clerk session token: " + err.Error()})
		return
	}

	// Now, try to associate this verified Clerk token with the user_code
	err = h.deviceManager.SetTokenForUserCode(req.UserCode, req.Token)
	if err != nil {
		// Handle specific errors from SetTokenForUserCode
		if err.Error() == "invalid user code" || err.Error() == "user code expired" || err.Error() == "user code already activated" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot link device: " + err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link device: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Device linking initiated successfully. TUI can now poll for the token."})
}
