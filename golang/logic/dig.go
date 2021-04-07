package logic

import (
	"goldclient/httpclient"
	"goldclient/models"
)

func SendLicenseIfNeeded(w *models.World) {
	go httpclient.PostLicence(w.LicensesChannel, w, -1)
}

func GetLicense(w *models.World) models.LicenseData {
	license := <-w.LicensesChannel
	return license
}
