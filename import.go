package utilities

import (
	"bytes"
	"encoding/base64"

	"github.com/Roshankumar0808/goinsta"
)


func ImportFromBytes(inputBytes []byte) (*goinsta.Instagram, error) {
	return goinsta.ImportReader(bytes.NewReader(inputBytes))
}


func ImportFromBase64String(base64String string) (*goinsta.Instagram, error) {
	sDec, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		return nil, err
	}

	return ImportFromBytes(sDec)
}
