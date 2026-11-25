using HypeDomino.UI.Cosmetics;
using HypeDomino.UI.Layout;
using HypeDomino.UI.Lists;
using HypeDomino.UI.Localization;
using HypeDomino.UI.Motion;
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

        public static Toggle Toggle(Transform parent)
        {
            var box = Child(parent, "Toggle", typeof(Image), typeof(LayoutElement));
            var image = box.GetComponent<Image>();
            image.color = TokenColor.SurfaceDefault;
            box.GetComponent<LayoutElement>().minHeight = TokenSize.TouchTargetMin;
            var check = Child(box, "Check", typeof(Image));
            Stretch(check);
            check.offsetMin = new Vector2(TokenSpacing.Sm, TokenSpacing.Sm);
            check.offsetMax = new Vector2(-TokenSpacing.Sm, -TokenSpacing.Sm);
            var mark = check.GetComponent<Image>();
            mark.color = TokenColor.RolePrimaryFill;
            mark.raycastTarget = false;
            var toggle = box.gameObject.AddComponent<Toggle>();
            toggle.targetGraphic = image;
            toggle.graphic = mark;
            return toggle;
        }

        public static Slider Slider(Transform parent)
        {
            var slider = DefaultControls.CreateSlider(new DefaultControls.Resources()).GetComponent<Slider>();
            var rect = (RectTransform)slider.transform;
            rect.SetParent(parent, false);
            slider.gameObject.AddComponent<LayoutElement>().minHeight = TokenSize.TouchTargetMin;
            rect.Find("Background").GetComponent<Image>().color = TokenColor.SurfaceDefault;
            slider.fillRect.GetComponent<Image>().color = TokenColor.RolePrimaryFill;
            slider.handleRect.GetComponent<Image>().color = TokenColor.TextPrimary;
            return slider;
        }

        public static TMP_Dropdown Dropdown(Transform parent)
        {
            var dropdown = TMP_DefaultControls.CreateDropdown(new TMP_DefaultControls.Resources()).GetComponent<TMP_Dropdown>();
            var rect = (RectTransform)dropdown.transform;
            rect.SetParent(parent, false);
            dropdown.gameObject.AddComponent<LayoutElement>().minHeight = TokenSize.TouchTargetMin;
            dropdown.targetGraphic.color = TokenColor.SurfaceDefault;
            dropdown.template.GetComponent<Image>().color = TokenColor.SurfaceRaised;
            Style(dropdown.captionText);
            Style(dropdown.itemText);
            dropdown.itemText.color = TokenColor.RolePrimaryOn;
            return dropdown;
        }

        public static Transform Scroll(Transform parent)
        {
            var holder = Child(parent, "Scroll", typeof(LayoutElement));
            holder.GetComponent<LayoutElement>().flexibleHeight = 1f;
            var viewport = Child(holder, "Viewport", typeof(Image), typeof(RectMask2D));
            Stretch(viewport);
            viewport.GetComponent<Image>().color = TokenColor.BackgroundBase;
            var content = Child(viewport, "Content", typeof(ContentSizeFitter));
            content.anchorMin = new Vector2(0f, 1f);
            content.anchorMax = Vector2.one;
            content.pivot = new Vector2(0.5f, 1f);
            content.sizeDelta = Vector2.zero;
            Lay(content);
            content.GetComponent<ContentSizeFitter>().verticalFit = ContentSizeFitter.FitMode.PreferredSize;
            var scroll = holder.gameObject.AddComponent<ScrollRect>();
            scroll.viewport = viewport;
            scroll.content = content;
            scroll.horizontal = false;
            return content;
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

        public static CosmeticImage Icon(Transform parent)
        {
            var icon = Child(parent, "Icon", typeof(Image), typeof(LayoutElement));
            var image = icon.GetComponent<Image>();
            image.preserveAspect = true;
            image.raycastTarget = false;
            image.enabled = false;
            icon.GetComponent<LayoutElement>().minHeight = TokenSize.TouchTargetMin;
            icon.gameObject.AddComponent<AnimatedImage>();
            return icon.gameObject.AddComponent<CosmeticImage>();
        }

        static void Style(TMP_Text text)
        {
            TypeSet.Loaded.Apply(text, TokenTextStyle.Label);
            text.color = TokenColor.TextPrimary;
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
