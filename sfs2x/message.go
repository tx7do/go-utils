package sfs2x

import "fmt"

// Protocol envelope helpers: the standard SFS2X message wrapper
// ({controller, action, params}) plus the system-controller action codes
// and handshake/encryption message builders, mirroring the official client
// request classes.

// ControllerID enumerates the well-known request controllers.
const (
	ControllerSystem      byte = 0 // handshake, login, room/user management...
	ControllerExtension   byte = 1 // user extension calls
	ControllerSecurity    byte = 2 // encryption init, UDP fallback
	ControllerBuddy       byte = 3 // buddy list
	ControllerRoomManager byte = 4
	ControllerAdmin       byte = 5 // admin tool / console
)

// SysAction enumerates the system-controller action codes.
const (
	SysActionHandshake            int16 = 0
	SysActionLogin                int16 = 1
	SysActionLogout               int16 = 2
	SysActionGetRoomList          int16 = 3
	SysActionJoinRoom             int16 = 4
	SysActionAutoJoin             int16 = 5
	SysActionCreateRoom           int16 = 6
	SysActionGenericMessage       int16 = 7
	SysActionChangeRoomName       int16 = 8
	SysActionChangeRoomPassword   int16 = 9
	SysActionObjectMessage        int16 = 10
	SysActionSetRoomVariables     int16 = 11
	SysActionSetUserVariables     int16 = 12
	SysActionCallExtension        int16 = 13
	SysActionLeaveRoom            int16 = 14
	SysActionSubscribeRoomGroup   int16 = 15
	SysActionUnsubscribeRoomGroup int16 = 16
	SysActionSpectatorToPlayer    int16 = 17
	SysActionPlayerToSpectator    int16 = 18
	SysActionChangeRoomCapacity   int16 = 19
	SysActionPublicMessage        int16 = 20
	SysActionPrivateMessage       int16 = 21
	SysActionModeratorMessage     int16 = 22
	SysActionAdminMessage         int16 = 23
	SysActionKickUser             int16 = 24
	SysActionBanUser              int16 = 25
	SysActionManualDisconnection  int16 = 26
	SysActionFindRooms            int16 = 27
	SysActionFindUsers            int16 = 28
	SysActionPingPong             int16 = 29
	SysActionSetUserPosition      int16 = 30
)

// Message is the standard SFS2X request/response envelope:
// {c: BYTE controller, a: SHORT action, p: SFS_OBJECT params}.
type Message struct {
	Controller byte
	Action     int16
	Params     *Object
}

// NewMessage builds an envelope around params (a fresh empty Object when
// nil).
func NewMessage(controller byte, action int16, params *Object) *Message {
	if params == nil {
		params = New()
	}
	return &Message{Controller: controller, Action: action, Params: params}
}

// ToObject serializes the envelope (the payload wrapped inside a frame).
func (m *Message) ToObject() *Object {
	o := New()
	o.Put("c", TByte, m.Controller)
	o.Put("a", TShort, m.Action)
	if m.Params == nil {
		m.Params = New()
	}
	o.Put("p", TSFSObject, m.Params)
	return o
}

// EncodeOrderedFrame frames the envelope preserving param key order
// (official-client byte fidelity).
func (m *Message) EncodeOrderedFrame() ([]byte, error) {
	return EncodeFrameWithOptions(m.ToObject(), FrameOptions{Ordered: true})
}

// ParseMessage decodes a framed envelope.
func ParseMessage(data []byte) (*Message, error) {
	return ParseMessageWithOptions(data, FrameOptions{})
}

// ParseMessageWithOptions decodes a framed envelope applying the given
// transforms (decryption key/IV for encrypted sessions).
func ParseMessageWithOptions(data []byte, opts FrameOptions) (*Message, error) {
	o, err := DecodeFrameWithOptions(data, opts)
	if err != nil {
		return nil, err
	}
	return ParseMessageObject(o)
}

// ParseMessageObject interprets a decoded SFSObject as an envelope.
func ParseMessageObject(o *Object) (*Message, error) {
	if o.GetType("c") != TByte || o.GetType("a") != TShort {
		return nil, fmt.Errorf("not a message envelope (c/a fields missing or mistyped)")
	}
	p := o.GetSFSObject("p")
	if p == nil {
		p = New()
	}
	return &Message{Controller: o.GetByte("c"), Action: o.GetShort("a"), Params: p}, nil
}

// Handshake request/response field names (official client conventions).
const (
	HandshakeKeyAPI          = "api" // client API version (request)
	HandshakeKeyClientType   = "cl"  // client type string (request)
	HandshakeKeyReconnection = "rt"  // reconnection token (request, optional)
	HandshakeKeySessionToken = "tk"  // session token (response; drives the HTTPS crypto-key servlet call)
	HandshakeKeyCompression  = "ct"  // compression threshold (response)
	HandshakeKeyMaxMessage   = "ms"  // max message size (response)
)

// NewHandshakeRequest builds the connection-time handshake request
// (controller 0 / action 0). The session token returned in the response is
// what the caller presents to the HTTPS crypto-key servlet.
func NewHandshakeRequest(apiVersion, clientType string) *Message {
	params := New()
	params.Put(HandshakeKeyAPI, TUTFString, apiVersion)
	params.Put(HandshakeKeyClientType, TUTFString, clientType)
	return NewMessage(ControllerSystem, SysActionHandshake, params)
}

// HandshakeResponse is the parsed server handshake reply.
type HandshakeResponse struct {
	SessionToken         string // "tk" — present to the crypto-key servlet
	CompressionThreshold int    // "ct" — 0 when absent
	MaxMessageSize       int    // "ms" — 0 when absent
}

// ParseHandshakeResponse extracts the handshake reply fields.
func ParseHandshakeResponse(o *Object) (*HandshakeResponse, error) {
	if o == nil {
		return nil, fmt.Errorf("nil handshake response")
	}
	resp := &HandshakeResponse{SessionToken: o.GetUTFString(HandshakeKeySessionToken)}
	if o.GetType(HandshakeKeyCompression) == TInt {
		resp.CompressionThreshold = int(o.GetInt(HandshakeKeyCompression))
	}
	if o.GetType(HandshakeKeyMaxMessage) == TInt {
		resp.MaxMessageSize = int(o.GetInt(HandshakeKeyMaxMessage))
	}
	if resp.SessionToken == "" {
		return nil, fmt.Errorf("handshake response without session token")
	}
	return resp, nil
}
