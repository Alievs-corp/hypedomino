namespace HypeDomino.Net.Auth
{
    public enum AuthOutcome
    {
        SignedIn,
        SignInRequired,
        Cancelled,
        Busy,
        Unsupported,
        ProviderFailed,
        Offline,
        Unavailable,
        RateLimited,
        Rejected,
        Banned,
        Deleted,
        DeletionPending,
    }

    public readonly struct AuthResult
    {
        public readonly AuthOutcome Outcome;
        public readonly bool NewAccount;
        public readonly string Error;
        public readonly string DeleteAt;

        public AuthResult(AuthOutcome outcome, bool newAccount = false, string error = null, string deleteAt = null)
        {
            Outcome = outcome;
            NewAccount = newAccount;
            Error = error;
            DeleteAt = deleteAt;
        }
    }
}
