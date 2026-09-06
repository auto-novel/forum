package handler

import (
	"net/http"
	"strings"

	"auth/internal/httpx"
	"auth/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func (h *categoryHandler) RegisterAdminRoutes(router chi.Router) {
	router.Post("/", httpx.EH(h.createCategory))
	router.Put("/{id}", httpx.EH(h.updateCategory))
	router.Post("/{cid}/tag", httpx.EH(h.createTag))
	router.Put("/{cid}/tag/{id}", httpx.EH(h.updateTag))
}

type categoryInput struct {
	Slug      string  `json:"slug" validate:"required,max=255"`
	BannerURL *string `json:"bannerUrl"`
}

func (h *categoryHandler) createCategory(w http.ResponseWriter, r *http.Request) error {
	input, err := httpx.Body[categoryInput](r)
	if err != nil {
		return err
	}

	category, err := h.categoryRepo.Create(
		strings.TrimSpace(input.Slug),
		input.BannerURL, "{}",
	)
	if repository.IsUniqueViolation(err) {
		return httpx.Conflict("分类已存在")
	}
	if err != nil {
		return httpx.InternalError(err, "创建分类失败")
	}

	response := categoryResponse{
		ID:        category.ID,
		Slug:      category.Slug,
		BannerURL: category.BannerURL,
	}
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, response)
	return nil
}

func (h *categoryHandler) updateCategory(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.ParseParamPositiveInt(r, "id")
	if err != nil {
		return err
	}
	input, err := httpx.Body[categoryInput](r)
	if err != nil {
		return err
	}

	category, err := h.categoryRepo.Update(
		id,
		strings.TrimSpace(input.Slug),
		input.BannerURL,
	)
	if repository.IsNotFound(err) {
		return httpx.NotFound("分类不存在")
	} else if repository.IsUniqueViolation(err) {
		return httpx.Conflict("分类已存在")
	} else if err != nil {
		return httpx.InternalError(err, "更新分类失败")
	}

	render.JSON(w, r, categoryResponse{
		ID:        category.ID,
		Slug:      category.Slug,
		BannerURL: category.BannerURL,
	})
	return nil
}

type tagInput struct {
	Name      string `json:"name" validate:"required,max=64"`
	Color     int16  `json:"color" validate:"gte=0"`
	IsActive  bool   `json:"isActive"`
	SortOrder int32  `json:"sortOrder"`
}

func (h *categoryHandler) createTag(w http.ResponseWriter, r *http.Request) error {
	categoryID, err := httpx.ParseParamPositiveInt(r, "cid")
	if err != nil {
		return err
	}
	if _, err := h.categoryRepo.Find(categoryID); repository.IsNotFound(err) {
		return httpx.NotFound("分类不存在")
	} else if err != nil {
		return httpx.InternalError(err, "查询分类失败")
	}

	input, err := httpx.Body[tagInput](r)
	if err != nil {
		return err
	}

	tag, err := h.tagRepo.Create(
		categoryID,
		strings.TrimSpace(input.Name),
		input.Color,
		input.SortOrder,
		"{}",
	)
	if repository.IsUniqueViolation(err) {
		return httpx.Conflict("标签已存在")
	} else if err != nil {
		return httpx.InternalError(err, "创建标签失败")
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, tagResponse{
		ID:        tag.ID,
		Name:      tag.Name,
		Color:     tag.Color,
		IsActive:  tag.IsActive,
		SortOrder: tag.SortOrder,
		CreatedAt: tag.CreatedAt,
		UpdatedAt: tag.UpdatedAt,
	})
	return nil
}

func (h *categoryHandler) updateTag(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.ParseParamPositiveInt(r, "id")
	if err != nil {
		return err
	}
	input, err := httpx.Body[tagInput](r)
	if err != nil {
		return err
	}

	tag, err := h.tagRepo.Update(
		id,
		strings.TrimSpace(input.Name),
		input.Color,
		input.IsActive,
		input.SortOrder,
	)
	if repository.IsNotFound(err) {
		return httpx.NotFound("标签不存在")
	} else if err != nil {
		return httpx.InternalError(err, "更新标签失败")
	}

	render.JSON(w, r, tagResponse{
		ID:        tag.ID,
		Name:      tag.Name,
		Color:     tag.Color,
		IsActive:  tag.IsActive,
		SortOrder: tag.SortOrder,
		CreatedAt: tag.CreatedAt,
		UpdatedAt: tag.UpdatedAt,
	})
	return nil
}
