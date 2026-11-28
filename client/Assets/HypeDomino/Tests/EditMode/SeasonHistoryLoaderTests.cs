using System;
using System.Collections.Generic;
using HypeDomino.Game.Ranking;
using HypeDomino.Net.Api;
using HypeDomino.Net.Auth;
using NUnit.Framework;

namespace HypeDomino.Tests.EditMode
{
    public sealed class SeasonHistoryLoaderTests
    {
        const string HistoryJson = @"{""seasons"":[" +
            @"{""id"":""season.2026_08"",""name_key"":""season.2026_08.name"",""starts_at"":""2026-08-01T00:00:00Z"",""ends_at"":""2026-11-01T00:00:00Z"",""modes"":[" +
            @"{""mode"":""1v1"",""rank"":""Gold II"",""tier"":""Gold"",""lp"":42,""position"":17,""reward"":{""granted"":true,""coins"":700,""bundle_id"":""bundle.rank_gold"",""emblems"":[]}}," +
            @"{""mode"":""2v2"",""rank"":""Challenger"",""tier"":""Challenger"",""lp"":640,""position"":1,""reward"":{""granted"":false,""coins"":0,""emblems"":[]}}]}," +
            @"{""id"":""season.2026_05"",""name_key"":""season.2026_05.name"",""starts_at"":""2026-05-01T00:00:00Z"",""ends_at"":""2026-08-01T00:00:00Z"",""modes"":[" +
            @"{""mode"":""1v1"",""rank"":""Silver III"",""tier"":""Silver"",""lp"":77,""position"":9,""reward"":{""granted"":true,""coins"":400,""emblems"":[]}}]}]}";

        sealed class Sent
        {
            public string Method;
            public string PathAndQuery;
            public string Body;
            public Action<HttpReply> Done;
        }

        sealed class Api : IApiTransport
        {
            public readonly List<Sent> Requests = new();

            public void Send(string method, string pathAndQuery, string body, string authorization, string mac, Action<HttpReply> done) =>
                Requests.Add(new Sent { Method = method, PathAndQuery = pathAndQuery, Body = body, Done = done });
        }

        sealed class Auth : IAuthTransport
        {
            public readonly List<Action<HttpReply>> Refreshes = new();

            public void Post(string path, string json, Action<HttpReply> done) => Refreshes.Add(done);
        }

        sealed class Identity : IIdentityProvider
        {
            public bool Supports(AuthProvider provider) => false;

            public void RequestIdToken(AuthProvider provider, string nonce, Action<IdTokenResult> done) => throw new NotSupportedException();
        }

        sealed class Store : ISecureStore
        {
            readonly Dictionary<string, string> values = new();

            public string Read(string key) => values.TryGetValue(key, out var value) ? value : null;

            public void Write(string key, string value) => values[key] = value;

            public void Delete(string key) => values.Remove(key);
        }

        Api api;
        SeasonHistoryLoader loader;
        List<SeasonHistoryLoadResult> results;

        [SetUp]
        public void SignIn()
        {
            api = new Api();
            var auth = new Auth();
            var store = new Store();
            var session = new AuthSession(auth, new Identity(), store);
            store.Write(AuthSession.RefreshTokenKey, "refresh-0");
            session.Resume(_ => { });
            auth.Refreshes[0](new HttpReply(200, "{\"access_token\":\"access-1\",\"refresh_token\":\"refresh-1\",\"expires_in\":900,\"request_key\":\"" + Convert.ToBase64String(new byte[32]) + "\",\"account_id\":\"acc-1\"}"));
            loader = new SeasonHistoryLoader(new ApiClient(api, session));
            results = new List<SeasonHistoryLoadResult>();
        }

        [Test]
        public void LoadReadsThePlayersOwnPastSeasonsNewestFirst()
        {
            Assert.IsTrue(loader.Load(results.Add));
            Assert.IsTrue(loader.Busy);
            Assert.AreEqual(1, api.Requests.Count);
            Assert.AreEqual("GET", api.Requests[0].Method);
            Assert.AreEqual("/ranked/seasons/history", api.Requests[0].PathAndQuery);
            Assert.IsNull(api.Requests[0].Body);

            api.Requests[0].Done(new HttpReply(200, HistoryJson));
            Assert.IsFalse(loader.Busy);
            Assert.AreEqual(SeasonHistoryLoadOutcome.Loaded, results[0].Outcome);
            var history = results[0].History;
            Assert.AreEqual(2, history.Count);
            Assert.AreEqual("season.2026_08", history.SeasonAt(0).Id);
            Assert.AreEqual("season.2026_08.name", history.SeasonAt(0).NameKey);
            Assert.AreEqual("season.2026_05", history.SeasonAt(1).Id);

            Assert.IsTrue(history.SeasonAt(0).TryFind("1v1", out var solo));
            Assert.AreEqual(RankTier.Gold, solo.Badge.Tier);
            Assert.AreEqual(2, solo.Badge.Division);
            Assert.AreEqual(42, solo.LP);
            Assert.AreEqual(17, solo.Position);
            Assert.IsTrue(solo.Rewarded);
            Assert.AreEqual(700, solo.RewardCoins);

            Assert.IsTrue(history.SeasonAt(0).TryFind("2v2", out var teams));
            Assert.AreEqual(RankTier.Challenger, teams.Badge.Tier);
            Assert.AreEqual(1, teams.Position);
            Assert.IsFalse(teams.Rewarded);

            Assert.IsFalse(history.SeasonAt(1).TryFind("2v2", out _));
        }

        [Test]
        public void AnEmptyHistoryIsLoaded()
        {
            loader.Load(results.Add);
            api.Requests[0].Done(new HttpReply(200, "{\"seasons\":[]}"));

            Assert.AreEqual(SeasonHistoryLoadOutcome.Loaded, results[0].Outcome);
            Assert.AreEqual(0, results[0].History.Count);
        }

        [Test]
        public void ASecondLoadWhileOneIsInFlightSendsNothing()
        {
            Assert.IsTrue(loader.Load(results.Add));
            Assert.IsFalse(loader.Load(results.Add));
            Assert.AreEqual(1, api.Requests.Count);

            api.Requests[0].Done(new HttpReply(200, HistoryJson));
            Assert.AreEqual(1, results.Count);
            Assert.IsTrue(loader.Load(results.Add));
            Assert.AreEqual(2, api.Requests.Count);
        }

        [Test]
        public void ALoadWithoutACallbackSendsNothing()
        {
            Assert.IsFalse(loader.Load(null));
            Assert.IsFalse(loader.Busy);
            Assert.IsEmpty(api.Requests);
        }

        [TestCase(HttpReply.NoResponse, null, SeasonHistoryLoadOutcome.Offline)]
        [TestCase(429, "{\"error\":\"rate_limited\"}", SeasonHistoryLoadOutcome.RateLimited)]
        [TestCase(401, "{\"error\":\"access_rejected\"}", SeasonHistoryLoadOutcome.SignInRequired)]
        [TestCase(500, "{\"error\":\"internal\"}", SeasonHistoryLoadOutcome.Failed)]
        [TestCase(200, "", SeasonHistoryLoadOutcome.Failed)]
        [TestCase(200, "not json", SeasonHistoryLoadOutcome.Failed)]
        public void AFailedReplyEndsTheLoadWithoutAHistory(int status, string body, SeasonHistoryLoadOutcome outcome)
        {
            loader.Load(results.Add);
            api.Requests[0].Done(new HttpReply(status, body));

            Assert.IsFalse(loader.Busy);
            Assert.AreEqual(outcome, results[0].Outcome);
            Assert.IsNull(results[0].History);
        }

        [TestCase(@"{""id"":"""",""starts_at"":""2026-08-01T00:00:00Z"",""ends_at"":""2026-11-01T00:00:00Z"",""modes"":[{""mode"":""1v1"",""rank"":""Gold II"",""tier"":""Gold"",""lp"":1,""position"":1}]}")]
        [TestCase(@"{""id"":""s"",""starts_at"":""2026-11-01T00:00:00Z"",""ends_at"":""2026-08-01T00:00:00Z"",""modes"":[{""mode"":""1v1"",""rank"":""Gold II"",""tier"":""Gold"",""lp"":1,""position"":1}]}")]
        [TestCase(@"{""id"":""s"",""starts_at"":""2026-08-01T00:00:00Z"",""ends_at"":""2026-11-01T00:00:00Z"",""modes"":[]}")]
        [TestCase(@"{""id"":""s"",""starts_at"":""2026-08-01T00:00:00Z"",""ends_at"":""2026-11-01T00:00:00Z"",""modes"":[{""mode"":""1v1"",""rank"":"""",""tier"":"""",""lp"":0,""position"":1}]}")]
        [TestCase(@"{""id"":""s"",""starts_at"":""2026-08-01T00:00:00Z"",""ends_at"":""2026-11-01T00:00:00Z"",""modes"":[{""mode"":""1v1"",""rank"":""Gold II"",""tier"":""Gold"",""lp"":1,""position"":0}]}")]
        [TestCase(@"{""id"":""s"",""starts_at"":""2026-08-01T00:00:00Z"",""ends_at"":""2026-11-01T00:00:00Z"",""modes"":[{""mode"":""1v1"",""rank"":""Gold II"",""tier"":""Gold"",""lp"":-1,""position"":1}]}")]
        [TestCase(@"{""id"":""s"",""starts_at"":""2026-08-01T00:00:00Z"",""ends_at"":""2026-11-01T00:00:00Z"",""modes"":[{""mode"":""1v1"",""rank"":""Master"",""tier"":""Gold"",""lp"":1,""position"":1}]}")]
        [TestCase(@"{""id"":""s"",""starts_at"":""2026-08-01T00:00:00Z"",""ends_at"":""2026-11-01T00:00:00Z"",""modes"":[{""mode"":""1v1"",""rank"":""Gold II"",""tier"":""Gold"",""lp"":1,""position"":1,""reward"":{""granted"":false,""coins"":5}}]}")]
        [TestCase(@"{""id"":""s"",""starts_at"":""2026-08-01T00:00:00Z"",""ends_at"":""2026-11-01T00:00:00Z"",""modes"":[{""mode"":""1v1"",""rank"":""Gold II"",""tier"":""Gold"",""lp"":1,""position"":1},{""mode"":""1v1"",""rank"":""Gold II"",""tier"":""Gold"",""lp"":1,""position"":1}]}")]
        public void AnInconsistentSeasonRefusesTheWholeReply(string season)
        {
            Assert.IsFalse(SeasonHistory.TryRead("{\"seasons\":[" + season + "]}", out var history));
            Assert.IsNull(history);
        }

        [Test]
        public void ReadingTheSeasonsAllocatesNothing()
        {
            SeasonHistory.TryRead(HistoryJson, out var history);
            TestDelegate step = () =>
            {
                for (var i = 0; i < history.Count; i++)
                {
                    var season = history.SeasonAt(i);
                    if (season.TryFind("1v1", out var solo)) _ = solo.LP + solo.Position + solo.RewardCoins;
                    _ = season.ModeCount + season.ModeAt(0).Position;
                }
            };
            step();
            Assert.That(step, Allocates.Nothing);
        }
    }
}
