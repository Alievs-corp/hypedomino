using System;
using HypeDomino.Editor;
using HypeDomino.Game;
using HypeDomino.Net;
using Hypedomino.V1;
using NUnit.Framework;
using UnityEngine;

namespace HypeDomino.Tests.EditMode
{
    public sealed class ServerEndpointTests
    {
        [TestCase("http://localhost:8080", "ws://localhost:8080/ws")]
        [TestCase("https://play.example.com", "wss://play.example.com/ws")]
        [TestCase("https://play.example.com/", "wss://play.example.com/ws")]
        [TestCase("https://play.example.com:8443/game", "wss://play.example.com:8443/game/ws")]
        public void TheSocketIsTheWsRouteOfTheServerUrl(string serverUrl, string socket)
        {
            Assert.AreEqual(new Uri(socket), ServerEndpoint.SocketUriOf(serverUrl));
        }

        [Test]
        public void TheEditorTalksToTheLocalServerWithoutATimeout()
        {
            var endpoint = ServerEndpoint.Editor;

            Assert.AreEqual("http://localhost:8080", endpoint.serverUrl);
            Assert.AreEqual(ServerEndpoint.NoTimeout, endpoint.httpTimeoutSeconds);
            Assert.AreEqual(new Uri("ws://localhost:8080/ws"), endpoint.SocketUri);
        }

        [Test]
        public void AnEndpointSurvivesTheBuildFile()
        {
            var built = ServerEndpoint.Parse("https://play.example.com/", "7");

            var read = ServerEndpoint.FromJson(JsonUtility.ToJson(built));

            Assert.AreEqual("https://play.example.com", read.serverUrl);
            Assert.AreEqual(7, read.httpTimeoutSeconds);
        }

        [TestCase(null, "7")]
        [TestCase("", "7")]
        [TestCase("play.example.com", "7")]
        [TestCase("ftp://play.example.com", "7")]
        [TestCase("https://play.example.com?x=1", "7")]
        [TestCase("https://user@play.example.com", "7")]
        [TestCase("https://play.example.com", "0")]
        [TestCase("https://play.example.com", "-3")]
        [TestCase("https://play.example.com", "2.5")]
        [TestCase("https://play.example.com", "")]
        public void ABadEndpointIsRefused(string serverUrl, string timeout)
        {
            Assert.Throws<ArgumentException>(() => ServerEndpoint.Parse(serverUrl, timeout));
        }

        [Test]
        public void ABuildTakesBothEndpointArgumentsOrNeither()
        {
            Assert.IsNull(ReleaseBuild.ParseEndpoint(null, null, null));
            Assert.Throws<ArgumentException>(() => ReleaseBuild.ParseEndpoint("https://play.example.com", null, TlsPinsTests.BothPins));
            Assert.Throws<ArgumentException>(() => ReleaseBuild.ParseEndpoint(null, "7", TlsPinsTests.BothPins));
            Assert.Throws<ArgumentException>(() => ReleaseBuild.ParseEndpoint("https://play.example.com", "7", null));
            Assert.AreEqual(7, ReleaseBuild.ParseEndpoint("https://play.example.com", "7", TlsPinsTests.BothPins).httpTimeoutSeconds);
        }

        [TestCase("https://play.example.com", "")]
        [TestCase("https://play.example.com", TlsPinsTests.LivePin)]
        [TestCase("http://play.example.com", TlsPinsTests.BothPins)]
        public void AReleaseBuildNeedsHttpsAndALiveAndABackupPin(string serverUrl, string tlsPins)
        {
            Assert.Throws<ArgumentException>(() => ReleaseBuild.ParseEndpoint(serverUrl, "7", tlsPins));
        }

        [Test]
        public void PinsSurviveTheBuildFile()
        {
            var built = ReleaseBuild.ParseEndpoint("https://play.example.com", "7", TlsPinsTests.BothPins);

            var read = ServerEndpoint.FromJson(JsonUtility.ToJson(built));

            Assert.IsFalse(read.Pins.IsOff);
            Assert.IsTrue(read.MayConnect(false));
        }

        [Test]
        public void OnlyADevelopmentBuildConnectsWithoutPins()
        {
            var unpinned = ServerEndpoint.Parse("https://play.example.com", "7");

            Assert.IsTrue(unpinned.Pins.IsOff);
            Assert.IsFalse(unpinned.MayConnect(false));
            Assert.IsTrue(unpinned.MayConnect(true));
            Assert.IsTrue(ServerEndpoint.Editor.MayConnect(true));
        }

        [TestCase(RuntimePlatform.Android, Platform.Android)]
        [TestCase(RuntimePlatform.IPhonePlayer, Platform.Ios)]
        [TestCase(RuntimePlatform.WindowsEditor, Platform.Unspecified)]
        public void TheHelloNamesThePlatform(RuntimePlatform platform, Platform expected)
        {
            Assert.AreEqual(expected, NetBoot.PlatformOf(platform));
        }

        [Test]
        public void TheHelloCarriesTheAppTheDeviceAndTheAccessToken()
        {
            var hello = NetBoot.Hello("access-1", RuntimePlatform.Android, 1700000000000);

            Assert.AreEqual(Application.version, hello.AppVersion);
            Assert.AreEqual(Platform.Android, hello.Platform);
            Assert.AreEqual(SystemInfo.operatingSystem, hello.OsVersion);
            Assert.AreEqual(SystemInfo.deviceModel, hello.DeviceModel);
            Assert.AreEqual("access-1", hello.SessionToken);
            Assert.AreEqual(1700000000000UL, hello.ClientTimeUnixMs);
            Assert.AreEqual(string.Empty, hello.BuildHash);
            Assert.IsNull(hello.Integrity);
        }

        [Test]
        public void AHelloWithoutASessionSendsAnEmptyToken()
        {
            Assert.AreEqual(string.Empty, NetBoot.Hello(null, RuntimePlatform.Android, 0).SessionToken);
        }
    }
}
