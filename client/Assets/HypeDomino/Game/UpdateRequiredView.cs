using System;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Game
{
    public sealed class UpdateRequiredView : MonoBehaviour
    {
        [SerializeField] GameObject panel;
        [SerializeField] Button update;

        AppFlow flow;
        string storeLink;
        Action<string> open;

        public void Configure(GameObject screenPanel, Button updateButton)
        {
            Unbind();
            panel = screenPanel;
            update = updateButton;
        }

        public void Bind(AppFlow appFlow, string link, Action<string> openLink)
        {
            Unbind();
            flow = appFlow;
            storeLink = link;
            open = openLink;
            flow.Changed += Show;
            update.onClick.AddListener(OpenStore);
            Show(flow.State);
        }

        void Start()
        {
            if (flow == null && AppBoot.Flow != null) Bind(AppBoot.Flow, StoreLink.Of(Application.platform, Application.identifier), Application.OpenURL);
        }

        void OnDestroy() => Unbind();

        void Unbind()
        {
            if (flow == null) return;
            flow.Changed -= Show;
            update.onClick.RemoveListener(OpenStore);
            flow = null;
            open = null;
        }

        void OpenStore() => open(storeLink);

        void Show(AppState state)
        {
            panel.SetActive(state == AppState.UpdateRequired);
            update.gameObject.SetActive(!string.IsNullOrEmpty(storeLink));
        }
    }
}
