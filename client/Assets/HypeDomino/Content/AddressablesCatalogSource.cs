using System;
using System.IO;
using UnityEngine;
using UnityEngine.AddressableAssets;
using UnityEngine.AddressableAssets.ResourceLocators;
using UnityEngine.Networking;
using UnityEngine.ResourceManagement.AsyncOperations;

namespace HypeDomino.Content
{
    public sealed class AddressablesCatalogSource : ICatalogSource
    {
        const string VersionKey = "content.catalogVersion";
        const string BaseUrlKey = "content.baseUrl";
        const int TimeoutSeconds = 20;

        readonly string storedPath;
        readonly string downloadPath;
        IResourceLocator remote;

        public AddressablesCatalogSource(string folder)
        {
            Directory.CreateDirectory(folder);
            storedPath = Path.Combine(folder, "catalog.bin");
            downloadPath = Path.Combine(folder, "catalog-new.bin");
        }

        public string StoredVersion => File.Exists(storedPath) ? PlayerPrefs.GetString(VersionKey, null) : null;

        public string StoredBaseUrl => PlayerPrefs.GetString(BaseUrlKey, null);

        public void LoadStored(Action<bool> done) => Load(storedPath, done);

        public void Download(string url, Action<bool> done)
        {
            var request = new UnityWebRequest(url, UnityWebRequest.kHttpVerbGET, new DownloadHandlerFile(downloadPath) { removeFileOnAbort = true }, null)
            {
                timeout = TimeoutSeconds,
            };
            request.SendWebRequest().completed += _ =>
            {
                var ok = request.result == UnityWebRequest.Result.Success;
                request.Dispose();
                done(ok);
            };
        }

        public void Commit(string version, string baseUrl, Action<bool> done) => Load(downloadPath, ok =>
        {
            if (!ok)
            {
                done(false);
                return;
            }
            try
            {
                File.Copy(downloadPath, storedPath, true);
                File.Delete(downloadPath);
            }
            catch (IOException)
            {
                done(false);
                return;
            }
            PlayerPrefs.SetString(VersionKey, version);
            PlayerPrefs.SetString(BaseUrlKey, baseUrl);
            PlayerPrefs.Save();
            done(true);
        });

        void Load(string path, Action<bool> done)
        {
            Addressables.LoadContentCatalogAsync(path, false).Completed += loaded =>
            {
                var ok = loaded.Status == AsyncOperationStatus.Succeeded;
                if (ok)
                {
                    if (remote != null) Addressables.RemoveResourceLocator(remote);
                    remote = loaded.Result;
                }
                Addressables.Release(loaded);
                done(ok);
            };
        }
    }
}
