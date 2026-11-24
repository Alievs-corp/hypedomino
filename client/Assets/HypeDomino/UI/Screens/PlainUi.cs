using HypeDomino.UI.Layout;
using HypeDomino.UI.Lists;
using HypeDomino.UI.Localization;
using HypeDomino.UI.Tokens;
using HypeDomino.UI.Typography;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.UI.Screens
{
    public static class PlainUi
    {
        public static ScreenView Screen(Transform host, string name, out Transform column)
        {
            var screen = new GameObject(name, typeof(RectTransform)).AddComponent<ScreenView>();
            screen.transform.SetParent(host, false);
            var content = Child(screen.transform, "Content", typeof(CanvasGroup), typeof(Image));
            Stretch(content);
            content.GetComponent<Image>().color = TokenColor.BackgroundBase;
            var safe = Child(content, "Safe", typeof(SafeAreaFitter));
            var layout = Lay(safe);
            layout.padding = new RectOffset(TokenSpacing.Lg, TokenSpacing.Lg, TokenSpacing.Lg, TokenSpacing.Lg);
            column = safe;
            screen.Configure(content.GetComponent<CanvasGroup>(), null);
            return screen;
        }

        public static GameObject Column(Transform parent)
        {
            var column = Child(parent, "Column");
            Lay(column);
            return column.gameObject;
        }

        public static GameObject Row(Transform parent)
        {
            var row = Child(parent, "Row", typeof(LayoutElement));
            row.sizeDelta = new Vector2(0f, TokenSize.TouchTargetMin);
            row.GetComponent<LayoutElement>().minHeight = TokenSize.TouchTargetMin;
            var layout = row.gameObject.AddComponent<HorizontalLayoutGroup>();
            layout.spacing = TokenSpacing.Md;
            layout.childAlignment = TextAnchor.MiddleCenter;
            layout.childControlWidth = true;
            layout.childControlHeight = true;
            layout.childForceExpandWidth = true;
            layout.childForceExpandHeight = true;
            return row.gameObject;
        }

        public static TMP_Text Text(Transform parent, TokenTextStyle style)
        {
            var text = Child(parent, "Text").gameObject.AddComponent<TextMeshProUGUI>();
            TypeSet.Loaded.Apply(text, style);
            text.color = TokenColor.TextPrimary;
            text.alignment = TextAlignmentOptions.Center;
            text.raycastTarget = false;
            text.text = string.Empty;
            return text;
        }

        public static TMP_Text Text(Transform parent, TokenTextStyle style, string labelKey, GameStrings strings)
        {
            var text = Text(parent, style);
            text.gameObject.AddComponent<PlainLabel>().Configure(text, labelKey, strings);
            return text;
        }

        public static Button Button(Transform parent)
        {
            var button = Child(parent, "Button", typeof(Image), typeof(LayoutElement));
            var image = button.GetComponent<Image>();
            image.color = TokenColor.RolePrimaryFill;
            button.GetComponent<LayoutElement>().minHeight = TokenSize.TouchTargetMin;
            var pressable = button.gameObject.AddComponent<Button>();
            pressable.targetGraphic = image;
            return pressable;
        }

        public static Button Button(Transform parent, string labelKey, GameStrings strings)
        {
            var button = Button(parent);
            button.gameObject.AddComponent<PlainLabel>().Configure(Label(button), labelKey, strings);
            return button;
        }

        public static TMP_Text Label(Button button)
        {
            var label = Text(button.transform, TokenTextStyle.Label);
            label.color = TokenColor.RolePrimaryOn;
            Stretch((RectTransform)label.transform);
            return label;
        }

        public static TMP_InputField Input(Transform parent)
        {
            var box = Child(parent, "Input", typeof(Image), typeof(LayoutElement));
            var image = box.GetComponent<Image>();
            image.color = TokenColor.SurfaceDefault;
            box.GetComponent<LayoutElement>().minHeight = TokenSize.TouchTargetMin;
            var area = Child(box, "Area", typeof(RectMask2D));
            Stretch(area);
            var text = Text(area, TokenTextStyle.Body);
            Stretch((RectTransform)text.transform);
            box.gameObject.SetActive(false);
            var field = box.gameObject.AddComponent<TMP_InputField>();
            field.textViewport = area;
            field.textComponent = text;
            field.targetGraphic = image;
            box.gameObject.SetActive(true);
            return field;
        }

        public static RecyclingList List(Transform parent)
        {
            var holder = Child(parent, "List", typeof(LayoutElement));
            holder.sizeDelta = new Vector2(0f, ScreenView.ReferenceShortSide);
            holder.GetComponent<LayoutElement>().preferredHeight = ScreenView.ReferenceShortSide;
            var viewport = Child(holder, "Viewport", typeof(Image), typeof(RectMask2D));
            Stretch(viewport);
            viewport.GetComponent<Image>().color = TokenColor.BackgroundBase;
            var content = Child(viewport, "Content");
            Stretch(content);
            var scroll = holder.gameObject.AddComponent<ScrollRect>();
            scroll.viewport = viewport;
            scroll.content = content;
            scroll.horizontal = false;
            return holder.gameObject.AddComponent<RecyclingList>();
        }

        public static GameObject Mark(Transform parent)
        {
            var mark = Child(parent, "Mark", typeof(Image), typeof(LayoutElement));
            var image = mark.GetComponent<Image>();
            image.color = TokenColor.RoleWarningFill;
            image.raycastTarget = false;
            mark.GetComponent<LayoutElement>().minHeight = TokenSpacing.Xs;
            return mark.gameObject;
        }

        static RectTransform Child(Transform parent, string name, params System.Type[] components)
        {
            var child = new GameObject(name, typeof(RectTransform));
            foreach (var component in components) child.AddComponent(component);
            var rect = (RectTransform)child.transform;
            rect.SetParent(parent, false);
            return rect;
        }

        static void Stretch(RectTransform rect)
        {
            rect.anchorMin = Vector2.zero;
            rect.anchorMax = Vector2.one;
            rect.offsetMin = Vector2.zero;
            rect.offsetMax = Vector2.zero;
        }

        static VerticalLayoutGroup Lay(Transform target)
        {
            var layout = target.gameObject.AddComponent<VerticalLayoutGroup>();
            layout.spacing = TokenSpacing.Md;
            layout.childAlignment = TextAnchor.UpperCenter;
            layout.childControlWidth = true;
            layout.childControlHeight = true;
            layout.childForceExpandWidth = true;
            layout.childForceExpandHeight = false;
            return layout;
        }
    }
}
