using UnityEngine.Networking;

namespace HypeDomino.Net.Api
{
    public sealed class PinnedCertificateHandler : CertificateHandler
    {
        readonly TlsPins pins;

        public PinnedCertificateHandler(TlsPins pins) => this.pins = pins;

        protected override bool ValidateCertificate(byte[] certificateData) => pins.Allows(certificateData);
    }
}
