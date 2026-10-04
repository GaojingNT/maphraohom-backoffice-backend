//go:debug x509negativeserial=1
package main

//	@title						Maphraohom Backoffice API
//	@version					1.0
//	@description				REST API for the Maphraohom backoffice: authentication, users, stores, products, customers, bills and files.
//
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Enter the JWT from /api/v1/auth/sign-in as "Bearer <token>".

import "maphraohom.app/maphraohom-backoffice/app"

func main() {
	// Start application server
	app.Boot()
}
