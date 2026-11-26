using System.Collections.Generic;
using HypeDomino.Game.Privacy;
using NUnit.Framework;
using UnityEngine;
using UnityEngine.UI;
using Object = UnityEngine.Object;

namespace HypeDomino.Tests.EditMode
{
    public sealed class PrivacyNoticeViewTests
    {
        const string PolicyUrl = "https://policy.test/privacy";
        const string TermsUrl = "https://policy.test/terms";

        GameObject root;
        Button policy;
        Button terms;
        Button proceed;
        PrivacyNoticeView view;
        List<string> opened;
        int continued;

        [SetUp]
        public void Build()
        {
            root = new GameObject(nameof(PrivacyNoticeViewTests), typeof(RectTransform));
            policy = ButtonUnder(root, "Policy");
            terms = ButtonUnder(root, "Terms");
            proceed = ButtonUnder(root, "Continue");
            view = root.AddComponent<PrivacyNoticeView>();
            view.Configure(policy, terms, proceed);
            opened = new List<string>();
            continued = 0;
        }

        [TearDown]
        public void Destroy() => Object.DestroyImmediate(root);

        static Button ButtonUnder(GameObject parent, string name)
        {
            var button = new GameObject(name, typeof(RectTransform), typeof(Image)).AddComponent<Button>();
            button.transform.SetParent(parent.transform, false);
            return button;
        }

        void Bind(string policyUrl, string termsUrl) => view.Bind(policyUrl, termsUrl, opened.Add, () => continued++);

        [Test]
        public void BeforeBindingNoLinkButtonShows()
        {
            Assert.IsFalse(policy.gameObject.activeSelf);
            Assert.IsFalse(terms.gameObject.activeSelf);
            Assert.IsTrue(proceed.gameObject.activeSelf);
        }

        [Test]
        public void WithoutConfiguredUrlsTheLinkButtonsAreHidden()
        {
            Bind(null, string.Empty);

            Assert.IsFalse(policy.gameObject.activeSelf);
            Assert.IsFalse(terms.gameObject.activeSelf);
            Assert.IsTrue(proceed.gameObject.activeSelf);
        }

        [Test]
        public void OnlyALinkWithAUrlShows()
        {
            Bind(PolicyUrl, null);

            Assert.IsTrue(policy.gameObject.activeSelf);
            Assert.IsFalse(terms.gameObject.activeSelf);
        }

        [Test]
        public void EachLinkButtonOpensItsOwnUrl()
        {
            Bind(PolicyUrl, TermsUrl);

            policy.onClick.Invoke();
            terms.onClick.Invoke();

            CollectionAssert.AreEqual(new[] { PolicyUrl, TermsUrl }, opened);
            Assert.AreEqual(0, continued);
        }

        [Test]
        public void ContinueMovesOnOnceAndUnbindsTheLinks()
        {
            Bind(PolicyUrl, TermsUrl);

            proceed.onClick.Invoke();
            proceed.onClick.Invoke();
            policy.onClick.Invoke();

            Assert.AreEqual(1, continued);
            Assert.IsEmpty(opened);
        }

        [Test]
        public void ABindReplacesTheEarlierOne()
        {
            var earlier = 0;
            view.Bind(PolicyUrl, TermsUrl, opened.Add, () => earlier++);
            Bind(null, null);

            proceed.onClick.Invoke();

            Assert.AreEqual(0, earlier);
            Assert.AreEqual(1, continued);
            Assert.IsFalse(policy.gameObject.activeSelf);
        }
    }
}
