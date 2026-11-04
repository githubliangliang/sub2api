"""Inspect the emitted SQL's plan on an isolated migrated SQLite copy.

The performance benchmark uses Go/modernc. This supplementary planner check uses
Python's SQLite and records its version separately.
"""
import hashlib
import json
from pathlib import Path
import re
import sqlite3
import tempfile

out = Path(__file__).resolve().parent
root = out.parents[2]
source = (root / "backend/internal/repository/ops_repo_output_tps.go").read_text()
expression = source.split("query :=", 1)[1].split("// SQLite", 1)[0]
parts = re.findall(r"`([^`]+)`", expression)
assert len(parts) == 4
constants = (root / "backend/internal/service/usage_log.go").read_text()
kinds = [int(re.search(name + r"\s+RequestType\s*=\s*(\d+)", constants)[1])
         for name in ["RequestTypeSync", "RequestTypeStream", "RequestTypeWSV2", "RequestTypeCyberBlocked"]]
request_filter = "AND ul.request_type IN (" + ", ".join(map(str, kinds)) + ")"

with tempfile.TemporaryDirectory(prefix="output-tps-query-plan-") as tmp:
    origin = sqlite3.connect("file:/tmp/sub2api-output-tps-local/sub2api.db?mode=ro", uri=True)
    db = sqlite3.connect(str(Path(tmp) / "plan.db"))
    origin.backup(db)
    origin.close()
    db.execute("DELETE FROM usage_logs")
    db.execute("""WITH RECURSIVE n(i) AS (SELECT 1 UNION ALL SELECT i+1 FROM n WHERE i < 100000)
    INSERT INTO usage_logs (user_id,api_key_id,account_id,group_id,model,output_tokens,duration_ms,request_type,created_at)
    SELECT 1,9001,9001,9001,'gpt-test',1+i%1000,1000+(i%400)*10,2,'2026-01-17 00:30:00+00:00' FROM n""")
    db.commit()
    plans = []
    for filtered in [False, True]:
        join = "LEFT JOIN groups g ON g.id=ul.group_id LEFT JOIN accounts a ON a.id=ul.account_id" if filtered else ""
        where = "WHERE ul.created_at >= $1 AND ul.created_at < $2"
        args = ["2026-01-17 00:00:00+00:00", "2026-01-17 01:00:00+00:00"]
        if filtered:
            where += " AND ul.group_id=$3 AND COALESCE(NULLIF(g.platform,''), a.platform)=$4"
            args += [9001, "openai"]
        query = parts[0] + join + parts[1] + where + parts[2] + request_filter + parts[3]
        plan = [dict(id=row[0], parent=row[1], detail=row[3]) for row in db.execute("EXPLAIN QUERY PLAN " + query, args)]
        assert any("INDEX" in row["detail"] and "ul" in row["detail"] for row in plan), plan
        result = db.execute(query, args).fetchone()
        assert result[-1] == 100000
        plans.append(dict(filtered=filtered, query=query, args=args, plan=plan, result=result))
    record = dict(sqlite_version=sqlite3.sqlite_version, runtime="Python sqlite3 (supplementary planner evidence; timings use modernc)",
                  product_query_sha256=hashlib.sha256(source.encode()).hexdigest(), rows=100000, plans=plans)
    (out / "query-plan-details.json").write_text(json.dumps(record, ensure_ascii=False, indent=2) + "\n")
    print(json.dumps({"rows": record["rows"], "sqlite_version": record["sqlite_version"], "plans": plans}, ensure_ascii=False, indent=2))
