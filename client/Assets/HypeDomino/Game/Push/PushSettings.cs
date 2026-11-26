using System;
using HypeDomino.Net.Api;

namespace HypeDomino.Game.Push
{
    public sealed class PushSettings
    {
        public const string Path = "/push/settings";

        readonly ApiClient api;
        readonly Func<string> policyVersion;
        readonly string appVersion;

        public PushSettings(ApiClient api, Func<string> policyVersion, string appVersion)
        {
            if (string.IsNullOrEmpty(appVersion)) throw new ArgumentException("app version required", nameof(appVersion));
            this.api = api;
            this.policyVersion = policyVersion ?? throw new ArgumentNullException(nameof(policyVersion));
            this.appVersion = appVersion;
        }

        public event Action Changed;

        public bool Busy { get; private set; }

        public bool Loaded { get; private set; }

        public bool Gameplay { get; private set; }

        public bool Promotional { get; private set; }

        public bool Load(Action<PushSettingsResult> done)
        {
            if (Busy)
            {
                return false;
            }
            Busy = true;
            api.Get(Path, reply => Answer(reply, done));
            return true;
        }

        public bool SetGameplay(bool on, Action<PushSettingsResult> done) => Save(on, Promotional, done);

        public bool SetPromotional(bool on, Action<PushSettingsResult> done) => Save(Gameplay, on, done);

        bool Save(bool gameplay, bool promotional, Action<PushSettingsResult> done)
        {
            var version = policyVersion();
            if (Busy || !Loaded || (promotional != Promotional && string.IsNullOrEmpty(version)))
            {
                return false;
            }
            Busy = true;
            var request = new PushSettingsRequest { gameplay = gameplay, promotional = promotional, policyVersion = version, appVersion = appVersion };
            api.Put(Path, request.ToJson(), reply => Answer(reply, done));
            return true;
        }

        void Answer(ApiReply reply, Action<PushSettingsResult> done)
        {
            Busy = false;
            var result = PushSettingsResult.From(reply);
            if (result.Outcome == PushSettingsOutcome.Done)
            {
                Apply(result.Settings);
            }
            done(result);
        }

        void Apply(PushSettingsReply settings)
        {
            var changed = !Loaded || settings.gameplay != Gameplay || settings.promotional != Promotional;
            Loaded = true;
            Gameplay = settings.gameplay;
            Promotional = settings.promotional;
            if (changed) Changed?.Invoke();
        }
    }
}
