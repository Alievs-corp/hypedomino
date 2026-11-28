using System;
using System.Collections.Generic;
using Google.Protobuf;
using HypeDomino.Game.Match;
using HypeDomino.Game.Ranking;
using HypeDomino.Net;
using HypeDomino.Net.Api;
using HypeDomino.Net.Auth;
using Hypedomino.V1;
using NUnit.Framework;

namespace HypeDomino.Tests.EditMode
{
    public sealed class MatchRewardsLoaderTests
    {
        const string MatchId = "0b9f6c1e-4d2a-4f61-9a57-3c8e2d7b5a10";
        const string ResultPath = "/matches/" + MatchId + "/result";
        const string CoinsOnlyJson = @"{""match"":""" + MatchId + @""",""kind"":""ranked"",""result"":""win"",""pending"":[""rating"",""pass_xp"",""missions""],""retry_after_ms"":2000,""coins"":{""match"":30,""first_win"":100},""pass_xp"":0,""missions"":[],""rating"":null}";
        const string CompleteJson = @"{""match"":""" + MatchId + @""",""kind"":""ranked"",""result"":""win"",""pending"":[],""coins"":{""match"":30,""first_win"":100},""pass_xp"":40,""missions"":[{""mission"":""mission.chain_1"",""period"":""once"",""coins"":0,""pass_xp"":0,""medals"":[""medal.chain_1""]},{""mission"":""mission.win_3"",""period"":""2026-11-03"",""coins"":50,""pass_xp"":20,""medals"":[]}],""rating"":{""placement"":false,""lp_delta"":21,""rank_before"":""Gold III"",""lp_before"":90,""rank_after"":""Gold II"",""lp_after"":11}}";

        sealed class Sent
        {
            public string Method;
            public string PathAndQuery;
            public Action<HttpReply> Done;
        }

        sealed class Api : IApiTransport
        {
            public readonly List<Sent> Requests = new();

            public void Send(string method, string pathAndQuery, string body, string authorization, string mac, Action<HttpReply> done) =>
                Requests.Add(new Sent { Method = method, PathAndQuery = pathAndQuery, Done = done });
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
        MatchRewardsLoader loader;
        MatchResult result;
        List<MatchRewardsLoadResult> results;

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
            loader = new MatchRewardsLoader(new ApiClient(api, session));
            result = Finished(MatchId);
            results = new List<MatchRewardsLoadResult>();
        }

        static MatchResult Finished(string matchId)
        {
            var view = new MatchView { State = MatchState.Finished, Seat = Seat.X, Mode = Mode._1V1, Winner = Team.A, MatchId = matchId };
            var snapshot = new MatchSnapshot();
            Assert.IsTrue(snapshot.TryRead(view.ToByteArray()));
            var finished = new MatchResult();
            Assert.IsTrue(finished.Read(snapshot));
            return finished;
        }

        void Answer(int request, string json) => api.Requests[request].Done(new HttpReply(200, json));

        static string Pending(string json, int index) => json.Replace(@"""pending"":[],", @"""pending"":[""rating""],""retry_after_ms"":" + (1000 + index) + ",");

        [Test]
        public void AFinishedViewNamesTheMatchToRead()
        {
            Assert.AreEqual(MatchId, result.MatchId);
            Assert.IsNull(result.Rewards);
            Assert.IsFalse(loader.Load(Finished(string.Empty), results.Add));
            Assert.IsFalse(loader.Load(result, null));
            Assert.IsEmpty(api.Requests);
        }

        [Test]
        public void ACompleteReplyFillsTheResultAndEndsTheLoad()
        {
            Assert.IsTrue(loader.Load(result, results.Add));
            Assert.IsTrue(loader.Busy);
            Assert.AreEqual(1, api.Requests.Count);
            Assert.AreEqual("GET", api.Requests[0].Method);
            Assert.AreEqual(ResultPath, api.Requests[0].PathAndQuery);

            Answer(0, CompleteJson);
            Assert.IsFalse(loader.Busy);
            Assert.AreEqual(1, results.Count);
            Assert.AreEqual(MatchRewardsLoadOutcome.Loaded, results[0].Outcome);
            Assert.IsTrue(results[0].Final);
            var rewards = result.Rewards;
            Assert.AreSame(rewards, results[0].Rewards);
            Assert.IsTrue(rewards.Complete);
            Assert.AreEqual(MatchRewardParts.None, rewards.Pending);
            Assert.AreEqual(30, rewards.MatchCoins);
            Assert.AreEqual(100, rewards.FirstWinCoins);
            Assert.AreEqual(40, rewards.PassXp);
            Assert.AreEqual(2, rewards.Missions.Count);
            Assert.AreEqual("mission.chain_1", rewards.Missions[0].Mission);
            CollectionAssert.AreEqual(new[] { "medal.chain_1" }, rewards.Missions[0].Medals);
            Assert.AreEqual(50, rewards.Missions[1].Coins);
            Assert.AreEqual(20, rewards.Missions[1].PassXp);
            Assert.IsTrue(rewards.HasRating);
            Assert.IsFalse(rewards.Rating.Placement);
            Assert.AreEqual(21, rewards.Rating.LpDelta);
            Assert.AreEqual(RankTier.Gold, rewards.Rating.Before.Tier);
            Assert.AreEqual(3, rewards.Rating.Before.Division);
            Assert.AreEqual(2, rewards.Rating.After.Division);
            Assert.AreEqual(90, rewards.Rating.LpBefore);
            Assert.AreEqual(11, rewards.Rating.LpAfter);

            loader.Tick(0);
            loader.Tick(100);
            Assert.AreEqual(1, api.Requests.Count);
        }

        [Test]
        public void APartialReplyIsShownAndReadAgainAfterTheServersDelay()
        {
            loader.Load(result, results.Add);
            Answer(0, CoinsOnlyJson);
            Assert.IsTrue(loader.Busy);
            Assert.AreEqual(MatchRewardsLoadOutcome.Partial, results[0].Outcome);
            Assert.IsFalse(results[0].Final);
            Assert.IsFalse(result.Rewards.Complete);
            Assert.IsTrue(result.Rewards.IsPending(MatchRewardParts.Rating));
            Assert.IsTrue(result.Rewards.IsPending(MatchRewardParts.PassXp));
            Assert.IsTrue(result.Rewards.IsPending(MatchRewardParts.Missions));
            Assert.IsFalse(result.Rewards.IsPending(MatchRewardParts.Coins));
            Assert.AreEqual(30, result.Rewards.MatchCoins);
            Assert.IsFalse(result.Rewards.HasRating);

            loader.Tick(10);
            loader.Tick(11.9);
            Assert.AreEqual(1, api.Requests.Count);
            loader.Tick(12);
            Assert.AreEqual(2, api.Requests.Count);
            Assert.AreEqual(ResultPath, api.Requests[1].PathAndQuery);
            loader.Tick(20);
            Assert.AreEqual(2, api.Requests.Count);

            Answer(1, CompleteJson);
            Assert.IsFalse(loader.Busy);
            Assert.AreEqual(2, results.Count);
            Assert.AreEqual(MatchRewardsLoadOutcome.Loaded, results[1].Outcome);
            Assert.IsTrue(results[1].Final);
            Assert.AreEqual(40, result.Rewards.PassXp);
        }

        [Test]
        public void TheLoadGivesUpAfterTheLastReadWithThePartsItHas()
        {
            loader.Load(result, results.Add);
            var now = 0.0;
            for (var read = 0; read < MatchRewardsLoader.MaxReads; read++)
            {
                Assert.AreEqual(read + 1, api.Requests.Count);
                Answer(read, Pending(CompleteJson, read));
                loader.Tick(now);
                now += 10;
                loader.Tick(now);
            }
            Assert.AreEqual(MatchRewardsLoader.MaxReads, api.Requests.Count);
            Assert.AreEqual(MatchRewardsLoader.MaxReads, results.Count);
            Assert.IsFalse(loader.Busy);
            var last = results[MatchRewardsLoader.MaxReads - 1];
            Assert.AreEqual(MatchRewardsLoadOutcome.Partial, last.Outcome);
            Assert.IsTrue(last.Final);
            Assert.IsTrue(result.Rewards.IsPending(MatchRewardParts.Rating));
            Assert.AreEqual(1004 / 1000.0, result.Rewards.RetryAfterSeconds, 1e-9);
            for (var read = 0; read < MatchRewardsLoader.MaxReads - 1; read++) Assert.IsFalse(results[read].Final);
        }

        [TestCase(404, "{\"error\":\"match.not_played\"}", MatchRewardsLoadOutcome.NotPlayed)]
        [TestCase(HttpReply.NoResponse, null, MatchRewardsLoadOutcome.Offline)]
        [TestCase(429, "{\"error\":\"rate_limited\"}", MatchRewardsLoadOutcome.RateLimited)]
        [TestCase(401, "{\"error\":\"access_rejected\"}", MatchRewardsLoadOutcome.SignInRequired)]
        [TestCase(500, "{\"error\":\"internal\"}", MatchRewardsLoadOutcome.Failed)]
        [TestCase(200, "not json", MatchRewardsLoadOutcome.Failed)]
        [TestCase(200, "{\"match\":\"" + MatchId + "\",\"pending\":[\"rating\"],\"coins\":{\"match\":0,\"first_win\":0},\"pass_xp\":0,\"missions\":[],\"rating\":null}", MatchRewardsLoadOutcome.Failed)]
        [TestCase(200, "{\"match\":\"6d1c7a52-0e8b-4c43-b1f5-5f0a9b2c4e77\",\"pending\":[],\"coins\":{\"match\":0,\"first_win\":0},\"pass_xp\":0,\"missions\":[],\"rating\":null}", MatchRewardsLoadOutcome.Failed)]
        public void ARefusedOrUnreadableReplyEndsTheLoadWithoutRewards(int status, string body, MatchRewardsLoadOutcome outcome)
        {
            loader.Load(result, results.Add);
            api.Requests[0].Done(new HttpReply(status, body));

            Assert.IsFalse(loader.Busy);
            Assert.AreEqual(1, results.Count);
            Assert.AreEqual(outcome, results[0].Outcome);
            Assert.IsTrue(results[0].Final);
            Assert.IsNull(results[0].Rewards);
            Assert.IsNull(result.Rewards);
            loader.Tick(0);
            loader.Tick(100);
            Assert.AreEqual(1, api.Requests.Count);
        }

        [Test]
        public void APlacementMatchCarriesNoRank()
        {
            loader.Load(result, results.Add);
            Answer(0, CompleteJson.Replace(@"{""placement"":false,""lp_delta"":21,""rank_before"":""Gold III"",""lp_before"":90,""rank_after"":""Gold II"",""lp_after"":11}", @"{""placement"":true,""lp_delta"":0,""lp_before"":0,""lp_after"":0}"));
            Assert.AreEqual(MatchRewardsLoadOutcome.Loaded, results[0].Outcome);
            Assert.IsTrue(result.Rewards.HasRating);
            Assert.IsTrue(result.Rewards.Rating.Placement);
            Assert.IsFalse(result.Rewards.Rating.After.Placed);
        }

        [Test]
        public void RewardsOfAnotherMatchAreNotShown()
        {
            Assert.IsTrue(MatchRewards.TryRead(CompleteJson, out var rewards));
            Assert.IsFalse(Finished("6d1c7a52-0e8b-4c43-b1f5-5f0a9b2c4e77").ShowRewards(rewards));
            Assert.IsTrue(result.ShowRewards(rewards));
        }

        [Test]
        public void WaitingForTheNextReadAllocatesNothing()
        {
            loader.Load(result, results.Add);
            Answer(0, CoinsOnlyJson);
            loader.Tick(0);
            TestDelegate step = () => loader.Tick(1);
            step();
            Assert.That(step, Allocates.Nothing);
            Assert.AreEqual(1, api.Requests.Count);
        }
    }
}
