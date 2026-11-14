using System;
using System.Collections.Generic;
using HypeDomino.Game.Promo;
using HypeDomino.Net.Api;
using HypeDomino.Net.Auth;
using NUnit.Framework;

namespace HypeDomino.Tests.EditMode
{
    public sealed class PromoRedeemTests
    {
        const string RedemptionJson = "{\"promo\":\"promo-1\",\"campaign\":\"campaign.backers\",\"grant\":{\"id\":7,\"account_id\":\"acc-1\",\"bundle_id\":\"bundle.backers\",\"revision\":1,\"source\":{\"type\":\"promo\",\"ref\":\"campaign.backers\"},\"reason\":\"promo\",\"key\":\"promo:promo-1:acc-1\",\"config_version\":3,\"coins\":500,\"cosmetics\":[\"avatar.comet_fox\",\"frame.paper_kite\"],\"duplicates\":{\"tile.moon_owl\":40},\"medals\":[],\"medals_skipped\":[],\"pass_xp\":300,\"pass_xp_dropped\":0,\"at\":\"2026-11-14T10:00:00Z\"}}";

        [TestCase("ABCD2345", "ABCD2345")]
        [TestCase("abcd2345", "ABCD2345")]
        [TestCase("ab-cd 23_45", "ABCD2345")]
        [TestCase("  wxyz  ", "WXYZ")]
        [TestCase("ABCDEFGHJKMNPQRSTUVWXYZ234567892", "ABCDEFGHJKMNPQRSTUVWXYZ234567892")]
        public void ATypedCodeIsNormalisedAsTheServerDoes(string typed, string expected)
        {
            Assert.IsTrue(PromoCode.TryNormalize(typed, out var code));
            Assert.AreEqual(expected, code);
        }

        [TestCase(null)]
        [TestCase("")]
        [TestCase("- _ ")]
        [TestCase("ABC")]
        [TestCase("ABCDEFGHJKMNPQRSTUVWXYZ2345678923")]
        [TestCase("ABCD0234")]
        [TestCase("ABCDO234")]
        [TestCase("ABCD1234")]
        [TestCase("ABCDI234")]
        [TestCase("ABCDL234")]
        [TestCase("abcdi234")]
        [TestCase("ABCDı234")]
        [TestCase("ABCD.234")]
        [TestCase("ABCD\n2345")]
        [TestCase("ABCDÉ234")]
        public void ACodeTheServerWouldRefuseForItsFormatIsRefusedBeforeSending(string typed)
        {
            Assert.IsFalse(PromoCode.TryNormalize(typed, out var code));
            Assert.IsNull(code);
        }

        [Test]
        public void AGrantShowsWhatTheServerGranted()
        {
            var result = PromoRedeemResult.From(new ApiReply(ApiOutcome.Ok, 200, RedemptionJson));
            Assert.AreEqual(PromoRedeemOutcome.Granted, result.Outcome);
            Assert.AreEqual(PromoRedeemKeys.Granted, result.MessageKey);
            Assert.AreEqual(500, result.Coins);
            Assert.AreEqual(new[] { "avatar.comet_fox", "frame.paper_kite" }, result.Cosmetics);
            Assert.AreEqual(300, result.PassXp);
            Assert.AreEqual(0, result.PassXpDropped);
        }

        [Test]
        public void AGrantWithoutABodyIsAFailureAndShowsNothingGranted()
        {
            var result = PromoRedeemResult.From(new ApiReply(ApiOutcome.Ok, 200, "not json"));
            Assert.AreEqual(PromoRedeemOutcome.Failed, result.Outcome);
            Assert.AreEqual(0, result.Coins);
            Assert.IsEmpty(result.Cosmetics);
        }

        [TestCase(400, "promo.code_invalid", PromoRedeemOutcome.CodeInvalid, "promo.redeem.code_invalid")]
        [TestCase(404, "promo.unknown", PromoRedeemOutcome.Unknown, "promo.redeem.unknown")]
        [TestCase(409, "promo.not_active", PromoRedeemOutcome.NotActive, "promo.redeem.not_active")]
        [TestCase(409, "promo.disabled", PromoRedeemOutcome.Disabled, "promo.redeem.disabled")]
        [TestCase(409, "promo.not_started", PromoRedeemOutcome.NotStarted, "promo.redeem.not_started")]
        [TestCase(409, "promo.expired", PromoRedeemOutcome.Expired, "promo.redeem.expired")]
        [TestCase(409, "promo.exhausted", PromoRedeemOutcome.Exhausted, "promo.redeem.exhausted")]
        [TestCase(409, "promo.already_redeemed", PromoRedeemOutcome.AlreadyRedeemed, "promo.redeem.already_redeemed")]
        [TestCase(403, "promo.not_eligible", PromoRedeemOutcome.NotEligible, "promo.redeem.not_eligible")]
        [TestCase(409, "promo.no_pass_season", PromoRedeemOutcome.NoPassSeason, "promo.redeem.no_pass_season")]
        [TestCase(409, "promo.pass_premium_owned", PromoRedeemOutcome.PassPremiumOwned, "promo.redeem.pass_premium_owned")]
        [TestCase(409, "item.source_not_listed", PromoRedeemOutcome.Failed, "notice.generic")]
        [TestCase(400, "malformed", PromoRedeemOutcome.Failed, "notice.generic")]
        [TestCase(401, "bad_mac", PromoRedeemOutcome.Failed, "notice.generic")]
        [TestCase(409, "promo.reason_from_a_newer_server", PromoRedeemOutcome.Failed, "notice.generic")]
        public void EveryServerRefusalMapsToItsOwnKey(int status, string error, PromoRedeemOutcome outcome, string key)
        {
            var result = PromoRedeemResult.From(new ApiReply(ApiOutcome.Rejected, status, "{\"error\":\"" + error + "\"}", error));
            Assert.AreEqual(outcome, result.Outcome);
            Assert.AreEqual(key, result.MessageKey);
            Assert.AreEqual(0, result.Coins);
            Assert.IsEmpty(result.Cosmetics);
        }

        [TestCase(ApiOutcome.RateLimited, 429, "promo.rate_limited", PromoRedeemOutcome.RateLimited, "promo.redeem.rate_limited")]
        [TestCase(ApiOutcome.RateLimited, 429, "rate_limited", PromoRedeemOutcome.RateLimited, "promo.redeem.rate_limited")]
        [TestCase(ApiOutcome.RateLimited, 429, "promo.locked", PromoRedeemOutcome.Locked, "promo.redeem.locked")]
        [TestCase(ApiOutcome.Unavailable, 503, "promo.unavailable", PromoRedeemOutcome.Failed, "notice.generic")]
        [TestCase(ApiOutcome.Unavailable, 500, "internal", PromoRedeemOutcome.Failed, "notice.generic")]
        [TestCase(ApiOutcome.Offline, HttpReply.NoResponse, null, PromoRedeemOutcome.Offline, "promo.redeem.offline")]
        [TestCase(ApiOutcome.SignInRequired, 401, "access_rejected", PromoRedeemOutcome.SignInRequired, null)]
        public void LimitsLocksAndTransportFailuresMapToTheirKeys(ApiOutcome api, int status, string error, PromoRedeemOutcome outcome, string key)
        {
            var result = PromoRedeemResult.From(new ApiReply(api, status, null, error));
            Assert.AreEqual(outcome, result.Outcome);
            Assert.AreEqual(key, result.MessageKey);
        }

        [Test]
        public void EveryOutcomeButSignInHasAMessageKey()
        {
            foreach (PromoRedeemOutcome outcome in Enum.GetValues(typeof(PromoRedeemOutcome)))
            {
                var key = PromoRedeemKeys.Of(outcome);
                if (outcome == PromoRedeemOutcome.SignInRequired)
                {
                    Assert.IsNull(key);
                    continue;
                }
                StringAssert.IsMatch("^[a-z0-9_]+(\\.[a-z0-9_]+)+$", key, outcome.ToString());
            }
        }

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
        PromoRedeemer redeemer;
        List<PromoRedeemResult> results;

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
            redeemer = new PromoRedeemer(new ApiClient(api, session));
            results = new List<PromoRedeemResult>();
        }

        [Test]
        public void TheNormalisedCodeIsPostedAsTheOnlyField()
        {
            Assert.IsTrue(redeemer.Redeem("abcd-2345", results.Add));
            Assert.AreEqual(1, api.Requests.Count);
            Assert.AreEqual("POST", api.Requests[0].Method);
            Assert.AreEqual("/promo/redeem", api.Requests[0].PathAndQuery);
            Assert.AreEqual("{\"code\":\"ABCD2345\"}", api.Requests[0].Body);
            Assert.IsEmpty(results);

            api.Requests[0].Done(new HttpReply(200, RedemptionJson));
            Assert.AreEqual(PromoRedeemOutcome.Granted, results[0].Outcome);
            Assert.AreEqual(500, results[0].Coins);
        }

        [Test]
        public void AMalformedCodeIsAnsweredWithoutARequest()
        {
            Assert.IsTrue(redeemer.Redeem("ab1", results.Add));
            Assert.AreEqual(0, api.Requests.Count);
            Assert.AreEqual(PromoRedeemOutcome.CodeInvalid, results[0].Outcome);
            Assert.AreEqual(PromoRedeemKeys.CodeInvalid, results[0].MessageKey);
            Assert.IsFalse(redeemer.Busy);
        }

        [Test]
        public void ASecondSubmitWhileOneIsInFlightSendsNothing()
        {
            Assert.IsTrue(redeemer.Redeem("ABCD2345", results.Add));
            Assert.IsTrue(redeemer.Busy);
            Assert.IsFalse(redeemer.Redeem("ABCD2345", results.Add));
            Assert.AreEqual(1, api.Requests.Count);

            api.Requests[0].Done(new HttpReply(409, "{\"error\":\"promo.already_redeemed\"}"));
            Assert.IsFalse(redeemer.Busy);
            Assert.AreEqual(1, results.Count);
            Assert.AreEqual(PromoRedeemOutcome.AlreadyRedeemed, results[0].Outcome);

            Assert.IsTrue(redeemer.Redeem("WXYZ2345", results.Add));
            Assert.AreEqual(2, api.Requests.Count);
        }
    }
}
