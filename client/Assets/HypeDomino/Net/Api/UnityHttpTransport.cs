using System;
using System.Text;
using HypeDomino.Net.Auth;
using UnityEngine.Networking;

namespace HypeDomino.Net.Api
{
    public sealed class UnityHttpTransport : IApiTransport, IAuthTransport
    {
        readonly string baseUrl;
        readonly int timeoutSeconds;

        public UnityHttpTransport(string baseUrl, int timeoutSeconds)
        {
            this.baseUrl = baseUrl.TrimEnd('/');
            this.timeoutSeconds = timeoutSeconds;
        }

        public void Post(string path, string json, Action<HttpReply> done) => Send("POST", path, json, null, null, done);

        public void Send(string method, string pathAndQuery, string body, string authorization, string mac, Action<HttpReply> done)
        {
            var request = new UnityWebRequest(baseUrl + pathAndQuery, method, new DownloadHandlerBuffer(), null) { timeout = timeoutSeconds };
            if (body != null)
            {
                request.uploadHandler = new UploadHandlerRaw(Encoding.UTF8.GetBytes(body));
                request.SetRequestHeader("Content-Type", "application/json");
            }
            if (authorization != null) request.SetRequestHeader("Authorization", authorization);
            if (mac != null) request.SetRequestHeader(RequestMac.Header, mac);
            request.SendWebRequest().completed += _ =>
            {
                var answered = request.result is UnityWebRequest.Result.Success or UnityWebRequest.Result.ProtocolError;
                var reply = answered ? new HttpReply((int)request.responseCode, request.downloadHandler.text) : new HttpReply(HttpReply.NoResponse, null);
                request.Dispose();
                done(reply);
            };
        }
    }
}
