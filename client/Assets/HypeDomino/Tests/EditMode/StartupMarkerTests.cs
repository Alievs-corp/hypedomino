using System.Collections.Generic;
using HypeDomino.Game;
using Hypedomino.V1;
using NUnit.Framework;

namespace HypeDomino.Tests.EditMode
{
    public sealed class StartupMarkerTests
    {
        readonly List<string> lines = new List<string>();
        AppFlow flow;
        double seconds;

        [SetUp]
        public void SetUp()
        {
            lines.Clear();
            seconds = 0d;
            flow = new AppFlow();
            new StartupMarker(flow, () => seconds, lines.Add);
        }

        [Test]
        public void NothingIsLoggedWhileTheAppBoots()
        {
            seconds = 1.5d;
            Assert.IsEmpty(lines);
        }

        [Test]
        public void TheFirstScreenAfterBootIsLoggedWithTheTimeSinceStartup()
        {
            seconds = 2.5d;
            flow.Hello(new ServerHello { Status = HelloStatus.Accepted });
            CollectionAssert.AreEqual(new[] { "Startup state=Lobby ms=2500" }, lines);
        }

        [Test]
        public void OnlyTheFirstScreenIsLogged()
        {
            seconds = 0.75d;
            flow.SignInNeeded();
            seconds = 9d;
            flow.Hello(new ServerHello { Status = HelloStatus.Accepted, InMatch = true });
            flow.View(MatchState.Finished);
            flow.LeaveMatch();
            CollectionAssert.AreEqual(new[] { "Startup state=Login ms=750" }, lines);
        }

        [TestCase(HelloStatus.UpdateRequired, "Startup state=UpdateRequired ms=1250")]
        [TestCase(HelloStatus.Maintenance, "Startup state=Maintenance ms=1250")]
        public void AScreenThatBlocksPlayIsLoggedAsItself(HelloStatus status, string line)
        {
            seconds = 1.25d;
            flow.Hello(new ServerHello { Status = status });
            CollectionAssert.AreEqual(new[] { line }, lines);
        }
    }
}
