package handler

import (
	"fmt"
	"net/http"
	"time"

	"auth/internal/httpx"
	"auth/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type commentHandler struct{ repo repository.CommentRepository }

func NewCommentHandler(repo repository.CommentRepository) *commentHandler {
	return &commentHandler{repo: repo}
}

func (h *commentHandler) RegisterRoutes(router chi.Router) {
	router.Use(httpx.RequireAccessToken)
	router.Patch("/{id}", httpx.EH(h.update))
	router.Delete("/{id}", httpx.EH(h.delete))
}

type commentResponse struct {
	ID             int64     `json:"id"`
	PostID         int64     `json:"postId"`
	RootID         *int64    `json:"rootId"`
	Content        string    `json:"content"`
	AuthorID       int64     `json:"authorId"`
	AuthorUsername string    `json:"authorUsername"`
	Status         int16     `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func newCommentResponse(r *http.Request, value repository.Comment) (commentResponse, error) {
	postID, err := repository.PostIDFromSubjectKey(value.SubjectKey)
	if err != nil {
		return commentResponse{}, fmt.Errorf("comment %d: %w", value.ID, err)
	}
	return commentResponse{
		ID:             value.ID,
		PostID:         postID,
		RootID:         value.RootID,
		Content:        publicCommentContent(r, value),
		AuthorID:       value.AuthorID,
		AuthorUsername: value.AuthorUsername,
		Status:         value.Status,
		CreatedAt:      value.CreatedAt,
		UpdatedAt:      value.UpdatedAt,
	}, nil
}

func publicCommentContent(r *http.Request, value repository.Comment) string {
	principal, err := httpx.AuthenticatedPrincipal(r)
	if value.Status == repository.StatusPublished || (err == nil && principal.IsAdmin()) {
		return value.Content
	}
	return ""
}

func (h *commentHandler) modifiableID(r *http.Request) (int64, error) {
	id, err := httpx.ParseParamPositiveInt(r, "id")
	if err != nil {
		return 0, err
	}
	comment, err := h.repo.Find(repository.CommentSubjectPost, id)
	if repository.IsNotFound(err) {
		return 0, httpx.NotFound("评论不存在")
	}
	if err != nil {
		return 0, httpx.InternalError(err, "查询评论失败")
	}
	principal, _ := httpx.AuthenticatedPrincipal(r)
	if comment.AuthorID != principal.UserID && !principal.IsAdmin() {
		return 0, httpx.Forbidden("只能修改自己的评论")
	}

	const modificationWindow = 20 * time.Minute
	if time.Now().After(comment.CreatedAt.Add(modificationWindow)) {
		return 0, httpx.Forbidden("评论只能在发布后 20 分钟内编辑或删除")
	}
	return comment.ID, nil
}

func (h *commentHandler) update(w http.ResponseWriter, r *http.Request) error {
	id, err := h.modifiableID(r)
	if err != nil {
		return err
	}
	input, err := httpx.Body[commentInput](r)
	if err != nil {
		return err
	}
	comment, err := h.repo.Update(repository.CommentSubjectPost, id, input.Content)
	if repository.IsNotFound(err) {
		return httpx.NotFound("评论不存在")
	}
	if err != nil {
		return httpx.InternalError(err, "更新评论失败")
	}
	response, err := newCommentResponse(r, *comment)
	if err != nil {
		return httpx.InternalError(err, "转换评论数据失败")
	}
	render.JSON(w, r, response)
	return nil
}

func (h *commentHandler) delete(w http.ResponseWriter, r *http.Request) error {
	id, err := h.modifiableID(r)
	if err != nil {
		return err
	}
	err = h.repo.SetStatus(repository.CommentSubjectPost, id, repository.StatusDeleted)
	if repository.IsNotFound(err) {
		return httpx.NotFound("评论不存在")
	}
	if err != nil {
		return httpx.InternalError(err, "删除评论失败")
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
