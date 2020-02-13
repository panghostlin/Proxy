/*******************************************************************************
** @Author:					Thomas Bouder <Tbouder>
** @Email:					Tbouder@protonmail.com
** @Date:					Tuesday 04 February 2020 - 15:40:06
** @Filename:				Router.go
**
** @Last modified by:		Tbouder
** @Last modified time:		Thursday 13 February 2020 - 19:47:51
*******************************************************************************/

package			main

import			_ "os"
import			"github.com/microgolang/logs"
import			"github.com/valyala/fasthttp"
import			"github.com/buaazp/fasthttprouter"
import			"github.com/fasthttp/websocket"
import			"encoding/json"

var fastupgrader = websocket.FastHTTPUpgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(ctx *fasthttp.RequestCtx) bool {
		return true
	},
}

func	resolve(ctx *fasthttp.RequestCtx, data interface{}, err error) {
	if (err != nil) {
		ctx.Response.Header.SetContentType(`application/json`)
		ctx.Response.SetStatusCode(404)
		json.NewEncoder(ctx).Encode(false)
		return
	}
	ctx.Response.Header.SetContentType(`application/json`)
	ctx.Response.SetStatusCode(200)
	json.NewEncoder(ctx).Encode(data)
}
func	resolvePicture(ctx *fasthttp.RequestCtx, data []byte, contentType string, err error) {
	if (err != nil) {
		ctx.Response.SetStatusCode(404)
		ctx.Write([]byte{})
		return
	}
	ctx.Response.Header.SetContentType(contentType)
	ctx.Response.SetStatusCode(200)
	ctx.Write(data)
}

func	withAuth(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		accessToken := ctx.Request.Header.Cookie(`accessToken`)

		if (accessToken == nil) {
			logs.Error(`No token`)
			ctx.Error(fasthttp.StatusMessage(fasthttp.StatusUnauthorized), fasthttp.StatusUnauthorized)
			return
		}

		isSuccess, memberID := CheckMemberCookie(ctx, string(accessToken))
		if (!isSuccess) {
			ctx.Error(fasthttp.StatusMessage(fasthttp.StatusUnauthorized), fasthttp.StatusUnauthorized)
			return
		}
		hashKey := ctx.Request.Header.Cookie(`hashKey`)

		ctx.SetUserValue(`memberID`, memberID)
		ctx.SetUserValue(`hashKey`, hashKey)
		h(ctx)
	})
}

func	InitRouter() func(*fasthttp.RequestCtx) {
	router := fasthttprouter.New()
	router.POST("/newMember/", createNewMember)
	router.POST("/loginMember/", loginMember)
	router.POST("/checkMember/", withAuth(checkMember))

	router.POST("/uploadPicture/", withAuth(uploadPicture))
	router.GET("/ws/uploadPicture/", wsUploadPicture)
	router.GET("/downloadPicture/:pictureSize/:pictureID", withAuth(downloadPicture))
	router.POST("/deletePictures/", withAuth(deletePictures))

	router.POST("/pictures/getby/member/", withAuth(listPicturesByMember))
	router.POST("/pictures/getby/album/", withAuth(listPicturesByAlbum))
	router.POST("/pictures/set/album/", withAuth(setPicturesAlbum))

	router.POST("/albums/create/", withAuth(createAlbum))
	router.POST("/albums/list/", withAuth(listAlbums))
	router.POST("/albums/get/", withAuth(getAlbum))
	router.POST("/albums/delete/", withAuth(deleteAlbum))
	router.POST("/albums/set/cover/", withAuth(setAlbumCover))
	router.POST("/albums/set/name/", withAuth(setAlbumName))

	return router.Handler
}
