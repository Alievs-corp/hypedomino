using System;
using HypeDomino.Content;
using HypeDomino.Game.Inventory;
using HypeDomino.Net;
using HypeDomino.UI.Board;
using HypeDomino.UI.Cosmetics;
using HypeDomino.UI.Effects;
using HypeDomino.UI.Emoji;
using HypeDomino.UI.Graphics;
using HypeDomino.UI.Localization;
using HypeDomino.UI.Match;
using HypeDomino.UI.Motion;
using HypeDomino.UI.Rarity;
using UnityEngine;

namespace HypeDomino.Game.Match
{
    public sealed class MatchSceneViews : MonoBehaviour, IMatchViews
    {
        public const float NoTileSizeCap = float.MaxValue;

        [SerializeField] MatchLayoutView layout;
        [SerializeField] ScorePanelView score;
        [SerializeField] BoardTilesView tiles;
        [SerializeField] AvatarView[] seats = new AvatarView[TableSides.Count];
        [SerializeField] TableView table;
        [SerializeField] LoadingScreenView loadingScreen;
        [SerializeField] MatchHudView hud;
        [SerializeField] SeatEmojiLayer emojiLayer;
        [SerializeField] EmojiPopup emojiPopup;
        [SerializeField] TextAsset rarities;
        [SerializeField] Sprite builtInAvatar;
        [SerializeField] Sprite builtInFrame;
        [SerializeField] ScorePopupView scorePopups;

        MatchHost host;
        CosmeticStage stage;
        CosmeticEffectsLink effects;
        EffectSettingsLink settings;
        EffectSettingsLink popupSettings;
        MatchTable matchTable;
        LoadingScreen loading;
        ActiveCosmeticManifest cosmetics;
        GameStrings texts;
        InventoryReply worn;
        int wheelSlots;
        bool emojiLayerBuilt;
        Action<InventoryLoadResult> wearLoaded;

        public bool Ready => host != null;
        public IMatchTable Table => matchTable;
        public ILoadingScreen LoadingScreen => loading;
        public CosmeticArtCache<Sprite> Stills => stage?.Stills;
        public MatchEmoji Emoji { get; private set; }

        public void Configure(MatchLayoutView layoutView, ScorePanelView scorePanel, BoardTilesView boardTiles, AvatarView[] seatAvatars, TableView tableView,
            LoadingScreenView loadingView, MatchHudView hudView, SeatEmojiLayer seatEmoji, EmojiPopup popup, TextAsset rarityTable, Sprite avatar, Sprite frame,
            ScorePopupView popups = null)
        {
            Detach();
            layout = layoutView;
            score = scorePanel;
            tiles = boardTiles;
            seats = seatAvatars;
            table = tableView;
            loadingScreen = loadingView;
            hud = hudView;
            emojiLayer = seatEmoji;
            emojiPopup = popup;
            rarities = rarityTable;
            builtInAvatar = avatar;
            builtInFrame = frame;
            scorePopups = popups;
        }

        public void Attach(MatchHost matchHost, Connection connection, ActiveCosmeticManifest manifest, GameStrings strings,
            ICosmeticArtSource<Sprite> stills, ICosmeticArtSource<LoopArt> loops, GraphicsQuality graphics, MotionPreference motion)
        {
            if (matchHost == null) throw new ArgumentNullException(nameof(matchHost));
            Detach();
            BuildEmojiLayer();
            cosmetics = manifest;
            texts = strings;
            stage = new CosmeticStage(stills, loops, RarityTable.Parse(rarities.text), builtInAvatar, builtInFrame, graphics.Effects, motion.ReduceMotion);
            matchTable = new MatchTable(new SeatAvatars(stage, seats), new TableBackground(stage, table), manifest, score, tiles, layout, NoTileSizeCap);
            loading = new LoadingScreen(loadingScreen, stage, manifest, strings);
            BuildEmoji();
            effects = new CosmeticEffectsLink(stage, graphics, motion, RefreshEffects);
            settings = new EffectSettingsLink(graphics, motion, hud, emojiLayer);
            popupSettings = scorePopups != null ? new EffectSettingsLink(graphics, motion, scorePopups) : null;
            host = matchHost;
            hud.Bind(host, connection, strings);
            host.Attach(this);
        }

        public void Detach()
        {
            if (host == null) return;
            host.Detach(this);
            hud.Unbind();
            effects.Dispose();
            settings.Dispose();
            popupSettings?.Dispose();
            host = null;
        }

        public void Wear(InventoryReply inventory)
        {
            if (inventory == null) return;
            worn = inventory;
            if (host == null) return;
            if (inventory.emoji_slots > wheelSlots) BuildEmoji();
            else Emoji.Wear(inventory);
        }

        void Start()
        {
            if (host != null || MatchBoot.Host == null) return;
            var link = FriendLobbyBoot.Link;
            if (link != null)
            {
                wearLoaded = result => Wear(result.Inventory);
                link.InventoryLoaded += wearLoaded;
                Wear(link.Worn);
            }
            Attach(MatchBoot.Host, NetBoot.Connection, ContentBoot.Cosmetics, LanguageBoot.Strings,
                new AddressablesArtSource<Sprite>(), new StillsOnly(), GraphicsBoot.Quality, MotionBoot.Motion);
        }

        void OnDestroy()
        {
            if (wearLoaded != null && FriendLobbyBoot.Link != null) FriendLobbyBoot.Link.InventoryLoaded -= wearLoaded;
            Detach();
        }

        void BuildEmojiLayer()
        {
            if (emojiLayerBuilt) return;
            var anchors = new RectTransform[seats.Length];
            for (var side = 0; side < seats.Length; side++) anchors[side] = (RectTransform)seats[side].transform;
            emojiLayer.Build(emojiPopup, anchors);
            emojiLayerBuilt = true;
        }

        void BuildEmoji()
        {
            wheelSlots = Math.Max(0, worn?.emoji_slots ?? 0);
            Emoji = new MatchEmoji(new EmojiWheel(wheelSlots), new SeatMessages(), texts, new EmojiArtShelf(cosmetics.Current, stage, ShowEmoji), ClockMs);
            Emoji.Wear(worn);
        }

        void RefreshEffects()
        {
            matchTable.RefreshEffects();
            loading.RefreshEffects();
        }

        void ShowEmoji(TableSide side, LoopArt art) => emojiLayer.Show((int)side, art, Time.unscaledTime);

        static long ClockMs() => (long)(Time.realtimeSinceStartupAsDouble * 1000d);

        sealed class StillsOnly : ICosmeticArtSource<LoopArt>
        {
            public void Load(string address, Action<LoopArt> done) => done(null);

            public void Release(string address)
            {
            }
        }
    }
}
