package handler

import (
	"net/http"

	"auth/internal/httpx"
	"auth/internal/repository"

	"github.com/go-chi/chi/v5"
)

type meHandler struct{ postRepo repository.PostRepository }

func NewMeHandler(postRepo repository.PostRepository) *meHandler {
	return &meHandler{postRepo: postRepo}
}

func (h *meHandler) RegisterRoutes(router chi.Router) {
	router.Use(httpx.RequireAccessToken)
	router.Get("/post", httpx.EH(h.listPosts))
	router.Get("/favorite", httpx.EH(h.listFavorites))
}

func (h *meHandler) listPosts(w http.ResponseWriter, r *http.Request) error {
	principal, _ := httpx.AuthenticatedPrincipal(r)
	return respondPosts(w, r, h.postRepo, repository.PostFilter{AuthorID: principal.UserID})
}

func (h *meHandler) listFavorites(w http.ResponseWriter, r *http.Request) error {
	principal, _ := httpx.AuthenticatedPrincipal(r)
	return respondPosts(w, r, h.postRepo, repository.PostFilter{FavoriteUserID: principal.UserID})
}
