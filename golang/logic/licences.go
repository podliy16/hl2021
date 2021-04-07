package logic

import (
	"goldclient/httpclient"
	"goldclient/models"
)

const MAX_LICENSES = 10

const QUERY_LICENSES_INIT = 7

func SetupLicences(w *models.World) {

	for i := 0; i < QUERY_LICENSES_INIT; i++ {
		go httpclient.PostLicence(w.LicensesChannel, w, -1)
	}

}
