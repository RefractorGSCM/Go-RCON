package presets

import "github.com/refractorgscm/rcon/packet"

func MordhauBroadcastChecker(p packet.Packet) bool {
	for _, v := range MordhauRestrictedPacketIDs {
		if v == p.ID() {
			return true
		}
	}

	return false
}
