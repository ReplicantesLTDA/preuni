package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	apperrors "github.com/preuni/pkg/errors"
	pkgmw "github.com/preuni/pkg/middleware"
	"github.com/preuni/svc/user/repository"
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
	s3Bucket    string
	s3Region    string
}

// NewAvatarHandler constructs an AvatarHandler.
func NewAvatarHandler(studentRepo *repository.StudentRepository, s3Bucket, s3Region string) *AvatarHandler {
	return &AvatarHandler{studentRepo: studentRepo, s3Bucket: s3Bucket, s3Region: s3Region}
}

// ServeUpload handles PUT /students/me/avatar — returns a presigned S3 PUT URL.
func (h *AvatarHandler) ServeUpload(w http.ResponseWriter, r *http.Request) {
	userID := pkgmw.UserIDFromContext(r.Context())

	// Build the object key: avatars/{user_id}/{timestamp}.webp
	objectKey := fmt.Sprintf("avatars/%s/%d.webp", userID, time.Now().UnixMilli())

	// In production this would use AWS SDK to generate a presigned URL.
	// For v1 we return a stub that documents the expected format.
	uploadURL := fmt.Sprintf(
		"https://%s.s3.%s.amazonaws.com/%s?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=PRESIGNED",
		h.s3Bucket, h.s3Region, objectKey,
	)

	pkgmw.JSON(w, http.StatusOK, AvatarUploadURLResponse{
		UploadURL: uploadURL,
		ObjectKey: objectKey,
	})
}

// ServeConfirm handles POST /students/me/avatar/confirm — updates avatar_url after successful S3 upload.
func (h *AvatarHandler) ServeConfirm(w http.ResponseWriter, r *http.Request) {
	var req AvatarConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ObjectKey == "" {
		pkgmw.ErrorResponse(w, apperrors.Validation("object_key", "object_key is required"))
		return
	}

	userID := pkgmw.UserIDFromContext(r.Context())

	avatarURL := fmt.Sprintf(
		"https://%s.s3.%s.amazonaws.com/%s",
		h.s3Bucket, h.s3Region, req.ObjectKey,
	)

	patch := &repository.StudentPatch{AvatarURL: &avatarURL}
	updated, err := h.studentRepo.Update(r.Context(), userID, patch)
	if err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}
	pkgmw.JSON(w, http.StatusOK, toStudentResponse(updated))
}
