package handler

import (
	"net/http"

	"auth/internal/httpx"

	"github.com/go-chi/chi/v5"
)

func (h *postHandler) RegisterAdminRoutes(router chi.Router) {
	router.Patch("/{id}", httpx.EH(h.moderate))
}

type postModerationInput struct {
	Status         int16  `json:"status"`
	CommentsLocked bool   `json:"commentsLocked"`
	PinOrder       *int32 `json:"pinOrder"`
}

func (h *postHandler) moderate(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.ParseParamPositiveInt(r, "id")
	if err != nil {
		return err
	}
	input, err := httpx.Body[postModerationInput](r)
	if err != nil {
		return err
	}
	if !validStatus(input.Status) {
		return httpx.BadRequest("status 必须为 0、1 或 2")
	}
	if err := h.postRepo.SetModeration(id, input.Status, input.CommentsLocked, input.PinOrder); err != nil {
		return repoError(err, "管理帖子失败")
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
