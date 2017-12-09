package api

import (
	"strconv"
	"strings"

	"github.com/Nikita-Boyarskikh/DB/db"
	"github.com/Nikita-Boyarskikh/DB/models"
	"github.com/jackc/pgx"
	"github.com/mailru/easyjson"
	"github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
)

func PostRouter(post *routing.RouteGroup) {
	post.Post("/<id>/details", func(ctx *routing.Context) error {
		strId := ctx.Param("id")
		intId, _ := strconv.Atoi(strId)
		id := int64(intId)

		tx, err := db.GetConn().Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		exPost, err := db.GetPost(tx, id)
		if err != nil {
			if err == pgx.ErrNoRows {
				json, err := easyjson.Marshal(models.Error{"Post with requested ID is not found"})
				if err != nil {
					return err
				}

				ctx.SetStatusCode(fasthttp.StatusNotFound)
				ctx.SetContentType("application/json")
				ctx.SetBody(json)
				return tx.Commit()
			} else {
				return err
			}
		}

		var editPost models.EditPost
		if err := easyjson.Unmarshal(ctx.PostBody(), &editPost); err != nil {
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			ctx.WriteData(err.Error())
			return tx.Commit()
		}

		var post models.Post
		if editPost.Message.Defined {
			post, err = db.UpdatePostMessage(tx, id, editPost.Message.V, editPost.Message.V != exPost.Message)
			if err != nil {
				return err
			}
		} else {
			post = exPost
		}

		json, err := easyjson.Marshal(post)
		if err != nil {
			return err
		}

		ctx.Success("application/json", json)
		return tx.Commit()
	})

	post.Get("/<id>/details", func(ctx *routing.Context) error {
		strId := ctx.Param("id")
		intId, _ := strconv.Atoi(strId)
		id := int64(intId)

		tx, err := db.GetConn().Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		post, err := db.GetPost(tx, id)
		if err != nil {
			if err == pgx.ErrNoRows {
				json, err := easyjson.Marshal(models.Error{"Post with requested ID is not found"})
				if err != nil {
					return err
				}

				ctx.SetStatusCode(fasthttp.StatusNotFound)
				ctx.SetContentType("application/json")
				ctx.SetBody(json)
				return tx.Commit()
			} else {
				return err
			}
		}

		var json []byte
		rel := ctx.QueryArgs().Peek("related")
		relObjs := make(map[string]bool)
		for _, rel := range strings.Split(string(rel), ",") {
			relObjs[rel] = true
		}

		var (
			author *models.User
			thread *models.Thread
			forum  *models.Forum
		)

		if relObjs["user"] {
			authorObj, err := db.GetUserByNickname(tx, post.Author)
			if err != nil {
				return err
			}
			author = &authorObj
		}
		if relObjs["thread"] {
			threadObj, err := db.GetThreadBySlugOrID(tx, string(post.Thread.V), post.Thread.V)
			if err != nil {
				return err
			}
			thread = &threadObj
		}
		if relObjs["forum"] {
			_, forumObj, err := db.GetForumBySlug(tx, post.Forum.V)
			if err != nil {
				return err
			}
			forum = &forumObj
		}

		json, err = easyjson.Marshal(models.PostFull{
			Post:   post,
			Author: author,
			Thread: thread,
			Forum:  forum,
		})
		if err != nil {
			return err
		}

		ctx.Success("application/json", json)
		return tx.Commit()
	})
}
