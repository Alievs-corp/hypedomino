using System;
using HypeDomino.UI.Motion;

namespace HypeDomino.UI.Cosmetics
{
    public sealed class StillsOnly : ICosmeticArtSource<LoopArt>
    {
        public void Load(string address, Action<LoopArt> done) => done(null);

        public void Release(string address)
        {
        }
    }
}
