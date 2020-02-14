/*******************************************************************************
** @Author:					Thomas Bouder <Tbouder>
** @Email:					Tbouder@protonmail.com
** @Date:					Thursday 09 January 2020 - 19:39:17
** @Filename:				Members.go
**
** @Last modified by:		Tbouder
** @Last modified time:		Friday 14 February 2020 - 18:18:23
*******************************************************************************/

package			main

import			"context"
import			"encoding/json"
import			"github.com/microgolang/logs"
import			"github.com/valyala/fasthttp"
import			"github.com/panghostlin/SDK/Members"

func	setCookieAndResolve(ctx *fasthttp.RequestCtx, cookie, hashKey string, err error) {
	if (err != nil) {
		ctx.Response.SetStatusCode(500)
		json.NewEncoder(ctx).Encode(false)
		return
	}

	setCookie(ctx, `accessToken`, cookie)
	setCookie(ctx, `hashKey`, hashKey)
	ctx.Response.Header.SetContentType(`application/json`)
	ctx.Response.SetStatusCode(200)
	json.NewEncoder(ctx).Encode(true)
}

/******************************************************************************
**	Router proxy function to create a member.
**	Call the Members Microservice to create a new member.
******************************************************************************/
func	createNewMember(ctx *fasthttp.RequestCtx) {
	req := &members.CreateMemberRequest{}
	json.Unmarshal(ctx.PostBody(), &req)

	result, err := clients.members.CreateMember(context.Background(), req)
	if (err != nil) {
		logs.Error(`LoginMember : fail to login member`, err)
		setCookieAndResolve(ctx, ``, ``, err)
	} else {
		setCookieAndResolve(ctx, result.GetAccessToken().Value, result.GetHashKey(), err)
	}

}

/******************************************************************************
**	Router proxy function to login a member.
**	Call the Members Microservice to login an existing member.
******************************************************************************/
func	loginMember(ctx *fasthttp.RequestCtx) {
	req := &members.LoginMemberRequest{}
	json.Unmarshal(ctx.PostBody(), &req)

	result, err := clients.members.LoginMember(context.Background(), req)
	if (err != nil) {
		logs.Error(`LoginMember : fail to login member`, err)
		setCookieAndResolve(ctx, ``, ``, err)
	} else {
		setCookieAndResolve(ctx, result.GetAccessToken().Value, result.GetHashKey(), err)
	}
}

/******************************************************************************
**	Router proxy function to login a member.
**	Call the Members Microservice to login an existing member.
******************************************************************************/
func	getMember(ctx *fasthttp.RequestCtx) {
	req := &members.GetMemberRequest{}
	req.MemberID = ctx.UserValue(`memberID`).(string)
	req.HashKey = string(ctx.UserValue(`hashKey`).([]byte))

	result, err := clients.members.GetMember(context.Background(), req)
	resolve(ctx, result.GetMember(), err)
}


/******************************************************************************
**	CheckMember
**	Router proxy function to check if a member exists and is connected
******************************************************************************/
func	checkMember(ctx *fasthttp.RequestCtx) {
	ctx.Response.SetStatusCode(200)
	json.NewEncoder(ctx).Encode(true)
}
