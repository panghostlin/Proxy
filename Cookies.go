/*******************************************************************************
** @Author:					Thomas Bouder <Tbouder>
** @Email:					Tbouder@protonmail.com
** @Date:					Saturday 11 January 2020 - 17:11:49
** @Filename:				Cookies.go
**
** @Last modified by:		Tbouder
** @Last modified time:		Tuesday 11 February 2020 - 20:45:56
*******************************************************************************/

package			main

import			"time"
import			"context"
import			"github.com/valyala/fasthttp"
import			"github.com/microgolang/logs"
import			"github.com/panghostlin/SDK/Members"

const	REFRESH_TOKEN_EXPIRATION_DURATION = (24 * time.Hour) * 31

func	SetHashKey(ctx *fasthttp.RequestCtx, memberID, hashKey string) {
	cookie := &fasthttp.Cookie{}
	cookie.SetKey(`hashKey`)
	cookie.SetValue(hashKey)
	cookie.SetPath(`/`)
	cookie.SetHTTPOnly(true)
	// cookie.SetSecure(true)
	// cookie.SetSameSite(fasthttp.CookieSameSiteNoneMode)
	// cookie.SetExpire(time.Unix(memberCookie.Expiration, 0))
	cookie.SetExpire(time.Now().Add(REFRESH_TOKEN_EXPIRATION_DURATION))
	ctx.Response.Header.SetCookie(cookie)
}

func	SetAccessToken(ctx *fasthttp.RequestCtx, memberID string, memberCookie *members.Cookie) {
	cookie := &fasthttp.Cookie{}
	cookie.SetKey(`accessToken`)
	cookie.SetValue(memberCookie.Value)
	cookie.SetPath(`/`)
	cookie.SetHTTPOnly(true)
	// cookie.SetSecure(true)
	// cookie.SetSameSite(fasthttp.CookieSameSiteNoneMode)
	// cookie.SetExpire(time.Unix(memberCookie.Expiration, 0))
	cookie.SetExpire(time.Now().Add(REFRESH_TOKEN_EXPIRATION_DURATION))
	ctx.Response.Header.SetCookie(cookie)
}

func	CheckMemberCookie(ctx *fasthttp.RequestCtx, requestAccessToken string) (bool, string) {
	req := &members.CheckAccessTokenRequest{AccessToken: requestAccessToken}
	result, err := clients.members.CheckAccessToken(context.Background(), req)

	if (err != nil) {
		logs.Error(`Fail to communicate with microservice`, err)
		return false, ``
	}
	if (!result.Success) {
		logs.Error(`Failed to check token`)
		return false, ``
	}
	// SetAccessToken(ctx, result.GetMemberID(), result.GetAccessToken())
	return result.Success, result.GetMemberID()
}