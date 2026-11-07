using System.IO;
using UnityEngine;
using UnityEngine.AddressableAssets;

namespace HypeDomino.Content
{
    public static class ContentBoot
    {
        const string Folder = "content";
        const float MillisecondsPerSecond = 1000f;

        public static CatalogUpdater Catalog { get; private set; }

        [RuntimeInitializeOnLoadMethod(RuntimeInitializeLoadType.BeforeSceneLoad)]
        static void Start()
        {
            if (Application.isEditor) return;
            Catalog = new CatalogUpdater(new AddressablesCatalogSource(Path.Combine(Application.persistentDataPath, Folder)), Platform());
            Addressables.InternalIdTransformFunc = location => ContentUrls.Resolve(location.InternalId, Catalog.BaseUrl);
            Addressables.InitializeAsync().Completed += _ => Catalog.Start();
            var driver = new GameObject(nameof(ContentBoot)) { hideFlags = HideFlags.HideInHierarchy };
            Object.DontDestroyOnLoad(driver);
            driver.AddComponent<Driver>();
        }

        static string Platform() => Application.platform == RuntimePlatform.IPhonePlayer ? "iOS" : "Android";

        sealed class Driver : MonoBehaviour
        {
            void Update() => Catalog.Advance(Time.unscaledDeltaTime * MillisecondsPerSecond);
        }
    }
}
