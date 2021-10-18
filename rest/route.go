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

	// e.Use(middleware.CORSWithConfig(middleware.CORSConfig{ //CORS Middleware
	// 	AllowOrigins: []string{"*"},
	// 	AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	// }))

	// Serve SPA app
	// This serves static files from "Path" directory and enables directory browing.
	// Default behavior when using with non root URL(refresh, reload etc) paths is to append the URL path to filesystem path
	// For example, when an incoming request comes for '/somepath' the actual filesystem request goes to 'filesystempath/somepath instead of only 'filesystempath'
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:   "public/build",
		Index:  "index.html",
		Browse: false,
		HTML5:  true,
	}))

	// Router List
	user := e.Group("/user")
	{
		user.POST("/signin", h.SignIn)
		user.POST("/signup", h.SignUp)
	}

	monitor := e.Group("/monitor")
	{
		monitor.GET("/cams/:id/rtsp", h.StreamRTSP) //
		monitor.GET("/cams", h.GetAllCam)
		monitor.GET("/cams/:id", h.GetCam)
		monitor.POST("/cams", h.AddNewCam)
		monitor.DELETE("/cams", h.DeleteCam)

	}

	opcua := e.Group("/opcua")
	{
		opcua.GET("/client/:id/analysis", h.MonitoringOpcUA) //
		opcua.GET("/client", h.GetAllServer)
		opcua.GET("/client/:id", h.GetServer)
		opcua.POST("/client", h.AddNewServer)
		opcua.DELETE("/client", h.DeleteServer)
	}

	// Start server
	e.Logger.Fatal(e.Start(":5000"))
}
