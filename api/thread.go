package api

import (
	"strconv"

	"github.com/Nikita-Boyarskikh/DB/db"
	"github.com/jackc/pgx"
	"github.com/mailru/easyjson"
	"github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
)

func ThreadRouter(thread *routing.RouteGroup) {
	thread.Get("/<slug_or_id>/details", func(ctx *routing.Context) error {
		logApi(ctx)

		slug_or_id := ctx.Param("slug_or_id")
		id, _ := strconv.Atoi(slug_or_id)

		if thread, err := db.GetThreadBySlugOrID(slug_or_id, int32(id)); err == nil {
			json, err := easyjson.Marshal(thread)
			if err != nil {
				log.Println("\t500:\t", err)
				return err
			}

			ctx.Success("application/json", json)
			log.Println("\t200\t", string(json))
			return nil
		} else if err != pgx.ErrNoRows {
			log.Println("\t500:\t", err)
			return err
		}

		json, err := easyjson.Marshal(Error{"Thread with requested slug or id is not found"})
		if err != nil {
			log.Println("\t500:\t", err)
			return err
		}

		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetContentType("application/json")
		ctx.SetBody(json)
		log.Println("\t404\t", string(json))
		return nil
	})

	thread.Post("/<slug_or_id>/details", func(ctx *routing.Context) error {
		logApi(ctx)

		slug_or_id := ctx.Param("slug_or_id")

		id, _ := strconv.Atoi(slug_or_id)

		if _, err := db.GetThreadBySlugOrID(slug_or_id, int32(id)); err != nil {
			if err == pgx.ErrNoRows {
				json, err := easyjson.Marshal(Error{"Thread with requested slug or ID is not found"})
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

		var patch db.PatchThread
		if err := easyjson.Unmarshal(ctx.PostBody(), &patch); err != nil {
			log.Println("\t400:\t", err)
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			ctx.WriteData(err.Error())
			return nil
		}

		thread, err := db.UpdateThreadBySlugOrID(slug_or_id, patch)
		if err != nil {
			log.Println("\t500:\t", err)
			return err
		}

		json, err := easyjson.Marshal(thread)
		if err != nil {
			log.Println("\t500:\t", err)
			return err
		}

		ctx.Success("application/json", json)
		log.Println("\t200\t", string(json))
		return nil
	})

	thread.Post("/<slug_or_id>/vote", func(ctx *routing.Context) error {
		logApi(ctx)

		slug_or_id := ctx.Param("slug_or_id")
		id, _ := strconv.Atoi(slug_or_id)

		t, err := db.GetThreadBySlugOrID(slug_or_id, int32(id))
		if err != nil {
			if err == pgx.ErrNoRows {
				json, err := easyjson.Marshal(Error{"Thread with requested slug or ID is not found"})
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

		var vote db.Vote
		if err := easyjson.Unmarshal(ctx.PostBody(), &vote); err != nil {
			log.Println("\t400:\t", err)
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			ctx.WriteData(err.Error())
			return nil
		}

		_, err = db.GetUserByNickname(vote.Nickname)
		if err != nil {
			if err == pgx.ErrNoRows {
				json, err := easyjson.Marshal(Error{"User with requested nickname is not found"})
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

		thread, err := db.VoteForThread(t.ID.V, vote)
		if err != nil {
			log.Println("\t500:\t", err)
			return err
		}

		json, err := easyjson.Marshal(thread)
		if err != nil {
			log.Println("\t500:\t", err)
			return err
		}

		ctx.Success("application/json", json)
		log.Println("\t200\t", string(json))
		return nil
	})

	thread.Post("/<slug_or_id>/create", func(ctx *routing.Context) error {
		logApi(ctx)

		slug_or_id := ctx.Param("slug_or_id")
		id, _ := strconv.Atoi(slug_or_id)

		t, err := db.GetThreadBySlugOrID(slug_or_id, int32(id))
		if err != nil {
			if err == pgx.ErrNoRows {
				json, err := easyjson.Marshal(Error{"Thread with requested slug or ID is not found"})
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

		var posts db.Posts
		if err := easyjson.Unmarshal(ctx.PostBody(), &posts); err != nil {
			log.Println("\t400:\t", err)
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			ctx.WriteData(err.Error())
			return nil
		}

		var nicknames []string
		for _, post := range posts {
			nicknames = append(nicknames, post.Author)
		}

		if len(nicknames) > 0 {
			exists, _ := db.CheckAllUsersExists(nicknames)
			if !exists {
				json, err := easyjson.Marshal(Error{"Can't find any post authors"})
				if err != nil {
					log.Println("\t500:\t", err)
					return err
				}

				ctx.SetStatusCode(fasthttp.StatusNotFound)
				ctx.SetContentType("application/json")
				ctx.SetBody(json)
				log.Println("\t404\t", string(json))
				return nil
			}
		}

		var otherThread bool
		if otherThread, err = db.CheckAllPostsInOneThread(t.ID.V, posts); err != nil {
			log.Println("\t500:\t", err)
			return err
		} else if !otherThread {
			json, err := easyjson.Marshal(Error{"Parent post was created in another thread"})
			if err != nil {
				log.Println("\t500:\t", err)
				return err
			}

			ctx.SetStatusCode(fasthttp.StatusConflict)
			ctx.SetContentType("application/json")
			ctx.SetBody(json)

			log.Println("\t409:\t")
			return nil
		}

		createdPosts, err := db.CreatePostsInThread(t.Forum.V, t.ID.V, posts)
		if err != nil {
			log.Println("\t500:\t", err)
			return err
		}

		json, err := easyjson.Marshal(createdPosts)
		if err != nil {
			log.Println("\t500:\t", err)
			return err
		}

		ctx.SetStatusCode(fasthttp.StatusCreated)
		ctx.SetContentType("application/json")
		ctx.SetBody(json)
		log.Println("\t201\t", string(json))
		return nil
	})

	thread.Get("/<slug_or_id>/posts", func(ctx *routing.Context) error {
		logApi(ctx)

		slug_or_id := ctx.Param("slug_or_id")
		id, _ := strconv.Atoi(slug_or_id)

		t, err := db.GetThreadBySlugOrID(slug_or_id, int32(id))
		if err != nil {
			if err == pgx.ErrNoRows {
				json, err := easyjson.Marshal(Error{"Thread with requested slug or ID is not found"})
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

		limit, err := ctx.QueryArgs().GetUint("limit")
		if err != nil {
			log.Println("\t400:\t", err)
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			return nil
		}
		sinceRow := ctx.QueryArgs().Peek("since")
		descRaw := ctx.QueryArgs().Peek("desc")
		sort := ctx.QueryArgs().Peek("sort")
		desc := string(descRaw) == "true"

		var since int64
		sinceInt, err := strconv.Atoi(string(sinceRow))
		if err != nil {
			since = -1
		} else {
			since = int64(sinceInt)
		}

		posts, err := db.GetPosts(t.ID.V, limit, since, string(sort), desc)
		if err != nil {
			log.Println("\t500:\t", err)
			return err
		}

		if len(posts) == 0 {
			posts = db.Posts{}
		}

		json, err := easyjson.Marshal(posts)
		if err != nil {
			log.Println("\t500:\t", err)
			return err
		}

		ctx.Success("application/json", json)
		log.Println("\t200\t", string(json))
		return nil
	})
}
