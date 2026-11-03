using System;
using System.Collections.Generic;
using System.IO;
using System.Text.RegularExpressions;
using NUnit.Framework;
using UnityEngine;

namespace HypeDomino.Tests.EditMode
{
    public sealed class RuntimeCodeRulesTests
    {
        const string DesignTokensFile = "DesignTokens.cs";
        static readonly string[] NonRuntimeFolders = { "Editor", "Generated", "Tests" };
        static readonly Regex HandWrittenColour = new Regex(@"\bnew\s+Color(32)?\s*\(|\bColor\.(black|blue|cyan|gray|green|grey|magenta|red|white|yellow)\b");

        [Test]
        public void RuntimeCodeDoesNotUseLinq()
        {
            var offenders = new List<string>();
            foreach (var file in RuntimeSources())
                if (File.ReadAllText(file).Contains("System.Linq")) offenders.Add(file);
            Assert.IsEmpty(offenders);
        }

        [Test]
        public void RuntimeCodeTakesColoursFromDesignTokens()
        {
            var offenders = new List<string>();
            foreach (var file in RuntimeSources())
                if (Path.GetFileName(file) != DesignTokensFile && HandWrittenColour.IsMatch(File.ReadAllText(file))) offenders.Add(file);
            Assert.IsEmpty(offenders);
        }

        [TestCase("image.color = new Color(1f, 0f, 0f);")]
        [TestCase("var tint = new Color32(255, 0, 0, 255);")]
        [TestCase("label.color = Color.white;")]
        public void HandWrittenColoursAreRecognised(string line) => Assert.IsTrue(HandWrittenColour.IsMatch(line));

        [TestCase("image.color = TokenColor.RolePrimaryFill;")]
        [TestCase("camera.backgroundColor = Color.clear;")]
        [TestCase("var colourCount = palette.Length;")]
        public void TokenColoursAreNotFlagged(string line) => Assert.IsFalse(HandWrittenColour.IsMatch(line));

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
