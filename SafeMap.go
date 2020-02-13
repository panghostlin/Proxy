/*******************************************************************************
** @Author:					Thomas Bouder <Tbouder>
** @Email:					Tbouder@protonmail.com
** @Date:					Thursday 09 January 2020 - 19:35:28
** @Filename:				SafeMap.go
**
** @Last modified by:		Tbouder
** @Last modified time:		Thursday 13 February 2020 - 19:52:48
*******************************************************************************/

package			main

import			"sync"
import			"github.com/fasthttp/websocket"

type regularIntMap struct {
	sync.RWMutex
	internal	map[string][]([]byte)
	len			map[string](uint)
	wsConn		map[string](*websocket.Conn)
	wsConnOpen	map[string](bool)
}

func newRegularIntMap() *regularIntMap {
	return &regularIntMap{
		internal: make(map[string][]([]byte)),
		len: make(map[string](uint)),
		wsConn: make(map[string](*websocket.Conn)),
		wsConnOpen: make(map[string](bool)),
	}
}

func (rm *regularIntMap) loadContent(key string) (value []([]byte), ok bool) {
	rm.RLock()
	result, ok := rm.internal[key]
	rm.RUnlock()
	return result, ok
}
func (rm *regularIntMap) loadLen(key string) (value uint, ok bool) {
	rm.RLock()
	result, ok := rm.len[key]
	rm.RUnlock()
	return result, ok
}
func (rm *regularIntMap) loadWs(key string) (value *websocket.Conn, isOpen bool, ok bool) {
	rm.RLock()
	result, ok := rm.wsConn[key]
	resultOpen, ok := rm.wsConnOpen[key]
	rm.RUnlock()
	return result, resultOpen, ok
}


func (rm *regularIntMap) initLen(key string) {
	rm.Lock()
	rm.len[key] = 0
	rm.Unlock()
}
func (rm *regularIntMap) initWs(key string, value *websocket.Conn) {
	rm.Lock()
	rm.wsConn[key] = value
	rm.wsConnOpen[key] = true
	rm.Unlock()
}


func (rm *regularIntMap) setContent(key string, value []([]byte)) {
	rm.Lock()
	rm.internal[key] = value
	rm.Unlock()
}
func (rm *regularIntMap) setWs(key string, value *websocket.Conn) {
	rm.Lock()
	rm.wsConn[key] = value
	rm.Unlock()
}
func (rm *regularIntMap) incLen(key string) {
	rm.Lock()
	rm.len[key] += 1
	rm.Unlock()
}


func (rm *regularIntMap) delete(key string) {
	rm.Lock()
	delete(rm.internal, key)
	delete(rm.len, key)
	rm.wsConnOpen[key] = false
	rm.Unlock()
}


var rm = newRegularIntMap()

