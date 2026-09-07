package app

func Boot() {
	// Initialize the application server
	app := NewApplicationServer()

	// Start the application server and listen for incoming requests
	app.StartServer()

	// Stop the application server when the application is terminated
	app.StopServer()
}
