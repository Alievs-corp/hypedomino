using System;
using Google.Protobuf;
using Hypedomino.V1;

namespace HypeDomino.Net
{
    public sealed class Session
    {
        public const uint ProtocolVersion = 1;
        public const int MaxSentFrameBytes = 8192;

        readonly IFrameTransport transport;
        readonly FrameWriter writer = new(MaxSentFrameBytes);
        ulong lastSentSequence;

        public Session(IFrameTransport transport) => this.transport = transport;

        public MatchSnapshot View { get; } = new();
        public bool Accepted { get; private set; }
        public bool Broken { get; private set; }
        public ulong SessionNonce { get; private set; }
        public ulong AckedSequence { get; private set; }

        public event Action<ServerHello> HelloReceived;
        public event Action<MatchSnapshot> ViewReceived;
        public event Action<Error> ErrorReceived;
        public event Action<QueueState> QueueStateReceived;
        public event Action<EmojiShown> EmojiReceived;
        public event Action<QuickMessageShown> QuickMessageReceived;

        public void Hello(ClientHello hello)
        {
            Accepted = false;
            Broken = false;
            SessionNonce = 0;
            hello.LastAckedSequence = AckedSequence;
            transport.Send(writer.Write(new FrameHeader(ProtocolVersion, AckedSequence + 1, 0, 0), Envelope.PayloadOneofCase.ClientHello, hello));
        }

        public ulong Send(Envelope.PayloadOneofCase intent, IMessage message)
        {
            if (!Accepted) throw new InvalidOperationException($"{intent} before an accepted hello");
            if (!IsIntent(intent)) throw new ArgumentException($"{intent} is not an intent", nameof(intent));
            var sequence = IsSequenced(intent) ? ++lastSentSequence : 0;
            transport.Send(writer.Write(new FrameHeader(ProtocolVersion, sequence, 0, SessionNonce), intent, message));
            return sequence;
        }

        public void Poll()
        {
            while (!Broken && transport.TryReceive(out var bytes))
            {
                if (!FrameReader.TryRead(bytes, out var frame) || frame.Header.ProtocolVersion > ProtocolVersion)
                {
                    Break();
                    return;
                }
                AckedSequence = frame.Header.AckedSequence;
                if (AckedSequence > lastSentSequence) lastSentSequence = AckedSequence;
                if (!TryDispatch(frame)) Break();
            }
        }

        void Break()
        {
            Broken = true;
            Accepted = false;
        }

        bool TryDispatch(in Frame frame)
        {
            switch (frame.Payload)
            {
                case Envelope.PayloadOneofCase.ServerHello:
                    var hello = ServerHello.Parser.ParseFrom(frame.PayloadBytes);
                    Accepted = hello.Status == HelloStatus.Accepted;
                    SessionNonce = Accepted ? hello.SessionNonce : 0;
                    HelloReceived?.Invoke(hello);
                    return true;
                case Envelope.PayloadOneofCase.MatchView:
                    if (!View.TryRead(frame.PayloadBytes)) return false;
                    ViewReceived?.Invoke(View);
                    return true;
                case Envelope.PayloadOneofCase.Error:
                    ErrorReceived?.Invoke(Error.Parser.ParseFrom(frame.PayloadBytes));
                    return true;
                case Envelope.PayloadOneofCase.QueueState:
                    QueueStateReceived?.Invoke(QueueState.Parser.ParseFrom(frame.PayloadBytes));
                    return true;
                case Envelope.PayloadOneofCase.EmojiShown:
                    EmojiReceived?.Invoke(EmojiShown.Parser.ParseFrom(frame.PayloadBytes));
                    return true;
                case Envelope.PayloadOneofCase.QuickMessageShown:
                    QuickMessageReceived?.Invoke(QuickMessageShown.Parser.ParseFrom(frame.PayloadBytes));
                    return true;
                case Envelope.PayloadOneofCase.None:
                    return true;
                default:
                    return false;
            }
        }

        static bool IsIntent(Envelope.PayloadOneofCase payload) => payload >= Envelope.PayloadOneofCase.PlaceTileIntent;

        static bool IsSequenced(Envelope.PayloadOneofCase intent) =>
            intent == Envelope.PayloadOneofCase.PlaceTileIntent ||
            intent == Envelope.PayloadOneofCase.PassIntent ||
            intent == Envelope.PayloadOneofCase.DrawIntent ||
            intent == Envelope.PayloadOneofCase.EndTurnIntent;
    }
}
