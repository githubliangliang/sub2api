"""Extract frozen pure helpers and compare them through isolated Go overlays."""
import hashlib
import json
import subprocess
import tempfile
from pathlib import Path

OUT = Path(__file__).resolve().parent
ROOT = OUT.parents[2]
SOURCE = json.loads((OUT / "baseline.json").read_text())["refs"]["v0.2.23"]

def git_show(path):
    return subprocess.check_output(["rtk", "proxy", "git", "show", SOURCE + ":" + path], cwd=ROOT)

with tempfile.TemporaryDirectory(prefix="sub2api-023-perf-") as tmp:
    tmp = Path(tmp)
    extractor = tmp / "extract.go"
    extractor.write_bytes((OUT / "extract-functions.go.txt").read_bytes())
    source1 = tmp / "ollama-source.go"
    source1.write_bytes(git_show("backend/internal/service/openai_gateway_ollama_cloud_cc_reasoning.go"))
    source2 = tmp / "sse-source.go"
    source2.write_bytes(git_show("backend/internal/service/gateway_upstream_response.go"))
    funcs = json.loads(subprocess.check_output([
        "rtk", "proxy", "go", "run", str(extractor), "--",
        str(source1), "normalizeOllamaCloudChatCompletionsRequest",
        str(source2), "sseEventNeedsJSONParse",
    ], cwd=ROOT))
    helper = git_show("backend/internal/service/openai_json_object_field_edit.go").decode()
    # Keep helper code exact; add imports needed by the two renamed consumers.
    helper = helper.replace('"encoding/json"', '"encoding/json"\n"strconv"\n"strings"\n"bytes"\n"github.com/tidwall/sjson"')
    candidate = helper + "\n" + funcs["normalizeOllamaCloudChatCompletionsRequest"].replace(
        "func normalizeOllamaCloudChatCompletionsRequest(", "func intake023CandidateOllama("
    ) + "\n"
    candidate += '\nvar (\nsseUsageFieldBytes=[]byte("\\"usage\\"")\nsseMessageFieldBytes=[]byte("\\"message\\"")\n)\n'
    candidate += funcs["sseEventNeedsJSONParse"].replace("func sseEventNeedsJSONParse(", "func intake023SourceSSEPredicate(") + "\n"
    source_out = OUT / "probes" / "performance-source.go.txt"
    source_out.write_text(candidate)
    subprocess.run(["rtk", "proxy", "gofmt", "-w", str(source_out)], check=True)
    overlay = tmp / "overlay.json"
    overlay.write_text(json.dumps({"Replace": {
        str(ROOT / "backend/internal/service/intake023_performance.go"): str(source_out),
        str(ROOT / "backend/internal/service/intake023_performance_test.go"): str(OUT / "probes/performance-test.go.txt"),
    }}))
    (OUT / "performance-source.json").write_text(json.dumps({
        "source_commit": SOURCE, "source_helper": "backend/internal/service/openai_json_object_field_edit.go",
        "consumer": "Ollama Cloud existing request normalizer; DeepSeek's absent placeholder consumer is excluded",
        "normalization": "rename consumers for comparison; add existing imports and two exact byte constants",
        "sha256": hashlib.sha256(source_out.read_bytes()).hexdigest(),
    }, indent=2)+"\n")
    for label, args in [
        ("performance-equivalence", ["-run", "^TestIntake023Perf"]),
        ("performance-benchmark", ["-run", "^$", "-bench", "^BenchmarkIntake023", "-benchtime=100ms", "-benchmem"]),
    ]:
        subprocess.run([
            "rtk", "proxy", "python3", str(OUT / "run-check.py"), label, "backend",
            "env", "GOMAXPROCS=1", "go", "test", "-tags=unit", "-overlay="+str(overlay),
            *args, "-count=1", "-v", "./internal/service",
        ], check=False)
