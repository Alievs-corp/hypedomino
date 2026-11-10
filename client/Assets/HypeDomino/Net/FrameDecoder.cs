using System;
using System.Buffers.Binary;
using Hypedomino.V1;

namespace HypeDomino.Net
{
    public static class FrameDecoder
    {
        const int VarintWire = 0;
        const int Fixed64Wire = 1;
        const int LengthWire = 2;
        const int Fixed32Wire = 5;
        const int MaxVarintBytes = 10;

        static readonly bool[] PayloadFields = BuildPayloadFields();

        public static bool TryRead(ReadOnlySpan<byte> frame, out InboundFrame header)
        {
            header = default;
            var position = 0;
            while (position < frame.Length)
            {
                if (!TryReadVarint(frame, ref position, out var tag)) return false;
                var field = (int)(tag >> 3);
                if (field == 0 || !TrySkipOrRead(frame, ref position, field, (int)(tag & 7), ref header)) return false;
            }
            return true;
        }

        static bool TrySkipOrRead(ReadOnlySpan<byte> frame, ref int position, int field, int wire, ref InboundFrame header)
        {
            switch (wire)
            {
                case VarintWire:
                    if (!TryReadVarint(frame, ref position, out var value)) return false;
                    SetVarint(ref header, field, value);
                    return true;
                case Fixed64Wire:
                    if (frame.Length - position < sizeof(ulong)) return false;
                    if (field == Envelope.SessionNonceFieldNumber)
                        header.SessionNonce = BinaryPrimitives.ReadUInt64LittleEndian(frame.Slice(position));
                    position += sizeof(ulong);
                    return true;
                case LengthWire:
                    if (!TryReadVarint(frame, ref position, out var length) || length > (ulong)(frame.Length - position)) return false;
                    if (IsPayload(field))
                    {
                        header.Payload = (Envelope.PayloadOneofCase)field;
                        header.PayloadStart = position;
                        header.PayloadLength = (int)length;
                    }
                    position += (int)length;
                    return true;
                case Fixed32Wire:
                    if (frame.Length - position < sizeof(uint)) return false;
                    position += sizeof(uint);
                    return true;
            }
            return false;
        }

        static void SetVarint(ref InboundFrame header, int field, ulong value)
        {
            switch (field)
            {
                case Envelope.ProtocolVersionFieldNumber:
                    header.ProtocolVersion = (uint)value;
                    return;
                case Envelope.SequenceFieldNumber:
                    header.Sequence = value;
                    return;
                case Envelope.AckedSequenceFieldNumber:
                    header.AckedSequence = value;
                    return;
            }
        }

        static bool TryReadVarint(ReadOnlySpan<byte> frame, ref int position, out ulong value)
        {
            value = 0;
            for (var i = 0; i < MaxVarintBytes && position < frame.Length; i++)
            {
                var current = frame[position++];
                value |= (ulong)(current & 0x7F) << (7 * i);
                if (current < 0x80) return true;
            }
            return false;
        }

        static bool IsPayload(int field) => field < PayloadFields.Length && PayloadFields[field];

        static bool[] BuildPayloadFields()
        {
            var fields = Envelope.Descriptor.Oneofs[0].Fields;
            var highest = 0;
            for (var i = 0; i < fields.Count; i++) highest = Math.Max(highest, fields[i].FieldNumber);
            var payload = new bool[highest + 1];
            for (var i = 0; i < fields.Count; i++) payload[fields[i].FieldNumber] = true;
            return payload;
        }
    }
}
