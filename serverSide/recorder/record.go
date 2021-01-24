package recorder

import (
	"fmt"
	"time"

	"github.com/linexjlin/simple-log"
	. "github.com/linexjlin/webRTCRecorder/serverSide/common"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/ivfwriter"
	"github.com/pion/webrtc/v3/pkg/media/oggwriter"
)

type Recorder struct{}

func NewRecorder() *Recorder {
	return &Recorder{}
}

func (r *Recorder) saveToDisk(i media.Writer, track *webrtc.TrackRemote) {
	defer func() {
		if err := i.Close(); err != nil {
			panic(err)
		}
	}()

	for {
		rtpPacket, _, err := track.ReadRTP()
		if err != nil {
			panic(err)
		}
		if err := i.WriteRTP(rtpPacket); err != nil {
			panic(err)
		}
	}
}

func (r *Recorder) Init(peer string, channel string) (locDes string, err error) {
	// Everything below is the Pion WebRTC API! Thanks for using it ❤️.

	// Create a MediaEngine object to configure the supported codec
	m := webrtc.MediaEngine{}

	// Setup the codecs you want to use.
	// We'll use a VP8 and Opus but you can also define your own
	if err := m.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: "video/VP8", ClockRate: 90000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil},
		PayloadType:        96,
	}, webrtc.RTPCodecTypeVideo); err != nil {
		panic(err)
	}
	if err := m.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: "audio/opus", ClockRate: 48000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil},
		PayloadType:        111,
	}, webrtc.RTPCodecTypeAudio); err != nil {
		panic(err)
	}

	// Create the API object with the MediaEngine
	api := webrtc.NewAPI(webrtc.WithMediaEngine(&m))

	// Prepare the configuration
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	// Create a new RTCPeerConnection
	peerConnection, err := api.NewPeerConnection(config)
	if err != nil {
		log.Error(err)
		return "", err
	}

	// Allow us to receive 1 audio track, and 1 video track
	if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil {
		log.Error(err)
		return "", err
	} else if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
		log.Error(err)
		return "", err
	}

	oggFile, err := oggwriter.New("./data/"+channel+".ogg", 48000, 2)
	if err != nil {
		log.Error(err)
		return "", err
	}
	ivfFile, err := ivfwriter.New("./data/" + channel + ".ivf")
	if err != nil {
		log.Error(err)
		return "", err
	}

	// Set a handler for when a new remote track starts, this handler saves buffers to disk as
	// an ivf file, since we could have multiple video tracks we provide a counter.
	// In your application this is where you would handle/process video
	peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		// Send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
		go func() {
			ticker := time.NewTicker(time.Second * 3)
			for range ticker.C {
				errSend := peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(track.SSRC())}})
				if errSend != nil {
					fmt.Println(errSend)
				}
			}
		}()

		codec := track.Codec()
		if codec.MimeType == "audio/opus" {
			log.Debug("Got Opus track, saving to disk as output.opus (48 kHz, 2 channels)")
			r.saveToDisk(oggFile, track)
		} else if codec.MimeType == "video/VP8" {
			log.Debug("Got VP8 track, saving to disk as output.ivf")
			r.saveToDisk(ivfFile, track)
		}
	})

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		log.Debug("Connection State has changed %s \n", connectionState.String())

		if connectionState == webrtc.ICEConnectionStateConnected {
			log.Debug("Ctrl+C the remote client to stop the demo")
		} else if connectionState == webrtc.ICEConnectionStateFailed ||
			connectionState == webrtc.ICEConnectionStateDisconnected {
			closeErr := oggFile.Close()
			if closeErr != nil {
				log.Error(closeErr)
			}

			closeErr = ivfFile.Close()
			if closeErr != nil {
				log.Error(closeErr)
			}

			log.Println("Done writing media files")
			//os.Exit(0)
		}
	})

	// Wait for the offer to be pasted
	offer := webrtc.SessionDescription{}
	if err := Decode(peer, &offer); err != nil {
		return "", err
	}

	// Set the remote SessionDescription
	err = peerConnection.SetRemoteDescription(offer)
	if err != nil {
		log.Error(err)
		return "", err
	}

	// Create answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		log.Error(err)
		return "", err
	}

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	// Sets the LocalDescription, and starts our UDP listeners
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		log.Error(err)
		return "", err
	}

	// Block until ICE Gathering is complete, disabling trickle ICE
	// we do this because we only can exchange one signaling message
	// in a production application you should exchange ICE Candidates via OnICECandidate
	<-gatherComplete

	// Output the answer in base64 so we can paste it in browser
	if locDes, err := Encode(*peerConnection.LocalDescription()); err != nil {
		return "", err
	} else {
		return fmt.Sprint(locDes), nil
	}

	// Block forever
	//select {}
}
