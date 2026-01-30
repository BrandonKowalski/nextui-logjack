package utils

import (
	"image/color"
	"os"

	"github.com/skip2/go-qrcode"
)

func CreateTempQRCode(content string, size int) (string, error) {
	qr, err := qrcode.New(content, qrcode.Medium)
	if err != nil {
		return "", err
	}

	qr.BackgroundColor = color.Black
	qr.ForegroundColor = color.White
	qr.DisableBorder = true

	tempFile, err := os.CreateTemp("", "qrcode-*")
	if err != nil {
		return "", err
	}

	err = qr.Write(size, tempFile)
	if err != nil {
		tempFile.Close()
		return "", err
	}
	tempFile.Close()

	return tempFile.Name(), nil
}
