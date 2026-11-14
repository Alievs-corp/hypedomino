using System.IO;
using UnityEngine;
using UnityEngine.AddressableAssets;

namespace HypeDomino.Content
{
    public static class ContentBoot
    {
        const string FolderName = "content";
        const float MillisecondsPerSecond = 1000f;

        public static string Folder => Path.Combine(Application.persistentDataPath, FolderName);

        public static CatalogUpdater Catalog { get; private set; }

        public static ContentCache Cache { get; private set; }

        [RuntimeInitializeOnLoadMethod(RuntimeInitializeLoadType.BeforeSceneLoad)]
        static void Start()
        {
            if (Application.isEditor) return;
            var folder = Folder;
            Catalog = new CatalogUpdater(new AddressablesCatalogSource(folder), Platform());
            Cache = new ContentCache(new UnityCacheStore(folder), () => Catalog.InMatch);
            Cache.Start();
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
