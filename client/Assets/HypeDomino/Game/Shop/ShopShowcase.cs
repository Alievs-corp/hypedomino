using System;
using System.Collections.Generic;
using HypeDomino.Net;
using HypeDomino.UI.Shop;

namespace HypeDomino.Game.Shop
{
    public static class ShopSectionKeys
    {
        public const string Featured = "shop.section.featured.title";
        public const string DailyRotation = "shop.section.daily_rotation.title";
        public const string WeeklyRotation = "shop.section.weekly_rotation.title";
        public const string Event = "shop.section.event.title";
        public const string Chests = "shop.section.chests.title";
        public const string Pass = "shop.section.pass.title";
        public const string Crystals = "shop.section.crystals.title";

        public static string Title(string section) => section switch
        {
            "featured" => Featured,
            "daily_rotation" => DailyRotation,
            "weekly_rotation" => WeeklyRotation,
            "event" => Event,
            "chests" => Chests,
            "pass" => Pass,
            "crystals" => Crystals,
            _ => null,
        };
    }

    public readonly struct OfferPrice
    {
        OfferPrice(Balance amount, string productId)
        {
            Amount = amount;
            ProductId = productId;
        }

        public Balance Amount { get; }
        public string ProductId { get; }
        public bool InStore => ProductId != null;

        public static bool TryFrom(OfferReply offer, out OfferPrice price)
        {
            price = default;
            var inStore = !string.IsNullOrEmpty(offer.product_id);
            var prices = (offer.price_coins > 0 ? 1 : 0) + (offer.price_crystals > 0 ? 1 : 0) + (inStore ? 1 : 0);
            if (prices != 1 || offer.price_coins < 0 || offer.price_crystals < 0) return false;
            if (inStore)
            {
                price = new OfferPrice(default, offer.product_id);
                return true;
            }
            var amount = offer.price_crystals > 0 ? new Balance(Currency.Crystal, offer.price_crystals) : new Balance(Currency.Coin, offer.price_coins);
            price = new OfferPrice(amount, null);
            return true;
        }
    }

    public sealed class OfferCard
    {
        OfferCard(OfferReply offer, OfferPrice price, ulong endsUnixMs)
        {
            OfferId = offer.id;
            Kind = offer.kind;
            NameKey = offer.name_key;
            Price = price;
            ReferencePrice = ReferenceOf(offer, price);
            GrantCoins = offer.coins;
            GrantCrystals = offer.crystals;
            GrantPassXp = offer.pass_xp;
            Cosmetics = offer.cosmetics ?? Array.Empty<string>();
            Owned = offer.owned ?? Array.Empty<string>();
            Limit = offer.limit;
            Bought = offer.bought;
            Purchasable = offer.purchasable;
            Refusal = offer.purchasable ? string.Empty : offer.reason ?? string.Empty;
            RefusalKey = offer.purchasable ? null : ShopBuyKeys.Of(ShopBuyResult.Refusal(offer.reason));
            EndsUnixMs = endsUnixMs;
            ChestCard.TryFrom(offer, out var chest);
            Chest = chest;
        }

        public string OfferId { get; }
        public string Kind { get; }
        public string NameKey { get; }
        public OfferPrice Price { get; }
        public Balance ReferencePrice { get; }
        public bool HasReferencePrice => ReferencePrice.Amount > 0;
        public long GrantCoins { get; }
        public long GrantCrystals { get; }
        public int GrantPassXp { get; }
        public string[] Cosmetics { get; }
        public string[] Owned { get; }
        public int Limit { get; }
        public int Bought { get; }
        public bool Purchasable { get; }
        public string Refusal { get; }
        public string RefusalKey { get; }
        public ulong EndsUnixMs { get; }
        public ChestCard Chest { get; }

        public bool IsOwned(string item) => Array.IndexOf(Owned, item) >= 0;

        public bool TrySecondsLeft(ServerClock clock, double nowSeconds, out double seconds)
        {
            seconds = 0;
            return clock != null && clock.TrySecondsLeft(EndsUnixMs, nowSeconds, out seconds);
        }

        public static bool TryFrom(OfferReply offer, ulong sectionEndsUnixMs, out OfferCard card)
        {
            card = null;
            if (offer == null || string.IsNullOrEmpty(offer.id) || string.IsNullOrEmpty(offer.name_key)) return false;
            if (!OfferPrice.TryFrom(offer, out var price)) return false;
            var offerEndsUnixMs = 0UL;
            if (!string.IsNullOrEmpty(offer.ends_at) && !ServerClock.TryReadUnixMs(offer.ends_at, out offerEndsUnixMs)) return false;
            card = new OfferCard(offer, price, Earlier(offerEndsUnixMs, sectionEndsUnixMs));
            return true;
        }

        static Balance ReferenceOf(OfferReply offer, OfferPrice price)
        {
            if (price.InStore) return default;
            var currency = price.Amount.Currency;
            return new Balance(currency, currency == Currency.Crystal ? offer.reference_price_crystals : offer.reference_price_coins);
        }

        static ulong Earlier(ulong a, ulong b)
        {
            if (a == 0) return b;
            if (b == 0) return a;
            return Math.Min(a, b);
        }
    }

    public sealed class ShopSection
    {
        ShopSection(string id, string titleKey, ulong endsUnixMs, OfferCard[] offers)
        {
            Id = id;
            TitleKey = titleKey;
            EndsUnixMs = endsUnixMs;
            Offers = offers;
        }

        public string Id { get; }
        public string TitleKey { get; }
        public ulong EndsUnixMs { get; }
        public OfferCard[] Offers { get; }

        public bool TrySecondsLeft(ServerClock clock, double nowSeconds, out double seconds)
        {
            seconds = 0;
            return clock != null && clock.TrySecondsLeft(EndsUnixMs, nowSeconds, out seconds);
        }

        public static bool TryFrom(SectionReply section, out ShopSection shown)
        {
            shown = null;
            if (section == null) return false;
            var endsUnixMs = 0UL;
            if (!string.IsNullOrEmpty(section.ends_at) && !ServerClock.TryReadUnixMs(section.ends_at, out endsUnixMs)) return false;
            var replies = section.offers ?? Array.Empty<OfferReply>();
            var offers = new OfferCard[replies.Length];
            for (var i = 0; i < replies.Length; i++)
            {
                if (!OfferCard.TryFrom(replies[i], endsUnixMs, out offers[i])) return false;
            }
            shown = new ShopSection(section.section, ShopSectionKeys.Title(section.section), endsUnixMs, offers);
            return true;
        }
    }

    public sealed class ShopShowcase
    {
        public const string Path = "/shop";

        ShopShowcase(long version, ShopSection[] sections)
        {
            Version = version;
            Sections = sections;
        }

        public long Version { get; }
        public ShopSection[] Sections { get; }

        public static bool TryRead(string body, out ShopShowcase showcase)
        {
            showcase = null;
            if (string.IsNullOrEmpty(body)) return false;
            ShopReply reply;
            try
            {
                reply = ShopReply.FromJson(body);
            }
            catch (ArgumentException)
            {
                return false;
            }
            return TryRead(reply, out showcase);
        }

        public static bool TryRead(ShopReply reply, out ShopShowcase showcase)
        {
            showcase = null;
            if (reply?.sections == null) return false;
            var sections = new List<ShopSection>(reply.sections.Length);
            foreach (var section in reply.sections)
            {
                if (!ShopSection.TryFrom(section, out var shown)) return false;
                if (shown.TitleKey == null || shown.Offers.Length == 0) continue;
                sections.Add(shown);
            }
            showcase = new ShopShowcase(reply.version, sections.ToArray());
            return true;
        }
    }
}
