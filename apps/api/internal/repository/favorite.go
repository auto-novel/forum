package repository

import (
	"auth/.gen/main/public/model"
	"auth/.gen/main/public/table"
	"database/sql"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
)

type FavoriteRepository interface {
	Has(postID, userID int64) (bool, error)
	Set(postID, userID int64, favorite bool) error
}

type favoriteRepository struct{ db *sql.DB }

func NewFavoriteRepository(db *sql.DB) FavoriteRepository { return &favoriteRepository{db: db} }

func (r *favoriteRepository) Has(postID, userID int64) (bool, error) {
	var result struct{ Exists bool }
	stmt := SELECT(EXISTS(SELECT(table.PostFavorite.PostID).
		FROM(table.PostFavorite).
		WHERE(table.PostFavorite.PostID.EQ(Int64(postID)).
			AND(table.PostFavorite.UserID.EQ(Int64(userID))))).AS("Exists"))
	if err := stmt.Query(r.db, &result); err != nil {
		return false, err
	}
	return result.Exists, nil
}

func (r *favoriteRepository) Set(postID, userID int64, favorite bool) error {
	if !favorite {
		stmt := table.PostFavorite.DELETE().
			WHERE(table.PostFavorite.PostID.EQ(Int64(postID)).AND(table.PostFavorite.UserID.EQ(Int64(userID))))
		_, err := stmt.Exec(r.db)
		return err
	}
	var exists struct{ Exists bool }
	findPost := SELECT(EXISTS(SELECT(table.Post.ID).
		FROM(table.Post).
		WHERE(table.Post.ID.EQ(Int64(postID)).AND(table.Post.Status.EQ(Int16(StatusPublished))))).AS("Exists"))
	if err := findPost.Query(r.db, &exists); err != nil {
		return err
	}
	if !exists.Exists {
		return qrm.ErrNoRows
	}
	record := model.PostFavorite{PostID: postID, UserID: userID}
	stmt := table.PostFavorite.INSERT(table.PostFavorite.PostID, table.PostFavorite.UserID).
		MODEL(record).
		ON_CONFLICT(table.PostFavorite.PostID, table.PostFavorite.UserID).
		DO_NOTHING()
	_, err := stmt.Exec(r.db)
	return err
}
