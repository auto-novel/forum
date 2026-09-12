package handler

import (
	"net/http"

	"auth/internal/httpx"
	"auth/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type categoryHandler struct {
	categoryRepo repository.CategoryRepository
	tagRepo      repository.TagRepository
}

func NewCategoryHandler(
	categoryRepo repository.CategoryRepository,
	tagRepo repository.TagRepository,
) *categoryHandler {
	return &categoryHandler{
		categoryRepo: categoryRepo,
		tagRepo:      tagRepo,
	}
}

func (h *categoryHandler) RegisterRoutes(router chi.Router) {
	router.Get("/", httpx.EH(h.list))
}

type categoryListResponse struct {
	ID        int64                 `json:"id"`
	Slug      string                `json:"slug"`
	BannerURL *string               `json:"bannerUrl,omitempty"`
	Tags      []categoryTagResponse `json:"tags"`
}

type categoryTagResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Color     int16  `json:"color"`
	SortOrder int32  `json:"sortOrder"`
}

func (h *categoryHandler) list(w http.ResponseWriter, r *http.Request) error {
	items, err := h.categoryRepo.List()
	if err != nil {
		return httpx.InternalError(err, "查询分类失败")
	}
	tags, err := h.tagRepo.ListActive()
	if err != nil {
		return httpx.InternalError(err, "查询标签失败")
	}
	render.JSON(w, r, newCategoryResponses(items, tags))
	return nil
}

func newCategoryResponses(items []repository.Category, tags []repository.Tag) []categoryListResponse {
	tagsByCategory := make(map[int64][]categoryTagResponse, len(items))
	for _, tag := range tags {
		tagsByCategory[tag.CategoryID] = append(tagsByCategory[tag.CategoryID], categoryTagResponse{
			ID:        tag.ID,
			Name:      tag.Name,
			Color:     tag.Color,
			SortOrder: tag.SortOrder,
		})
	}
	response := make([]categoryListResponse, len(items))
	for i, item := range items {
		categoryTags := tagsByCategory[item.ID]
		if categoryTags == nil {
			categoryTags = []categoryTagResponse{}
		}
		response[i] = categoryListResponse{
			ID:        item.ID,
			Slug:      item.Slug,
			BannerURL: item.BannerURL,
			Tags:      categoryTags,
		}
	}
	return response
}
