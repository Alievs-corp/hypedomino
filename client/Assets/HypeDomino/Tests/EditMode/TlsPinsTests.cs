using System;
using HypeDomino.Net;
using NUnit.Framework;

namespace HypeDomino.Tests.EditMode
{
    public sealed class TlsPinsTests
    {
        public const string Certificate =
            "MIIBezCCASGgAwIBAgIUZYz3uymgyZAfjrH0bv9RN46SmBEwCgYIKoZIzj0EAwIwEzERMA8GA1UEAwwIcGluLnRlc3QwHhcNMjYxMTI0MTYxNDQxWhcNMzYxMTIxMTYxNDQxWjATMREwDwYDVQQDDAhwaW4udGVzdDBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABLkjhOLe1D6v92U025xOZLYjZUU7mMDKmskIqPmd9AScpOAG4nptgIIY3A7V3ik6KfQA4ZrYU/Ym4dAAnrdayuejUzBRMB0GA1UdDgQWBBSeuCMrsynb/GS+cWyDx/ZnO4ILmjAfBgNVHSMEGDAWgBSeuCMrsynb/GS+cWyDx/ZnO4ILmjAPBgNVHRMBAf8EBTADAQH/MAoGCCqGSM49BAMCA0gAMEUCIHKQx2JpqVm06x9lNRAZ0TvSDhxcuYPkZQYhRv5Wbb9mAiEA/Ct32ke46wschO1C6WelMsge1qGZ4QkNk+JPOWuVAyw=";
        public const string LivePin = "Kvnv9UcsTIEuN960Y+JIKmcqfSfkig/mWtAp8IAUyRo=";
        public const string BackupPin = "8uh8n3Pcp3e71PLYlNlbYVBH12GTyGBtqWT99GhjAZY=";
        public const string BothPins = LivePin + "," + BackupPin;

        static byte[] Der => Convert.FromBase64String(Certificate);

        [Test]
        public void ThePinIsTheSha256OfTheSubjectPublicKeyInfo()
        {
            Assert.AreEqual(LivePin, TlsPins.PinOf(Der));
        }

        [Test]
        public void ACertificateWithAPinnedKeyIsAllowed()
        {
            Assert.IsTrue(TlsPins.Parse(new[] { LivePin, BackupPin }).Allows(Der));
            Assert.IsTrue(TlsPins.Parse(new[] { BackupPin, LivePin }).Allows(Der));
        }

        [Test]
        public void ACertificateWhoseKeyIsNotPinnedIsRefused()
        {
            var otherPins = TlsPins.Parse(new[] { BackupPin, Convert.ToBase64String(new byte[TlsPins.PinBytes]) });

            Assert.IsFalse(otherPins.Allows(Der));
        }

        [Test]
        public void AMalformedOrMissingCertificateIsRefused()
        {
            var pins = TlsPins.Parse(TlsPins.Split(BothPins));
            var truncated = new byte[Der.Length / 2];
            Array.Copy(Der, truncated, truncated.Length);

            Assert.IsFalse(pins.Allows(null));
            Assert.IsFalse(pins.Allows(Array.Empty<byte>()));
            Assert.IsFalse(pins.Allows(truncated));
            Assert.IsFalse(pins.Allows(new byte[] { 0x30, 0x84, 0xFF, 0xFF, 0xFF, 0xFF }));
        }

        [Test]
        public void NoPinsTurnPinningOffAndAllowNothing()
        {
            Assert.IsTrue(TlsPins.Parse(null).IsOff);
            Assert.IsTrue(TlsPins.Parse(TlsPins.Split("")).IsOff);
            Assert.IsFalse(TlsPins.Off.Allows(Der));
        }

        [Test]
        public void TheBuildVariableSplitsOnCommas()
        {
            CollectionAssert.AreEqual(new[] { LivePin, BackupPin }, TlsPins.Split(BothPins));
            CollectionAssert.IsEmpty(TlsPins.Split(null));
        }

        [TestCase(LivePin)]
        [TestCase(LivePin + ",not-base64")]
        [TestCase(LivePin + ",AAAA")]
        [TestCase(LivePin + ", " + BackupPin)]
        public void APinSetWithoutABackupOrWithABadPinIsRefused(string pins)
        {
            Assert.Throws<ArgumentException>(() => TlsPins.Parse(TlsPins.Split(pins)));
        }
    }
}
