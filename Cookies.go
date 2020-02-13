/*******************************************************************************
** @Author:					Thomas Bouder <Tbouder>
** @Email:					Tbouder@protonmail.com
** @Date:					Saturday 11 January 2020 - 17:11:49
** @Filename:				Cookies.go
**
** @Last modified by:		Tbouder
** @Last modified time:		Thursday 13 February 2020 - 19:50:20
*******************************************************************************/

package			main

import			"time"
import			"context"
import			"github.com/valyala/fasthttp"
import			"github.com/microgolang/logs"
import			"github.com/panghostlin/SDK/Members"

const	REFRESH_TOKEN_EXPIRATION_DURATION = (24 * time.Hour) * 31

func	setCookie(ctx *fasthttp.RequestCtx, key, value string) {
	cookie := &fasthttp.Cookie{}
	cookie.SetKey(key)
	cookie.SetValue(value)
	cookie.SetPath(`/`)
	cookie.SetHTTPOnly(true)
	// cookie.SetSecure(true)
	// cookie.SetSameSite(fasthttp.CookieSameSiteNoneMode)
	// cookie.SetExpire(time.Unix(memberCookie.Expiration, 0))
	cookie.SetExpire(time.Now().Add(REFRESH_TOKEN_EXPIRATION_DURATION))
	ctx.Response.Header.SetCookie(cookie)
}

func	checkMemberCookie(ctx *fasthttp.RequestCtx, requestAccessToken string) (bool, string) {
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
	return result.Success, result.GetMemberID()
}