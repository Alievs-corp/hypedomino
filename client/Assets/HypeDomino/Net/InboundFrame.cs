using Hypedomino.V1;

namespace HypeDomino.Net
{
    public struct InboundFrame
    {
        public uint ProtocolVersion;
        public ulong Sequence;
        public ulong AckedSequence;
        public ulong SessionNonce;
        public Envelope.PayloadOneofCase Payload;
        public int PayloadStart;
        public int PayloadLength;
    }
}
