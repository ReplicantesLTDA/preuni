package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/preuni/app/internal/storage"
	"github.com/preuni/app/internal/user/handler"
	"github.com/preuni/app/internal/user/handler/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAvatarUpload_StorageErrorReturns500 covers ServeUpload's error branch
// (previously untested): when presigning fails, the handler must surface a
// 500, not panic or return a malformed response. No DB or live MinIO server
// needed -- an invalid bucket name makes minio-go reject the request
// locally, before any network call.
func TestAvatarUpload_StorageErrorReturns500(t *testing.T) {
	storageClient, err := storage.New(storage.Config{
		Endpoint:  "localhost:9000",
		AccessKey: "test",
		SecretKey: "test",
		Bucket:    "Invalid_Bucket_Name!", // rejected locally by S3 bucket-naming rules
	})
	require.NoError(t, err)

	avatarH := handler.NewAvatarHandler(nil, storageClient)
	h := withAuth(http.HandlerFunc(avatarH.ServeUpload))

	req := httptest.NewRequest(http.MethodPut, "/v1/students/me/avatar", nil)
	req.Header.Set("Authorization", "Bearer "+testhelper.MakeTestJWT(t, "00000000-0000-4000-a000-000000000001"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code, rec.Body.String())
}
