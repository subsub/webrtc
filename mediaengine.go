package webrtc

import (
	"github.com/pions/rtp"
	"github.com/pions/rtp/codecs"
)

// PayloadTypes for the default codecs
const (
	DefaultPayloadTypeG722 = 9
	DefaultPayloadTypeOpus = 111
	DefaultPayloadTypeVP8  = 96
	DefaultPayloadTypeVP9  = 98
	DefaultPayloadTypeH264 = 100
)

// MediaEngine defines the codecs supported by a RTCPeerConnection
type MediaEngine struct {
	codecs []*RTCRtpCodec
}

// RegisterCodec registers a codec to a media engine
func (m *MediaEngine) RegisterCodec(codec *RTCRtpCodec) uint8 {
	// TODO: generate PayloadType if not set
	m.codecs = append(m.codecs, codec)
	return codec.PayloadType
}

// RegisterDefaultCodecs is a helper that registers the default codecs supported by pions-webrtc
func (m *MediaEngine) RegisterDefaultCodecs() {
	m.RegisterCodec(NewRTCRtpOpusCodec(DefaultPayloadTypeOpus, 48000, 2))
	m.RegisterCodec(NewRTCRtpG722Codec(DefaultPayloadTypeG722, 8000))
	m.RegisterCodec(NewRTCRtpVP8Codec(DefaultPayloadTypeVP8, 90000))
	m.RegisterCodec(NewRTCRtpH264Codec(DefaultPayloadTypeH264, 90000))
	m.RegisterCodec(NewRTCRtpVP9Codec(DefaultPayloadTypeVP9, 90000))
}

func (m *MediaEngine) getCodec(payloadType uint8) (*RTCRtpCodec, error) {
	for _, codec := range m.codecs {
		if codec.PayloadType == payloadType {
			return codec, nil
		}
	}
	return nil, ErrCodecNotFound
}

func (m *MediaEngine) getCodecSDP(codecParams RTCRtpCodecParameters) (*RTCRtpCodec, error) {
	for _, codec := range m.codecs {
		if codec.Name == codecParams.Name &&
			codec.ClockRate == codecParams.ClockRate &&
			codec.Channels == codecParams.Channels {
			ok, err := codecParams.equalFMTP(codec.SdpFmtpLine) // TODO: Protocol specific matching?
			if err != nil {
				pcLog.Warnf("failed compare FMTP: %v", err)
				continue
			}
			if !ok {
				continue
			}
			return codec, nil
		}

	}
	return nil, ErrCodecNotFound
}

func (m *MediaEngine) getCodecsByKind(kind RTCRtpCodecType) []*RTCRtpCodec {
	var codecs []*RTCRtpCodec
	for _, codec := range m.codecs {
		if codec.Type == kind {
			codecs = append(codecs, codec)
		}
	}
	return codecs
}

// Names for the default codecs supported by pions-webrtc
const (
	G722 = "G722"
	Opus = "opus"
	VP8  = "VP8"
	VP9  = "VP9"
	H264 = "H264"
)

// NewRTCRtpG722Codec is a helper to create a G722 codec
func NewRTCRtpG722Codec(payloadType uint8, clockrate uint32) *RTCRtpCodec {
	c := NewRTCRtpCodec(RTCRtpCodecTypeAudio,
		G722,
		clockrate,
		0,
		"",
		payloadType,
		&codecs.G722Payloader{})
	return c
}

// NewRTCRtpOpusCodec is a helper to create an Opus codec
func NewRTCRtpOpusCodec(payloadType uint8, clockrate uint32, channels uint32) *RTCRtpCodec {
	c := NewRTCRtpCodec(RTCRtpCodecTypeAudio,
		Opus,
		clockrate,
		channels,
		"minptime=10;useinbandfec=1",
		payloadType,
		&codecs.OpusPayloader{})
	return c
}

// NewRTCRtpVP8Codec is a helper to create an VP8 codec
func NewRTCRtpVP8Codec(payloadType uint8, clockrate uint32) *RTCRtpCodec {
	c := NewRTCRtpCodec(RTCRtpCodecTypeVideo,
		VP8,
		clockrate,
		0,
		"",
		payloadType,
		&codecs.VP8Payloader{})
	return c
}

// NewRTCRtpVP9Codec is a helper to create an VP9 codec
func NewRTCRtpVP9Codec(payloadType uint8, clockrate uint32) *RTCRtpCodec {
	c := NewRTCRtpCodec(RTCRtpCodecTypeVideo,
		VP9,
		clockrate,
		0,
		"",
		payloadType,
		nil) // TODO
	return c
}

// NewRTCRtpH264Codec is a helper to create an H264 codec
func NewRTCRtpH264Codec(payloadType uint8, clockrate uint32) *RTCRtpCodec {
	c := NewRTCRtpCodec(RTCRtpCodecTypeVideo,
		H264,
		clockrate,
		0,
		"level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42001f",
		payloadType,
		&codecs.H264Payloader{})
	return c
}

// RTCRtpCodecType determines the type of a codec
type RTCRtpCodecType int

const (

	// RTCRtpCodecTypeAudio indicates this is an audio codec
	RTCRtpCodecTypeAudio RTCRtpCodecType = iota + 1

	// RTCRtpCodecTypeVideo indicates this is a video codec
	RTCRtpCodecTypeVideo
)

func (t RTCRtpCodecType) String() string {
	switch t {
	case RTCRtpCodecTypeAudio:
		return "audio"
	case RTCRtpCodecTypeVideo:
		return "video"
	default:
		return ErrUnknownType.Error()
	}
}

// RTCRtpCodec represents a codec supported by the PeerConnection
type RTCRtpCodec struct {
	RTCRtpCodecCapability
	Type        RTCRtpCodecType
	Name        string
	PayloadType uint8
	Payloader   rtp.Payloader
}

// NewRTCRtpCodec is used to define a new codec
func NewRTCRtpCodec(
	codecType RTCRtpCodecType,
	name string,
	clockrate uint32,
	channels uint32,
	fmtp string,
	payloadType uint8,
	payloader rtp.Payloader,
) *RTCRtpCodec {
	return &RTCRtpCodec{
		RTCRtpCodecCapability: RTCRtpCodecCapability{
			MimeType:    codecType.String() + "/" + name,
			ClockRate:   clockrate,
			Channels:    channels,
			SdpFmtpLine: fmtp,
		},
		PayloadType: payloadType,
		Payloader:   payloader,
		Type:        codecType,
		Name:        name,
	}
}

// RTCRtpCodecCapability provides information about codec capabilities.
type RTCRtpCodecCapability struct {
	MimeType    string
	ClockRate   uint32
	Channels    uint32 // uint32 as specified in ORTC
	SdpFmtpLine string
}

// RTCRtpHeaderExtensionCapability is used to define a RFC5285 RTP header extension supported by the codec.
type RTCRtpHeaderExtensionCapability struct {
	URI string
}

// RTCRtpCapabilities represents the capabilities of a transceiver
type RTCRtpCapabilities struct {
	Codecs           []RTCRtpCodecCapability
	HeaderExtensions []RTCRtpHeaderExtensionCapability
}
