using HypeDomino.Core.Diagnostics;
using NUnit.Framework;
using UnityEngine.TestTools.Constraints;
using Is = UnityEngine.TestTools.Constraints.Is;

namespace HypeDomino.Tests.EditMode
{
    public sealed class FrameWindowTests
    {
        const double SixtyHzThreshold = 1.5 / 60;
        const double OnTime = 1.0 / 60;
        const double Late = 2.0 / 60;

        [Test]
        public void ReportsWorkPercentilesMissedVsyncsAllocationsAndBatches()
        {
            var window = new FrameWindow(20, SixtyHzThreshold);
            for (var i = 1; i <= 20; i++)
                window.Add(i, i == 20 ? Late : OnTime, i % 10 == 0 ? 64 * i : 0, 40 + i);
            Assert.IsTrue(window.IsFull);

            var report = window.Close();
            Assert.AreEqual(20, report.Frames);
            Assert.AreEqual(11, report.WorkP50Ms);
            Assert.AreEqual(20, report.WorkP95Ms);
            Assert.AreEqual(5, report.MissedVsyncPercent);
            Assert.AreEqual(2, report.AllocatingFrames);
            Assert.AreEqual(1280, report.MaxGcBytes);
            Assert.AreEqual(60, report.MaxBatches);
        }

        [Test]
        public void CloseStartsAnEmptyWindow()
        {
            var window = new FrameWindow(2, SixtyHzThreshold);
            window.Add(9, Late, 128, 90);
            window.Add(9, Late, 128, 90);
            window.Close();
            Assert.IsFalse(window.IsFull);

            window.Add(3, OnTime, 0, 10);
            window.Add(4, OnTime, 0, 12);
            var report = window.Close();
            Assert.AreEqual(4, report.WorkP50Ms);
            Assert.AreEqual(0, report.MissedVsyncPercent);
            Assert.AreEqual(0, report.AllocatingFrames);
            Assert.AreEqual(0, report.MaxGcBytes);
            Assert.AreEqual(12, report.MaxBatches);
        }

        [Test]
        public void AddingAFrameDoesNotAllocate()
        {
            var window = new FrameWindow(1001, SixtyHzThreshold);
            window.Add(0, OnTime, 0, 1);
            Assert.That(() =>
            {
                for (var i = 0; i < 1000; i++) window.Add(i, OnTime, 0, 1);
            }, Is.Not.AllocatingGCMemory());
        }
    }
}
