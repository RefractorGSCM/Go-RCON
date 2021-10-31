package packet

type Packet interface {
	Size() int32
	ID() int32
	Type() PacketType
	Body() []byte
	Build() ([]byte, error)
}
