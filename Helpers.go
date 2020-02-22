/*******************************************************************************
** @Author:					Thomas Bouder <Tbouder>
** @Email:					Tbouder@protonmail.com
** @Date:					Sunday 12 January 2020 - 17:45:47
** @Filename:				Helpers.go
**
** @Last modified by:		Tbouder
** @Last modified time:		Saturday 22 February 2020 - 11:46:08
*******************************************************************************/

package			main

import			"fmt"
import			"crypto/rand"
import			"encoding/base64"

func	generateByte(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	return b, err

}
func	generateUUID(n uint32) (string, error) {
	b, err := generateByte(n)
	if (err != nil) {
		return ``, err
	}
	uuid := fmt.Sprintf(
		"%x-%x-%x-%x-%x-%x-%x-%x-%x-%x-%x-%x-%x-%x-%x-%x",
		b[0:2], b[2:4], b[4:6], b[6:8], b[8:10], b[10:12], b[12:14], b[14:16], b[16:18], b[18:20], b[20:22], b[22:24], b[24:26], b[26:28], b[28:30], b[30:32], 
	)
	return uuid, nil
}

func	generateNonce(n uint32) (string, error) {
	b, err := generateByte(n)
    if (err != nil) {
        return ``, err
	}
    ciphertext := base64.RawStdEncoding.EncodeToString(b)
    return ciphertext, nil
}
