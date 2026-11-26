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
        [SerializeField] EffectPlayer effectPlayer;
        [SerializeField] RectTransform banner;

        MatchHost host;
        CosmeticStage stage;
        CosmeticEffectsLink effects;
        EffectSettingsLink settings;
        EffectSettingsLink popupSettings;
        MatchTable matchTable;
        TilePlaceLayer tilePlacing;
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
        public TilePlaceLayer TilePlacing => tilePlacing;

        public void Configure(MatchLayoutView layoutView, ScorePanelView scorePanel, BoardTilesView boardTiles, AvatarView[] seatAvatars, TableView tableView,
            LoadingScreenView loadingView, MatchHudView hudView, SeatEmojiLayer seatEmoji, EmojiPopup popup, TextAsset rarityTable, Sprite avatar, Sprite frame,
            ScorePopupView popups = null, EffectPlayer player = null, RectTransform bannerAnchor = null)
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
            effectPlayer = player;
            banner = bannerAnchor;
            tilePlacing = null;
        }

        public void Attach(MatchHost matchHost, Connection connection, ActiveCosmeticManifest manifest, GameStrings strings,
            ICosmeticArtSource<Sprite> stills, ICosmeticArtSource<LoopArt> loops, GraphicsQuality graphics, MotionPreference motion)
        {
            if (matchHost == null) throw new ArgumentNullException(nameof(matchHost));
            Detach();
            BuildEmojiLayer();
            BuildTilePlacing();
            cosmetics = manifest;
            texts = strings;
            stage = new CosmeticStage(stills, loops, RarityTable.Parse(rarities.text), builtInAvatar, builtInFrame, graphics.Effects, motion.ReduceMotion);
            var board = new MatchBoard(tiles, tilePlacing, SeatOrigins(), NewEffects(), NowMs, AudioBoot.PlayMatch);
            matchTable = new MatchTable(new SeatAvatars(stage, seats), new TableBackground(stage, table), manifest, score, board, layout, NoTileSizeCap);
            loading = new LoadingScreen(loadingScreen, stage, manifest, strings);
            BuildEmoji();
            effects = new CosmeticEffectsLink(stage, graphics, motion, RefreshEffects);
            settings = new EffectSettingsLink(graphics, motion, EffectTargets());
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
            matchTable.Dispose();
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
                new AddressablesArtSource<Sprite>(), new StillsOnlyArtSource(), GraphicsBoot.Quality, MotionBoot.Motion);
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

        void BuildTilePlacing()
        {
            if (tilePlacing != null) return;
            if (!tiles.TryGetComponent(out tilePlacing)) tilePlacing = tiles.gameObject.AddComponent<TilePlaceLayer>();
            var moved = new Transform[tiles.Capacity];
            var shadows = new Transform[tiles.Capacity];
            for (var slot = 0; slot < moved.Length; slot++)
            {
                moved[slot] = tiles.TileAt(slot);
                shadows[slot] = tiles.ShadowAt(slot);
            }
            tilePlacing.Build(moved, shadows);
            if (effectPlayer != null) tilePlacing.Attach(effectPlayer);
        }

        Transform[] SeatOrigins()
        {
            var origins = new Transform[seats.Length];
            for (var side = 0; side < seats.Length; side++) origins[side] = seats[side] != null ? seats[side].transform : null;
            return origins;
        }

        MatchEffects NewEffects()
        {
            if (effectPlayer == null || banner == null || scorePopups == null) return null;
            return new MatchEffects(effectPlayer, tilePlacing, (RectTransform)tiles.transform, banner, score.Mine.Meter, score.Theirs.Meter, scorePopups);
        }

        IEffectSettingsTarget[] EffectTargets() =>
            effectPlayer != null
                ? new IEffectSettingsTarget[] { hud, emojiLayer, tilePlacing, effectPlayer }
                : new IEffectSettingsTarget[] { hud, emojiLayer, tilePlacing };

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

        static float NowMs() => Time.unscaledTime * 1000f;
    }
}
