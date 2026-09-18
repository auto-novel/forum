package handler

import (
	"net/http"
	"strconv"
	"strings"

	"auth/internal/httpx"
	"auth/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func (h *commentHandler) RegisterAdminRoutes(router chi.Router) {
	router.Get("/", httpx.EH(h.listAdmin))
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

func (h *commentHandler) listAdmin(w http.ResponseWriter, r *http.Request) error {
	query := r.URL.Query()
	pagination, err := parsePagination(query, 20, 100)
	if err != nil {
		return err
	}
	filter := repository.CommentFilter{
		Search:     strings.TrimSpace(query.Get("q")),
		AuthorName: strings.TrimSpace(query.Get("author_name")),
		Status:     repository.CommentStatusAll,
	}
	if values, ok := query["post_id"]; ok {
		if len(values) != 1 {
			return httpx.BadRequest("post_id 必须为正整数")
		}
		id, err := strconv.ParseInt(values[0], 10, 64)
		if err != nil || id <= 0 {
			return httpx.BadRequest("post_id 必须为正整数")
		}
		filter.PostID = id
	}
	if values, ok := query["status"]; ok {
		if len(values) != 1 {
			return httpx.BadRequest("status 必须为 all、0、1 或 2")
		}
		if values[0] != "all" {
			status, err := strconv.ParseInt(values[0], 10, 16)
			if err != nil || !validStatus(int16(status)) {
				return httpx.BadRequest("status 必须为 all、0、1 或 2")
			}
			filter.Status = int16(status)
		}
	}
	total, items, err := h.repo.ListAdmin(filter, pagination.Limit, pagination.Offset)
	if err != nil {
		return httpx.InternalError(err, "查询评论失败")
	}
	responses := make([]commentResponse, len(items))
	for i, item := range items {
		response, err := newCommentResponse(r, item)
		if err != nil {
			return httpx.InternalError(err, "转换评论数据失败")
		}
		responses[i] = response
	}
	render.JSON(w, r, page[commentResponse]{Total: total, Items: responses})
	return nil
}
