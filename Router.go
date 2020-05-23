/*******************************************************************************
** @Author:					Thomas Bouder <Tbouder>
** @Email:					Tbouder@protonmail.com
** @Date:					Tuesday 04 February 2020 - 15:40:06
** @Filename:				Router.go
**
** @Last modified by:		Tbouder
** @Last modified time:		Wednesday 01 April 2020 - 12:03:26
*******************************************************************************/

package			main

import			"github.com/microgolang/logs"
import			"github.com/valyala/fasthttp"
import			"github.com/buaazp/fasthttprouter"
import			"github.com/fasthttp/websocket"
import			"github.com/panghostlin/SDK/Pictures"
import			"encoding/json"
// import			"bytes"

var fastupgrader = websocket.FastHTTPUpgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(ctx *fasthttp.RequestCtx) bool {
		return true
	},
}


func	resolve(ctx *fasthttp.RequestCtx, data interface{}, err error, errCode ...int) {
	if (err != nil) {
		ctx.Response.Header.SetContentType(`application/json`)
		if (len(errCode) == 0) {
			ctx.Response.SetStatusCode(404)
		} else {
			ctx.Response.SetStatusCode(errCode[0])
		}
		json.NewEncoder(ctx).Encode(false)
		return
	}
	ctx.Response.Header.SetContentType(`application/json`)
	ctx.Response.SetStatusCode(200)
	json.NewEncoder(ctx).Encode(data)
}
func	resolvePicture(ctx *fasthttp.RequestCtx, resp *pictures.DownloadPictureResponse, err error) {
	if (err != nil) {
		ctx.Response.SetStatusCode(403)
		ctx.Write([]byte{})
		return
	}
	// ctx.Response.Header.SetContentType(`application/octet-stream`)
	// ctx.Response.SetStatusCode(200)
	type	sRet struct {
		Picture		[]byte
		Key			string
		IV			string
		Preview		string
	}

	// response.GetChunk(), response.GetContentType()
	toRet := &sRet{
		Picture: resp.GetChunk(),
		Key: resp.GetCrypto().GetKey(),
		IV: resp.GetCrypto().GetIV(),
		Preview: resp.GetPreview(), //`LEHV6nWB2yk8pyo0adR*.7kCMdnj`,
	}
	// ctx.Write(reqBodyBytes.Bytes())


	ctx.Response.Header.SetContentType(`application/json`)
	ctx.Response.SetStatusCode(200)
	json.NewEncoder(ctx).Encode(toRet)
}

func	withAuth(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		accessToken := ctx.Request.Header.Cookie(`accessToken`)

		if (accessToken == nil) {
			logs.Error(`No token`)
			ctx.Error(fasthttp.StatusMessage(fasthttp.StatusUnauthorized), fasthttp.StatusUnauthorized)
			return
		}

		isSuccess, memberID := checkMemberCookie(ctx, string(accessToken))
		if (!isSuccess && false) {
			ctx.Error(fasthttp.StatusMessage(fasthttp.StatusUnauthorized), fasthttp.StatusUnauthorized)
			return
		}
		hashKey := ctx.Request.Header.Cookie(`hashKey`)

		ctx.SetUserValue(`memberID`, memberID)
		ctx.SetUserValue(`hashKey`, hashKey)
		h(ctx)
	})
}

func	initRouter() func(*fasthttp.RequestCtx) {
	router := fasthttprouter.New()
	router.POST("/newMember/", createNewMember)
	router.POST("/loginMember/", loginMember)
	router.POST("/checkMember/", withAuth(checkMember))
	router.POST("/getMember/", withAuth(getMember))

	router.POST("/uploadPicture/", withAuth(uploadPicture))
	router.GET("/ws/uploadPicture/:fileUUID", wsUploadPicture)
	router.GET("/ws/uploadPicture/", wsUploadPicture)

	router.GET("/downloadPreview/:pictureID", withAuth(downloadPreview))
	router.GET("/downloadPicture/:pictureSize/:pictureID", withAuth(downloadPicture))
	router.POST("/deletePictures/", withAuth(deletePictures))

	router.POST("/pictures/getby/member/", withAuth(listPicturesByMember))
	router.POST("/pictures/getby/album/", withAuth(listPicturesByAlbum))
	router.POST("/pictures/set/album/", withAuth(setPicturesAlbum))
	router.POST("/pictures/set/date/", withAuth(setPicturesDate))

	router.POST("/albums/create/", withAuth(createAlbum))
	router.POST("/albums/list/", withAuth(listAlbums))
	router.POST("/albums/get/", withAuth(getAlbum))
	router.POST("/albums/delete/", withAuth(deleteAlbum))
	router.POST("/albums/set/cover/", withAuth(setAlbumCover))
	router.POST("/albums/set/name/", withAuth(setAlbumName))

	return router.Handler
}
