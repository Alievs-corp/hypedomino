using System;
using System.Diagnostics;
using System.Text;
using System.Threading;
using HypeDomino.Net;
using Hypedomino.V1;
using NUnit.Framework;
using Unity.Profiling;
using UnityEngine.Profiling;
using UnityEngine.TestTools.Constraints;
using Is = UnityEngine.TestTools.Constraints.Is;

namespace HypeDomino.Tests.EditMode
{
    public sealed class AAllocationProbeTests
    {
        const int Iterations = 3000;
        static object sink;
        static object noiseSink;
        static volatile bool noiseRunning;

        sealed class QuietSocket : ISocketTransport
        {
            public TransportState State => TransportState.Open;
            public void Connect(Uri uri) { }
            public void Send(ReadOnlySpan<byte> frame) { }
            public bool TryReceive(out ReadOnlySpan<byte> frame)
            {
                frame = default;
                return false;
            }
            public void Dispose() { }
        }

        sealed class Tally
        {
            int failures;
            int firstFailure = -1;
            long maxValue;
            int failuresWithCollection;
            long slowestFailureTicks;

            public void Add(int iteration, long value, long ticks, bool collected)
            {
                if (value == 0) return;
                failures++;
                if (firstFailure < 0) firstFailure = iteration;
                maxValue = Math.Max(maxValue, value);
                slowestFailureTicks = Math.Max(slowestFailureTicks, ticks);
                if (collected) failuresWithCollection++;
            }

            public string Describe(string name) =>
                $"{name}: nonzero={failures}/{Iterations} first={firstFailure} max={maxValue} withGc={failuresWithCollection} slowestMs={slowestFailureTicks * 1000.0 / Stopwatch.Frequency:F3}";
        }

        static void Nothing() { }

        static void AllocateThree()
        {
            sink = new object();
            sink = new object();
            sink = new object();
        }

        static long ViaConstraint(TestDelegate step) => Is.Not.AllocatingGCMemory().ApplyTo(step).IsSuccess ? 0 : 1;

        static long ViaRecorder(TestDelegate step)
        {
            var recorder = Recorder.Get("GC.Alloc");
            recorder.enabled = false;
            recorder.FilterToCurrentThread();
            recorder.enabled = true;
            step();
            recorder.enabled = false;
            recorder.CollectFromAllThreads();
            return recorder.sampleBlockCount;
        }

        static long ViaThreadRecorder(TestDelegate step)
        {
            var recorder = new ProfilerRecorder("GC.Alloc", 64, ProfilerRecorderOptions.CollectOnlyOnCurrentThread);
            try
            {
                recorder.Start();
                step();
                recorder.Stop();
                return recorder.Valid ? recorder.Count : -1;
            }
            finally
            {
                recorder.Dispose();
            }
        }

        static long ViaThreadBytes(TestDelegate step)
        {
            var before = GC.GetAllocatedBytesForCurrentThread();
            step();
            return GC.GetAllocatedBytesForCurrentThread() - before;
        }

        static string Run(string name, Func<TestDelegate, long> measure, TestDelegate step, bool noise, bool collectOften)
        {
            step();
            var tally = new Tally();
            Thread thread = null;
            if (noise)
            {
                noiseRunning = true;
                thread = new Thread(() =>
                {
                    while (noiseRunning) noiseSink = new byte[48];
                }) { IsBackground = true };
                thread.Start();
                Thread.Sleep(20);
            }
            try
            {
                for (var i = 0; i < Iterations; i++)
                {
                    if (collectOften && i % 50 == 0) GC.Collect();
                    var collections = GC.CollectionCount(0);
                    var started = Stopwatch.GetTimestamp();
                    var value = measure(step);
                    var ticks = Stopwatch.GetTimestamp() - started;
                    tally.Add(i, value, ticks, GC.CollectionCount(0) != collections);
                }
            }
            finally
            {
                noiseRunning = false;
                thread?.Join();
            }
            return tally.Describe(name + (noise ? " +noise" : "") + (collectOften ? " +gc" : ""));
        }

        [Test]
        public void ProbeAllocationMeasurements()
        {
            var first = new StringBuilder();
            first.Append("firstConstraint=").Append(ViaConstraint(Nothing));
            first.Append(" thread=").Append(Thread.CurrentThread.ManagedThreadId);
            first.Append(" threadRecorderCalibration=").Append(ViaThreadRecorder(AllocateThree));
            first.Append(" threadBytesCalibration=").Append(ViaThreadBytes(AllocateThree));
            first.Append(" recorderCalibration=").Append(ViaRecorder(AllocateThree));

            var connection = new Connection(() => new QuietSocket(), () => new ClientHello(), () => 1.0, new Random(1));
            connection.Start(new Uri("wss://play.test/ws"));
            TestDelegate tick = connection.Tick;

            var report = new StringBuilder();
            report.AppendLine(first.ToString());
            foreach (var (label, measure) in new (string, Func<TestDelegate, long>)[]
                     {
                         ("constraint", ViaConstraint),
                         ("recorder", ViaRecorder),
                         ("threadRecorder", ViaThreadRecorder),
                         ("threadBytes", ViaThreadBytes),
                     })
            {
                report.AppendLine(Run(label + " allocateThree", measure, AllocateThree, false, false));
                report.AppendLine(Run(label + " nothing", measure, Nothing, false, false));
                report.AppendLine(Run(label + " nothing", measure, Nothing, true, false));
                report.AppendLine(Run(label + " nothing", measure, Nothing, false, true));
                report.AppendLine(Run(label + " tick", measure, tick, false, false));
                report.AppendLine(Run(label + " tick", measure, tick, true, false));
                report.AppendLine(Run(label + " tick", measure, tick, true, true));
            }

            Assert.Fail(report.ToString());
        }
    }
}
