using System.Collections.Generic;
using UnityEngine;

namespace HypeDomino.Core.Diagnostics
{
    public sealed class ErrorReports
    {
        public const int MaxMissingKeysPerSession = 20;

        readonly IErrorSender sender;
        readonly HashSet<string> reportedMissingKeys = new HashSet<string>();
        volatile bool consented;

        public ErrorReports(IErrorSender sender)
        {
            this.sender = sender;
        }

        public bool Consented => consented;

        public void Consent() => consented = true;

        public void Withdraw() => consented = false;

        public void OnLog(string message, string stackTrace, LogType type)
        {
            if (!consented || !IsReported(type)) return;
            sender.Send(new ErrorReport(type, message, stackTrace));
        }

        public void MissingKey(string key, string language, int bundleVersion)
        {
            if (!consented || reportedMissingKeys.Count >= MaxMissingKeysPerSession || !reportedMissingKeys.Add(key)) return;
            sender.Send(new ErrorReport(LogType.Warning, $"Strings missing key={key} language={language} version={bundleVersion}", string.Empty));
        }

        static bool IsReported(LogType type) => type == LogType.Exception || type == LogType.Error || type == LogType.Assert;
    }
}
