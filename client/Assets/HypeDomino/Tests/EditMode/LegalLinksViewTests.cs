using System.Collections.Generic;
using HypeDomino.Game.Privacy;
using Hypedomino.V1;
using NUnit.Framework;
using UnityEngine;
using UnityEngine.UI;
using Object = UnityEngine.Object;

namespace HypeDomino.Tests.EditMode
{
    public sealed class LegalLinksViewTests
    {
        const string PolicyUrl = "https://policy.test/privacy";
        const string TermsUrl = "https://policy.test/terms";

        GameObject root;
        Button policy;
        Button terms;
        LegalLinksView view;
        PolicyVersions versions;
        List<string> opened;

        [SetUp]
        public void Build()
        {
            root = new GameObject(nameof(LegalLinksViewTests), typeof(RectTransform));
            policy = ButtonUnder(root, "Policy");
            terms = ButtonUnder(root, "Terms");
            view = root.AddComponent<LegalLinksView>();
            view.Configure(policy, terms);
            versions = new PolicyVersions();
            opened = new List<string>();
        }

        [TearDown]
        public void Destroy()
        {
            if (root != null) Object.DestroyImmediate(root);
        }

        static Button ButtonUnder(GameObject parent, string name)
        {
            var button = new GameObject(name, typeof(RectTransform), typeof(Image)).AddComponent<Button>();
            button.transform.SetParent(parent.transform, false);
            return button;
        }

        void Announce(string policyUrl, string termsUrl, HelloStatus status = HelloStatus.Accepted) =>
            versions.Read(new ServerHello
            {
                Status = status,
                PrivacyVersions = new PrivacyVersions { PrivacyPolicyVersion = "2026-11-01", PrivacyPolicyUrl = policyUrl, TermsOfUseUrl = termsUrl },
            });

        [Test]
        public void BeforeBindingNoLinkShows()
        {
            Assert.IsFalse(policy.gameObject.activeSelf);
            Assert.IsFalse(terms.gameObject.activeSelf);
        }

        [Test]
        public void WithoutAnnouncedUrlsTheLinksStayHidden()
        {
            view.Bind(versions, opened.Add);
            Announce(string.Empty, string.Empty);
            Assert.IsFalse(policy.gameObject.activeSelf);
            Assert.IsFalse(terms.gameObject.activeSelf);
        }

        [Test]
        public void AnAnnouncedUrlShowsItsLinkAndOpensIt()
        {
            view.Bind(versions, opened.Add);
            Announce(PolicyUrl, string.Empty);
            Assert.IsTrue(policy.gameObject.activeSelf);
            Assert.IsFalse(terms.gameObject.activeSelf);
            Announce(PolicyUrl, TermsUrl);
            Assert.IsTrue(terms.gameObject.activeSelf);
            policy.onClick.Invoke();
            terms.onClick.Invoke();
            CollectionAssert.AreEqual(new[] { PolicyUrl, TermsUrl }, opened);
        }

        [Test]
        public void UrlsKnownBeforeBindingShowAtOnce()
        {
            Announce(PolicyUrl, TermsUrl);
            view.Bind(versions, opened.Add);
            Assert.IsTrue(policy.gameObject.activeSelf);
            Assert.IsTrue(terms.gameObject.activeSelf);
        }

        [Test]
        public void AWithdrawnUrlHidesItsLinkAgain()
        {
            view.Bind(versions, opened.Add);
            Announce(PolicyUrl, TermsUrl);
            Announce(string.Empty, TermsUrl);
            Assert.IsFalse(policy.gameObject.activeSelf);
            Assert.IsTrue(terms.gameObject.activeSelf);
        }

        [Test]
        public void AHelloThatIsNotAcceptedChangesNoLink()
        {
            view.Bind(versions, opened.Add);
            Announce(PolicyUrl, TermsUrl, HelloStatus.UpdateRequired);
            Assert.IsFalse(policy.gameObject.activeSelf);
            Assert.AreEqual(string.Empty, versions.PrivacyPolicyUrl);
        }

        [Test]
        public void ADestroyedViewNoLongerFollowsTheVersions()
        {
            view.Bind(versions, opened.Add);
            Object.DestroyImmediate(root);
            Assert.DoesNotThrow(() => Announce(PolicyUrl, TermsUrl));
        }
    }
}
