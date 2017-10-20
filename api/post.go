package api

import (
	"github.com/Nikita-Boyarskikh/DB/db"
	"github.com/jackc/pgx"
	"github.com/mailru/easyjson"
	"github.com/mailru/easyjson/opt"
	"github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
	"strconv"
	"strings"
)

//easyjson:json
type EditPost struct {
	Message opt.String
}

func PostRouter(post *routing.RouteGroup) {
	post.Post("/<id>/details", func(ctx *routing.Context) error {
		logApi(ctx)

		strId := ctx.Param("id")
		intId, _ := strconv.Atoi(strId)
		id := int64(intId)

		exPost, err := db.GetPost(id)
		if err != nil {
			if err == pgx.ErrNoRows {
				json, err := easyjson.Marshal(Error{"Post with requested ID is not found"})
				if err != nil {
					log.Println("\t500:\t", err)
					return err
				}

				ctx.SetStatusCode(fasthttp.StatusNotFound)
				ctx.SetContentType("application/json")
				ctx.SetBody(json)
				log.Println("\t404\t", string(json))
				return nil
			} else {
				log.Println("\t500:\t", err)
				return err
			}
		}

		var editPost EditPost
		if err := easyjson.Unmarshal(ctx.PostBody(), &editPost); err != nil {
			log.Println("\t400:\t", err)
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			ctx.WriteData(err.Error())
			return nil
		}

		var post db.Post
		if editPost.Message.Defined {
			post, err = db.UpdatePostMessage(id, editPost.Message.V, editPost.Message.V != exPost.Message)
			if err != nil {
				log.Println("\t500:\t", err)
				return err
			}
		} else {
			post, err = db.GetPost(id)
			if err != nil {
				log.Println("\t500:\t", err)
				return err
			}
		}

		json, err := easyjson.Marshal(post)
		if err != nil {
			log.Println("\t500:\t", err)
			return err
		}

		ctx.Success("application/json", json)
		log.Println("\t200\t", string(json))
		return nil
	})

	post.Get("/<id>/details", func(ctx *routing.Context) error {
		logApi(ctx)

		strId := ctx.Param("id")
		intId, _ := strconv.Atoi(strId)
		id := int64(intId)

		post, err := db.GetPost(id)
		if err != nil {
			if err == pgx.ErrNoRows {
				json, err := easyjson.Marshal(Error{"Post with requested ID is not found"})
				if err != nil {
					log.Println("\t500:\t", err)
					return err
				}

				ctx.SetStatusCode(fasthttp.StatusNotFound)
				ctx.SetContentType("application/json")
				ctx.SetBody(json)
				log.Println("\t404\t", string(json))
				return nil
			} else {
				log.Println("\t500:\t", err)
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
			author *db.User
			thread *db.Thread
			forum  *db.Forum
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

		json, err = easyjson.Marshal(db.PostFull{
			Post:   post,
			Author: author,
			Thread: thread,
			Forum:  forum,
		})
		if err != nil {
			log.Println("\t500:\t", err)
			return err
		}

		ctx.Success("application/json", json)
		log.Println("\t200\t", string(json))
		return nil
	})
}
