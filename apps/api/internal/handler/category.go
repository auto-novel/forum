package handler

import (
	"net/http"
	"time"

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
	router.Get("/{id}/tag", httpx.EH(h.listTags))
}

type categoryResponse struct {
	ID        int64   `json:"id"`
	Slug      string  `json:"slug"`
	BannerURL *string `json:"bannerUrl,omitempty"`
}

type tagResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Color     int16     `json:"color"`
	IsActive  bool      `json:"isActive"`
	SortOrder int32     `json:"sortOrder"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (h *categoryHandler) list(w http.ResponseWriter, r *http.Request) error {
	items, err := h.categoryRepo.List()
	if err != nil {
		return httpx.InternalError(err, "查询分类失败")
	}
	response := make([]categoryResponse, len(items))
	for i, item := range items {
		response[i] = categoryResponse{
			ID:        item.ID,
			Slug:      item.Slug,
			BannerURL: item.BannerURL,
		}
	}
	render.JSON(w, r, response)
	return nil
}

func (h *categoryHandler) listTags(w http.ResponseWriter, r *http.Request) error {
	categoryID, err := httpx.ParseParamPositiveInt(r, "id")
	if err != nil {
		return err
	}
	items, err := h.tagRepo.List(categoryID, false)
	if err != nil {
		return httpx.InternalError(err, "查询标签失败")
	}
	response := make([]tagResponse, len(items))
	for i, item := range items {
		response[i] = tagResponse{
			ID:        item.ID,
			Name:      item.Name,
			Color:     item.Color,
			IsActive:  item.IsActive,
			SortOrder: item.SortOrder,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		}
	}
	render.JSON(w, r, response)
	return nil
}
