using System;

namespace HypeDomino.Game.Shop
{
    public enum ShopCheckoutStep
    {
        Closed,
        Preview,
        Buying,
        Done,
    }

    public sealed class ShopCheckout
    {
        readonly ShopBuyer buyer;
        readonly Action<ShopBuyResult> settle;
        Action<ShopBuyResult> done;

        public ShopCheckout(ShopBuyer buyer)
        {
            this.buyer = buyer;
            settle = Settle;
        }

        public ShopCheckoutStep Step { get; private set; }
        public OfferCard Offer { get; private set; }
        public ShopBuyResult Result { get; private set; }

        public bool CanConfirm => Step switch
        {
            ShopCheckoutStep.Preview => !buyer.Busy && Offer.Purchasable && !Offer.Price.InStore,
            ShopCheckoutStep.Done => !buyer.Busy && buyer.IsUnsettled(Offer.OfferId),
            _ => false,
        };

        public bool Open(OfferCard offer)
        {
            if (offer == null || Step == ShopCheckoutStep.Buying) return false;
            Offer = offer;
            Result = default;
            Step = ShopCheckoutStep.Preview;
            return true;
        }

        public bool Confirm(Action<ShopBuyResult> done)
        {
            if (!CanConfirm) return false;
            var previous = Step;
            Step = ShopCheckoutStep.Buying;
            this.done = done;
            if (buyer.Buy(Offer.OfferId, settle)) return true;
            Step = previous;
            this.done = null;
            return false;
        }

        public bool Close()
        {
            if (Step == ShopCheckoutStep.Buying) return false;
            Step = ShopCheckoutStep.Closed;
            Offer = null;
            Result = default;
            return true;
        }

        void Settle(ShopBuyResult result)
        {
            Result = result;
            Step = ShopCheckoutStep.Done;
            var notify = done;
            done = null;
            notify?.Invoke(result);
        }
    }
}
