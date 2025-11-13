package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/filehash/internal/config"
	"github.com/filehash/internal/usecase"
	"github.com/filehash/pkg/validator"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type Handlers struct {
	cfg         config.Config
	log         *zap.Logger
	authUseCase *usecase.AuthUseCase
	fileUseCase *usecase.FileUseCase
	excelUseCase *usecase.ExcelUseCase
}

func NewHandlers(
	cfg config.Config,
	log *zap.Logger,
	authUseCase *usecase.AuthUseCase,
	fileUseCase *usecase.FileUseCase,
	excelUseCase *usecase.ExcelUseCase,
) *Handlers {
	return &Handlers{
		cfg:         cfg,
		log:         log,
		authUseCase: authUseCase,
		fileUseCase: fileUseCase,
		excelUseCase: excelUseCase,
	}
}

func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	limited := io.LimitReader(r.Body, 1<<20)
	defer r.Body.Close()

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(limited)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		h.log.Warn("json decode failed", zap.Error(err))
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	registerReq := usecase.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	resp, err := h.authUseCase.Register(ctx, registerReq)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		h.log.Error("register failed", zap.Error(err))
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"status":  "success",
		"user_id": resp.UserID,
		"token":   resp.Token,
	})
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	limited := io.LimitReader(r.Body, 1<<20)
	defer r.Body.Close()

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(limited)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		h.log.Warn("json decode failed", zap.Error(err))
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	loginReq := usecase.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	resp, err := h.authUseCase.Login(ctx, loginReq)
	if err != nil {
		if strings.Contains(err.Error(), "invalid credentials") {
			writeError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		h.log.Error("login failed", zap.Error(err))
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "success",
		"user_id": resp.UserID,
		"token":   resp.Token,
	})
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *Handlers) Upload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	maxBody := h.cfg.MaxUpload + (1 << 20)
	r.Body = http.MaxBytesReader(w, r.Body, maxBody)

	if err := r.ParseMultipartForm(maxBody); err != nil {
		h.log.Warn("multipart parse failed", zap.Error(err))
		writeError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file field is required")
		return
	}
	defer file.Close()

	payload, err := readFilePayload(file, h.cfg.MaxUpload)
	if err != nil {
		h.log.Warn("file read failed", zap.Error(err))
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	contentType := http.DetectContentType(payload.peek)
	if !validator.ValidateContentType(contentType) {
		writeError(w, http.StatusUnsupportedMediaType, "unsupported content type")
		return
	}

	userID := r.FormValue("user_id")
	var userPtr *string
	if strings.TrimSpace(userID) != "" {
		uid := strings.TrimSpace(userID)
		if err := validator.ValidateUserID(uid); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		userPtr = &uid
	}

	req := usecase.UploadFileRequest{
		Filename:    header.Filename,
		Content:     payload.data,
		ContentType: contentType,
		UserID:      userPtr,
	}

	resp, err := h.fileUseCase.UploadFile(ctx, req)
	if err != nil {
		h.log.Error("upload file failed", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "upload failed")
		return
	}

	h.log.Info("file uploaded",
		zap.String("file_id", resp.FileID),
		zap.String("request_id", getRequestID(ctx)),
	)

	writeJSON(w, http.StatusCreated, map[string]any{
		"status":      "success",
		"file_id":     resp.FileID,
		"token":       resp.Token,
		"expires_in":  resp.ExpiresIn,
		"content_type": contentType,
		"size_bytes":  len(payload.data),
	})
}

func (h *Handlers) GetImage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fileID := chi.URLParam(r, "id")
	if strings.TrimSpace(fileID) == "" {
		writeError(w, http.StatusBadRequest, "file id required")
		return
	}

	tokenStr, err := bearerToken(r.Header.Get("Authorization"))
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	req := usecase.GetFileRequest{
		FileID: fileID,
		Token:  tokenStr,
	}

	resp, err := h.fileUseCase.GetFile(ctx, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "file not found")
			return
		}
		if strings.Contains(err.Error(), "token") || strings.Contains(err.Error(), "mismatch") {
			writeError(w, http.StatusForbidden, "invalid token")
			return
		}
		h.log.Error("get file failed", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "retrieval failed")
		return
	}

	w.Header().Set("Content-Type", resp.ContentType)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, validator.SanitizeFilename(resp.Filename)))
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(resp.Content); err != nil {
		h.log.Warn("write response failed", zap.Error(err))
	}
}

func (h *Handlers) GetFileMetadata(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fileID := chi.URLParam(r, "id")
	if strings.TrimSpace(fileID) == "" {
		writeError(w, http.StatusBadRequest, "file id required")
		return
	}

	tokenStr, err := bearerToken(r.Header.Get("Authorization"))
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	req := usecase.GetFileMetadataRequest{
		FileID: fileID,
		Token:  tokenStr,
	}

	asset, err := h.fileUseCase.GetFileMetadata(ctx, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "file not found")
			return
		}
		if strings.Contains(err.Error(), "token") || strings.Contains(err.Error(), "mismatch") {
			writeError(w, http.StatusForbidden, "invalid token")
			return
		}
		h.log.Error("get metadata failed", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "retrieval failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"file_id":         asset.ID,
		"original_name":   asset.OriginalName,
		"content_type":    asset.ContentType,
		"size_bytes":      asset.SizeBytes,
		"encryption_alg":  asset.EncryptionAlg,
		"created_at":      asset.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":      asset.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

func (h *Handlers) DeleteFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fileID := chi.URLParam(r, "id")
	if strings.TrimSpace(fileID) == "" {
		writeError(w, http.StatusBadRequest, "file id required")
		return
	}

	tokenStr, err := bearerToken(r.Header.Get("Authorization"))
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	req := usecase.DeleteFileRequest{
		FileID: fileID,
		Token:  tokenStr,
	}

	if err := h.fileUseCase.DeleteFile(ctx, req); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "file not found")
			return
		}
		if strings.Contains(err.Error(), "token") || strings.Contains(err.Error(), "mismatch") {
			writeError(w, http.StatusForbidden, "invalid token")
			return
		}
		h.log.Error("delete file failed", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "deletion failed")
		return
	}

	h.log.Info("file deleted",
		zap.String("file_id", fileID),
		zap.String("request_id", getRequestID(ctx)),
	)

	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "file deleted",
	})
}

func (h *Handlers) ListFiles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := r.URL.Query().Get("user_id")

	if userID == "" {
		writeError(w, http.StatusBadRequest, "user_id query parameter required")
		return
	}

	if err := validator.ValidateUserID(userID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	req := usecase.ListFilesRequest{
		UserID: userID,
	}

	assets, err := h.fileUseCase.ListFiles(ctx, req)
	if err != nil {
		h.log.Error("list files failed", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "list failed")
		return
	}

	results := make([]map[string]any, 0, len(assets))
	for _, asset := range assets {
		results = append(results, map[string]any{
			"file_id":       asset.ID,
			"original_name": asset.OriginalName,
			"content_type":  asset.ContentType,
			"size_bytes":    asset.SizeBytes,
			"created_at":    asset.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status": "success",
		"files":  results,
		"count":  len(results),
	})
}

func (h *Handlers) JSONToExcel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	limited := io.LimitReader(r.Body, 5<<20)
	defer r.Body.Close()

	var payload map[string][]any
	decoder := json.NewDecoder(limited)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		h.log.Warn("json decode failed", zap.Error(err))
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	req := usecase.GenerateExcelRequest{
		Data: payload,
	}

	resp, err := h.excelUseCase.GenerateExcel(ctx, req)
	if err != nil {
		h.log.Warn("excel generation failed", zap.Error(err))
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"status":       "success",
		"path":         resp.Path,
		"rows":         resp.Rows,
		"excel_id":     resp.ExcelID,
		"generated_at": time.Now().UTC().Format(time.RFC3339),
	})
}

func readFilePayload(file multipart.File, maxSize int64) (*filePayload, error) {
	var buf bytes.Buffer
	limited := io.LimitReader(file, maxSize+1)
	n, err := buf.ReadFrom(limited)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	if n > maxSize {
		return nil, fmt.Errorf("file exceeds limit of %d bytes", maxSize)
	}
	data := buf.Bytes()
	peekLen := len(data)
	if peekLen > 512 {
		peekLen = 512
	}
	return &filePayload{
		data: data,
		peek: data[:peekLen],
	}, nil
}

type filePayload struct {
	data []byte
	peek []byte
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		// Writing to response writer failure cannot be handled gracefully here.
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{
		"status":  "error",
		"message": message,
	})
}

func bearerToken(header string) (string, error) {
	if header == "" {
		return "", errors.New("authorization header required")
	}
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("invalid authorization header")
	}
	return parts[1], nil
}

func getRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	return middleware.GetReqID(ctx)
}

