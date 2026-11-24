using HypeDomino.Game;
using HypeDomino.Game.Lobby;
using HypeDomino.Game.Login;
using HypeDomino.UI.Screens;
using HypeDomino.UI.Tokens;
using NUnit.Framework;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Tests.PlayMode
{
    public sealed class ScreenHostTests
    {
        GameObject root;
        ScreenStack stack;
        ScreenHost host;

        [SetUp]
        public void Build()
        {
            root = new GameObject("Host");
            stack = root.AddComponent<ScreenStack>();
            stack.enabled = false;
            host = new ScreenHost(stack);
        }

        [TearDown]
        public void DestroyHost() => Object.Destroy(root);

        [Test]
        public void TheHostBuildsTheLoginAndLobbyScreensHidden()
        {
            var login = host.ScreenOf(AppState.Login);
            var lobby = host.ScreenOf(AppState.Lobby);

            Assert.IsNotNull(login.GetComponent<LoginScreenView>());
            Assert.IsNotNull(lobby.GetComponent<LobbyScreenView>());
            Assert.AreEqual(PlayModes.Count + 2, lobby.GetComponentsInChildren<Button>(true).Length);
            Assert.IsFalse(login.IsShown);
            Assert.IsFalse(lobby.IsShown);
            Assert.IsNull(host.ScreenOf(AppState.Match));
            Assert.AreEqual(0, stack.Count);
        }

        [Test]
        public void TheShownScreenFollowsTheAppState()
        {
            var login = host.ScreenOf(AppState.Login);
            var lobby = host.ScreenOf(AppState.Lobby);

            host.Follow(AppState.Login);
            stack.Advance(TokenDurationMs.Slow);
            Assert.AreSame(login, stack.Top);
            Assert.IsTrue(login.IsShown);
            Assert.IsTrue(login.TakesInput);

            host.Follow(AppState.Lobby);
            Assert.AreSame(lobby, stack.Top);
            Assert.AreEqual(1, stack.Count);
            Assert.IsTrue(lobby.TakesInput);
            Assert.IsFalse(login.TakesInput);
            stack.Advance(TokenDurationMs.Slow);
            Assert.IsTrue(lobby.IsShown);
            Assert.IsFalse(login.IsShown);

            host.Follow(AppState.Match);
            Assert.AreEqual(0, stack.Count);
            Assert.IsFalse(lobby.TakesInput);
            stack.Advance(TokenDurationMs.Base);
            Assert.IsFalse(lobby.IsShown);
        }
    }
}
