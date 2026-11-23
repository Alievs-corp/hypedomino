using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Game
{
    public sealed class MaintenanceScreenView : MonoBehaviour
    {
        [SerializeField] GameObject panel;
        [SerializeField] Button retry;

        MaintenanceScreen screen;

        public void Configure(GameObject screenPanel, Button retryButton)
        {
            Unbind();
            panel = screenPanel;
            retry = retryButton;
        }

        public void Bind(MaintenanceScreen maintenance)
        {
            Unbind();
            screen = maintenance;
            screen.Changed += Refresh;
            retry.onClick.AddListener(Retry);
            Refresh();
        }

        void Start()
        {
            if (screen == null && NetBoot.Maintenance != null) Bind(NetBoot.Maintenance);
        }

        void OnDestroy() => Unbind();

        void Unbind()
        {
            if (screen == null) return;
            screen.Changed -= Refresh;
            retry.onClick.RemoveListener(Retry);
            screen = null;
        }

        void Retry() => screen.Retry();

        void Refresh()
        {
            panel.SetActive(screen.Active);
            retry.interactable = !screen.Checking;
        }
    }
}
