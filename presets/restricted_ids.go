package presets

// MordhauRestrictedPacketIDs is a slice of restricted packet IDs which should not be used when writing packets to RCON
// connections. Since Mordhau supports broadcasts, these restricted IDs are all used by the server when sending us
// broadcast messages.
//
// Most of these IDs belong to a respective broadcast channel. There are some gaps in the increments where no channels
// currently exist (for example, 54322) but just to be sure the entire range from the minimum observed broadcast packet
// ID up to maximum is included.
var MordhauRestrictedPacketIDs = []int32{54321, 54322, 54323, 54324, 54325, 54326, 54327, 54328, 54329, 54330}

// 54321: Matchstate
// 54324: Scorefeed
// 54325: Chat
// 54326: Login
// 54330: Punishment
