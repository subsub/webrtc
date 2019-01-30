package webrtc

import (
	"fmt"

	"github.com/pkg/errors"
)

// RTCRtpTransceiver represents a combination of an RTCRtpSender and an RTCRtpReceiver that share a common mid.
type RTCRtpTransceiver struct {
	Mid       string
	Sender    *RTCRtpSender
	Receiver  *RTCRtpReceiver
	Direction RTCRtpTransceiverDirection
	// currentDirection RTCRtpTransceiverDirection
	// firedDirection   RTCRtpTransceiverDirection
	// receptive bool
	stopped bool

	remoteCapabilities     RTCRtpParameters
	recvDecodingParameters []RTCRtpDecodingParameters

	peerConnection *RTCPeerConnection
}

func (t *RTCRtpTransceiver) setSendingTrack(track *RTCTrack) error {
	t.Sender.Track = track

	switch t.Direction {
	case RTCRtpTransceiverDirectionRecvonly:
		t.Direction = RTCRtpTransceiverDirectionSendrecv
	case RTCRtpTransceiverDirectionInactive:
		t.Direction = RTCRtpTransceiverDirectionSendonly
	default:
		return errors.Errorf("Invalid state change in RTCRtpTransceiver.setSending")
	}
	return nil
}

func (t *RTCRtpTransceiver) start() error {
	// Start the sender
	sender := t.Sender
	if sender != nil {
		sender.Send(RTCRtpSendParameters{
			encodings: RTCRtpEncodingParameters{
				RTCRtpCodingParameters{SSRC: sender.Track.Ssrc, PayloadType: sender.Track.PayloadType},
			}})
	}

	// Start the receiver
	receiver := t.Receiver
	if receiver != nil {
		params := RTCRtpReceiveParameters{
			RTCRtpParameters: t.remoteCapabilities,
			Encodings:        t.recvDecodingParameters,
		}

		err := receiver.Receive(params)
		if err != nil {
			return fmt.Errorf("failed to receive: %v", err)
		}

		t.peerConnection.onTrack(receiver.Track)
	}

	return nil
}

// Stop irreversibly stops the RTCRtpTransceiver
func (t *RTCRtpTransceiver) Stop() error {
	if t.Sender != nil {
		t.Sender.Stop()
	}
	if t.Receiver != nil {
		if err := t.Receiver.Stop(); err != nil {
			return err
		}
	}
	return nil
}
