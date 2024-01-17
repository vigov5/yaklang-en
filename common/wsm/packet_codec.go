package wsm

type PacketCodecI interface {
	// ClientRequestEncode encodes the payload in the request package
	ClientRequestEncode(raw []byte) ([]byte, error)
	// ServerResponseDecode webshell server gets the payload in the request package
	ServerResponseDecode(raw []byte) ([]byte, error)
	SetPacketScriptContent(content string)
}
