using System;
using HypeDomino.Net.Auth;

namespace HypeDomino.Net.Api
{
    public interface IApiTransport
    {
        void Send(string method, string pathAndQuery, string body, string authorization, string mac, Action<HttpReply> done);
    }
}
