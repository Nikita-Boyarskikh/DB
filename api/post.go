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

		exPost, err := db.GetPost(id)
		if err != nil {
			if err == pgx.ErrNoRows {
				json, err := easyjson.Marshal(models.Error{"Post with requested ID is not found"})
				if err != nil {
					return err
				}

				ctx.SetStatusCode(fasthttp.StatusNotFound)
				ctx.SetContentType("application/json")
				ctx.SetBody(json)
				return nil
			} else {
				return err
			}
		}

		var editPost models.EditPost
		if err := easyjson.Unmarshal(ctx.PostBody(), &editPost); err != nil {
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			ctx.WriteData(err.Error())
			return nil
		}

		var post models.Post
		if editPost.Message.Defined {
			post, err = db.UpdatePostMessage(id, editPost.Message.V, editPost.Message.V != exPost.Message)
			if err != nil {
				return err
			}
		} else {
			post, err = db.GetPost(id)
			if err != nil {
				return err
			}
		}

		json, err := easyjson.Marshal(post)
		if err != nil {
			return err
		}

		ctx.Success("application/json", json)
		return nil
	})

	post.Get("/<id>/details", func(ctx *routing.Context) error {
		strId := ctx.Param("id")
		intId, _ := strconv.Atoi(strId)
		id := int64(intId)

		post, err := db.GetPost(id)
		if err != nil {
			if err == pgx.ErrNoRows {
				json, err := easyjson.Marshal(models.Error{"Post with requested ID is not found"})
				if err != nil {
					return err
				}

				ctx.SetStatusCode(fasthttp.StatusNotFound)
				ctx.SetContentType("application/json")
				ctx.SetBody(json)
				return nil
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
			authorObj, _ := db.GetUserByNickname(post.Author)
			author = &authorObj
		}
		if relObjs["thread"] {
			threadObj, _ := db.GetThreadBySlugOrID(string(post.Thread.V), post.Thread.V)
			thread = &threadObj
		}
		if relObjs["forum"] {
			_, forumObj, _ := db.GetForumBySlug(post.Forum.V)
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
		return nil
	})
}
