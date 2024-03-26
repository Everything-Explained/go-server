package routes

import (
	"github.com/Everything-Explained/go-server/configs"
	"github.com/Everything-Explained/go-server/internal/router"
)

func HandleAssets(r *router.Router) {
	assetDir := configs.GetConfig().ClientPath + "/assets"
	r.GetStatic("/assets", assetDir)
}
