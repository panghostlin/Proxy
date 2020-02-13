/*******************************************************************************
** @Author:					Thomas Bouder <Tbouder>
** @Email:					Tbouder@protonmail.com
** @Date:					Thursday 09 January 2020 - 19:39:17
** @Filename:				Members.go
**
** @Last modified by:		Tbouder
** @Last modified time:		Thursday 13 February 2020 - 19:12:45
*******************************************************************************/

package			main

import			"context"
import			"errors"
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
**	createMemberGRPC
**	Call the Members Microservice to create a new member.
**
**	CreateNewMember
**	Router proxy function to create a member.
******************************************************************************/
func	createMemberGRPC(data []byte) (string, *members.Cookie, string, error) {
	req := &members.CreateMemberRequest{}
	json.Unmarshal(data, &req)

	result, err := clients.members.CreateMember(context.Background(), req)
	if (err != nil) {
		logs.Error(`CreateMember : fail to communicate with microservice`, err)
		return ``, &members.Cookie{}, ``, err
	} else if (!result.GetSuccess()) {
		logs.Error(`CreateMember : fail to create the member`)
		return ``, &members.Cookie{}, ``, errors.New(`Failed to create member`)
	}
	return result.GetMemberID(), result.GetAccessToken(), result.GetHashKey(), nil
}
func	createNewMember(ctx *fasthttp.RequestCtx) {
	_, cookie, hashKey, err := createMemberGRPC(ctx.PostBody())
	setCookieAndResolve(ctx, cookie.Value, hashKey, err)
}


/******************************************************************************
**	loginMemberGRPC
**	Call the Members Microservice to login an existing member.
**
**	LoginMember
**	Router proxy function to login a member.
******************************************************************************/
func	loginMemberGRPC(data []byte) (string, *members.Cookie, string, error) {
	req := &members.LoginMemberRequest{}
	json.Unmarshal(data, &req)

	result, err := clients.members.LoginMember(context.Background(), req)
	if (err != nil) {
		logs.Error(`LoginMember : fail to communicate with microservice`, err)
		return ``, &members.Cookie{}, ``, err
	} else if (!result.GetSuccess()) {
		logs.Error(`LoginMember : fail to login the member`)
		return ``, &members.Cookie{}, ``, errors.New(`Failed to create member`)
	}
	return result.GetMemberID(), result.GetAccessToken(), result.GetHashKey(), nil
}
func	loginMember(ctx *fasthttp.RequestCtx) {
	_, cookie, hashKey, err := loginMemberGRPC(ctx.PostBody())
	setCookieAndResolve(ctx, cookie.Value, hashKey, err)
}

/******************************************************************************
**	CheckMember
**	Router proxy function to check if a member exists and is connected
******************************************************************************/
func	checkMember(ctx *fasthttp.RequestCtx) {
	ctx.Response.SetStatusCode(200)
	json.NewEncoder(ctx).Encode(true)
}