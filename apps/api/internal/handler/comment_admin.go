package handler

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"forum/internal/domain"
	"forum/internal/httpx"
	"forum/internal/usecase"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func (h *commentHandler) RegisterAdminRoutes(router chi.Router) {
	router.Get("/", httpx.EH(h.listAdmin))
	router.Put("/{id}/status", httpx.EH(h.setStatus))
	router.Delete("/author/{authorId}", httpx.EH(h.deleteAllByAuthor))
}

type commentStatusInput struct {
	Status string `json:"status" label:"评论状态" validate:"required"`
}

func (h *commentHandler) setStatus(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.ParseParamPositiveInt(r, "id", "评论 ID")
	if err != nil {
		return err
	}
	subjectType, err := parseAdminCommentSubjectType(r.URL.Query())
	if err != nil {
		return err
	}
	kind := domain.CommentSubjectPost
	if subjectType != nil {
		kind = *subjectType
	}
	input, err := httpx.Body[commentStatusInput](r)
	if err != nil {
		return err
	}
	status, err := parseCommentStatus(input.Status)
	if err != nil {
		return err
	}
	principal, _ := httpx.AuthenticatedPrincipal(r)
	if err := h.commentUsecase.SetStatus(
		r.Context(),
		actorFromPrincipal(principal),
		usecase.SetCommentStatusCommand{
			SubjectType: kind,
			CommentID:   id,
			Status:      status,
		},
	); err != nil {
		return transportError(err)
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *commentHandler) deleteAllByAuthor(w http.ResponseWriter, r *http.Request) error {
	authorID, err := httpx.ParseParamPositiveInt(r, "authorId", "作者 ID")
	if err != nil {
		return err
	}
	principal, _ := httpx.AuthenticatedPrincipal(r)
	if err := h.commentUsecase.DeleteAllByAuthor(
		r.Context(),
		actorFromPrincipal(principal),
		usecase.DeleteCommentsByAuthorCommand{
			AuthorID: authorID,
		},
	); err != nil {
		return transportError(err)
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
	subjectType, err := parseAdminCommentSubjectType(query)
	if err != nil {
		return err
	}
	listQuery := usecase.ListAdminCommentsQuery{
		SubjectType: subjectType,
		Search:      strings.TrimSpace(query.Get("q")),
		AuthorName:  strings.TrimSpace(query.Get("author_name")),
		Limit:       pagination.Limit,
		Offset:      pagination.Offset,
	}
	if values, ok := query["subject_key"]; ok {
		if len(values) != 1 || values[0] == "" {
			return httpx.BadRequest("评论对象 key 无效")
		}
		listQuery.SubjectKey = &values[0]
	}
	if values, ok := query["post_id"]; ok {
		if subjectType != nil && *subjectType != domain.CommentSubjectPost {
			return httpx.BadRequest("帖子 ID 不能用于非帖子评论")
		}
		if len(values) != 1 {
			return httpx.BadRequest("帖子 ID 必须为正整数")
		}
		id, err := strconv.ParseInt(values[0], 10, 64)
		if err != nil || id <= 0 {
			return httpx.BadRequest("帖子 ID 必须为正整数")
		}
		listQuery.PostID = id
	}
	if values, ok := query["status"]; ok {
		if len(values) != 1 {
			return httpx.BadRequest("状态无效")
		}
		if values[0] != "all" {
			status, err := strconv.ParseInt(values[0], 10, 16)
			if err != nil || !domain.CommentStatus(status).Valid() {
				return httpx.BadRequest("状态无效")
			}
			commentStatus := domain.CommentStatus(status)
			listQuery.Status = &commentStatus
		}
	}
	principal, _ := httpx.AuthenticatedPrincipal(r)
	total, items, err := h.commentUsecase.ListAdmin(
		r.Context(),
		actorFromPrincipal(principal),
		listQuery,
	)
	if err != nil {
		return transportError(err)
	}
	responses := make([]adminCommentResponse, len(items))
	for i, item := range items {
		response, err := newAdminCommentResponse(item)
		if err != nil {
			return httpx.InternalError(err, "转换评论数据失败")
		}
		responses[i] = response
	}
	render.JSON(w, r, page[adminCommentResponse]{Total: total, Items: responses})
	return nil
}

// Admin responses carry subject metadata without changing the public post contract.
type adminCommentResponse struct {
	ID             int64                       `json:"id"`
	SubjectType    domain.CommentSubjectType   `json:"subjectType"`
	SubjectKey     string                      `json:"subjectKey"`
	PostID         *int64                      `json:"postId"`
	RootID         *int64                      `json:"rootId"`
	Content        string                      `json:"content"`
	AuthorID       int64                       `json:"authorId"`
	AuthorUsername string                      `json:"authorUsername"`
	AuthorBlocked  bool                        `json:"authorBlocked"`
	Status         int16                       `json:"status"`
	CreatedAt      time.Time                   `json:"createdAt"`
	UpdatedAt      time.Time                   `json:"updatedAt"`
	ReplyCount     int64                       `json:"replyCount"`
	Replies        *page[adminCommentResponse] `json:"replies,omitempty"`
}

func newAdminCommentResponse(value domain.CommentReadModel) (adminCommentResponse, error) {
	response := adminCommentResponse{
		ID: value.ID, SubjectType: value.SubjectType, SubjectKey: value.SubjectKey,
		RootID: value.RootID, Content: value.Content, AuthorID: value.AuthorID,
		AuthorUsername: value.AuthorUsername, AuthorBlocked: value.AuthorBlocked,
		Status: int16(value.Status), CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt,
	}
	if value.SubjectType == domain.CommentSubjectPost {
		postID, err := domain.PostIDFromCommentSubjectKey(value.SubjectKey)
		if err != nil {
			return adminCommentResponse{}, err
		}
		response.PostID = &postID
	}
	return response, nil
}

func parseAdminCommentSubjectType(query url.Values) (*domain.CommentSubjectType, error) {
	values, ok := query["subject_type"]
	if !ok {
		return nil, nil
	}
	if len(values) != 1 || (values[0] != "0" && values[0] != "1") {
		return nil, httpx.BadRequest("评论对象类型无效")
	}
	kind := domain.CommentSubjectPost
	if values[0] == "1" {
		kind = domain.CommentSubjectNovel
	}
	return &kind, nil
}

func parseCommentStatus(value string) (domain.CommentStatus, error) {
	switch value {
	case "published":
		return domain.CommentStatusPublished, nil
	case "hidden":
		return domain.CommentStatusHidden, nil
	case "deleted":
		return domain.CommentStatusDeleted, nil
	default:
		return 0, httpx.BadRequest("状态无效")
	}
}
