using System;
using HypeDomino.Content;
using HypeDomino.Game.Medals;
using HypeDomino.UI.Cosmetics;
using HypeDomino.UI.Localization;
using HypeDomino.UI.Medals;
using HypeDomino.UI.Profile;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Game.Profile
{
    public sealed class ProfileScreenView : MonoBehaviour
    {
        [SerializeField] ProfileSummaryView summary;
        [SerializeField] MedalCellSlot[] showcase;
        [SerializeField] TMP_InputField nicknameInput;
        [SerializeField] Button saveNickname;
        [SerializeField] TMP_Text message;

        Action<ProfileLoadResult> loaded;
        Action<MedalLoadResult> medalsLoaded;
        Action<NicknameSaveResult> saved;
        NicknameSaver saver;
        CosmeticStage stage;
        ActiveCosmeticManifest manifest;
        GameStrings strings;
        MedalCollection medals;
        PlayerProfile profile;
        string messageKey;

        public void Configure(ProfileSummaryView summaryView, MedalCellSlot[] showcaseCells, TMP_InputField nicknameField, Button saveButton, TMP_Text messageLine)
        {
            Unbind();
            summary = summaryView;
            showcase = showcaseCells;
            nicknameInput = nicknameField;
            saveNickname = saveButton;
            message = messageLine;
        }

        public void Bind(ProfileLoader loader, MedalLoader medalLoader, NicknameSaver nicknameSaver, CosmeticStage cosmetics, ActiveCosmeticManifest cosmeticManifest, GameStrings texts)
        {
            Unbind();
            saver = nicknameSaver;
            stage = cosmetics;
            manifest = cosmeticManifest;
            strings = texts;
            medals = new MedalCollection(cosmetics.Rarities);
            profile = null;
            messageKey = null;
            loaded ??= OnLoaded;
            medalsLoaded ??= OnMedalsLoaded;
            saved ??= OnSaved;
            saveNickname.onClick.AddListener(OnSave);
            strings.Changed += Redraw;
            manifest.Changed += Redraw;
            Redraw();
            loader.Load(loaded);
            medalLoader.Load(medalsLoaded);
        }

        void Start()
        {
            if (strings == null && NetBoot.Profile != null)
                Bind(NetBoot.Profile, NetBoot.Medals, new NicknameSaver(NetBoot.Api, AnalyticsBoot.Queue), StillsStage(), ContentBoot.Cosmetics, LanguageBoot.Strings);
        }

        static CosmeticStage StillsStage() =>
            new(new AddressablesArtSource<Sprite>(), new StillsOnlyArtSource(), null, null, null, GraphicsBoot.Quality.Effects, MotionBoot.Motion.ReduceMotion);

        void OnDestroy() => Unbind();

        void Unbind()
        {
            if (strings == null) return;
            saveNickname.onClick.RemoveListener(OnSave);
            strings.Changed -= Redraw;
            manifest.Changed -= Redraw;
            strings = null;
            manifest = null;
            saver = null;
        }

        void OnLoaded(ProfileLoadResult result)
        {
            if (strings == null) return;
            if (result.Outcome == ProfileLoadOutcome.Loaded) profile = result.Profile;
            messageKey = result.MessageKey;
            Redraw();
        }

        void OnMedalsLoaded(MedalLoadResult result)
        {
            if (strings == null) return;
            medals.Show(result.Outcome == MedalLoadOutcome.Loaded ? result.Collection : null);
            ShowMedals();
        }

        void OnSave()
        {
            if (!saver.Save(nicknameInput.text, saved)) return;
            ShowSave();
        }

        void OnSaved(NicknameSaveResult result)
        {
            if (strings == null) return;
            if (result.Outcome == NicknameSaveOutcome.Saved) profile = result.Profile;
            messageKey = result.MessageKey;
            Redraw();
        }

        void Redraw()
        {
            summary.Show(profile, stage, manifest.Current, strings);
            ShowMedals();
            ShowSave();
            message.enabled = messageKey != null;
            message.text = messageKey == null ? string.Empty : strings.Text(messageKey);
        }

        void ShowSave()
        {
            nicknameInput.interactable = profile != null;
            saveNickname.interactable = profile != null && !saver.Busy;
        }

        void ShowMedals()
        {
            for (var at = 0; at < showcase.Length; at++)
            {
                if (at >= medals.ShowcaseCount)
                {
                    showcase[at].Hide();
                    continue;
                }
                var medal = medals.ShowcaseAt(at);
                showcase[at].Show(stage, medal.staticAddress, strings.Text(medal.nameKey), true);
            }
        }
    }
}
