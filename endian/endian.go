package endian

import "encoding/binary"

type Mode binary.ByteOrder

var Big = binary.BigEndian
var Little = binary.LittleEndian
