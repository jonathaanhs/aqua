package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/Masterminds/squirrel"
	"github.com/bxcodec/aqua/models"
	"github.com/labstack/echo/v4"
	"golang.org/x/sync/errgroup"
)

// ErroResponse ...
type ErroResponse struct {
	Message string `json:"message"`
}

// ArticleHandler ...
type ArticleHandler struct {
	DB *sql.DB
}

// InitArticle ...
func InitArticle(db *sql.DB) ArticleHandler {
	return ArticleHandler{
		DB: db,
	}
}

// FetchArticles ...
func (h ArticleHandler) FetchArticles(c echo.Context) (err error) {
	authorID := c.QueryParam("author_id")
	limit := c.QueryParam("limit")

	limitNumber := int64(20) // default
	if limit != "" {
		limitNumber, err = strconv.ParseInt(limit, 10, 64)
		if err != nil {
			resp := ErroResponse{
				Message: fmt.Sprintf("parameter 'limit' is not valid, should be a number. Error message: %s", err.Error()),
			}
			return c.JSON(http.StatusBadRequest, resp)
		}
	}

	data, err := h.fetchArticles(limitNumber, authorID)
	if err != nil {
		resp := ErroResponse{
			Message: err.Error(),
		}
		return c.JSON(http.StatusInternalServerError, resp)
	}

	g, _ := errgroup.WithContext(c.Request().Context())
	mutex := sync.Mutex{}
	for i, item := range data {
		i, item := i, item
		g.Go(func() (err error) {
			author, er := h.getAuthorByID(item.Author.ID)
			if er != nil {
				return er
			}
			mutex.Lock()
			data[i].Author = author
			mutex.Unlock()
			return nil
		})
	}

	err = g.Wait()
	if err != nil {
		resp := ErroResponse{
			Message: fmt.Sprintf("Failed when fetching the author's detail. Got Err: %s", err.Error()),
		}
		return c.JSON(http.StatusInternalServerError, resp)
	}

	return c.JSON(http.StatusOK, data)
}

func (h ArticleHandler) fetchArticles(limit int64, authorID string) (res []models.Article, err error) {
	queryBuilder := squirrel.Select("id", "title", "body", "author_id").From("articles")
	queryBuilder = queryBuilder.Limit(uint64(limit))

	if authorID != "" {
		queryBuilder = queryBuilder.Where(squirrel.Eq{
			"author_id": authorID,
		})
	}

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return
	}

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		return
	}
	defer rows.Close()

	res = []models.Article{}
	for rows.Next() {
		var item models.Article
		var authorID string
		er := rows.Scan(
			&item.ID,
			&item.Title,
			&item.Body,
			&authorID,
		)
		if er != nil {
			err = er
			return
		}

		item.Author = models.Author{ID: authorID}
		res = append(res, item)
	}

	return
}

func (h ArticleHandler) getAuthorByID(authorID string) (res models.Author, err error) {
	query := `SELECT id, name FROM authors WHERE id=?`
	row := h.DB.QueryRow(query, authorID)
	err = row.Scan(
		&res.ID,
		&res.Name,
	)
	return
}
