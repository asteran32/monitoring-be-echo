package rest

import (
	"app/db"
	"app/model"
	"app/service"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type Handler struct {
	db db.DBInterface
}

type HandlerInterface interface {
	// users
	SignUp(c echo.Context) error
	SignIn(c echo.Context) error
	// cameras
	StreamRTSP(c echo.Context) error
	AddNewCam(c echo.Context) error
	DeleteCam(c echo.Context) error
	GetAllCam(c echo.Context) error
	GetCam(c echo.Context) error
	// opcua servers
	MonitoringOpcUA(c echo.Context) error
	AddNewServer(c echo.Context) error
	DeleteServer(c echo.Context) error
	GetAllServer(c echo.Context) error
	GetServer(c echo.Context) error
}

func NewHandler() (HandlerInterface, error) {
	client, err := db.NewClient()
	if err != nil {
		return nil, err
	}
	return &Handler{db: client}, nil
}

// User Sign in
func (h *Handler) SignIn(c echo.Context) error {
	if h.db == nil {
		return c.JSON(http.StatusInternalServerError, "server database error")
	}

	var user model.User
	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	customer, err := h.db.UserSignIn(user.Email, user.Password)
	if err != nil {
		if err == db.ErrINVALIDPASSWORD {
			return c.JSON(http.StatusForbidden, err.Error())
		}

		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, customer)
}

// User Sign up
func (h *Handler) SignUp(c echo.Context) error {
	if h.db == nil {
		return c.JSON(http.StatusInternalServerError, "server database error")
	}

	var user model.User
	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	err := h.db.UserSignUp(user)
	if err != nil {
		if err == db.ErrINVALIDDATA {
			return c.JSON(http.StatusForbidden, err.Error())
		}
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, "Success")
}

// Streaming RTSP
func (h *Handler) StreamRTSP(c echo.Context) error {
	if h.db == nil {
		return c.JSON(http.StatusInternalServerError, "server database error")
	}

	cams, err := h.db.GetAllCam()
	if err != nil || len(cams) == 0 {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	// websocket
	unSafeconn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	ws := &service.ThreadSafeWriter{Conn: unSafeconn}

	go ws.WebRTCStreamH264() // controller 보내기

	return c.JSON(http.StatusOK, cams)
}

// Add RTSP Camera
func (h *Handler) AddNewCam(c echo.Context) error {
	if h.db == nil {
		return c.JSON(http.StatusInternalServerError, "server database error")
	}

	var cam model.Camera
	if err := c.Bind(&cam); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if err := h.db.AddNewCam(cam); err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, "Success")
}

// Delete RTSP Camera
func (h *Handler) DeleteCam(c echo.Context) error {
	if h.db == nil {
		return c.JSON(http.StatusInternalServerError, "server database error")
	}

	var cam model.Camera
	if err := c.Bind(&cam); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if err := h.db.DeleteCam(cam.Name); err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, "Success")
}

// Get RTSP Camera Information
func (h *Handler) GetCam(c echo.Context) error {
	if h.db == nil {
		return c.JSON(http.StatusInternalServerError, "server database error")
	}
	param := c.Param("id")

	cam, err := h.db.GetCamByID(param)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, cam)
}

// Get RTSP Camera Information
func (h *Handler) GetAllCam(c echo.Context) error {
	if h.db == nil {
		return c.JSON(http.StatusInternalServerError, "server database error")
	}

	cams, err := h.db.GetAllCam()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, cams)
}

// Monitoring OPC UA Server
func (h *Handler) MonitoringOpcUA(c echo.Context) error {
	if h.db == nil {
		return c.JSON(http.StatusInternalServerError, "server database error")
	}
	opcs, err := h.db.GetAllServer()
	if err != nil || len(opcs) == 0 {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	unSafeconn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	ws := &service.ThreadSafeWriter{Conn: unSafeconn}

	go ws.SetConfigurationAndRun(opcs[0])

	return c.JSON(http.StatusOK, opcs)
}

// Add OPC UA Server
func (h *Handler) AddNewServer(c echo.Context) error {
	if h.db == nil {
		return c.JSON(http.StatusInternalServerError, "server database error")
	}

	var opc model.OpcUAServer
	if err := c.Bind(&opc); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if err := h.db.AddNewServer(opc); err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, "Success")
}

// Delete OPC UA Server
func (h *Handler) DeleteServer(c echo.Context) error {
	if h.db == nil {
		return c.JSON(http.StatusInternalServerError, "server database error")
	}

	var opc model.OpcUAServer
	if err := c.Bind(&opc); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if err := h.db.DeleteServer(opc.Name); err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, "Success")
}

// Get OPC UA Server
func (h *Handler) GetServer(c echo.Context) error {
	if h.db == nil {
		return c.JSON(http.StatusInternalServerError, "server database error")
	}
	param := c.Param("id")

	opc, err := h.db.GetServerByID(param)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, opc)
}

// Get ALL OPC UA Server
func (h *Handler) GetAllServer(c echo.Context) error {
	if h.db == nil {
		return c.JSON(http.StatusInternalServerError, "server database error")
	}

	opcs, err := h.db.GetAllServer()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, opcs)
}
