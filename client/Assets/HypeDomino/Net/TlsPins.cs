using System;
using System.Collections.Generic;
using System.Security.Cryptography;

namespace HypeDomino.Net
{
    public sealed class TlsPins
    {
        public const int PinBytes = 32;
        public const int MinPins = 2;
        const int PinBase64Chars = 44;

        public static readonly TlsPins Off = new(Array.Empty<string>());

        readonly HashSet<string> pins;

        TlsPins(IEnumerable<string> pins) => this.pins = new HashSet<string>(pins, StringComparer.Ordinal);

        public bool IsOff => pins.Count == 0;

        public static TlsPins Parse(IReadOnlyList<string> pins)
        {
            if (pins == null || pins.Count == 0) return Off;
            if (pins.Count < MinPins) throw new ArgumentException($"pinning needs at least {MinPins} pins, the live key and a backup key, got {pins.Count}");
            foreach (var pin in pins)
                if (!IsPin(pin)) throw new ArgumentException($"a pin is the standard base64 SHA-256 of a SubjectPublicKeyInfo, got '{pin}'");
            return new TlsPins(pins);
        }

        public static string[] Split(string commaSeparated) =>
            string.IsNullOrWhiteSpace(commaSeparated) ? Array.Empty<string>() : commaSeparated.Split(new[] { ',' }, StringSplitOptions.RemoveEmptyEntries);

        public bool Allows(byte[] certificate)
        {
            if (IsOff || certificate == null) return false;
            try
            {
                return pins.Contains(PinOf(certificate));
            }
            catch (ArgumentException)
            {
                return false;
            }
        }

        public static string PinOf(byte[] certificate)
        {
            var spki = SubjectPublicKeyInfo(certificate);
            using var sha = SHA256.Create();
            return Convert.ToBase64String(sha.ComputeHash(spki.Array, spki.Offset, spki.Count));
        }

        public static ArraySegment<byte> SubjectPublicKeyInfo(byte[] certificate)
        {
            var at = Element(certificate, 0, Sequence).ContentStart;
            at = Element(certificate, at, Sequence).ContentStart;
            if (at < certificate.Length && certificate[at] == ExplicitVersion) at = Element(certificate, at, ExplicitVersion).End;
            at = Element(certificate, at, Integer).End;
            for (var skipped = 0; skipped < FieldsBeforeKey; skipped++) at = Element(certificate, at, Sequence).End;
            var key = Element(certificate, at, Sequence);
            return new ArraySegment<byte>(certificate, at, key.End - at);
        }

        const byte Sequence = 0x30;
        const byte Integer = 0x02;
        const byte ExplicitVersion = 0xA0;
        const int FieldsBeforeKey = 4;

        readonly struct DerElement
        {
            public readonly int ContentStart;
            public readonly int End;

            public DerElement(int contentStart, int end)
            {
                ContentStart = contentStart;
                End = end;
            }
        }

        static DerElement Element(byte[] der, int at, byte tag)
        {
            if (at + 2 > der.Length || der[at] != tag) throw new ArgumentException("the certificate is not the expected DER");
            var first = der[at + 1];
            var content = at + 2;
            long length = first;
            if (first > 0x7F)
            {
                var lengthBytes = first & 0x7F;
                if (lengthBytes == 0 || lengthBytes > 4 || content + lengthBytes > der.Length) throw new ArgumentException("the certificate has a bad DER length");
                length = 0;
                for (var i = 0; i < lengthBytes; i++) length = (length << 8) | der[content + i];
                content += lengthBytes;
            }
            if (content + length > der.Length) throw new ArgumentException("the certificate is truncated");
            return new DerElement(content, (int)(content + length));
        }

        static bool IsPin(string pin)
        {
            if (pin == null || pin.Length != PinBase64Chars) return false;
            try
            {
                return Convert.FromBase64String(pin).Length == PinBytes;
            }
            catch (FormatException)
            {
                return false;
            }
        }
    }
}
