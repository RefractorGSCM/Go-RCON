package packet

import (
	"bytes"
	"github.com/franela/goblin"
	. "github.com/onsi/gomega"
	"github.com/refractorgscm/rcon/endian"
	"math"
	"testing"
)

func Test(t *testing.T) {
	g := goblin.Goblin(t)

	// Special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Packets", func() {
		g.Describe("Client IPacket", func() {
			var packet *ClientPacket
			var rawPacket []byte

			g.BeforeEach(func() {
				packet = &ClientPacket{
					mode:  endian.Little,
					pType: TypeCommand,
					body:  []byte("Hello, world!"),
					id:    1,
				}

				rawPacket = []byte{'\x17', '\x00', '\x00', '\x00', '\x01', '\x00', '\x00', '\x00', '\x02', '\x00',
					'\x00', '\x00', '\x48', '\x65', '\x6c', '\x6c', '\x6f', '\x2c', '\x20', '\x77', '\x6f', '\x72', '\x6c',
					'\x64', '\x21', '\x00', '\x00'}
			})

			g.Describe("NewClientPacket()", func() {
				g.It("Should return the expected packet", func() {
					got := NewClientPacket(endian.Little, TypeCommand, "Hello, world!", nil)

					Expect(got).To(Equal(packet))
				})

				g.It("Should reset packet ID counter if nest value with cause overflow", func() {
					nextClientPacketID = math.MaxInt32 - 1

					got := NewClientPacket(endian.Little, TypeCommand, "Hello, world!", nil)
					Expect(got.ID()).To(Equal(int32(1)))
				})

				g.It("Should skip restricted packet IDs", func() {
					nextClientPacketID = 10
					restrictedIDs := []int32{10, 11, 12}

					got := NewClientPacket(endian.Little, TypeCommand, "Hello, world!", restrictedIDs)
					Expect(got.ID()).To(Equal(int32(13)))
				})
			})

			g.Describe("Body()", func() {
				g.It("Should return the correct null terminated body", func() {
					expected := append(packet.body, '\x00')
					got := packet.Body()

					Expect(got).To(Equal(expected))
				})
			})

			g.Describe("Size()", func() {
				g.It("Should return the correct length", func() {
					expected := int32(4 + 4 + len("Hello, world!") + 1 + 1)

					length := packet.Size()
					Expect(length).To(Equal(expected))
				})
			})

			g.Describe("ID()", func() {
				g.It("Should return the correct ID", func() {
					Expect(packet.ID()).To(Equal(packet.id))
				})
			})

			g.Describe("Build()", func() {
				g.It("Should not return an error", func() {
					_, err := packet.Build()

					Expect(err).To(BeNil())
				})

				g.It("Should generate the correct output", func() {
					got, err := packet.Build()

					Expect(err).To(BeNil())
					Expect(got).To(Equal(rawPacket))
				})
			})

			g.Describe("DecodeClientPacket()", func() {
				g.It("Should not return an error", func() {
					_, err := DecodeClientPacket(packet.mode, bytes.NewReader(rawPacket))

					Expect(err).To(BeNil())
				})

				g.It("Should decode the correct packet", func() {
					decoded, err := DecodeClientPacket(packet.mode, bytes.NewReader(rawPacket))

					Expect(err).To(BeNil())
					Expect(decoded).To(Equal(packet))
				})
			})
		})
	})
}
