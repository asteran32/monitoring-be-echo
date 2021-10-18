package service

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os/exec"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/h264reader"
)

type message struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}

// Helper to make Gorilla Websockets threadsafe
type ThreadSafeWriter struct {
	Conn *websocket.Conn
}

func (t *ThreadSafeWriter) WriteJSON(v interface{}) error {
	var mutex = &sync.Mutex{}
	mutex.Lock()
	defer mutex.Unlock()

	return t.Conn.WriteJSON(v)
}

var (
	peerConnection = &webrtc.PeerConnection{}
	test_rtsp_url  = "rtsp://wowzaec2demo.streamlock.net/vod/mp4:BigBuckBunny_115k.mov"
)

func (t *ThreadSafeWriter) WebRTCStreamH264() error {
	defer t.Conn.Close()

	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return err
	}
	defer peerConnection.Close()

	// Add track and streaming h264 codec rtsp video using ffmepg
	// Create a video track and rtpSender
	videoTrack, videoTrackErr := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "pion")
	if videoTrackErr != nil {
		return videoTrackErr
	}

	rtpSender, rtpSenderErr := peerConnection.AddTrack(videoTrack)
	if rtpSenderErr != nil {
		return rtpSenderErr
	}

	iceConnectedCtx, iceConnectedCtxCancel := context.WithCancel(context.Background())

	// Read incoming RTCP packets
	// Before these packets are returned they are processed by interceptors. For things
	// like NACK this needs to be called.
	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()

	// Trickle ICE. Emit server candidate to client
	peerConnection.OnICECandidate(func(i *webrtc.ICECandidate) {
		if i == nil {
			return
		}
		candidateString, err := json.Marshal(i.ToJSON())
		if err != nil {
			log.Print(err)
			return
		}
		if writeErr := t.Conn.WriteJSON(&message{Event: "candidate", Data: string(candidateString)}); writeErr != nil {
			log.Print(writeErr)
			return
		}
	})

	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		if connectionState == webrtc.ICEConnectionStateConnected {
			log.Print("Peer has connected")
			iceConnectedCtxCancel()
		} else if connectionState == webrtc.ICEConnectionStateFailed {
			if closeErr := peerConnection.Close(); closeErr != nil {
				panic(closeErr)
			}
		}
	})

	go func() {
		// Open a H264 file and start reading using our IVFReader
		cmd := exec.Command("ffmpeg", "-i", test_rtsp_url, "-c:v", "libx264",
			"-an", "-bsf:v", "h264_mp4toannexb", "-b:v", "2M", "-max_delay", "0",
			"-bf", "0", "-f", "h264", "pipe:1")

		cmdStdOut, err := cmd.StdoutPipe()
		if err != nil {
			panic(err)
		}

		h264, h264Err := h264reader.NewReader(cmdStdOut)
		if h264Err != nil {
			panic(h264Err)
		}

		// Wait for connection established
		<-iceConnectedCtx.Done()

		cmd.Start()
		// Send our video file frame at a time. Pace our sending so we send it at the same speed it should be played back as.
		// This isn't required since the video is timestamped, but we will such much higher loss if we send all at once.
		//
		// It is important to use a time.Ticker instead of time.Sleep because
		// * avoids accumulating skew, just calling time.Sleep didn't compensate for the time spent parsing the data
		// * works around latency issues with Sleep (see https://github.com/golang/go/issues/44343)
		ticker := time.NewTicker(time.Millisecond * 33)
		for ; true; <-ticker.C {
			nal, h264Err := h264.NextNAL()
			if h264Err == io.EOF {
				log.Print("All video frames parsed and sent \n")
				return
			}
			if h264Err != nil {
				return
			}
			if h264Err = videoTrack.WriteSample(media.Sample{Data: nal.Data, Duration: time.Second}); h264Err != nil {
				return
			}
		}
		log.Print("End stream video \n")
	}()

	// Send and Get JSON message for signaling
	done := make(chan bool)
	msg := &message{}

	go func() {
		for {
			_, raw, err := t.Conn.ReadMessage()
			if err != nil {
				log.Println(err)
				return

			} else if err := json.Unmarshal(raw, &msg); err != nil {
				log.Println(err)
				return
			}

			switch msg.Event {
			case "candidate":
				candidate := webrtc.ICECandidateInit{}
				if err := json.Unmarshal([]byte(msg.Data), &candidate); err != nil {
					log.Print(err)
					return
				}
				if err := peerConnection.AddICECandidate(candidate); err != nil {
					return
				}

			case "offer":
				offer := webrtc.SessionDescription{}
				if err := json.Unmarshal([]byte(msg.Data), &offer); err != nil {
					log.Println(err)
					return
				}
				if err := peerConnection.SetRemoteDescription(offer); err != nil {
					return
				}
				answer, err := peerConnection.CreateAnswer(nil)
				if err != nil {
					return
				}
				if err = peerConnection.SetLocalDescription(answer); err != nil {
					return
				}
				answerString, err := json.Marshal(answer)
				if err != nil {
					return
				}
				if writeErr := t.Conn.WriteJSON(&message{Event: "answer", Data: string(answerString)}); writeErr != nil {
					return
				}
			}
		}
	}()
	<-done
	return nil
}
