package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/preuni/app/internal/storage"
	"github.com/preuni/app/internal/user/repository"
	apperrors "github.com/preuni/pkg/errors"
	pkgmw "github.com/preuni/pkg/middleware"
)

// AvatarUploadURLResponse is returned from PUT /students/me/avatar.
type AvatarUploadURLResponse struct {
	UploadURL string `json:"upload_url"`
	ObjectKey string `json:"object_key"`
}

// AvatarConfirmRequest is the body for POST /students/me/avatar/confirm.
type AvatarConfirmRequest struct {
	ObjectKey string `json:"object_key"`
}

// AvatarHandler handles avatar upload URL generation and confirmation.
type AvatarHandler struct {
	studentRepo *repository.StudentRepository
	storage     *storage.Client
}

// NewAvatarHandler constructs an AvatarHandler.
func NewAvatarHandler(studentRepo *repository.StudentRepository, storageClient *storage.Client) *AvatarHandler {
	return &AvatarHandler{studentRepo: studentRepo, storage: storageClient}
}

// ServeUpload handles PUT /students/me/avatar — returns a presigned PUT
// URL the client uploads the image bytes to directly.
func (h *AvatarHandler) ServeUpload(w http.ResponseWriter, r *http.Request) {
	userID := pkgmw.UserIDFromContext(r.Context())

	objectKey := fmt.Sprintf("avatars/%s/%d.webp", userID, time.Now().UnixMilli())

	uploadURL, err := h.storage.PresignedPutURL(r.Context(), objectKey)
	if err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	pkgmw.JSON(w, http.StatusOK, AvatarUploadURLResponse{
		UploadURL: uploadURL,
		ObjectKey: objectKey,
	})
}

// ServeConfirm handles POST /students/me/avatar/confirm — updates
// avatar_url after the client has successfully PUT the object.
func (h *AvatarHandler) ServeConfirm(w http.ResponseWriter, r *http.Request) {
	var req AvatarConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ObjectKey == "" {
		pkgmw.ErrorResponse(w, apperrors.Validation("object_key", "object_key is required"))
		return
	}

	userID := pkgmw.UserIDFromContext(r.Context())

	avatarURL := h.storage.PublicURL(req.ObjectKey)

	patch := &repository.StudentPatch{AvatarURL: &avatarURL}
	updated, err := h.studentRepo.Update(r.Context(), userID, patch)
	if err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}
	pkgmw.JSON(w, http.StatusOK, toStudentResponse(updated))
}
