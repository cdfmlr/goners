package api

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/cdfmlr/goners"
	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slog"
	"golang.org/x/net/websocket"
)

// goners http api:
//
// devicse:
//   GET  /devices: lookup devices
// pcap:
//   POST   /pcap:  start a capturing
//   DELETE /pcap:  stop a capturing
//   WS     /pcap/{sessionID}: get packets
//

// wssessions holds sessions' ws output handler
var wssessions sync.Map // map[SessionID]websocket.Handler

type GetDevicesRequest struct{}

type GetDevicesResponse []*goners.Device

// GET /devices
func GetDevices(c *gin.Context) {
	req := GetDevicesRequest{}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	resp, err := getDevices(req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func getDevices(req GetDevicesRequest) (GetDevicesResponse, error) {
	devices, err := goners.LookupDevices()
	if err != nil {
		return nil, err
	}
	return devices, nil
}

type StartPcapRequest struct {
	Device  string        `json:"device"`
	Filter  string        `json:"filter"`
	Snaplen int           `json:"snaplen"`
	Promisc bool          `json:"promisc"`
	Timeout time.Duration `json:"timeout"`
	Format  string        `json:"format"`
	Output  string        `json:"output"`
}

type StartPcapResponse struct {
	SessionID goners.SessionID `json:"session_id"`
}

// POST /pcap
func StartPcap(c *gin.Context) {
	req := StartPcapRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	resp, err := startPcap(req)

	if err != nil {
		slog.Warn("startPcap failed.", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func startPcap(req StartPcapRequest) (StartPcapResponse, error) {
	config := goners.PcapSessionConfig{
		Device:  req.Device,
		Filter:  req.Filter,
		Snaplen: req.Snaplen,
		Promisc: req.Promisc,
		Timeout: req.Timeout,
	}

	var formater goners.PacketsFormater
	switch req.Format {
	case "text":
		formater = goners.StringPacketsFormater
	case "json":
		formater = goners.JsonPacketsFormater
	}
	config.Format = formater

	// TODO: outputer choice
	out, ws := goners.NewWebSocketOutputer()
	config.Output = out

	sessionID, err := goners.GetPcapSessionsManager().StartSession(&config)
	if err != nil {
		return StartPcapResponse{}, err
	}

	wssessions.Store(sessionID, ws)

	return StartPcapResponse{SessionID: sessionID}, nil
}

type StopPcapRequest struct {
	SessionID goners.SessionID `json:"session_id"`
}

type StopPcapResponse struct{}

// DELETE /pcap
func StopPcap(c *gin.Context) {
	req := StopPcapRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	resp, err := stopPcap(req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func stopPcap(req StopPcapRequest) (StopPcapResponse, error) {
	err := goners.GetPcapSessionsManager().CloseSession(req.SessionID)
	wssessions.Delete(req.SessionID)
	return StopPcapResponse{}, err
}

// WS /pcap/{sessionID}
func WsPcap(c *gin.Context) {
	sessionID := goners.SessionID(c.Param("sessionID"))
	ws, ok := wssessions.Load(sessionID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "session not found",
		})
		return
	}

	wsHandler, ok := ws.(websocket.Handler)
	if !ok {
		slog.Error("wssessions stored ws type error.",
			"type", fmt.Sprintf("%T", ws),
			"ws", ws)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal error",
		})
		return
	}

	wsHandler.ServeHTTP(c.Writer, c.Request)
}

// register http api
func RegisterHttpApi(r *gin.Engine) {
	r.GET("/devices", GetDevices)
	r.POST("/pcap", StartPcap)
	r.DELETE("/pcap", StopPcap)
	r.Any("/pcap/:sessionID", WsPcap)
}

// router
func NewHttp() *gin.Engine {
	r := gin.Default()
	RegisterHttpApi(r)
	return r
}
