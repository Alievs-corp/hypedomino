using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Game.Lobby
{
    public sealed class LobbyTrainingCardView : MonoBehaviour
    {
        [SerializeField] Button start;

        LobbyTraining training;
        LobbyPlay play;

        public void Configure(Button startButton)
        {
            Unbind();
            start = startButton;
        }

        public void Bind(LobbyTraining lobbyTraining, LobbyPlay lobbyPlay)
        {
            Unbind();
            training = lobbyTraining;
            play = lobbyPlay;
            play.Changed += Refresh;
            start.onClick.AddListener(OnStart);
            Refresh();
        }

        void OnDestroy() => Unbind();

        void Unbind()
        {
            if (training == null) return;
            play.Changed -= Refresh;
            start.onClick.RemoveListener(OnStart);
            training = null;
            play = null;
        }

        void OnStart() => training.Start();

        void Refresh() => start.interactable = training.State == PlayCardState.Ready;
    }
}
