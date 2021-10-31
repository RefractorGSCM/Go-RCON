package packet

type PacketType int32

const TypeAuth = PacketType(3)
const TypeAuthRes = PacketType(2)
const TypeCommand = PacketType(2)
const TypeCommandRes = PacketType(0)
const AuthFailedID = -1
