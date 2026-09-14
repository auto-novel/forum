package handler

import (
	"net/http"

	"auth/internal/httpx"
	"auth/internal/repository"

	"github.com/go-chi/chi/v5"
)

func (h *commentHandler) RegisterAdminRoutes(router chi.Router) {
	router.Put("/{id}/status", httpx.EH(h.setStatus))
	router.Delete("/author/{authorId}", httpx.EH(h.deleteAllByAuthor))
}

type commentStatusInput struct {
	Status string `json:"status" validate:"required"`
}

func (h *commentHandler) setStatus(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.ParseParamPositiveInt(r, "id")
	if err != nil {
		return err
	}
	input, err := httpx.Body[commentStatusInput](r)
	if err != nil {
		return err
	}
	var status int16
	switch input.Status {
	case "published":
		status = repository.StatusPublished
	case "hidden":
		status = repository.StatusHidden
	case "deleted":
		status = repository.StatusDeleted
	default:
		return httpx.BadRequest("status 必须为 published、hidden 或 deleted")
	}
	err = h.repo.SetStatus(repository.CommentSubjectPost, id, status)
	if repository.IsNotFound(err) {
		return httpx.NotFound("评论不存在")
	}
	if err != nil {
		return httpx.InternalError(err, "设置评论状态失败")
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *commentHandler) deleteAllByAuthor(w http.ResponseWriter, r *http.Request) error {
	authorID, err := httpx.ParseParamPositiveInt(r, "authorId")
	if err != nil {
		return err
	}
	if err := h.repo.DeleteAllByAuthor(authorID); err != nil {
		return httpx.InternalError(err, "删除用户评论失败")
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
