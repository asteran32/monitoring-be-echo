package rest

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func RunAPIWithHandler() {
	e := echo.New()

	// Handler
	h, err := NewHandler()
	if err != nil {
		return
	}

	// Middleware
	e.Use(middleware.Logger()) // Log all request log
	e.Use(middleware.Recover())

	//CORS Middleware
	// e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
	// 	AllowOrigins: []string{"*"},
	// 	AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	// }))

	// Serve SPA app
	// This serves static files from "Path" directory and enables directory browing.
	// Default behavior when using with non root URL(refresh, reload etc) paths is to append the URL path to filesystem path
	// For example, when an incoming request comes for '/somepath' the actual filesystem request goes to 'filesystempath/somepath instead of only 'filesystempath'
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:   "build",
		Index:  "index.html",
		Browse: false,
		HTML5:  true,
	}))

	// Router
	user := e.Group("/user")
	{
		user.POST("/signin", h.SignIn)
		user.POST("/signup", h.SignUp)
	}

	monitor := e.Group("/monitor")
	{
		monitor.GET("/cams", h.GetAllCam)
		monitor.POST("/cams", h.AddNewCam)
		monitor.GET("/cams/:id", h.GetCurrentCam)
		monitor.DELETE("/cams/:id", h.DeleteCurrentCam)
		monitor.GET("/stream/:id", h.StreamRTSP)

	}

	opcua := e.Group("/server")
	{
		// opcua.GET("/stream/:id", h.MonitoringOpcUA) //-> client/:id랑 병합
		opcua.GET("/client", h.GetAllServer)
		opcua.POST("/client", h.AddNewServer)
		opcua.GET("/client/:id", h.GetCurrentServer)
		opcua.DELETE("/client/:id", h.DeleteCurrentServer)
	}

	// Start server
	e.Logger.Fatal(e.Start(":5000"))
}
