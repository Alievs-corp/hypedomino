using System;
using Google.Protobuf;
using Hypedomino.V1;

namespace HypeDomino.Net
{
    public sealed class ProtocolSession
    {
        public const uint ProtocolVersion = 1;

        readonly Envelope hello = new Envelope { ProtocolVersion = ProtocolVersion, ClientHello = new ClientHello() };
        readonly Envelope placeTile = new Envelope { ProtocolVersion = ProtocolVersion, PlaceTileIntent = new PlaceTileIntent { Tile = new Tile() } };
        readonly Envelope pass = new Envelope { ProtocolVersion = ProtocolVersion, PassIntent = new PassIntent() };
        readonly Envelope draw = new Envelope { ProtocolVersion = ProtocolVersion, DrawIntent = new DrawIntent() };
        readonly Envelope endTurn = new Envelope { ProtocolVersion = ProtocolVersion, EndTurnIntent = new EndTurnIntent() };
        readonly Envelope loading = new Envelope { ProtocolVersion = ProtocolVersion, LoadingIntent = new LoadingIntent() };
        readonly Envelope emoji = new Envelope { ProtocolVersion = ProtocolVersion, EmojiIntent = new EmojiIntent() };
        readonly Envelope quickMessage = new Envelope { ProtocolVersion = ProtocolVersion, QuickMessageIntent = new QuickMessageIntent() };
        readonly Envelope joinQueue = new Envelope { ProtocolVersion = ProtocolVersion, JoinQueueIntent = new JoinQueueIntent() };
        readonly Envelope leaveQueue = new Envelope { ProtocolVersion = ProtocolVersion, LeaveQueueIntent = new LeaveQueueIntent() };

        byte[] buffer = Array.Empty<byte>();
        ulong nonce;
        long helloSentAtMs;
        ulong awaitedSequence;
        long awaitedSentAtMs;

        public ulong Acked { get; private set; }
        public ulong Sequence { get; private set; }
        public long ClockOffsetMs { get; private set; }
        public long RoundTripMs { get; private set; }

        public event Action<ServerHello> HelloReceived;
        public event Action<Error> ErrorReceived;
        public event Action<MatchView> ViewReceived;
        public event Action<EmojiShown> EmojiReceived;
        public event Action<QuickMessageShown> QuickMessageReceived;
        public event Action<QueueState> QueueReceived;

        public long ServerNowMs(long nowMs) => nowMs + ClockOffsetMs;

        public ArraySegment<byte> Hello(ClientHello clientHello, long nowMs)
        {
            clientHello.LastAckedSequence = Acked;
            hello.ClientHello = clientHello;
            hello.Sequence = Acked;
            helloSentAtMs = nowMs;
            awaitedSequence = 0;
            return Write(hello);
        }

        public ArraySegment<byte> PlaceTile(uint low, uint high, BoardEnd end, long nowMs)
        {
            var intent = placeTile.PlaceTileIntent;
            intent.Tile.Low = low;
            intent.Tile.High = high;
            intent.End = end;
            return Sequenced(placeTile, nowMs);
        }

        public ArraySegment<byte> Pass(long nowMs) => Sequenced(pass, nowMs);

        public ArraySegment<byte> Draw(long nowMs) => Sequenced(draw, nowMs);

        public ArraySegment<byte> EndTurn(long nowMs) => Sequenced(endTurn, nowMs);

        public ArraySegment<byte> Loading(uint progress, bool ready)
        {
            loading.LoadingIntent.Progress = progress;
            loading.LoadingIntent.Ready = ready;
            return Unsequenced(loading);
        }

        public ArraySegment<byte> Emoji(string itemId)
        {
            emoji.EmojiIntent.ItemId = itemId;
            return Unsequenced(emoji);
        }

        public ArraySegment<byte> QuickMessage(string key, bool teamOnly)
        {
            quickMessage.QuickMessageIntent.Key = key;
            quickMessage.QuickMessageIntent.TeamOnly = teamOnly;
            return Unsequenced(quickMessage);
        }

        public ArraySegment<byte> JoinQueue(Mode mode, QueueKind queue, IntegrityProof integrity)
        {
            var intent = joinQueue.JoinQueueIntent;
            intent.Mode = mode;
            intent.Queue = queue;
            intent.Integrity = integrity;
            return Unsequenced(joinQueue);
        }

        public ArraySegment<byte> LeaveQueue() => Unsequenced(leaveQueue);

        public bool Receive(ReadOnlySpan<byte> frame, long nowMs)
        {
            if (!FrameDecoder.TryRead(frame, out var header)) return false;
            var payload = frame.Slice(header.PayloadStart, header.PayloadLength);
            try
            {
                Dispatch(header, payload, nowMs);
            }
            catch (InvalidProtocolBufferException)
            {
                return false;
            }
            return true;
        }

        void Dispatch(in InboundFrame header, ReadOnlySpan<byte> payload, long nowMs)
        {
            switch (header.Payload)
            {
                case Envelope.PayloadOneofCase.ServerHello:
                    var serverHello = ServerHello.Parser.ParseFrom(payload);
                    Accept(header, serverHello, nowMs);
                    HelloReceived?.Invoke(serverHello);
                    return;
                case Envelope.PayloadOneofCase.MatchView:
                    var view = MatchView.Parser.ParseFrom(payload);
                    Acknowledge(header.AckedSequence, nowMs);
                    ViewReceived?.Invoke(view);
                    return;
                case Envelope.PayloadOneofCase.Error:
                    ErrorReceived?.Invoke(Error.Parser.ParseFrom(payload));
                    return;
                case Envelope.PayloadOneofCase.EmojiShown:
                    EmojiReceived?.Invoke(EmojiShown.Parser.ParseFrom(payload));
                    return;
                case Envelope.PayloadOneofCase.QuickMessageShown:
                    QuickMessageReceived?.Invoke(QuickMessageShown.Parser.ParseFrom(payload));
                    return;
                case Envelope.PayloadOneofCase.QueueState:
                    QueueReceived?.Invoke(QueueState.Parser.ParseFrom(payload));
                    return;
            }
        }

        void Accept(in InboundFrame header, ServerHello serverHello, long nowMs)
        {
            if (serverHello.Status != HelloStatus.Accepted) return;
            nonce = serverHello.SessionNonce;
            Acked = header.AckedSequence;
            Sequence = Acked;
            RoundTripMs = nowMs - helloSentAtMs;
            ClockOffsetMs = (long)serverHello.ServerTimeUnixMs - (helloSentAtMs + RoundTripMs / 2);
        }

        void Acknowledge(ulong acked, long nowMs)
        {
            Acked = acked;
            Sequence = Math.Max(Sequence, acked);
            if (awaitedSequence == 0 || acked < awaitedSequence) return;
            RoundTripMs = nowMs - awaitedSentAtMs;
            awaitedSequence = 0;
        }

        ArraySegment<byte> Sequenced(Envelope envelope, long nowMs)
        {
            envelope.Sequence = ++Sequence;
            awaitedSequence = Sequence;
            awaitedSentAtMs = nowMs;
            return Unsequenced(envelope);
        }

        ArraySegment<byte> Unsequenced(Envelope envelope)
        {
            envelope.SessionNonce = nonce;
            return Write(envelope);
        }

        ArraySegment<byte> Write(Envelope envelope)
        {
            var size = envelope.CalculateSize();
            if (buffer.Length < size) buffer = new byte[size];
            envelope.WriteTo(new Span<byte>(buffer, 0, size));
            return new ArraySegment<byte>(buffer, 0, size);
        }
    }
}
