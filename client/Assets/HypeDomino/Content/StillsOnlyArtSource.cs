using System;
using HypeDomino.UI.Cosmetics;
using HypeDomino.UI.Motion;

namespace HypeDomino.Content
{
    public sealed class StillsOnlyArtSource : ICosmeticArtSource<LoopArt>
    {
        public void Load(string address, Action<LoopArt> done) => done(null);

        public void Release(string address)
        {
        }
    }
}
