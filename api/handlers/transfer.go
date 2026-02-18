//
// COPYRIGHT OpenDI
//

package handlers

import (
	"net/http"
	"opendi/model-hub/api/apiTypes"
	"opendi/model-hub/api/database"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GetTransfer godoc
// @Summary      Get ownership transfer request for model
// @Description  Get ownership transfer request for model
// @Tags         models
// @Produce      json
// @Param        tag  path  string  true  "Model tag"
// @Success      200   {object}  gin.H  "Transfer request details"
// @Failure      401   {object}  gin.H  "Unauthorized"
// @Failure      403   {object}  gin.H  "Forbidden"
// @Failure      404   {object}  gin.H  "Transfer request not found"
// @Router       /v0/models/transfer/{tag} [get]
func (h *ModelHandler) GetTransfer(c *gin.Context) {
	tag := c.Param("tag")

	// make sure the user has authorization
	actingUserID, err := getUserIDFromToken(c, h.engMode)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// get the transfer
	model, err := database.GetModelByTag(tag)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	transfer, err := database.GetTransferByModelUUID(model.Meta.UUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// make sure the user can do this action
	if transfer.FromUserID != actingUserID && transfer.ToUserID != actingUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid permissions for this action"})
		return
	}

	c.JSON(http.StatusOK, transfer)
}

// GetUserPendingTransfers godoc
// @Summary      Get pending transfers for the authenticated user
// @Description  Get list of models waiting to be transferred to the current user
// @Tags         user
// @Produce      json
// @Success      200  {array}   gin.H
// @Failure      401  {object}  gin.H  "Unauthorized"
// @Failure      500  {object}  gin.H  "Internal server error"
// @Router       /v0/user/transfers [get]
func (h *ModelHandler) GetUserPendingTransfers(c *gin.Context) {
	actingUserID, err := getUserIDFromToken(c, h.engMode)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	transfers, err := database.GetTransfersByTargetUserID(actingUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transfers)
}

// PostTransfer godoc
// @Summary      Create ownership transfer request for model
// @Description  Create ownership transfer request for model
// @Tags         models
// @Produce      json
// @Param        tag  path  string  true  "Model tag"
// @Param        transfer  body  object  true  "Transfer request details"
// @Success      200   {object}  gin.H  "Transfer request created"
// @Failure      400   {object}  gin.H  "Bad request"
// @Failure      403   {object}  gin.H  "Forbidden"
// @Failure      404   {object}  gin.H  "Model not found"
// @Failure      500   {object}  gin.H  "Internal server error"
// @Router       /v0/models/transfer/{tag} [post]
func (h *ModelHandler) PostTransfer(c *gin.Context) {
	tag := c.Param("tag")
	owner := c.Query("owner")

	// make sure the user has authorization
	actingUserID, err := getUserIDFromToken(c, h.engMode)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// make sure the owner ID is valid
	if owner == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "owner parameter is required"})
		return
	}
	toUser, err := database.GetUserByEmail(owner)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid new owner email"})
		return
	}

	// make sure the model exists
	model, err := database.GetModelByTag(tag)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// make sure the requesting user owns the model
	if model.Addons.OwnerID != actingUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid permissions for this action"})
		return
	}

	// make sure an existing transfer does not exist
	existingTransfer, err := database.GetTransferByModelUUID(model.Meta.UUID)
	if err == nil && existingTransfer != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "transfer request already exists for this model"})
		return
	}

	// create the transfer
	transfer := &apiTypes.Transfer{
		CDMUUID:    model.Meta.UUID,
		ToUserID:   toUser.ID,
		FromUserID: actingUserID,
		CreatedAt:  time.Now(),
	}
	if err := database.CreateTransfer(transfer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transfer)
}

// DeleteTransfer godoc
// @Summary      Accept/decline ownership transfer request for model
// @Description  Accept/decline ownership transfer request for model
// @Tags         models
// @Produce      json
// @Param        tag  path  string  true  "Model tag"
// @Success      200   {object}  gin.H  "Transfer request processed"
// @Failure      400   {object}  gin.H  "Bad request"
// @Failure      403   {object}  gin.H  "Forbidden"
// @Failure      404   {object}  gin.H  "Transfer request not found"
// @Failure      500   {object}  gin.H  "Internal server error"
// @Router       /v0/models/transfer/{tag} [delete]
func (h *ModelHandler) DeleteTransfer(c *gin.Context) {
	tag := c.Param("tag")
	accept := c.Query("accept")

	// make sure the user has authorization
	actingUserID, err := getUserIDFromToken(c, h.engMode)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// make sure the accept parameter is valid
	if accept == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "accept parameter is required"})
		return
	}
	acceptBool, err := strconv.ParseBool(accept)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "accept parameter must be true or false"})
		return
	}

	// make sure the transfer exists
	model, err := database.GetModelByTag(tag)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	transfer, err := database.GetTransferByModelUUID(model.Meta.UUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// make sure the acting user can actually accept this transfer
	if transfer.ToUserID != actingUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid permissions for this action"})
		return
	}

	// delete the transfer
	if err := database.DeleteTransfer(transfer, acceptBool); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}
