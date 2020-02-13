/*******************************************************************************
** @Author:					Thomas Bouder <Tbouder>
** @Email:					Tbouder@protonmail.com
** @Date:					Thursday 09 January 2020 - 19:35:28
** @Filename:				SafeMap.go
**
** @Last modified by:		Tbouder
** @Last modified time:		Thursday 13 February 2020 - 12:44:20
*******************************************************************************/

package			main

import			"sync"
// import			"github.com/fasthttp/websocket"
import			"github.com/gorilla/websocket"

type RegularIntMap struct {
	sync.RWMutex
	internal	map[string][]([]byte)
	len			map[string](uint)
	wsConn		map[string](*websocket.Conn)
	wsConnOpen	map[string](bool)
}

func NewRegularIntMap() *RegularIntMap {
	return &RegularIntMap{
		internal: make(map[string][]([]byte)),
		len: make(map[string](uint)),
		wsConn: make(map[string](*websocket.Conn)),
		wsConnOpen: make(map[string](bool)),
	}
}

func (rm *RegularIntMap) LoadContent(key string) (value []([]byte), ok bool) {
	rm.RLock()
	result, ok := rm.internal[key]
	rm.RUnlock()
	return result, ok
}
func (rm *RegularIntMap) LoadLen(key string) (value uint, ok bool) {
	rm.RLock()
	result, ok := rm.len[key]
	rm.RUnlock()
	return result, ok
}
func (rm *RegularIntMap) LoadWs(key string) (value *websocket.Conn, isOpen bool, ok bool) {
	rm.RLock()
	result, ok := rm.wsConn[key]
	resultOpen, ok := rm.wsConnOpen[key]
	rm.RUnlock()
	return result, resultOpen, ok
}


func (rm *RegularIntMap) InitLen(key string) {
	rm.Lock()
	rm.len[key] = 0
	rm.Unlock()
}
func (rm *RegularIntMap) InitWs(key string, value *websocket.Conn) {
	rm.Lock()
	rm.wsConn[key] = value
	rm.wsConnOpen[key] = true
	rm.Unlock()
}


func (rm *RegularIntMap) SetContent(key string, value []([]byte)) {
	rm.Lock()
	rm.internal[key] = value
	rm.Unlock()
}
func (rm *RegularIntMap) SetWs(key string, value *websocket.Conn) {
	rm.Lock()
	rm.wsConn[key] = value
	rm.Unlock()
}
func (rm *RegularIntMap) IncLen(key string) {
	rm.Lock()
	rm.len[key] += 1
	rm.Unlock()
}


func (rm *RegularIntMap) Delete(key string) {
	rm.Lock()
	delete(rm.internal, key)
	delete(rm.len, key)
	rm.wsConnOpen[key] = false
	rm.Unlock()
}


var rm = NewRegularIntMap()

