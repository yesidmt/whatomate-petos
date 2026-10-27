package handlers

import (
	"strings"

	"github.com/shridarpatil/whatomate/internal/models"
	"github.com/valyala/fasthttp"
	"github.com/zerodha/fastglue"
)

func isWAHAAccount(account *models.WhatsAppAccount) bool {
	return account != nil && strings.HasPrefix(account.PhoneID, "waha:")
}

func accountProvider(account *models.WhatsAppAccount) string {
	if isWAHAAccount(account) {
		return "waha"
	}
	return "meta"
}

func (a *App) WAHAStatus(r *fastglue.Request) error {
	if _, _, err := a.requireAuth(r, models.ResourceAccounts, models.ActionRead); err != nil {
		return nil
	}
	if a.WAHA == nil {
		return r.SendEnvelope(map[string]any{"configured": false, "connected": false})
	}
	err := a.WAHA.VerifySession(r.RequestCtx)
	return r.SendEnvelope(map[string]any{
		"configured": true,
		"connected":  err == nil,
		"phone":      a.WAHA.ExpectedPhone,
	})
}

// CreateWAHAAccount attaches the one configured temporary session to the
// organization. The sender is confirmed by WAHA before creating any record.
// No API key or linked-device credentials are stored in PostgreSQL.
func (a *App) CreateWAHAAccount(r *fastglue.Request) error {
	orgID, userID, err := a.requireAuth(r, models.ResourceAccounts, models.ActionWrite)
	if err != nil {
		return nil
	}
	if a.WAHA == nil {
		return r.SendErrorEnvelope(fasthttp.StatusServiceUnavailable, "Temporary WAHA transport is not configured", nil, "")
	}
	var req struct {
		Name              string `json:"name"`
		IsDefaultOutgoing bool   `json:"is_default_outgoing"`
	}
	if err := a.decodeRequest(r, &req); err != nil {
		return nil
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || len(req.Name) > 100 {
		return r.SendErrorEnvelope(fasthttp.StatusBadRequest, "A name of up to 100 characters is required", nil, "")
	}
	if err := a.WAHA.VerifySession(r.RequestCtx); err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusServiceUnavailable, "WAHA session is not connected to the approved phone", nil, "")
	}
	account := models.WhatsAppAccount{
		OrganizationID:    orgID,
		Name:              req.Name,
		PhoneID:           "waha:" + a.WAHA.Session,
		BusinessID:        "waha",
		AccessToken:       "waha-managed", // schema requires a value; not an API credential
		APIVersion:        "waha",
		IsDefaultOutgoing: req.IsDefaultOutgoing,
		Status:            "active",
		CreatedByID:       &userID,
		UpdatedByID:       &userID,
	}
	if err := a.encryptAccountSecrets(&account); err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusInternalServerError, "Failed to create account", nil, "")
	}
	tx := a.DB.Begin()
	if tx.Error != nil {
		return r.SendErrorEnvelope(fasthttp.StatusInternalServerError, "Failed to create account", nil, "")
	}
	if req.IsDefaultOutgoing {
		if err := tx.Model(&models.WhatsAppAccount{}).Where("organization_id = ? AND is_default_outgoing = ?", orgID, true).Update("is_default_outgoing", false).Error; err != nil {
			tx.Rollback()
			return r.SendErrorEnvelope(fasthttp.StatusInternalServerError, "Failed to set default account", nil, "")
		}
	}
	if err := tx.Create(&account).Error; err != nil {
		tx.Rollback()
		return r.SendErrorEnvelope(fasthttp.StatusConflict, "Account name already exists or could not be created", nil, "")
	}
	if err := tx.Commit().Error; err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusInternalServerError, "Failed to create account", nil, "")
	}
	a.logAudit(orgID, userID, "account", account.ID, models.AuditActionCreated, nil, &account)
	return r.SendEnvelope(accountToResponse(account))
}
