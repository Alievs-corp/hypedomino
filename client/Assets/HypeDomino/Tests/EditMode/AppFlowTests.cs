using System.Collections.Generic;
using HypeDomino.Game;
using Hypedomino.V1;
using NUnit.Framework;

namespace HypeDomino.Tests.EditMode
{
    public sealed class AppFlowTests
    {
        static ServerHello Accepted(bool inMatch) => new ServerHello { Status = HelloStatus.Accepted, InMatch = inMatch };

        [Test]
        public void StartsInBootAndGoesToLoginWithoutASession()
        {
            var flow = new AppFlow();
            Assert.AreEqual(AppState.Boot, flow.State);
            flow.SignInNeeded();
            Assert.AreEqual(AppState.Login, flow.State);
        }

        [Test]
        public void AnAcceptedHelloOpensTheLobbyOrTheRunningMatch()
        {
            var lobby = new AppFlow();
            lobby.Hello(Accepted(false));
            Assert.AreEqual(AppState.Lobby, lobby.State);

            var match = new AppFlow();
            match.Hello(Accepted(true));
            Assert.AreEqual(AppState.Match, match.State);
        }

        [Test]
        public void UpdateRequiredBlocksPlayAndIgnoresSignIn()
        {
            var flow = new AppFlow();
            flow.Hello(new ServerHello { Status = HelloStatus.UpdateRequired });
            Assert.AreEqual(AppState.UpdateRequired, flow.State);
            flow.SignInNeeded();
            Assert.AreEqual(AppState.UpdateRequired, flow.State);
        }

        [Test]
        public void MaintenanceShowsTheMaintenanceScreenUntilAnAcceptedHello()
        {
            var flow = new AppFlow();
            flow.Hello(new ServerHello { Status = HelloStatus.Maintenance });
            Assert.AreEqual(AppState.Maintenance, flow.State);
            flow.Hello(Accepted(false));
            Assert.AreEqual(AppState.Lobby, flow.State);
        }

        [Test]
        public void AnUnknownHelloStatusChangesNothing()
        {
            var flow = new AppFlow();
            flow.Hello(new ServerHello { Status = (HelloStatus)99 });
            Assert.AreEqual(AppState.Boot, flow.State);
        }

        [Test]
        public void AMatchViewInTheLobbyEntersTheMatchUnlessItIsFinished()
        {
            var flow = new AppFlow();
            flow.Hello(Accepted(false));
            flow.View(MatchState.Finished);
            Assert.AreEqual(AppState.Lobby, flow.State);
            flow.View(MatchState.Loading);
            Assert.AreEqual(AppState.Match, flow.State);
        }

        [Test]
        public void TheMatchCanBeLeftOnlyOnceItIsFinished()
        {
            var flow = new AppFlow();
            flow.Hello(Accepted(true));
            Assert.IsFalse(flow.LeaveMatch());
            flow.View(MatchState.Playing);
            Assert.IsFalse(flow.LeaveMatch());
            Assert.AreEqual(AppState.Match, flow.State);
            flow.View(MatchState.Finished);
            Assert.IsTrue(flow.LeaveMatch());
            Assert.AreEqual(AppState.Lobby, flow.State);
        }

        [Test]
        public void ChangedFiresOncePerNewState()
        {
            var flow = new AppFlow();
            var seen = new List<AppState>();
            flow.Changed += seen.Add;
            flow.Hello(Accepted(false));
            flow.Hello(Accepted(false));
            flow.View(MatchState.Opening);
            flow.View(MatchState.Playing);
            CollectionAssert.AreEqual(new[] { AppState.Lobby, AppState.Match }, seen);
        }
    }
}
