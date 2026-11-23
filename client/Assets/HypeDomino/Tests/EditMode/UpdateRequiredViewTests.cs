using System.Collections.Generic;
using HypeDomino.Game;
using Hypedomino.V1;
using NUnit.Framework;
using UnityEngine;
using UnityEngine.UI;
using Object = UnityEngine.Object;

namespace HypeDomino.Tests.EditMode
{
    public sealed class UpdateRequiredViewTests
    {
        const string Link = "https://play.google.com/store/apps/details?id=com.example.game";

        GameObject root;
        GameObject panel;
        Button update;
        AppFlow flow;
        List<string> opened;

        [SetUp]
        public void Build()
        {
            root = new GameObject(nameof(UpdateRequiredViewTests), typeof(RectTransform));
            panel = new GameObject("Panel", typeof(RectTransform));
            panel.transform.SetParent(root.transform, false);
            update = new GameObject("Update", typeof(RectTransform), typeof(Image)).AddComponent<Button>();
            update.transform.SetParent(panel.transform, false);
            flow = new AppFlow();
            opened = new List<string>();
        }

        [TearDown]
        public void Destroy() => Object.DestroyImmediate(root);

        void Show(string link)
        {
            var view = root.AddComponent<UpdateRequiredView>();
            view.Configure(panel, update);
            view.Bind(flow, link, opened.Add);
        }

        [Test]
        public void ThePanelShowsOnlyWhenTheServerRequiresAnUpdate()
        {
            Show(Link);

            Assert.IsFalse(panel.activeSelf);

            flow.Hello(new ServerHello { Status = HelloStatus.UpdateRequired });

            Assert.IsTrue(panel.activeSelf);
            Assert.IsTrue(update.gameObject.activeSelf);
        }

        [Test]
        public void TheUpdateButtonOpensTheStorePage()
        {
            Show(Link);
            flow.Hello(new ServerHello { Status = HelloStatus.UpdateRequired });

            update.onClick.Invoke();

            CollectionAssert.AreEqual(new[] { Link }, opened);
        }

        [Test]
        public void WithoutAStoreLinkTheUpdateButtonIsHidden()
        {
            Show(string.Empty);
            flow.Hello(new ServerHello { Status = HelloStatus.UpdateRequired });

            Assert.IsTrue(panel.activeSelf);
            Assert.IsFalse(update.gameObject.activeSelf);
        }

        [Test]
        public void ASignInPromptCannotLeaveTheUpdateScreen()
        {
            Show(Link);
            flow.Hello(new ServerHello { Status = HelloStatus.UpdateRequired });

            flow.SignInNeeded();

            Assert.IsTrue(panel.activeSelf);
        }
    }
}
