package logic

import (
	"context"
)

func printLogo(ctx context.Context) {
	log := clog(ctx)
	log.Println(`/////////////////////////////////////////////////`)
	log.Println(`//   __   ____  __     ____  ____  ____  ____  //`)
	log.Println(`//  / _\ (  _ \(  )___(_  _)(  __)/ ___)(_  _) //`)
	log.Println(`// /    \ ) __/ )((___) )(   ) _) \___ \  )(   //`)
	log.Println(`// \_/\_/(__)  (__)    (__) (____)(____/ (__)  //`)
	log.Println(`/////////////////////////////////////////////////`)
}
