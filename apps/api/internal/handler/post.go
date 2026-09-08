package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"auth/internal/httpx"
	"auth/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type postTagResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Color int16  `json:"color"`
}

type postResponse struct {
	ID             int64             `json:"id"`
	CategoryID     int64             `json:"categoryId"`
	Title          string            `json:"title"`
	AuthorID       int64             `json:"authorId"`
	AuthorUsername string            `json:"authorUsername"`
	Content        string            `json:"content"`
	Status         int16             `json:"status"`
	ViewsCount     int32             `json:"viewsCount"`
	CommentsCount  int32             `json:"commentsCount"`
	CommentsLocked bool              `json:"commentsLocked"`
	PinOrder       *int32            `json:"pinOrder"`
	CreatedAt      time.Time         `json:"createdAt"`
	UpdatedAt      time.Time         `json:"updatedAt"`
	ActiveAt       time.Time         `json:"activeAt"`
	Tags           []postTagResponse `json:"tags"`
}

func newPostResponse(value repository.PostDetails) postResponse {
	tags := make([]postTagResponse, len(value.Tags))
	for i, tag := range value.Tags {
		tags[i] = postTagResponse{
			ID:    tag.ID,
			Name:  tag.Name,
			Color: tag.Color,
		}
	}
	return postResponse{
		ID:             value.ID,
		CategoryID:     value.CategoryID,
		Title:          value.Title,
		AuthorID:       value.AuthorID,
		AuthorUsername: value.AuthorUsername,
		Content:        value.Content,
		Status:         value.Status,
		ViewsCount:     value.ViewsCount,
		CommentsCount:  value.CommentsCount,
		CommentsLocked: value.CommentsLocked,
		PinOrder:       value.PinOrder,
		CreatedAt:      value.CreatedAt,
		UpdatedAt:      value.UpdatedAt,
		ActiveAt:       value.ActiveAt,
		Tags:           tags,
	}
}

type postHandler struct {
	postRepo     repository.PostRepository
	favoriteRepo repository.FavoriteRepository
	commentRepo  repository.CommentRepository
}

func NewPostHandler(
	postRepo repository.PostRepository,
	favoriteRepo repository.FavoriteRepository,
	commentRepo repository.CommentRepository,
) *postHandler {
	return &postHandler{postRepo: postRepo, favoriteRepo: favoriteRepo, commentRepo: commentRepo}
}

func (h *postHandler) RegisterRoutes(router chi.Router) {
	router.Get("/", httpx.EH(h.list))
	router.With(httpx.RequireAccessToken).Post("/", httpx.EH(h.create))
	router.Route("/{id}", func(router chi.Router) {
		router.Get("/", httpx.EH(h.get))
		router.With(httpx.RequireAccessToken).Patch("/", httpx.EH(h.update))
		router.With(httpx.RequireAccessToken).Delete("/", httpx.EH(h.delete))
		router.With(httpx.RequireAccessToken).Put("/favorite", httpx.EH(h.favorite))
		router.With(httpx.RequireAccessToken).Delete("/favorite", httpx.EH(h.unfavorite))
		router.Get("/comment", httpx.EH(h.listComments))
		router.With(httpx.RequireAccessToken).Post("/comment", httpx.EH(h.createComment))
	})
}

func postFilterFrom(r *http.Request) (repository.PostFilter, error) {
	filter := repository.PostFilter{
		CategorySlug: r.URL.Query().Get("category"),
		Search:       strings.TrimSpace(r.URL.Query().Get("q")),
	}
	for _, part := range r.URL.Query()["tag"] {
		for _, value := range strings.Split(part, ",") {
			id, err := strconv.ParseInt(value, 10, 64)
			if err != nil || id <= 0 {
				return filter, httpx.BadRequest("tag 必须为正整数")
			}
			filter.TagIDs = append(filter.TagIDs, id)
		}
	}
	if !uniquePositiveIDs(filter.TagIDs) {
		return filter, httpx.BadRequest("tag 不能重复")
	}
	return filter, nil
}

func respondPosts(
	w http.ResponseWriter,
	r *http.Request,
	repo repository.PostRepository,
	filter repository.PostFilter,
) error {
	pagination, err := parsePagination(r.URL.Query(), 20, 100)
	if err != nil {
		return err
	}
	total, items, err := repo.List(filter, pagination.Limit, pagination.Offset)
	if err != nil {
		return repoError(err, "查询帖子失败")
	}
	response := make([]postResponse, len(items))
	for i, item := range items {
		response[i] = newPostResponse(item)
	}
	render.JSON(w, r, page[postResponse]{Total: total, Items: response})
	return nil
}

func (h *postHandler) list(w http.ResponseWriter, r *http.Request) error {
	filter, err := postFilterFrom(r)
	if err != nil {
		return err
	}
	return respondPosts(w, r, h.postRepo, filter)
}

func (h *postHandler) get(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.ParseParamPositiveInt(r, "id")
	if err != nil {
		return err
	}
	post, err := h.postRepo.Find(id, true)
	if err != nil {
		return repoError(err, "查询帖子失败")
	}
	render.JSON(w, r, newPostResponse(*post))
	return nil
}

type postInput struct {
	Category string  `json:"category"`
	Title    string  `json:"title"`
	Content  string  `json:"content"`
	TagIDs   []int64 `json:"tagIds"`
}

func validatePost(input postInput, creating bool) error {
	if creating && !validText(input.Category, 1, 255) {
		return httpx.BadRequest("category 长度必须为 1 到 255")
	}
	if !validText(input.Title, 1, 500) {
		return httpx.BadRequest("title 长度必须为 1 到 500")
	}
	if !validText(input.Content, 1, 1000000) {
		return httpx.BadRequest("content 不能为空且不能超过 1000000 字")
	}
	if !uniquePositiveIDs(input.TagIDs) {
		return httpx.BadRequest("tagIds 必须为不重复的正整数")
	}
	return nil
}

func (h *postHandler) create(w http.ResponseWriter, r *http.Request) error {
	input, err := httpx.Body[postInput](r)
	if err != nil {
		return err
	}
	if err := validatePost(input, true); err != nil {
		return err
	}
	principal, _ := httpx.AuthenticatedPrincipal(r)
	post, err := h.postRepo.Create(repository.CreatePostInput{
		CategorySlug:   strings.TrimSpace(input.Category),
		Title:          strings.TrimSpace(input.Title),
		Content:        input.Content,
		TagIDs:         input.TagIDs,
		AuthorID:       principal.UserID,
		AuthorUsername: principal.Username,
		Attr:           "{}",
	})
	if err != nil {
		return repoError(err, "创建帖子失败")
	}
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, newPostResponse(*post))
	return nil
}

func (h *postHandler) ownedID(r *http.Request) (int64, error) {
	id, err := httpx.ParseParamPositiveInt(r, "id")
	if err != nil {
		return 0, err
	}
	post, err := h.postRepo.Find(id, false)
	if err != nil {
		return 0, repoError(err, "查询帖子失败")
	}
	principal, _ := httpx.AuthenticatedPrincipal(r)
	if post.AuthorID != principal.UserID && !principal.IsAdmin() {
		return 0, httpx.Forbidden("只能修改自己的帖子")
	}
	return id, nil
}

func (h *postHandler) update(w http.ResponseWriter, r *http.Request) error {
	id, err := h.ownedID(r)
	if err != nil {
		return err
	}
	input, err := httpx.Body[postInput](r)
	if err != nil {
		return err
	}
	if err := validatePost(input, false); err != nil {
		return err
	}
	post, err := h.postRepo.Update(id, repository.UpdatePostInput{
		Title:   strings.TrimSpace(input.Title),
		Content: input.Content,
		TagIDs:  input.TagIDs,
	})
	if err != nil {
		return repoError(err, "更新帖子失败")
	}
	render.JSON(w, r, newPostResponse(*post))
	return nil
}

func (h *postHandler) delete(w http.ResponseWriter, r *http.Request) error {
	id, err := h.ownedID(r)
	if err != nil {
		return err
	}
	if err := h.postRepo.SetStatus(id, repository.StatusDeleted); err != nil {
		return repoError(err, "删除帖子失败")
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *postHandler) setFavorite(w http.ResponseWriter, r *http.Request, favorite bool) error {
	postID, err := httpx.ParseParamPositiveInt(r, "id")
	if err != nil {
		return err
	}
	principal, _ := httpx.AuthenticatedPrincipal(r)
	if err := h.favoriteRepo.Set(postID, principal.UserID, favorite); err != nil {
		return repoError(err, "更新收藏失败")
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *postHandler) favorite(w http.ResponseWriter, r *http.Request) error {
	return h.setFavorite(w, r, true)
}

func (h *postHandler) unfavorite(w http.ResponseWriter, r *http.Request) error {
	return h.setFavorite(w, r, false)
}

func (h *postHandler) listComments(w http.ResponseWriter, r *http.Request) error {
	postID, err := httpx.ParseParamPositiveInt(r, "id")
	if err != nil {
		return err
	}
	if _, err := h.postRepo.Find(postID, false); err != nil {
		return repoError(err, "查询帖子失败")
	}
	pagination, err := parsePagination(r.URL.Query(), 20, 100)
	if err != nil {
		return err
	}
	total, items, err := h.commentRepo.List(
		repository.CommentSubjectPost,
		repository.PostSubjectKey(postID),
		pagination.Limit,
		pagination.Offset,
	)
	if err != nil {
		return httpx.InternalError(err, "查询评论失败")
	}
	response := make([]commentResponse, len(items))
	for i, item := range items {
		response[i], err = newCommentResponse(item)
		if err != nil {
			return httpx.InternalError(err, "转换评论数据失败")
		}
	}
	render.JSON(w, r, page[commentResponse]{Total: total, Items: response})
	return nil
}

type commentInput struct {
	Content string `json:"content" validate:"required,max=100000"`
	RootID  *int64 `json:"rootId" validate:"omitempty,gt=0"`
}

func (h *postHandler) createComment(w http.ResponseWriter, r *http.Request) error {
	postID, err := httpx.ParseParamPositiveInt(r, "id")
	if err != nil {
		return err
	}
	input, err := httpx.Body[commentInput](r)
	if err != nil {
		return err
	}
	principal, _ := httpx.AuthenticatedPrincipal(r)
	comment, err := h.commentRepo.Create(repository.CreateCommentInput{
		SubjectType:    repository.CommentSubjectPost,
		SubjectKey:     repository.PostSubjectKey(postID),
		RootID:         input.RootID,
		Content:        input.Content,
		AuthorID:       principal.UserID,
		AuthorUsername: principal.Username,
		Attr:           "{}",
	})
	if repository.IsNotFound(err) {
		return httpx.NotFound("帖子不存在")
	}
	if errors.Is(err, repository.ErrCommentsLocked) {
		return httpx.Conflict("评论区已锁定")
	}
	if err != nil {
		return httpx.InternalError(err, "创建评论失败")
	}
	response, err := newCommentResponse(*comment)
	if err != nil {
		return httpx.InternalError(err, "转换评论数据失败")
	}
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, response)
	return nil
}
