using HypeDomino.Game.Medals;
using HypeDomino.Game.Profile;
using HypeDomino.UI.Cosmetics;
using HypeDomino.UI.Motion;
using HypeDomino.UI.Screens;
using HypeDomino.UI.Tokens;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Game.Screens
{
    public static class ProfileScreens
    {
        public const string NavKey = "ui.plain.nav.profile";
        public const string LevelKey = "ui.plain.profile.level";
        public const string SaveNicknameKey = "ui.plain.profile.save_nickname";
        public const int ShowcaseCells = 3;

        public static readonly string[] StatKeys =
        {
            "ui.plain.profile.games",
            "ui.plain.profile.wins",
            "ui.plain.profile.losses",
            "ui.plain.profile.win_rate",
            "ui.plain.profile.streak",
            "ui.plain.profile.best_streak",
        };

        public static ScreenView Profile(ScreenHost host)
        {
            var screen = host.PushedScreen("Profile", out var column);
            var list = PlainUi.Scroll(column);
            var avatar = Avatar(list);
            var nickname = PlainUi.Text(list, TokenTextStyle.Title);
            var level = Value(list, host, LevelKey);
            var stats = new TMP_Text[StatKeys.Length];
            for (var stat = 0; stat < stats.Length; stat++) stats[stat] = Value(list, host, StatKeys[stat]);
            var summary = screen.gameObject.AddComponent<ProfileSummaryView>();
            summary.Configure(avatar, nickname, level, stats);
            var showcase = new MedalCellSlot[ShowcaseCells];
            for (var cell = 0; cell < showcase.Length; cell++) showcase[cell] = Medal(list);
            var nicknameInput = PlainUi.Input(list);
            var save = PlainUi.Button(list, SaveNicknameKey, host.Strings);
            var message = PlainUi.Text(list, TokenTextStyle.Body);
            screen.gameObject.AddComponent<ProfileScreenView>().Configure(summary, showcase, nicknameInput, save, message);
            return screen;
        }

        static TMP_Text Value(Transform list, ScreenHost host, string labelKey)
        {
            var row = PlainUi.Row(list).transform;
            PlainUi.Text(row, TokenTextStyle.Label, labelKey, host.Strings);
            return PlainUi.Text(row, TokenTextStyle.Counter);
        }

        static AvatarView Avatar(Transform list)
        {
            var holder = new GameObject("Avatar", typeof(RectTransform), typeof(LayoutElement));
            holder.transform.SetParent(list, false);
            holder.GetComponent<LayoutElement>().minHeight = TokenSize.TouchTargetMin;
            var avatar = holder.AddComponent<AvatarView>();
            avatar.Configure(Art(holder.transform), Art(holder.transform), null, null);
            return avatar;
        }

        static MedalCellSlot Medal(Transform list)
        {
            var row = PlainUi.Row(list);
            var icon = Art(row.transform);
            var name = PlainUi.Text(row.transform, TokenTextStyle.Label);
            var locked = PlainUi.Mark(row.transform);
            var cell = row.AddComponent<MedalCellSlot>();
            cell.Configure(icon, name, locked);
            cell.Hide();
            return cell;
        }

        static CosmeticImage Art(Transform parent)
        {
            var art = new GameObject("Art", typeof(RectTransform), typeof(Image), typeof(AnimatedImage), typeof(CosmeticImage), typeof(LayoutElement));
            art.transform.SetParent(parent, false);
            var image = art.GetComponent<Image>();
            image.raycastTarget = false;
            image.preserveAspect = true;
            image.enabled = false;
            art.GetComponent<LayoutElement>().minHeight = TokenSize.TouchTargetMin;
            return art.GetComponent<CosmeticImage>();
        }
    }
}
