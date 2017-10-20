package api

import (
	"github.com/Nikita-Boyarskikh/DB/db"
	"github.com/Nikita-Boyarskikh/DB/models"
	"github.com/jackc/pgx"
	"github.com/mailru/easyjson"
	"github.com/mailru/easyjson/opt"
	"github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
)

func UserRouter(user *routing.RouteGroup) {
	user.Post("/<nickname>/create", func(ctx *routing.Context) error {
		logApi(ctx)

		var (
			nickname = ctx.Param("nickname")
			user     = models.User{
				Nickname: opt.OString(nickname),
			}
		)

		if err := easyjson.Unmarshal(ctx.PostBody(), &user); err != nil {
			log.Println("\t400:\t", err)
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			ctx.WriteData(err.Error())
			return nil
		}

		users, err := db.GetUsersByEmailAndNickname(user.Email, nickname)
		if len(users) > 0 {
			json, err := easyjson.Marshal(users)
			if err != nil {
				log.Println("\t500:\t", err)
				return err
			}

			ctx.SetStatusCode(fasthttp.StatusConflict)
			ctx.SetContentType("application/json")
			ctx.SetBody(json)

			log.Println("\t409\t", string(json))
			return nil
		} else if err != nil {
			log.Println("\t500:\t", err)
			return err
		}

		if err := db.CreateUser(user); err != nil {
			log.Println("\t500:\t", err)
			return err
		}

		json, err := easyjson.Marshal(user)
		if err != nil {
			log.Println("\t500:\t", err)
			return err
		}

		ctx.SetStatusCode(fasthttp.StatusCreated)
		ctx.SetContentType("application/json")
		ctx.SetBody(json)
		log.Println("\t409\t", string(json))
		return nil
	})

	user.Post("/<nickname>/profile", func(ctx *routing.Context) error {
		logApi(ctx)

		var (
			nickname = ctx.Param("nickname")
			user     = models.User{
				Nickname: opt.OString(nickname),
			}
		)

		if err := easyjson.Unmarshal(ctx.PostBody(), &user); err != nil {
			log.Println("\t400:\t", err)
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			ctx.WriteData(err.Error())
			return nil
		}

		users, err := db.GetUsersByEmailAndNickname(user.Email, nickname)
		if len(users) > 1 {
			message, err := easyjson.Marshal(models.Error{"Email or nickname conflict with existing users"})
			if err != nil {
				log.Println("\t500:\t", err)
				return err
			}

			ctx.SetStatusCode(fasthttp.StatusConflict)
			ctx.SetContentType("application/json")
			ctx.SetBody(message)

			log.Println("\t409\t", string(message))
			return nil
		} else if len(users) == 0 {
			message, err := easyjson.Marshal(models.Error{"User with requested nickname is not found"})
			if err != nil {
				log.Println("\t500:\t", err)
				return err
			}

			ctx.SetStatusCode(fasthttp.StatusNotFound)
			ctx.SetContentType("application/json")
			ctx.SetBody(message)

			log.Println("\t404\t", string(message))
			return nil
		} else if err != nil {
			log.Println("\t500:\t", err)
			return err
		}

		updatedUser, err := db.UpdateUser(user)
		if err != nil {
			log.Println("\t500:\t", err)
			return err
		}

		json, err := easyjson.Marshal(updatedUser)
		if err != nil {
			log.Println("\t500:\t", err)
			return err
		}

		ctx.Success("application/json", json)

		log.Println("\t200\t", string(json))
		return nil
	})

	user.Get("/<nickname>/profile", func(ctx *routing.Context) error {
		logApi(ctx)

		nickname := ctx.Param("nickname")

		user, err := db.GetUserByNickname(nickname)
		if err != nil {
			if err == pgx.ErrNoRows {
				json, err := easyjson.Marshal(models.Error{"Requested user is not found"})
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

		json, err := easyjson.Marshal(user)
		if err != nil {
			log.Println("\t500:\t", err)
			return err
		}

		ctx.Success("application/json", json)

		log.Println("\t200\t", string(json))
		return nil
	})
}
