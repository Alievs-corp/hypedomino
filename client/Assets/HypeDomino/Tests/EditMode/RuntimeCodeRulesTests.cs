using System;
using System.Collections.Generic;
using System.IO;
using NUnit.Framework;
using UnityEngine;

namespace HypeDomino.Tests.EditMode
{
    public sealed class RuntimeCodeRulesTests
    {
        static readonly string[] NonRuntimeFolders = { "Editor", "Generated", "Tests" };

        [Test]
        public void RuntimeCodeDoesNotUseLinq()
        {
            var offenders = new List<string>();
            foreach (var file in RuntimeSources())
                if (File.ReadAllText(file).Contains("System.Linq")) offenders.Add(file);
            Assert.IsEmpty(offenders);
        }

        static IEnumerable<string> RuntimeSources()
        {
            var root = Path.Combine(Application.dataPath, "HypeDomino");
            foreach (var folder in Directory.GetDirectories(root))
            {
                if (Array.IndexOf(NonRuntimeFolders, Path.GetFileName(folder)) >= 0) continue;
                foreach (var file in Directory.GetFiles(folder, "*.cs", SearchOption.AllDirectories))
                    yield return file;
            }
        }
    }
}
